package executors

import (
	"github.com/bioflows/src/bioflows/helpers"
	"github.com/bioflows/src/bioflows/models"
	"github.com/bioflows/src/bioflows/models/pipelines"
	"strings"
)

/*

This function will use "URL" property in Biopipeline definition file and use it to download the tool

 */

var UseUrl TransformCall = func (b *pipelines.BioPipeline,config models.FlowConfig) error{
	if len(b.URL) > 0 {
		//check to see the pipeline
		if helpers.IsValidUrl(b.URL) {
			// That means it is a valid URL, so download the file
			t := &pipelines.BioPipeline{}
			fileDetails := &helpers.FileDetails{}
			err := helpers.GetFileDetails(fileDetails,b.URL)
			if err != nil {
				return err
			}
			currPath := strings.Join([]string{config["bf_tool_basepath"].(string),fileDetails.FileName},"")
			err = helpers.GetFileDetails(fileDetails,currPath)
			if err != nil {
				return err
			}
			//tool_is_local := config[config2.WF_BF_TOOL_LOCAL].(bool)
			if fileDetails.Local {
				// that means it is a local file , so read it
				err = helpers.ReadLocalBioFlowFile(t,currPath)
			}else{
				// It could be a remote http/https command, so download the file
				err = helpers.DownloadBioFlowFile(t,currPath)
			}
			if err != nil {
				return err
			}
			err = pipelines.Clone(b,t,config)
			return err
		}
	}
	return nil
}

var UseBioFlowId TransformCall = func (b *pipelines.BioPipeline,config models.FlowConfig) error {
	t := &pipelines.BioPipeline{}
	helpers.DownloadFromBioFlowsHub(t,b.BioflowId,b.Version)
	return pipelines.Clone(b,t,config)
}


