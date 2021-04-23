package cli

import (
	"bioflows/config"
	"bioflows/executors"
	"bioflows/helpers"
	"bioflows/models"
	"bioflows/models/pipelines"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func ReadConfig(cfgFile string) (models.FlowConfig,error) {
	workflowConfig := models.FlowConfig{}
	config_in , err := os.Open(cfgFile)
	config_contents , err := ioutil.ReadAll(config_in)
	var Configuration models.SystemConfig = models.SystemConfig{}
	err = yaml.Unmarshal(config_contents,&Configuration)
	if err != nil {
		fmt.Println(err.Error())
		return nil , err
	}
	workflowConfig.Fill(Configuration.ToMap())
	return workflowConfig,nil
}

func GetAttachableVolumes(step *pipelines.BioPipeline) ([]models.Parameter, error) {
	attachable := make(map[string]models.Parameter)
	//access the inputs of the parent workflow
	totalParams := make([]models.Parameter,1)
	totalParams = append(totalParams,step.Inputs...)
	totalParams = append(totalParams,step.Config...)
	for _ , param := range totalParams {
		paramName := strings.ToLower(param.Type)
		if (paramName == "dir" || paramName == "directory") && param.IsAttachable() {
			attachable[param.Name] = param
		}
	}
	volumes := make([]models.Parameter,1)
	for _,v := range attachable {
		volumes = append(volumes,v)
	}
	return volumes , nil
}

func RunTool(configFile string, toolPath string,workflowId string ,
	workflowName string,outputDir string,dataDir string,
	initialsConfig string,
	tconfig models.FlowConfig) error{
	tool := &pipelines.BioPipeline{}
	workflowConfig := models.FlowConfig{}
	if !helpers.IsValidUrl(toolPath){
		tool_in, err := os.Open(toolPath)
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
			return err
		}
	}else{
		//Download the tool remotely
		err := helpers.DownloadBioFlowFile(tool,toolPath)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error Downloading the file: %s",err.Error()))
			return err
		}
	}
	BfConfig , err := ReadConfig(configFile)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	workflowConfig.Fill(BfConfig)
	volumes , err := GetAttachableVolumes(tool)
	if err != nil {
		log.Fatal(fmt.Sprintf("Received Error: %s",err.Error()))
		return err
	}
	fmt.Println("Executing the current tool.")
	executor := executors.ToolExecutor{}
	executor.SetAttachableVolumes(volumes)
	executor.SetPipelineName(workflowName)
	workflowConfig[config.WF_INSTANCE_OUTDIR] = outputDir
	workflowConfig[config.WF_INSTANCE_DATADIR] = dataDir
	workflowConfig.Fill(tconfig)
	if len(initialsConfig) > 0 {
		initialParams, err := ReadParamsConfig(initialsConfig)
		if err != nil {
			return err
		}
		workflowConfig.Fill(initialParams)
	}
	tool_name := tool.Name
	if len(tool_name) <= 0 {
		tool_name = workflowName
	}
	var funcCall func(*models.ToolInstance , models.FlowConfig) (models.FlowConfig,error)
	if tool.Loop {
		funcCall = executor.RunToolLoop
	}else{
		// The tool is not loop
		funcCall = executor.Run
	}
	_ , err = funcCall(&models.ToolInstance{WorkflowID: workflowId,Name: workflowName ,WorkflowName: workflowName,Tool:tool.ToTool()},workflowConfig)
	if err != nil {
		fmt.Println(err)
	}
	return err

}
