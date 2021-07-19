package cli

import (
	"fmt"
	"github.com/bioflows/src/bioflows/config"
	"github.com/bioflows/src/bioflows/executors"
	"github.com/bioflows/src/bioflows/helpers"
	"github.com/bioflows/src/bioflows/models"
	"github.com/bioflows/src/bioflows/models/pipelines"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func RunPipeline(configFile,toolPath,outputDir,dataDir, initialsConfig string,clean bool,pconfig models.FlowConfig) error{
	fmt.Println(fmt.Sprintf("Using Configuration File: %s",configFile))
	pipeline := &pipelines.BioPipeline{}
	workflowConfig := models.FlowConfig{}
	fileDetails := &helpers.FileDetails{}
	err := helpers.GetFileDetails(fileDetails,toolPath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if fileDetails.Local {
		pipeline_in,err := os.Open(toolPath)
		if err != nil {
			fmt.Printf("There was an error opening the tool File: %s",err.Error())
			return err
		}
		//The tool is being run from a local directory

		mypipeline_contents , err := ioutil.ReadAll(pipeline_in)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error: %s",err.Error()))
			return err
		}
		err = yaml.Unmarshal([]byte(mypipeline_contents),pipeline)
		if err != nil {
			fmt.Printf("Error: %s",err.Error())
			return err
		}

	}else{
		//The mentioned tool is remote

		err = helpers.DownloadBioFlowFile(pipeline,toolPath)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error Downloading the file: %s",err.Error()))
			return err
		}
	}
	workflowConfig[config.WF_BF_TOOL_PATH] = toolPath
	workflowConfig[config.WF_BF_TOOL_BASEPATH] = fileDetails.Base
	workflowConfig[config.WF_BF_TOOL_LOCAL] = fileDetails.Local
	BfConfig, err := ReadConfig(configFile)
	if err != nil {
		fmt.Printf("Error Reading Config: %s",err.Error())
		return err
	}
	workflowConfig.Fill(BfConfig)
	workflowConfig[config.WF_INSTANCE_OUTDIR] = outputDir
	workflowConfig[config.WF_INSTANCE_DATADIR] = dataDir
	workflowConfig.Fill(pconfig)
	if len(initialsConfig) > 0 {
		initialParams, err := ReadParamsConfig(initialsConfig)
		if err != nil {
			return err
		}
		workflowConfig.Fill(initialParams)
	}
	fmt.Println(fmt.Sprintf("Executing Workflow: %s",pipeline.Name))
	executor := executors.DagExecutor{}
	err = executor.Setup(workflowConfig)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %s",err.Error()))
		return err
	}
	err =  executor.Run(pipeline,workflowConfig)
	if clean {
		executor.Clean()
	}
	return err
}
