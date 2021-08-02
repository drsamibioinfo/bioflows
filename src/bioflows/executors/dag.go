package executors

import (
	"errors"
	"fmt"
	"github.com/aidarkhanov/nanoid"
	config2 "github.com/bioflows/src/bioflows/config"
	"github.com/bioflows/src/bioflows/expr"
	"github.com/bioflows/src/bioflows/logs"
	"github.com/bioflows/src/bioflows/managers"
	"github.com/bioflows/src/bioflows/models"
	"github.com/bioflows/src/bioflows/models/pipelines"
	"github.com/bioflows/src/bioflows/resolver"
	"github.com/bioflows/src/bioflows/scripts"
	"github.com/goombaio/dag"
	"github.com/mbndr/logo"
	"os"
	"sort"
	"strings"
	"sync"
)

type DagExecutor struct {
	contextManager *managers.ContextManager
	planManager *managers.ExecutionPlanManager
	transformations []TransformCall
	parentPipeline *pipelines.BioPipeline
	logger *logo.Logger
	containerConfig *models.ContainerConfig
	scheduler *DagScheduler
	exprManager *expr.ExprManager
	rankedList [][]*dag.Vertex
	basePath string
	instanceId string
	finalStatus bool
	explain bool
	// this bucket represents all errors that might have been encountered during the execution of the current DagExecutor
	errors []error
}

func (p *DagExecutor) GetFinalStatus() bool {
	return p.finalStatus
}
func (p *DagExecutor) SetExplain(explain bool) {
	p.explain = explain
}
func (p *DagExecutor) init() error {
	p.basePath = strings.Join([]string{config2.BIOFLOWS_NAME,config2.BIOFLOWS_PIPELINES},"/")
	p.finalStatus = true
	p.errors = make([]error,1)
	instanceId := nanoid.New()

	p.instanceId = instanceId
	return nil
}
func (p *DagExecutor) GetInstanceId() string {
	return p.instanceId
}
func (p *DagExecutor) GetPipelineKey() string {
	return strings.Join([]string{p.basePath,p.GetInstanceId(),p.parentPipeline.ID},"/")
}
func (p *DagExecutor) SetBasePath(basePath string) {
	p.basePath = basePath
}
func (p *DagExecutor) SetContainerConfig(containerConfig *models.ContainerConfig) {
	p.containerConfig = containerConfig
}
func (p *DagExecutor) SetContext(c *managers.ContextManager) {
	p.contextManager = c
}

func (p *DagExecutor) copyParentParamsInto(step *pipelines.BioPipeline) {
	if len(p.parentPipeline.Inputs) > 0 {
		if step.Inputs == nil || len(step.Inputs) == 0{
			step.Inputs = make([]models.Parameter,len(p.parentPipeline.Inputs))
			copy(step.Inputs,p.parentPipeline.Inputs)
		}else{
			step.Inputs = append(step.Inputs,p.parentPipeline.Inputs...)
		}
	}
}
func (p *DagExecutor) GetContext() *managers.ContextManager {
	return p.contextManager
}
//This function returns the final result of the current pipeline
func (p *DagExecutor) GetPipelineOutput(pipelineId *string) models.FlowConfig {
	tempConfig := models.FlowConfig{}
	pipelineKey := p.GetPipelineKey()
	if pipelineId != nil {
		pipelineKey = *pipelineId
	}else{
		pipelineKey = resolver.ResolvePipelineKey(pipelineKey)
	}
	pipelineConfig , err := p.GetContext().GetStateManager().GetPipelineState(pipelineKey)
	if err != nil {
		p.Log(fmt.Sprintf("Unable to fetch Pipeline Configuration for %s",pipelineKey))
		return tempConfig
	}
	tempConfig.Fill(pipelineConfig)
	return tempConfig
}


func (p DagExecutor) SetPipelineGeneralConfig(b *pipelines.BioPipeline,originalConfig *models.FlowConfig) {
	// Read the pipeline general configuration section
	if b.Config != nil && len(b.Config) > 0 {
		internalConfig := make(map[string]interface{})
		for _ , param := range b.Config {
			internalConfig[param.Name] = param.Value
		}
		(*originalConfig)[config2.BIOFLOWS_INTERNAL_CONFIG] = internalConfig
	}
	//Attach the general container configuration if exists.
	if b.ContainerConfig != nil {
		p.containerConfig = b.ContainerConfig
	}
}
func (p *DagExecutor) Clean() bool {
	return p.contextManager.GetStateManager().RemoveConfigByID(config2.BIOFLOWS_NAME)
}

func (p *DagExecutor) CheckStatus(pipelineId string , step pipelines.BioPipeline) int {
	status := SHOULD_RUN
	toolKey := resolver.ResolveToolKey(step.ID,pipelineId)
	toolData, _ := p.contextManager.GetStateManager().GetStateByID(toolKey)
	// If toolData exists, this means the tool has already run before
	if toolData != nil {
		data := toolData.(map[string]interface{})
		if ok, found := data["status"]; found && ok.(bool){
			status = DONT_RUN
		}
	}
	//Check that all dependent steps have run successfully
	if len(step.Depends) > 0 {
		depends := strings.Split(step.Depends,",")
		result := true
		for _ , v := range depends {
			toolName := resolver.ResolveToolKey(v,pipelineId)
			data , _ := p.GetContext().GetStateManager().GetStateByID(toolName)
			if data != nil {
				toolConfig := data.(map[string]interface{})
				if statusVar , found := toolConfig["status"]; !found {
					status = SHOULD_QUEUE
				}else{
					result = result && (statusVar.(bool))
				}
			}else{
				status = SHOULD_QUEUE
			}

		}
		if !result{
			status = DONT_RUN
		}
	}
	return status
}

func (p *DagExecutor) Setup(config models.FlowConfig) error {
	err := p.init()
	if err != nil {
		return err
	}
	p.scheduler = &DagScheduler{}
	p.exprManager = &expr.ExprManager{}
	p.transformations = make([]TransformCall,0)
	p.contextManager = &managers.ContextManager{}
	p.planManager = &managers.ExecutionPlanManager{}
	err = p.contextManager.Setup(config)
	if err != nil {
		return err
	}
	p.planManager.SetContextManager(p.contextManager)
	p.createLogFile(config)
	return p.planManager.Setup(config)
}
func (p *DagExecutor) createLogFile(config models.FlowConfig) error {
	workflowOutputFile := strings.Join([]string{
		fmt.Sprintf("%v",config[config2.WF_INSTANCE_OUTDIR]),
		"workflow.logs",
	},"/")
	p.logger = logs.NewLogger(config)
	p.logger.SetPrefix(fmt.Sprintf("%v: ",config2.BIOFLOWS_DISPLAY_NAME))
	file,  err := os.Create(workflowOutputFile)
	if err != nil {
		return err
	}
	rec := logo.NewReceiver(file,config2.BIOFLOWS_DISPLAY_NAME)
	if p.parentPipeline != nil {
		rec = logo.NewReceiver(file,p.parentPipeline.ID)
	}
	p.logger.Receivers = append(p.logger.Receivers,rec)
	return nil
}

func (p *DagExecutor) Log(logs ...interface{}) {
	//p.logger.Println(logs...)
	p.logger.Info(logs...)
}

func (p *DagExecutor) Run(b *pipelines.BioPipeline,config models.FlowConfig) error {
	p.SetPipelineGeneralConfig(b,&config)
	var finalError error
	defer func() error{
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				finalError = r.(error)
				p.Log(fmt.Sprintf("Error: %s. Aborting.....",finalError.Error()))
			case string:
				finalError = errors.New(r.(string))
			default:
				finalError = errors.New("There was an exception while running the current pipeline....")
			}

		}
		return finalError
	}()
	p.parentPipeline = b
	finalError = p.runLocal(b,config)
	p.Log(fmt.Sprintf("Workflow: (%s) has finished....",b.Name))
	p.addError(finalError)
	//Finally add all errors
	return p.GetAllErrors()
}
func (p *DagExecutor) GetAllErrors() error {
	var errString string = ""
	if len(p.errors) > 0 {
		for _ , err := range p.errors {
			if err == nil {
				continue
			}
			errString += err.Error() + "\n";
		}
	}
	if len(errString) > 0{
		return fmt.Errorf(errString)
	}
	return nil
}
func (p *DagExecutor) runLocal(b *pipelines.BioPipeline, config models.FlowConfig) error {
	graph , err := pipelines.CreateGraph(b)
	if err != nil {
		return err
	}
	p.rankedList , err = p.scheduler.Rank(b,graph)
	if err != nil {
		return err
	}
	if p.rankedList == nil {
		return errors.New("Failed to rank the current pipeline. Aborting....")
	}
	// evaluate current pipeline parameters
	p.evaluateParameters(b,config)
	// try to execute any before scripts
	err = p.executeBeforeScripts(b,config)
	if err != nil {
		p.Log(fmt.Sprintf("Executing Script (%s) Error : %s",b.Name,err.Error()))
		return err
	}
	defer func(){
		err = p.executeAfterScripts(b,config)
		if err != nil {
			p.Log(fmt.Sprintf("Executing Script (%s) Error : %s",b.Name,err.Error()))
		}
	}()
	for _ , sublist := range p.rankedList {
		wg := sync.WaitGroup{}
		for _ , node := range sublist {
			if node == nil {
				continue
			}
			wg.Add(1)
			go p.execute(config,node,&wg)
		}
		wg.Wait()
	}
	return nil
}
func (p *DagExecutor) prepareConfig(b *pipelines.BioPipeline,config models.FlowConfig) models.FlowConfig {
	tempConfig := models.FlowConfig{}
	for k , v := range config{
		tempConfig[k] = v
	}
	// Get Parent Pipeline Configuration from KV Store
	pipelineConfig , err := p.GetContext().GetStateManager().GetPipelineState(p.GetPipelineKey())
	if err != nil {
		p.Log(fmt.Sprintf("Unable to fetch Pipeline Configuration for %s",p.GetPipelineKey()))
		return tempConfig
	}
	tempConfig.Fill(pipelineConfig)
	// ***** End: Get Parent Pipeline Configuration from KV Store ********
	return tempConfig
}
func (p *DagExecutor) GetStepOutputDirectory(config models.FlowConfig , currentFlow *pipelines.BioPipeline) (string,error) {
	self_dir := strings.Join([]string{p.parentPipeline.ID,currentFlow.ID},"_")
	workflowOutputDir , ok := config[config2.WF_INSTANCE_OUTDIR]
	if !ok {
		err := fmt.Errorf("Output_dir configuration parameter is not set. Please set this variable and try again.")
		return "" , err
	}
	return strings.Join([]string{fmt.Sprintf("%v",workflowOutputDir),self_dir},string(os.PathSeparator)) , nil
}
func (p *DagExecutor) evaluateParameters(step *pipelines.BioPipeline,config models.FlowConfig) {
	//Evaluate current Step inputs
	selfDir , err := p.GetStepOutputDirectory(config,step)
	if err != nil {
		return
	}
	config["self_dir"] = selfDir
	config["location"] = selfDir
	if step.Inputs != nil && len(step.Inputs) > 0 {
		for _ , param := range step.Inputs {
			if param.Value == nil {
				if _ , ok := config[param.Name] ; !ok {
					config[param.Name] = ""
				}
				continue
			}
			config[param.Name] = p.exprManager.Render(param.GetParamValue(),config)
		}
	}
	// Adding the current step Config internal parameters
	if step.Config != nil && len(step.Config) > 0 {
		for _ , param := range step.Config{
			config[param.Name] = p.exprManager.Render(param.GetParamValue(),config)
		}
	}
	//Evaluate current Step outputs
	if step.Outputs != nil && len(step.Outputs) > 0 {
		for _ , param := range step.Outputs {
			if param.Value == nil {
				if _ , ok := config[param.Name]; !ok {
					config[param.Name] = ""
				}
				continue
			}
			config[param.Name] = p.exprManager.Render(param.GetParamValue(),config)
		}
	}
}
func (p *DagExecutor) executeBeforeScripts(step *pipelines.BioPipeline , config models.FlowConfig) error{

	beforeScripts := make([]models.Script,0)
	for idx , script := range step.Scripts {
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
		}
		err := scriptManager.RunScript(beforeScript,config)

		if err != nil {
			return err
		}
	}
	//finally
	return nil
}
func (p *DagExecutor) executeAfterScripts (step *pipelines.BioPipeline,config models.FlowConfig) error {
	afterScripts := make([]models.Script,0)
	for idx, script := range step.Scripts {
		if script.IsAfter() {
			if script.Order <= 0 {
				script.Order = idx + 1
			}
			afterScripts = append(afterScripts,script)
		}
	}
	sort.Slice(afterScripts,func(i,j int) bool {
		return afterScripts[i].Order < afterScripts[j].Order
	})
	for _ , afterScript := range afterScripts {
		var scriptManager scripts.ScriptManager
		switch strings.ToLower(afterScript.Type) {
		case "js":
			fallthrough
		default:
			scriptManager = &scripts.JSScriptManager{}
		}
		err := scriptManager.RunAfter(afterScript,config)
		if err != nil {
			return err
		}
	}
	//finally
	return nil
}
func (p *DagExecutor) reportFailure(toolKey string , flowConfig models.FlowConfig) error{
	flowConfig["status"] = false
	flowConfig["exitCode"] = 1
	err := p.contextManager.SaveState(toolKey,flowConfig.GetAsMap())
	if err != nil {
		p.Log(fmt.Sprintf("Received Error: %s",err.Error()))
		return err
	}
	return nil

}
func (p *DagExecutor) getAttachableVolumes(step *pipelines.BioPipeline) ([]models.Parameter,error) {
	attachable := make(map[string]models.Parameter)
	//access the inputs of the parent workflow
	totalParams := make([]models.Parameter,1)
	totalParams = append(totalParams,p.parentPipeline.Inputs...)
	totalParams = append(totalParams,p.parentPipeline.Config...)
	totalParams = append(totalParams,step.Inputs...)
	totalParams = append(totalParams,step.Config...)
	for _ , param := range totalParams {
		paramName := strings.ToLower(param.Type)
		if (paramName == "dir" || paramName == "directory") && param.IsAttachable() {
			attachable[param.Name] = param
		}
	}
	volumes := make([]models.Parameter,1)
	for _,v := range attachable {
		volumes = append(volumes,v)
	}
	return volumes , nil
}
func (p *DagExecutor) runInloopScripts(InLineScripts []models.Script,config models.FlowConfig) {

	sort.Slice(InLineScripts, func(i, j int) bool {
		return InLineScripts[i].Order < InLineScripts[j].Order
	})
	var scriptManager *scripts.JSScriptManager = new(scripts.JSScriptManager)
	for _ , script := range InLineScripts {
		_ = scriptManager.RunScript(script,config)
	}
}
func (p *DagExecutor) execute(config models.FlowConfig,vertex *dag.Vertex,wg *sync.WaitGroup) {
	defer wg.Done()
	currentFlow := vertex.Value.(pipelines.BioPipeline)
	PreprocessPipeline(&currentFlow,config,p.transformations...)
	toolKey := resolver.ResolveToolKey(currentFlow.ID,p.GetPipelineKey())
	//pipelineKey := resolver.ResolvePipelineKey(p.parentPipeline.ID)
	status := p.CheckStatus(p.GetPipelineKey(),currentFlow)
	switch status {
	case SHOULD_RUN:
		p.copyParentParamsInto(&currentFlow)
		// Step 1: Evaluate input parameters and output parameters for the current step before executing it
		p.evaluateParameters(&currentFlow,config)
		//Step 2: Try to execute before scripts first
		err := p.executeBeforeScripts(&currentFlow,config)
		if err != nil {
			p.Log(fmt.Sprintf("Executing Scripts (%s) Error : %s",currentFlow.Name,err.Error()))
			return
		}
		defer func(){
			err := p.executeAfterScripts(&currentFlow,config)
			if err != nil {
				p.Log(fmt.Sprintf("Executing Scripts (%s) Error : %s",currentFlow.Name,err.Error()))
			}
		}()
		if currentFlow.IsTool() {
			// It is a tool
			if currentFlow.IsLoop() {
				// Get the loop variable
				if len(currentFlow.LoopVar) == 0 {
					p.Log(fmt.Sprintf("Tool is loop but no loop variable has been defined.. aborting..."))
					p.reportFailure(toolKey,config)
					return
				}
				// Get Loop Variable name
				if loop_elements , ok := config[currentFlow.LoopVar] ; ok {
					if elements , islist := loop_elements.([]interface{}); islist {
						stepTruth := true
						for idx , el := range elements {
							executor := ToolExecutor{}
							executor.SetBasePath(toolKey)
							executor.SetPipelineName(p.parentPipeline.ID)
							executor.SetContainerConfiguration(p.containerConfig)
							toolInstance := &models.ToolInstance{
								WorkflowID: p.parentPipeline.ID,
								WorkflowName: p.parentPipeline.Name,
								Tool:currentFlow.ToTool(),
							}
							toolInstance.Prepare()
							generalConfig := p.prepareConfig(p.parentPipeline,config)
							generalConfig[fmt.Sprintf("%s_item",currentFlow.LoopVar)] = el
							generalConfig[fmt.Sprintf("loop_index")] = idx
							// RunScript the given tool
							volumes , err := p.getAttachableVolumes(&currentFlow)
							if err != nil {
								executor.Log(fmt.Sprintf("Received Error : %s",err.Error()))
								return
							}
							executor.SetAttachableVolumes(volumes)
							// Run InLoop Scripts first
							inlineScripts := currentFlow.GetInLoopScripts()
							if len(inlineScripts) > 0{
								p.runInloopScripts(inlineScripts,generalConfig)
							}
							executor.SetExplain(p.explain)
							toolInstanceFlowConfig , err := executor.Run(toolInstance,generalConfig)
							if err != nil {

								executor.Log(fmt.Sprintf("Received Error : %s",err.Error()))
							}
							if toolInstanceFlowConfig != nil {
								stepTruth = stepTruth && toolInstanceFlowConfig["status"].(bool)

								if idx < len(elements) - 1{
									toolInstanceFlowConfig["status"] = false
								}else{
									toolInstanceFlowConfig["status"] = stepTruth
								}
								toolKeyInAloop := executor.GetToolKey()
								err = p.contextManager.SaveState(toolKeyInAloop,toolInstanceFlowConfig.GetAsMap())
								if err != nil {
									p.Log(fmt.Sprintf("Received Error: %s",err.Error()))
									return
								}
							}
						}
						config["status"] = stepTruth
						config["exitCode"] = 0
						p.finalStatus = p.finalStatus && stepTruth
						err = p.contextManager.SaveState(toolKey,config.GetAsMap())
						if err != nil {
							p.Log(fmt.Sprintf("Received Error: %s",err.Error()))
							return
						}

					}else{
						// The Loop variable contains non-array type data , i.e. it is not an array
						p.reportFailure(toolKey,config)
						p.finalStatus = p.finalStatus && false
						p.Log(fmt.Sprintf("Failing Tool : %s, The tool has no associated data in the loop variable.",
							currentFlow.Name))
						return
					}
				}else{
					p.addError(fmt.Errorf("Loop Variable is not defined: %v",currentFlow.LoopVar))
				}
			}else {
				// The current tool is not loop
				executor := ToolExecutor{}
				executor.SetBasePath(p.GetPipelineKey())
				executor.SetPipelineName(p.parentPipeline.ID)
				executor.SetContainerConfiguration(p.containerConfig)
				toolInstance := &models.ToolInstance{
					WorkflowID: p.parentPipeline.ID,
					WorkflowName: p.parentPipeline.Name,
					Tool:currentFlow.ToTool(),
				}
				toolInstance.Prepare()

				generalConfig := p.prepareConfig(p.parentPipeline,config)
				// RunScript the given tool
				volumes , err := p.getAttachableVolumes(&currentFlow)
				if err != nil {
					executor.Log(fmt.Sprintf("Received Error : %s",err.Error()))
					return
				}
				executor.SetAttachableVolumes(volumes)
				executor.SetExplain(p.explain)
				toolInstanceFlowConfig , err := executor.Run(toolInstance,generalConfig)
				if err != nil {

					executor.Log(fmt.Sprintf("Received Error : %s",err.Error()))
				}
				if toolInstanceFlowConfig != nil {

					err = p.contextManager.SaveState(toolKey,toolInstanceFlowConfig.GetAsMap())
					if err != nil {
						p.Log(fmt.Sprintf("Received Error: %s",err.Error()))
						return
					}
				}
				if status, ok := toolInstanceFlowConfig["status"]; ok {
					p.finalStatus = p.finalStatus && status.(bool)
				}
			}

		}else{
			//Step 3: Try to run the current nested pipeline
			//it is a nested pipeline
			if !currentFlow.IsLoop() {
				// It is a nested pipeline but not a loop
				nestedPipelineExecutor := DagExecutor{}
				nestedPipelineExecutor.SetContainerConfig(p.containerConfig)
				nestedPipelineConfig := models.FlowConfig{}
				pipelineConfig := p.prepareConfig(&currentFlow,config)
				nestedPipelineConfig.Fill(config)
				nestedPipelineConfig.Fill(pipelineConfig)
				nestedPipelineExecutor.Setup(nestedPipelineConfig)
				nestedPipelineExecutor.SetBasePath(toolKey)
				err := nestedPipelineExecutor.Run(&currentFlow,nestedPipelineConfig)
				if err != nil {

					nestedPipelineExecutor.Log(err.Error())
				}
				pipeConfig := nestedPipelineExecutor.GetPipelineOutput(&toolKey)
				pipeConfig["status"] = nestedPipelineExecutor.GetFinalStatus()
				if nestedPipelineExecutor.GetFinalStatus() {
					pipeConfig["exitCode"] = 0
				}else{
					pipeConfig["exitCode"] = 1
				}
				err = p.contextManager.SaveState(toolKey,pipeConfig.GetAsMap())
			}else{
				// It is a nested pipeline and a loop
				if len(currentFlow.LoopVar) == 0 {
					p.Log(fmt.Sprintf("%s is defined as loop but no loop variable has been defined.",
					currentFlow.Name))
					p.reportFailure(toolKey,config)
					return
				}
				if loop_elements , ok := config[currentFlow.LoopVar]; ok {
					if elements, islist := loop_elements.([]interface{}); islist{
						for idx , el := range elements{
							nestedPipelineExecutor := DagExecutor{}
							nestedPipelineExecutor.SetContainerConfig(p.containerConfig)
							nestedPipelineConfig := models.FlowConfig{}
							pipelineConfig := p.prepareConfig(&currentFlow,config)
							nestedPipelineConfig.Fill(config)
							nestedPipelineConfig.Fill(pipelineConfig)
							nestedPipelineExecutor.Setup(nestedPipelineConfig)
							nestedPipelineExecutor.SetBasePath(toolKey)
							nestedPipelineConfig[fmt.Sprintf("%s_item",currentFlow.LoopVar)] = el
							nestedPipelineConfig[fmt.Sprintf("loop_index")] = idx
							inlineScripts := currentFlow.GetInLoopScripts()
							if len(inlineScripts) > 0{
								p.runInloopScripts(inlineScripts,nestedPipelineConfig)
							}
							err := nestedPipelineExecutor.Run(&currentFlow,nestedPipelineConfig)
							if err != nil {
								nestedPipelineExecutor.Log(err.Error())
							}
							pipeConfig := nestedPipelineExecutor.GetPipelineOutput(nil)
							pipelineKeyInAloop := nestedPipelineExecutor.GetPipelineKey()
							err = p.contextManager.SaveState(pipelineKeyInAloop,pipeConfig.GetAsMap())
						}
						config["status"] = true
						config["exitCode"] = 0
						p.finalStatus = p.finalStatus && true
						err = p.contextManager.SaveState(toolKey,config.GetAsMap())
						if err != nil {
							p.Log(fmt.Sprintf("Received Error: %s",err.Error()))
							return
						}
					}
				}

			}

		}
	case SHOULD_QUEUE:
		fallthrough
	case DONT_RUN:
		fallthrough
	default:
		p.finalStatus = false
		p.Log(fmt.Sprintf("Flow: %s has already run before, deferring....",currentFlow.Name))
		return
	}
}

func (p *DagExecutor) addError(err error) {
	p.errors = append(p.errors,err)
}




