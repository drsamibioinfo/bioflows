package cli

import (
	"github.com/bioflows/src/bioflows/helpers"
	"github.com/bioflows/src/bioflows/models/pipelines"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
)

func ValidateYAML(filePath string) (bool,error){
	// Test if the current file is remote or local
	//if the file is remote download and save it
	//if the file is local, use it
	var data []byte = nil
	var err error = nil
	pipeline := &pipelines.BioPipeline{}
	if helpers.IsValidUrl(filePath) {
		//The file is remote URI
		//Download and save the file to temporary directory
		resp , err := http.Get(filePath)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error Downloading the file: %s",err.Error()))
			return false , err
		}
		data , err = ioutil.ReadAll(resp.Body)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error Validating the file: %s",err.Error()))
			return false , err
		}
	}else{
		//validate the file
		tool_in , err := os.Open(filePath)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error Opening the file: %s",err.Error()))
			return false,err
		}
		//read the entire contents of the file
		data , err = ioutil.ReadAll(tool_in)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error Reading the file: %s",err.Error()))
			return false, err
		}
	}
	err = yaml.Unmarshal(data,pipeline)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error Validating the file: %s",err.Error()))
		return false , err
	}
	return true , nil
}