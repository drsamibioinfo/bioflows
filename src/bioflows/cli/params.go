package cli

import (
	config2 "bioflows/config"
	"bioflows/executors"
	"bioflows/helpers"
	"bioflows/models"
	"bioflows/models/pipelines"
	"errors"
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/goombaio/dag"
	"strings"
)
func RecursiveProcessing(b *pipelines.BioPipeline,config models.FlowConfig) {
	if b.URL == "" && len(b.URL) == 0 && len(b.Steps) <= 0 {
		return
	}else{
		executors.PreprocessPipeline(b,config,nil)
		for idx , child := range b.Steps {
			RecursiveProcessing(&child,config)
			b.Steps[idx] = child
		}
	}

}

func GetRequirementsTableFor(toolPath string) (*simpletable.Table,error){
	table := simpletable.New()
	pipeline := &pipelines.BioPipeline{}
	config := models.FlowConfig{}
	err := helpers.ReadPipelineFile(pipeline,toolPath)
	if err != nil {
		return nil , err
	}
	filedetails := &helpers.FileDetails{}
	err = helpers.GetFileDetails(filedetails,toolPath)
	if err != nil {
		return nil , err
	}
	config[config2.WF_BF_TOOL_BASEPATH] = filedetails.Base
	config[config2.WF_BF_TOOL_PATH] = toolPath


	RecursiveProcessing(pipeline,config)
	graph , err := pipelines.CreateGraph(pipeline)
	if err != nil {
		return nil , err
	}
	var successors []*dag.Vertex

	if pipeline.Type == "tool" {

		successors = make([]*dag.Vertex,0)
		oneVertex := &dag.Vertex{
			Value: *pipeline,
		}
		successors = append(successors,oneVertex)
	}else{
		successors = graph.SourceVertices()
	}

	if len(successors) <= 0 {
		return nil , errors.New(fmt.Sprintf("BioFlows Definition File: %s is invalid.",pipeline.Name))
	}
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter,Text: "Step #"},
			{Align: simpletable.AlignCenter,Text: "Param. Name (required)"},
			{Align: simpletable.AlignCenter,Text: "Param. Description"},
		},
	}
	TotalInputs := make([]models.Parameter,0)
	if len(pipeline.Inputs) > 0 {
		TotalInputs = append(TotalInputs,pipeline.Inputs...)
	}
	for _ , parent := range successors {
		parentPipeline := parent.Value.(pipelines.BioPipeline)

		if strings.EqualFold(strings.ToLower(parentPipeline.Type),"pipeline") && len(parentPipeline.Steps) > 0 {

			nestedGraph , err  := pipelines.CreateGraph(&parentPipeline)
			if err != nil {
				return nil , errors.New("Unable to create graph for the nested pipeline.")
			}
			nested_successors := nestedGraph.SourceVertices()
			if len(nested_successors) <= 0 {
				return nil , errors.New(fmt.Sprintf("Nested Pipeline (%s) is invalid",parentPipeline.Name))
			}
			for _ , nested := range nested_successors {
				current := nested.Value.(pipelines.BioPipeline)
				TotalInputs = append(TotalInputs,current.Inputs...)
			}
		}else{
			TotalInputs = append(TotalInputs,parentPipeline.Inputs...)
		}
	}
	for index , param := range TotalInputs {
		r := []*simpletable.Cell {
			{Align: simpletable.AlignCenter,Text: fmt.Sprintf("%d",index+1)},
			{Align: simpletable.AlignCenter,Text: fmt.Sprintf("%s",param.Name)},
			{Align: simpletable.AlignCenter,Text: fmt.Sprintf("%s",param.GetDescription())},
		}
		table.Body.Cells = append(table.Body.Cells,r)
	}

	return table , nil
}
