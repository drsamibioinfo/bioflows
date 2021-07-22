package main

import (
	"github.com/bioflows/src/bioflows/config"
	"github.com/bioflows/src/bioflows/executors"
	"github.com/bioflows/src/bioflows/models"
	"github.com/bioflows/src/bioflows/models/pipelines"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func main(){
	pipeline := &pipelines.BioPipeline{}
	tool_in, err := os.Open("/home/snouto/projects/bioflows/scripts/secondpipe.yaml")

	if err != nil {
		fmt.Printf("There was an error opening the tool file, %v\n",err)
		os.Exit(1)
	}
	mytool_content, err := ioutil.ReadAll(tool_in)
	if err != nil {
		fmt.Printf("Error reading the contents of the tool , %v\n",err)
		os.Exit(1)
	}
	err = yaml.Unmarshal([]byte(mytool_content),pipeline)
	if err != nil {
		//fmt.Println("There was a problem unmarshaling the current tool")
		fmt.Println(err.Error())
		return
	}
	executor := executors.DagExecutor{}
	workflowConfig := models.FlowConfig{}
	var Configuration models.SystemConfig = models.SystemConfig{}
	config_in , err := os.Open(".bf.yaml")
	config_contents , err := ioutil.ReadAll(config_in)
	err = yaml.Unmarshal(config_contents,&Configuration)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	workflowConfig.Fill(Configuration.ToMap())
	workflowConfig[config.WF_INSTANCE_OUTDIR] = "/home/snouto/workflows"
	err = executor.Setup(workflowConfig)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	finalConfig := executor.Run(pipeline,workflowConfig)
	fmt.Println(finalConfig)
	executor.Clean()
	//b , err := pipelines.PreparePipeline(pipeline)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	//g , err := pipelines.CreateGraph(b)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	//gg, _ , err := pipelines.ToDotGraph(b,g)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	//fmt.Print(gg)
	//parents := g.SourceVertices()
	//for _, v := range parents {
	//	fmt.Println("Executing: ",v.ID)
	//	if v.Children.Size() > 0 {
	//		for _ , c := range v.Children.Values(){
	//			cv := c.(*dag.Vertex)
	//			fmt.Println("===Executing : ",cv.ID)
	//		}
	//	}
	//}

	//fmt.Println(pipeline.ToJson())
}

