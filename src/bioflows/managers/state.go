package managers

import "github.com/bioflows/src/bioflows/models"

type StateManager interface {
	Setup(map[string]interface{}) error
	GetStateByID(string) (interface{},error)
	GetPipelineState(string) (models.FlowConfig, error)
	SetStateByID(string,interface{}) error
	RemoveConfigByID(string) bool
}






