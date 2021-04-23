package executors

import "bioflows/models"

/*
  Base Interface to be implemented by different Tool Executors..
 */
type Executor interface {
	Run(*models.ToolInstance, models.FlowConfig) (models.FlowConfig, error)
}
