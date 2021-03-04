package executors

import (
	config2 "bioflows/config"
	"bioflows/expr"
	"bioflows/managers"
	"bioflows/models"
	"bioflows/models/pipelines"
	"bioflows/resolver"
	"bioflows/scripts"
	"errors"
	"fmt"
	"github.com/goombaio/dag"
	"log"
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
	logger *log.Logger
	containerConfig *models.ContainerConfig
	scheduler *DagScheduler
	exprManager *expr.ExprManager
	rankedList [][]*dag.Vertex
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
func (p *DagExecutor) GetPipelineOutput() models.FlowConfig {
	tempConfig := models.FlowConfig{}
	pipelineKey := resolver.ResolvePipelineKey(p.parentPipeline.ID)
	pipelineConfig , err := p.GetContext().GetStateManager().GetPipelineState(pipelineKey)
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to fetch Pipeline Configuration for %s",pipelineKey))
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
	return p.contextManager.GetStateManager().RemoveConfigByID(resolver.BIOFLOWS_NAME)
}

func (p *DagExecutor) CheckStatus(pipelineId string , step pipelines.BioPipeline) int {
	status := SHOULD_RUN
	toolKey := resolver.ResolveToolKey(step.ID,pipelineId)
	section , _ := p.contextManager.GetStateManager().GetStateByID(toolKey)
	if section != nil {
		data := section.(map[string]interface{})
		if _, found := data["status"]; found {
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

	p.scheduler = &DagScheduler{}
	p.exprManager = &expr.ExprManager{}
	p.transformations = make([]TransformCall,0)
	p.contextManager = &managers.ContextManager{}
	p.planManager = &managers.ExecutionPlanManager{}
	err := p.contextManager.Setup(config)
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
	p.logger = &log.Logger{}
	p.logger.SetPrefix(fmt.Sprintf("%v: ",config2.BIOFLOWS_NAME))
	file,  err := os.Create(workflowOutputFile)
	if err != nil {
		return err
	}
	p.logger.SetOutput(file)
	return nil
}

func (p *DagExecutor) Log(logs ...interface{}) {
	p.logger.Println(logs...)
	fmt.Println(logs...)
}

func (p *DagExecutor) Run(b *pipelines.BioPipeline,config models.FlowConfig) error {
	p.SetPipelineGeneralConfig(b,&config)
	var finalError error
	defer func() error{
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				finalError = r.(error)
				fmt.Println(fmt.Sprintf("Error: %s. Aborting.....",finalError.Error()))
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
	return finalError
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
	pipelineKey := resolver.ResolvePipelineKey(b.ID)
	pipelineConfig , err := p.GetContext().GetStateManager().GetPipelineState(pipelineKey)
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to fetch Pipeline Configuration for %s",pipelineKey))
		return tempConfig
	}
	tempConfig.Fill(pipelineConfig)
	// ***** End: Get Parent Pipeline Configuration from KV Store ********
	return tempConfig
}
func (p *DagExecutor) evaluateParameters(step *pipelines.BioPipeline,config models.FlowConfig) {
	//Evaluate current Step inputs
	if step.Inputs != nil && len(step.Inputs) > 0 {
		for _ , param := range step.Inputs {
			if param.Value == nil {
				continue
			}
			config[param.Name] = p.exprManager.Render(param.GetParamValue(),config)
		}
	}
	//Evaluate current Step outputs
	if step.Outputs != nil && len(step.Outputs) > 0 {
		for _ , param := range step.Outputs {
			if param.Value == nil {
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
		err := scriptManager.RunBefore(beforeScript,config)

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
		fmt.Println(fmt.Sprintf("Received Error: %s",err.Error()))
		return err
	}
	return nil

}
func (p *DagExecutor) execute(config models.FlowConfig,vertex *dag.Vertex,wg *sync.WaitGroup) {
	defer wg.Done()
	currentFlow := vertex.Value.(pipelines.BioPipeline)
	PreprocessPipeline(&currentFlow,config,p.transformations...)
	toolKey := resolver.ResolveToolKey(currentFlow.ID,p.parentPipeline.ID)
	//pipelineKey := resolver.ResolvePipelineKey(p.parentPipeline.ID)
	status := p.CheckStatus(p.parentPipeline.ID,currentFlow)
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

						for idx , el := range elements {
							executor := ToolExecutor{}
							executor.SetPipelineName(p.parentPipeline.Name)
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
							// Run the given tool
							toolInstanceFlowConfig , err := executor.Run(toolInstance,generalConfig)
							if err != nil {

								executor.Log(fmt.Sprintf("Received Error : %s",err.Error()))
							}
							if toolInstanceFlowConfig != nil {
								err = p.contextManager.SaveState(toolKey,toolInstanceFlowConfig.GetAsMap())
								if err != nil {
									fmt.Println(fmt.Sprintf("Received Error: %s",err.Error()))
									return
								}
							}
						}
					}else{
						// The Loop variable contains non-array type data , i.e. it is not an array
						p.reportFailure(toolKey,config)
						return
					}
				}
			}else {
				// The current tool is not loop
				executor := ToolExecutor{}
				executor.SetPipelineName(p.parentPipeline.Name)
				executor.SetContainerConfiguration(p.containerConfig)
				toolInstance := &models.ToolInstance{
					WorkflowID: p.parentPipeline.ID,
					WorkflowName: p.parentPipeline.Name,
					Tool:currentFlow.ToTool(),
				}
				toolInstance.Prepare()
				generalConfig := p.prepareConfig(p.parentPipeline,config)
				toolInstanceFlowConfig , err := executor.Run(toolInstance,generalConfig)
				if err != nil {

					executor.Log(fmt.Sprintf("Received Error : %s",err.Error()))
				}
				if toolInstanceFlowConfig != nil {

					err = p.contextManager.SaveState(toolKey,toolInstanceFlowConfig.GetAsMap())
					if err != nil {
						fmt.Println(fmt.Sprintf("Received Error: %s",err.Error()))
						return
					}
				}
			}

		}else{
			//Step 3: Try to run the current nested pipeline
			//it is a nested pipeline
			nestedPipelineExecutor := DagExecutor{}
			nestedPipelineExecutor.SetContainerConfig(p.containerConfig)
			nestedPipelineConfig := models.FlowConfig{}
			pipelineConfig := p.prepareConfig(&currentFlow,config)
			nestedPipelineConfig.Fill(config)
			nestedPipelineConfig.Fill(pipelineConfig)
			nestedPipelineExecutor.Setup(nestedPipelineConfig)
			err := nestedPipelineExecutor.Run(&currentFlow,nestedPipelineConfig)
			if err != nil {

				nestedPipelineExecutor.Log(err.Error())
			}
			pipeConfig := nestedPipelineExecutor.GetPipelineOutput()
			err = p.contextManager.SaveState(toolKey,pipeConfig)

		}
	case SHOULD_QUEUE:
		fallthrough
	case DONT_RUN:
		fallthrough
	default:
		p.Log(fmt.Sprintf("Flow: %s has already run before, deferring....",currentFlow.Name))
		return
	}
}




