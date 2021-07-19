package cli

import (
	"github.com/bioflows/src/bioflows/helpers"
	"github.com/bioflows/src/bioflows/models/pipelines"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func RenderGraphViz(pipelineFile string) (string,error) {
	pipeline := &pipelines.BioPipeline{}

	if !helpers.IsValidUrl(pipelineFile) {
		tool_in, err := os.Open(pipelineFile)

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
			return "" , err
		}
	}else{
		// that means the file is remote
		err :=helpers.DownloadBioFlowFile(pipeline,pipelineFile)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error Downloading the file: %s",err.Error()))
			return "", err
		}
	}
	b , err := pipelines.PreparePipeline(pipeline,nil)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	g , err := pipelines.CreateGraph(b)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return pipelines.ToDotGraph(b,g)
}
