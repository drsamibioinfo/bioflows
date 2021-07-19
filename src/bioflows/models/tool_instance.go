package models

import (
	"github.com/bioflows/src/bioflows/helpers/id"
	"strings"
	"time"
)

type ToolInstance struct {
	*Tool
	Name string `json:"name"`
	WorkflowID string `json:"workflowId"`
	WorkflowName string `json:"workflowName"`
	StartTime time.Duration
	EndTime time.Duration
	Status int `json:"status"`

}

func (instance *ToolInstance) Prepare() {
	//Generate random unique ID if the instance Tool ID is not set
	if len(instance.Tool.ID) <= 0 {
		instance.Tool.ID , _ = id.NewID()
	}
	//if the tool name is not set, then use the tool ID
	if len(instance.Tool.Name) <= 0{
		instance.Tool.Name = instance.Tool.ID
	}else{
		// If the tool name exists, replace whitespace with underscores
		instance.Tool.Name = strings.ReplaceAll(instance.Tool.Name," ","_")
	}
	//If the tool name is set , use that as the tool instance name
	instance.Name = instance.Tool.Name

	//If the workflow ID is not set, then generate new random unique ID
	if len(instance.WorkflowID) <= 0{
		instance.WorkflowID , _ = id.NewID()
	}
}

func (instance ToolInstance) PrepareCommand() []string{
	splitted_command := strings.Split(string(instance.Command)," ")
	if splitted_command[0] == "sudo"{
		splitted_command = splitted_command[1:]
	}
	instanceCommand := make([]string,0)
	instanceCommand = append(instanceCommand,splitted_command...)
	return instanceCommand
}

func (instance ToolInstance) PrepareInstallations() [][]string{

	installations := make([][]string,0)
	for _,installation := range instance.Dependencies {
		temp_installation := make([]string,0)
		splitted_installation := strings.Split(installation," ")
		if splitted_installation[0] == "sudo" {
			splitted_installation = splitted_installation[1:]
		}
		temp_installation = append(temp_installation,splitted_installation...)
		installations = append(installations,temp_installation)

	}
	return installations

}


func (instance ToolInstance) GetContainerName() string {
	return strings.Join([]string{instance.WorkflowID,instance.ID,instance.Name},"_")
}

