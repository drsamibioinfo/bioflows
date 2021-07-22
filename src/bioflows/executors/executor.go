package executors

import "github.com/bioflows/src/bioflows/models"

/*
  Base Interface to be implemented by different Tool Executors..
 */
type Executor interface {
	Run(*models.ToolInstance, models.FlowConfig) (models.FlowConfig, error)
}
