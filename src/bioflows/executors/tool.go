package executors

import (
	"bioflows/config"
	dockcontainer "bioflows/container"
	"bioflows/expr"
	"bioflows/helpers"
	"bioflows/models"
	"bioflows/process"
	"bioflows/scripts"
	"bioflows/virtualization"
	"fmt"
	"github.com/aidarkhanov/nanoid"
	"github.com/docker/docker/api/types/container"
	"log"
	"net/smtp"
	"os"
	"sort"
	"strings"
)

type ToolExecutor struct {
	ToolInstance            *models.ToolInstance
	ContainerManager        *virtualization.VirtualizationManager
	toolLogger              *log.Logger
	flowConfig              models.FlowConfig
	exprManager             *expr.ExprManager
	pipelineName            string
	dockerManager           *dockcontainer.DockerManager
	hostOutputDir           string
	hostDataDir             string
	pipelineContainerConfig *models.ContainerConfig
	AttachableVolumes []models.Parameter
	basePath string
	instanceId string
	explain bool
}
func (t *ToolExecutor) GetInstanceId() string{
	return t.instanceId
}
func (t *ToolExecutor) GetToolKey() string{
	return strings.Join([]string{t.basePath,t.GetInstanceId()},"/")
}
func (t *ToolExecutor) SetBasePath(basePath string) {
	t.basePath = basePath
}
func (e *ToolExecutor) SetContainerConfiguration(containerConfig *models.ContainerConfig){
	e.pipelineContainerConfig = containerConfig
}
func (e *ToolExecutor) SetPipelineName(name string) {
	//e.pipelineName = strings.
	e.pipelineName = strings.ReplaceAll(name," ","_")
}

func (e *ToolExecutor) notify(tool *models.ToolInstance) {
	if tool.Notification != nil {

		if EmailSection , ok := e.flowConfig["email"]; !ok {
			err := fmt.Errorf("Tool (%s) requires Email notification but BioFlows Configuration is missing The Email Section...",tool.Name)
			e.Log(err.Error())
		}else{
			email := EmailSection.(map[string]interface{})
			username := fmt.Sprintf("%v",email["username"])
			password := fmt.Sprintf("%v",email["password"])
			SMTPHost := fmt.Sprintf("%v",email["host"])
			SMTPPort := email["port"].(int)
			message := []byte(tool.Notification.Body)
			auth := smtp.PlainAuth("",username,password,SMTPHost)
			To := strings.Split(tool.Notification.To,",")
			e.Log("Start Sending Email Notifications....")
			err := smtp.SendMail(fmt.Sprintf("%s:%d",SMTPHost,SMTPPort),auth,username,To,message)
			if err != nil {
				e.Log(err.Error())
			}
			e.Log(fmt.Sprintf("Tool (%s): The Email was sent Successfully....",tool.Name))
		}

	}
}

func (e *ToolExecutor) prepareParameters() models.FlowConfig {

	flowConfig := make(models.FlowConfig)
	toolConfigKey , toolDir , _ := e.GetToolOutputDir()
	flowConfig[toolConfigKey] = toolDir
	flowConfig["self_dir"] = toolDir
	flowConfig["location"] = toolDir
	//Copy all flow configs at the workflow level into the current tool flowconfig
	if len(e.flowConfig) > 0 {
		for k,v := range e.flowConfig{
			flowConfig[k] = v
		}
	}
	//We should also copy the configuration details
	if len(e.ToolInstance.Config) > 0 {
		configs := make(map[string]interface{})
		for _ , param := range e.ToolInstance.Config {
			if param.Value == nil {
				configs[param.Name] = nil
				continue
			}
			paramValue := e.exprManager.Render(param.GetParamValue(),flowConfig)
			configs[param.Name] = paramValue
		}
		for k , v := range configs {
			flowConfig[k] = v
		}
	}
	if len(e.ToolInstance.Inputs) > 0 {
		inputs := make(map[string]string)
		for _ , param := range e.ToolInstance.Inputs{
			if param.Value == nil {
				continue
			}
			paramValue := e.exprManager.Render(param.GetParamValue(),flowConfig)
			inputs[param.Name] = paramValue
		}
		//Append the processed input parameters into the current flowConfig
		for k , v := range inputs {
			flowConfig[k] = v
		}
	}
	if len(e.ToolInstance.Outputs) > 0{

		//Prepare outputs
		outputs := make(map[string]string)
		for _ , param := range e.ToolInstance.Outputs {
			paramValue := e.exprManager.Render(param.GetParamValue(),flowConfig)
			outputs[param.Name] = paramValue
		}
		for k,v  := range outputs{
			flowConfig[k] = v
		}
	}
	e.addImplicitVariables(&flowConfig)
	return flowConfig
}
func (e *ToolExecutor) addImplicitVariables(config *models.FlowConfig){
	//This variable might be used by embedded scripts to impede the firing of the current tool
	//Defaults to false
	(*config)["impede"] = false
}
func (e *ToolExecutor) executeBeforeScripts() (map[string]interface{},error) {
	configuration := e.prepareParameters()
	configuration["command"] = e.ToolInstance.Command.ToString()
	beforeScripts := make([]models.Script,0)
	for idx , script := range e.ToolInstance.Scripts {
		if script.IsBefore() {
			if script.Order <= 0 {
				script.Order = idx + 1
			}
			beforeScripts = append(beforeScripts,script)
		}
	}
	//sort the scripts according to the assigned orders
	sort.Slice(beforeScripts, func(i, j int) bool {

		return beforeScripts[i].Order < beforeScripts[j].Order

	})
	for _ , beforeScript := range beforeScripts {
		var scriptManager scripts.ScriptManager
		switch strings.ToLower(beforeScript.Type) {
		case "js":
			fallthrough
		default:
			scriptManager = &scripts.JSScriptManager{}
			scriptManager.Prepare(e.ToolInstance)
		}
		err := scriptManager.RunScript(beforeScript,configuration)
		if err != nil {
			return configuration , err
		}
	}
	return configuration , nil
}
func (e *ToolExecutor) executeAfterScripts(configuration map[string]interface{}) (map[string]interface{},error)  {

	afterScripts := make([]models.Script,0)
	for idx , script := range e.ToolInstance.Scripts {
		if script.IsAfter() {
			if script.Order <= 0 {
				script.Order = idx + 1
			}
			afterScripts = append(afterScripts,script)
		}
	}
	//sort the scripts according to the assigned orders
	sort.Slice(afterScripts, func(i, j int) bool {

		return afterScripts[i].Order < afterScripts[j].Order

	})
	for _ , afterScript := range afterScripts {
		var scriptManager scripts.ScriptManager
		switch strings.ToLower(afterScript.Type) {
		case "js":
			fallthrough
		default:
			scriptManager = &scripts.JSScriptManager{}
			scriptManager.Prepare(e.ToolInstance)
		}
		err := scriptManager.RunAfter(afterScript,configuration)
		if err != nil {
			return configuration , err
		}
	}
	return configuration , nil
}
func (e *ToolExecutor) GetToolOutputDir() (toolConfigKey string,toolDir string,err error) {
	workflowOutputDir , ok := e.flowConfig[config.WF_INSTANCE_OUTDIR]
	if !ok {
		err = fmt.Errorf("Unable to get the Tool/Workflow Output Directory")
		return
	}
	toolOutputDir := strings.Join([]string{e.pipelineName,e.ToolInstance.ID},"_")
	toolDir = strings.Join([]string{fmt.Sprintf("%v",workflowOutputDir),toolOutputDir},"/")
	preparedToolName := strings.ReplaceAll(e.ToolInstance.ID," ","_")
	toolConfigKey = fmt.Sprintf("%s_dir",preparedToolName)
	return
}
func (e *ToolExecutor) CreateOutputFile(name string,ext string) (string,error) {

	outputFile := strings.Join([]string{e.ToolInstance.ID,name},"_")
	outputFile = strings.Join([]string{outputFile,ext},".")
	_ , toolOutputDir , err := e.GetToolOutputDir()
	if err != nil {
		return "" , err
	}
	os.Mkdir(toolOutputDir,config.FILE_MODE_WRITABLE_PERM)
	outputFile = strings.Join([]string{toolOutputDir,outputFile},"/")
	return outputFile , nil

}

func (e *ToolExecutor) SetExplain(explain bool){
	e.explain = explain
}

func (e *ToolExecutor) init(flowConfig models.FlowConfig) error {
	e.ContainerManager = nil
	instanceId , err := nanoid.New()
	if err != nil {
		return err
	}
	e.instanceId = instanceId
	e.flowConfig = flowConfig
	e.AttachableVolumes = make([]models.Parameter,1)
	e.hostDataDir = fmt.Sprintf("%v",e.flowConfig[config.WF_INSTANCE_DATADIR])
	e.hostOutputDir = fmt.Sprintf("%v",e.flowConfig[config.WF_INSTANCE_OUTDIR])
	e.exprManager = &expr.ExprManager{}
	// initialize the tool logger
	logFileName , err := e.CreateOutputFile("logs","logs")
	if err != nil {
		return err
	}
	e.toolLogger = &log.Logger{}
	e.toolLogger.SetPrefix(fmt.Sprintf("%v: ",config.BIOFLOWS_DISPLAY_NAME))
	file , err := os.Create(logFileName)
	if err != nil {
		fmt.Printf("Can't Create Tool (%s) log file %s",e.ToolInstance.Name, logFileName)
		return err
	}
	e.toolLogger.SetOutput(file)
	//initialize Docker
	hostConfig := &container.HostConfig{}
	hostConfig.Binds = append(hostConfig.Binds,fmt.Sprintf("%s:%s",e.hostOutputDir,
		e.hostOutputDir),
		fmt.Sprintf("%s:%s",e.hostDataDir,e.hostDataDir))
	e.dockerManager = &dockcontainer.DockerManager{
		DockerConfig:     nil,
		HostConfig:       hostConfig,
		NetworkingConfig: nil,
	}
	e.dockerManager.SetLogger(e.toolLogger)



	return nil
}
func (e *ToolExecutor) Log(logs ...interface{}) {
	e.toolLogger.Println(logs...)
	fmt.Println(logs...)
}
func (e *ToolExecutor) isDockerized() bool {
	result := e.ToolInstance.ImageId != "" && len(e.ToolInstance.ImageId) > 1
	return result
}
func (e *ToolExecutor) execute() (models.FlowConfig,error) {
	//prepare parameters
	toolConfig, err := e.executeBeforeScripts()
	if err != nil {
		return toolConfig,err
	}
	if toolConfig["impede"] == true{
		e.Log(fmt.Sprintf("Tool (%s) has been impeded.",e.ToolInstance.Name))
		toolConfig["exitCode"] = 0
		toolConfig["status"] = true
		return toolConfig,nil
	}
	//Defer the notification till the end of the execute method
	defer e.notify(e.ToolInstance)
	toolCommandStr := fmt.Sprintf("%v",toolConfig["command"])
	toolCommand := e.exprManager.Render(toolCommandStr,toolConfig)
	toolConfigKey, _ , _ := e.GetToolOutputDir()
	var exitCode int
	var toolErr error
	var outputBytes []byte
	var errorBytes []byte
	var tempContainerConfig *models.ContainerConfig = nil
	if e.ToolInstance.ContainerConfig != nil {
		tempContainerConfig = e.ToolInstance.ContainerConfig
	}else{
		tempContainerConfig = e.pipelineContainerConfig
	}
	e.Log(fmt.Sprintf("RunScript Command : %s",toolCommand))
	if e.explain{
		fmt.Printf("Explain => Tool Name: %s , Command: %s\n",e.ToolInstance.ID,toolCommand)
		goto AfterScriptsAndExit
	}
	if e.isDockerized() {
		var imageURL string
		if tempContainerConfig == nil {
			imageURL = fmt.Sprintf("%s/%s",dockcontainer.DOCKER_REPOSITORY,e.ToolInstance.ImageId)
		}else{
			imageURL = fmt.Sprintf("%s/%s",tempContainerConfig.URL,e.ToolInstance.ImageId)
		}
		//first try to pull the image
		output , err := e.dockerManager.PullImage(imageURL,tempContainerConfig)
		if err != nil {
			return nil , err
		}
		//Log the output
		e.Log(output)

		out,outErr,toolErr := e.dockerManager.RunContainer(toolConfigKey,e.ToolInstance.ImageId,[]string{
			"bash",
			"-c",
			toolCommand,
		},false)
		if toolErr != nil {
			errorBytes = []byte(toolErr.Error())
			exitCode = 1
		}else{
			exitCode = 0
		}
		if out != nil {
			outputBytes = out.Bytes()
		}
		if outErr != nil {
			errorBytes = outErr.Bytes()
		}
	}else{

		executor := &process.CommandExecutor{Command: toolCommand,CommandDir: fmt.Sprintf("%v",toolConfig[toolConfigKey])}
		executor.Init()
		exitCode , toolErr  = executor.Run()
		outputBytes = executor.GetOutput().Bytes()
		errorBytes = executor.GetError().Bytes()
	}
	AfterScriptsAndExit:
	toolConfig , err = e.executeAfterScripts(toolConfig)
	toolConfig["exitCode"] = exitCode
	if toolErr != nil {
		toolConfig["status"] = false
	}
	if exitCode == 0 {
		toolConfig["status"] = true
	}
	if exitCode > 0 {
		toolConfig["status"] = false
	}
	delete(toolConfig,"self_dir")

	defer func(){
		 // This deferred function will delete all config keys
		// from the current tool before it exits, because we don't want to store these configs
		// Configurations of a tool are meant to be inclusive to the execution instance of the tool
		// it is not meant to be stored or injected to the next tool
		if len(e.ToolInstance.Config) > 0 {
			for _ , param := range e.ToolInstance.Config {
				delete(toolConfig,param.Name)
			}
		}
	}()
	defer e.Log(fmt.Sprintf("Tool: %s has finished.",e.ToolInstance.Name))
	if e.ToolInstance.Shadow{
		return toolConfig,toolErr
	}
	//Create output file for the output of this tool
	toolOutputFile , err := e.CreateOutputFile("stdout","out")
	if err != nil {
		return toolConfig,err
	}
	err = helpers.WriteOrAppend(toolOutputFile,outputBytes,config.FILE_MODE_WRITABLE_PERM)
	if err != nil {
		return toolConfig,err
	}
	//Create err file for this tool
	toolErrFile , err := e.CreateOutputFile("stderr","err")
	if err != nil {
		return toolConfig,err
	}
	err = helpers.WriteOrAppend(toolErrFile,errorBytes,config.FILE_MODE_WRITABLE_PERM)
	if err != nil {
		return toolConfig,err
	}
	//Delete the temporary mapped self_dir key from the configuration
	return toolConfig,toolErr
}
func (e *ToolExecutor) RunToolLoop(t *models.ToolInstance , workflowConfig models.FlowConfig)  (models.FlowConfig,error) {
	e.ToolInstance = t
	err := e.init(workflowConfig)
	if err != nil {
		return nil , err
	}
	e.Log(fmt.Sprintf("Tool (%s) is prepared successfully. ",t.Name))
	return e.executeLoop()
}
func (e *ToolExecutor) executeLoop()  (models.FlowConfig,error) {
	toolConfig , err := e.executeBeforeScripts()
	if err != nil {
		return toolConfig , err
	}
	if toolConfig["impede"] == true{
		e.Log(fmt.Sprintf("Tool (%s) has been impeded.",e.ToolInstance.Name))
		toolConfig["exitCode"] = 0
		toolConfig["status"] = true
		return toolConfig,nil
	}
	defer e.notify(e.ToolInstance)
	if len(e.ToolInstance.LoopVar) == 0 {
		errStr := "Tool is loop but no loop variable has been defined.. aborting..."
		e.Log(fmt.Sprintf(errStr))
		return toolConfig , fmt.Errorf(errStr)
	}
	var exitCode int
	var toolErr error
	var outputBytes []byte
	var errorBytes []byte
	if loop_elements , ok := toolConfig[e.ToolInstance.LoopVar]; ok {
		if elements , islist := loop_elements.([]interface{}); islist {
			for idx , el := range elements {
				toolConfig[fmt.Sprintf("%s_item",e.ToolInstance.LoopVar)] = el
				toolConfig[fmt.Sprintf("loop_index")] = idx
				toolCommandStr := fmt.Sprintf("%v",toolConfig["command"])
				toolCommand := e.exprManager.Render(toolCommandStr,toolConfig)
				toolConfigKey, _ , _ := e.GetToolOutputDir()

				var tempContainerConfig *models.ContainerConfig = nil
				if e.ToolInstance.ContainerConfig != nil {
					tempContainerConfig = e.ToolInstance.ContainerConfig
				}else{
					tempContainerConfig = e.pipelineContainerConfig
				}
				e.Log(fmt.Sprintf("RunScript Command : %s",toolCommand))
				if e.explain {
					fmt.Printf("Explain => Tool Name: %s , Command: %s\n",e.ToolInstance.ID,toolCommand)
					goto AfterScriptsAndExit
				}
				if e.isDockerized() {
					var imageURL string
					if tempContainerConfig == nil {
						imageURL = fmt.Sprintf("%s/%s",dockcontainer.DOCKER_REPOSITORY,e.ToolInstance.ImageId)
					}else{
						imageURL = fmt.Sprintf("%s/%s",tempContainerConfig.URL,e.ToolInstance.ImageId)
					}
					//first try to pull the image
					output , err := e.dockerManager.PullImage(imageURL,tempContainerConfig)
					if err != nil {
						return nil , err
					}
					//Log the output
					e.Log(output)
					out,outErr,toolErr := e.dockerManager.RunContainer(toolConfigKey,e.ToolInstance.ImageId,[]string{
						"bash",
						"-c",
						toolCommand,
					},false)
					if toolErr != nil {
						errorBytes = append(errorBytes,[]byte(toolErr.Error())...)
						exitCode = 1
					}else{
						exitCode = 0
					}
					if out != nil {
						outputBytes = append(outputBytes,out.Bytes()...)
					}
					if outErr != nil {
						errorBytes = append(errorBytes,outErr.Bytes()...)
					}
				}else{

					executor := &process.CommandExecutor{Command: toolCommand,CommandDir: fmt.Sprintf("%v",toolConfig[toolConfigKey])}
					executor.Init()
					exitCode , toolErr  = executor.Run()
					outputBytes = append(outputBytes,executor.GetOutput().Bytes()...)
					errorBytes = append(errorBytes,executor.GetError().Bytes()...)
				}

			}
		}
	}
	AfterScriptsAndExit:
	toolConfig , err = e.executeAfterScripts(toolConfig)
	toolConfig["exitCode"] = exitCode
	if toolErr != nil {
		toolConfig["status"] = false
	}
	if exitCode == 0 {
		toolConfig["status"] = true
	}
	if exitCode > 0 {
		toolConfig["status"] = false
	}
	delete(toolConfig,"self_dir")
	defer e.Log(fmt.Sprintf("Tool: %s has finished.",e.ToolInstance.Name))
	if e.ToolInstance.Shadow{
		return toolConfig,toolErr
	}
	//Create output file for the output of this tool
	toolOutputFile , err := e.CreateOutputFile("stdout","out")
	if err != nil {
		return toolConfig,err
	}
	err = helpers.WriteOrAppend(toolOutputFile,outputBytes,config.FILE_MODE_WRITABLE_PERM)
	if err != nil {
		return toolConfig,err
	}
	//Create err file for this tool
	toolErrFile , err := e.CreateOutputFile("stderr","err")
	if err != nil {
		return toolConfig,err
	}
	err = helpers.WriteOrAppend(toolErrFile,errorBytes,config.FILE_MODE_WRITABLE_PERM)
	if err != nil {
		return toolConfig,err
	}
	//Delete the temporary mapped self_dir key from the configuration
	return toolConfig,toolErr

}
// This function will add attachable directory parameters which are going to be attached as volumes
//to the running container
func (e *ToolExecutor) addAttachableVolume(parameter *models.Parameter){
	if e.dockerManager != nil {
		if volumePath , ok := parameter.Value.(string); ok {
			e.dockerManager.AddAttachableVolume(volumePath)
		}
	}
}
func (e *ToolExecutor) SetAttachableVolumes(volumes []models.Parameter) {
	e.AttachableVolumes = append(e.AttachableVolumes,volumes...)
}
func (e *ToolExecutor) Run(t *models.ToolInstance, workflowConfig models.FlowConfig) (models.FlowConfig,error) {

	e.ToolInstance = t
	err := e.init(workflowConfig)
	if err != nil {
		return nil,err
	}
	fmt.Println(fmt.Sprintf("Running (%s) Tool...",t.Name))
	if e.AttachableVolumes != nil {
		for _ , volume := range e.AttachableVolumes {
			e.addAttachableVolume(&volume)
		}
	}
	return e.execute()
}

