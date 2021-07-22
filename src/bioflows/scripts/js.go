package scripts

import (
	config2 "github.com/bioflows/src/bioflows/config"
	"github.com/bioflows/src/bioflows/helpers"
	"github.com/bioflows/src/bioflows/models"
	"github.com/bioflows/src/bioflows/scripts/io"
	"errors"
	"github.com/dop251/goja"
	"io/ioutil"
	"strings"
)

type ScriptManager interface {
	Prepare(toolInstance *models.ToolInstance)
	RunScript(script models.Script,config models.FlowConfig) (error)
	RunAfter(script models.Script,config models.FlowConfig) error
	getCode(script models.Script , config models.FlowConfig) (string , error)
}

type JSScriptManager struct {

	toolInstance *models.ToolInstance

}
func (manager *JSScriptManager) Prepare(toolInstance *models.ToolInstance) {
	manager.toolInstance = toolInstance

}
func (manager *JSScriptManager) getCode(script models.Script , config models.FlowConfig) (string,error) {
	code := script.Code.ToString()
	if len(code) > 2 {
		return code, nil
	}
	if script.CodeFile != "" && len(script.CodeFile) > 1 {
		tool_is_local := config[config2.WF_BF_TOOL_LOCAL].(bool)
		details := helpers.FileDetails{}
		err := helpers.GetFileDetails(&details,script.CodeFile)
		if err != nil {
			return "" , err
		}
		tool_basePath := config[config2.WF_BF_TOOL_BASEPATH].(string)
		codeFile_location := strings.Join([]string{tool_basePath,details.Base},"")
		if !tool_is_local{
			data , err := helpers.DownloadRemoteFile(codeFile_location)
			if err != nil {
				return "" , err
			}
			return string(data) , nil
		}else{
			if details.Local {
				codefile_data , err := ioutil.ReadFile(codeFile_location)
				if err != nil {
					return "" , err
				}
				return string(codefile_data) , nil
			}else{
				data , err := helpers.DownloadRemoteFile(script.CodeFile)
				if err != nil {
					return "" , err
				}
				return string(data) , nil
			}
		}
	}
	return "" , errors.New("invalid script directive. no code found")
}
func (manager *JSScriptManager) RunScript(script models.Script,config models.FlowConfig) error {
	vm := goja.New()
	if manager.toolInstance != nil {
		config["command"] = manager.toolInstance.Command.ToString()
	}else{
		config["command"] = ""
	}
	vm.Set("self",config)
	vm.Set("io",&io.IO{
		VM: vm,
	})
	code , err := manager.getCode(script,config)
	if err != nil {
		return err
	}
	_ , err = vm.RunString(code)
	if err != nil {
		return  err
	}
	return nil
}

func (manager *JSScriptManager) RunAfter(script models.Script,config models.FlowConfig) error {
	return manager.RunScript(script,config)
}
