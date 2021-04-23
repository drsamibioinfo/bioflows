package main

import (
	"bioflows/models/pipelines"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func main(){

	tool := &pipelines.BioPipeline{}
	tool_in, err := os.Open("/home/snouto/bioflows/scripts/pip.yml")
	if err != nil {
		fmt.Printf("There was an error opening the tool file, %v\n",err)
		os.Exit(1)
	}
	mytool_content, err := ioutil.ReadAll(tool_in)
	if err != nil {
		fmt.Printf("Error reading the contents of the tool , %v\n",err)
		os.Exit(1)
	}

	err = yaml.Unmarshal([]byte(mytool_content),tool)
	if err != nil {
		//fmt.Println("There was a problem unmarshaling the current tool")
		fmt.Println(err.Error())
		return
	}

	//executor := executors.ToolExecutor{}
	//workflowConfig := models.FlowConfig{}
	//var Configuration map[string]interface{} = make(map[string]interface{})
	//config_in , err := os.Open("bf.yaml")
	//config_contents , err := ioutil.ReadAll(config_in)
	//err = yaml.Unmarshal(config_contents,&Configuration)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	//workflowConfig.Fill(Configuration)
	//workflowConfig[config.WF_INSTANCE_OUTDIR] = "/home/snouto/workflows"
	//currentTool := models.ToolInstance{WorkflowID: "myworkflowId",Name: "mytool",WorkflowName: "MyworkflowName",Tool:tool}
	//currentTool.Prepare()
	//_, err = executor.Run(&currentTool,workflowConfig)
	//if err != nil {
	//	fmt.Println(err)
	//}


//
//	fmt.Println(fmt.Sprintf(`
//	Before : %v,
//	After : %v
//`,tool.Scripts[0].IsBefore(),tool.Scripts[0].IsAfter()))
}
