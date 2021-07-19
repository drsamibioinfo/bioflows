package managers

import (
	"github.com/bioflows/src/bioflows/context"
	"github.com/bioflows/src/bioflows/helpers"
	"github.com/bioflows/src/bioflows/models"
	"fmt"
	"strings"
)

type LocalStateManager struct {

	context *context.BioContext

}
func (c *LocalStateManager) RemoveConfigByID(key string) bool {
	keys := c.filterKeys(key)
	if len(keys) > 0 {
		for _ , key := range keys {
			c.context.DeleteByKey(key)
		}
	}
	return false
}
func (c *LocalStateManager) filterKeys(query string) []string {
	filteredKeys := make([]string,0)
	keys := c.context.GetKeys()
	for _ , key := range keys{
		if strings.HasPrefix(key,query) {
			filteredKeys = append(filteredKeys,key)
		}
	}
	return filteredKeys
}
func (c *LocalStateManager) GetPipelineState(pipelineKey string) (models.FlowConfig, error) {
	finalConfig := models.FlowConfig{}
	filteredKeys := c.filterKeys(pipelineKey)
	if len(filteredKeys) <= 0 {
		return nil , ERR_NOT_FOUND
	}
	for _ , key := range filteredKeys {
		state , err := c.GetStateByID(key)
		if err != nil {
			continue
		}
		finalConfig[helpers.GetToolIdFromKey(key)] = state

	}
	return finalConfig , nil
}
func (c *LocalStateManager) GetStateByID(stepId string) (interface{},error){
	if c.context.HasKey(stepId){
		return c.context.GetKey(stepId)
	}
	return nil , nil
}
func (c *LocalStateManager) SetStateByID(stepId string,config interface{}) error {
	r := c.context.AddVar(stepId,config)
	if !r{
		return fmt.Errorf("Unable to add the state for the given stepId")
	}
	return nil
}

func (c *LocalStateManager) Setup(config map[string]interface{}) error {
	c.context = &context.BioContext{}
	return nil
}
