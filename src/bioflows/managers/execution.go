package managers

import "github.com/bioflows/src/bioflows/models"

type ExecutionPlanManager struct {
	contextManager *ContextManager
	config models.FlowConfig
	serviceManager *ClusterServiceManager

}

func (e *ExecutionPlanManager) Setup(config map[string]interface{}) error {
	e.config = config
	e.serviceManager = &ClusterServiceManager{}
	e.serviceManager.Setup(config)
	return nil
}

func (e *ExecutionPlanManager) SetContextManager(contextManager *ContextManager){
	e.contextManager = contextManager
}



