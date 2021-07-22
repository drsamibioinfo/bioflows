package pipelines

import (
	"github.com/bioflows/src/bioflows/expr"
	"github.com/bioflows/src/bioflows/models"
	"strings"
)

func getInputIndexWithName(l []models.Parameter,name string) int {
	var foundIndex int  = -1
	for index, param := range l {
		if strings.EqualFold(param.Name,name){
			foundIndex = index
			break
		}
	}
	return foundIndex
}
func evalTargetInputs(t *BioPipeline,config models.FlowConfig) {
	exprManager := &expr.ExprManager{}
	if t.Inputs != nil || len(t.Inputs) > 0 {
		for _ , param := range t.Inputs {
			paramValue := param.GetParamValue()
			param.Value = exprManager.Render(paramValue,config)
		}
	}
}

func Clone(o *BioPipeline, t *BioPipeline, config models.FlowConfig) error {
	// TODO: Perform the clone process
	// It overrides what is inside t into o
	o.URL = t.URL
	o.ImageId = t.ImageId
	o.Caps = t.Caps
	o.Type = t.Type
	o.BioflowId = t.BioflowId
	if len(o.Name) <= 0{
		o.Name = t.Name
	}
	o.Description = t.Description
	if o.Discussions == nil {
		o.Discussions = make([]string,0)
	}
	o.Discussions = append(o.Discussions,t.Discussions...)
	if o.Website == ""{
		o.Website = t.Website
	}
	o.Version = t.Version
	o.Website = t.Website
	o.Icon = t.Icon
	o.Maintainer = t.Maintainer
	if o.Scripts == nil {
		o.Scripts = make([]models.Script,0)

	}
	o.Scripts = append(o.Scripts,t.Scripts...)
	if o.References == nil {
		o.References = make([]models.Reference,0)
	}
	o.References = append(o.References,t.References...)
	// Evaluate target inputs before copying
	evalTargetInputs(t,config)
	// Copy the inputs section
	//Taking into account any new additional configuration
	if o.Inputs == nil {
		o.Inputs = make([]models.Parameter,0)
		o.Inputs = append(o.Inputs,t.Inputs...)
	}else{
		allParams := make([]models.Parameter,0)
		for _, param := range o.Inputs {
			targetIndex := getInputIndexWithName(t.Inputs,param.Name)
			if targetIndex == -1 {
				allParams = append(allParams,param)
				continue
			}
			t.Inputs[targetIndex].Fill(&param)
			allParams = append(allParams,t.Inputs[targetIndex])
		}
		o.Inputs = make([]models.Parameter,len(allParams))
		copy(o.Inputs,allParams)
	}

	// Copy the configuration section
	//Taking into account any new additional configuration
	if o.Config == nil {
		o.Config = make([]models.Parameter,0)
		o.Config = append(o.Config,t.Config...)
	}else{

		allParams := make([]models.Parameter,0)
		for _, param := range o.Config {
			targetIndex := getInputIndexWithName(t.Config,param.Name)
			if targetIndex == -1 {
				allParams = append(allParams,param)
				continue
			}
			t.Config[targetIndex].Fill(&param)
			allParams = append(allParams,t.Config[targetIndex])
		}
		o.Config = make([]models.Parameter,len(allParams))
		copy(o.Config,allParams)
	}

	// Copy the outputs section
	//Taking into account any new additional configuration
	if o.Outputs == nil {
		o.Outputs = make([]models.Parameter,0)
		o.Outputs = append(o.Outputs,t.Outputs...)
	}else{
		allParams := make([]models.Parameter,0)
		for _, param := range o.Outputs {
			targetIndex := getInputIndexWithName(t.Outputs,param.Name)
			if targetIndex == -1 {
				allParams = append(allParams,param)
				continue
			}
			t.Outputs[targetIndex].Fill(&param)
			allParams = append(allParams,t.Outputs[targetIndex])
		}
		o.Outputs = make([]models.Parameter,len(allParams))
		copy(o.Outputs,allParams)
	}
	o.Command = t.Command
	o.Dependencies = make([]string,len(t.Dependencies))
	copy(o.Dependencies,t.Dependencies)
	o.Deprecated = t.Deprecated
	o.Conditions = make([]models.Scriptable,len(t.Conditions))
	copy(o.Conditions,t.Conditions)

	o.Steps = make([]BioPipeline,len(t.Steps))
	copy(o.Steps,t.Steps)
	if len(t.Steps) > 0 {
		o.Type = "pipeline"
	}
	return nil
}
