package executors

import (
	config2 "bioflows/config"
	"bioflows/managers"
	"bioflows/models"
	"bioflows/models/pipelines"
	"bioflows/resolver"
	"errors"
	"fmt"
	"github.com/goombaio/dag"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SHOULD_RUN = iota
	DONT_RUN
	SHOULD_QUEUE

)

type PipelineExecutor struct {
	contextManager *managers.ContextManager
	planManager *managers.ExecutionPlanManager
	transformations []TransformCall
	waitGroup sync.WaitGroup
	mutex *sync.Mutex
	memory *ExecutorMemory
	stopChan chan interface{}
	parentPipeline *pipelines.BioPipeline
	ticker *time.Ticker
	logger *log.Logger
	containerConfig *models.ContainerConfig
}

func (p *PipelineExecutor) SetContainerConfig(containerConfig *models.ContainerConfig) {
	p.containerConfig = containerConfig
}

//This function returns the final result of the current pipeline
func (p *PipelineExecutor) GetPipelineOutput() models.FlowConfig {
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

func (p *PipelineExecutor) IsRemote() bool {
	return p.contextManager.IsRemote()
}

func (p *PipelineExecutor) handleWaitQueue(config models.FlowConfig) {

	go func(){

		for{
			select{
			case <- p.ticker.C:
				if task , ok := p.memory.PopFromMemory(); ok {
					p.executeSingleVertex(p.parentPipeline,config,task)

				}
			}
		}
	}()

}

func (p *PipelineExecutor) SetContext(c *managers.ContextManager) {
	p.contextManager = c
}

func (p *PipelineExecutor) GetContext() *managers.ContextManager {
	return p.contextManager
}
func (p PipelineExecutor) SetPipelineGeneralConfig(b *pipelines.BioPipeline,originalConfig *models.FlowConfig) {
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
func (p *PipelineExecutor) Clean() bool {
	return p.contextManager.GetStateManager().RemoveConfigByID(config2.BIOFLOWS_NAME)
}
func (p *PipelineExecutor) Run(b *pipelines.BioPipeline,config models.FlowConfig) error {
	//Set default pipeline general configuration if exists..
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
	//Start handling wait queue
	p.handleWaitQueue(config)
	// Start processing the current pipeline
	//PreprocessPipeline(b,config,p.transformations...)
	if p.IsRemote(){
		finalError = p.runOnCluster(b,config)
	}else{
		finalError = p.runLocally(b,config)
	}
	p.Log(fmt.Sprintf("Workflow: (%s) has finished....",b.Name))
	return finalError

}
func (p *PipelineExecutor) canRun(pipelineId string , step pipelines.BioPipeline) bool {
	result := true
	if len(step.Depends) <= 0 {
		return true
	}
	depends := strings.Split(step.Depends,",")

	for _ , v := range depends {
		toolName := resolver.ResolveToolKey(v,pipelineId)
		data , err := p.GetContext().GetStateManager().GetStateByID(toolName)
		if err != nil {
			result = false
			return result
		}
		toolConfig := data.(map[string]interface{})
		if status , ok := toolConfig["status"]; !ok {
			result = false
		}else{
			result = result && !(status.(bool))
		}
	}
	return result
}
func (p *PipelineExecutor) isAlreadyRun(toolKey string) bool{
	result := false
	section , err := p.contextManager.GetStateManager().GetStateByID(toolKey)
	if err != nil {
		result = false
		return result
	}
	if section == nil {
		return false
	}
	data := section.(map[string]interface{})
	if _ , ok := data["status"] ; ok {
		result = true
	}
	return result

}

func (p *PipelineExecutor) CheckStatus(pipelineId string , step pipelines.BioPipeline) int {
	status := SHOULD_RUN
	toolKey := resolver.ResolveToolKey(step.ID,pipelineId)
	section , _ := p.contextManager.GetStateManager().GetStateByID(toolKey)
	if section != nil {
		data := section.(map[string]interface{})
		if _, ok := data["status"]; !ok {
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
				if status , ok := toolConfig["status"]; !ok {
					status = SHOULD_QUEUE
				}else{
					result = result && (status.(bool))
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
func (p *PipelineExecutor) executeSingleVertex(b *pipelines.BioPipeline , config models.FlowConfig,vertex *dag.Vertex) {
	defer p.waitGroup.Done()
	var execStatus bool = true
	currentFlow := vertex.Value.(pipelines.BioPipeline)
	PreprocessPipeline(&currentFlow,config,p.transformations...)
	toolKey := resolver.ResolveToolKey(currentFlow.ID,b.ID)
	//pipelineKey := resolver.ResolvePipelineKey(p.parentPipeline.ID)
	status := p.CheckStatus(b.ID,currentFlow)
	switch status {
	case SHOULD_RUN:
		if currentFlow.IsTool() {
			executor := ToolExecutor{}
			executor.SetPipelineName(p.parentPipeline.Name)
			executor.SetContainerConfiguration(p.containerConfig)
			toolInstance := &models.ToolInstance{
				WorkflowID: b.ID,
				WorkflowName: b.Name,
				Tool:currentFlow.ToTool(),
			}
			toolInstance.Prepare()
			generalConfig := p.prepareConfig(p.parentPipeline,config)
			toolInstanceFlowConfig , err := executor.Run(toolInstance,generalConfig)
			if err != nil {
				execStatus = false
				executor.Log(fmt.Sprintf("Received Error : %s",err.Error()))
			}
			if toolInstanceFlowConfig != nil {
				exitCode := fmt.Sprintf("%v",toolInstanceFlowConfig["exitCode"])
				code , err := strconv.Atoi(exitCode)
				if err != nil {
					execStatus = false
				}else{
					if code > 0 {
						execStatus = false
					}
				}
				err = p.contextManager.SaveState(toolKey,toolInstanceFlowConfig.GetAsMap())
				if err != nil {
					fmt.Println(fmt.Sprintf("Received Error: %s",err.Error()))
					return
				}
			}

		}else{
			// It is a nested pipeline step.
			//it is a nested pipeline
			nestedPipelineExecutor := PipelineExecutor{}
			nestedPipelineExecutor.SetContainerConfig(p.containerConfig)
			nestedPipelineConfig := models.FlowConfig{}
			pipelineConfig := p.prepareConfig(&currentFlow,config)
			nestedPipelineConfig.Fill(config)
			nestedPipelineConfig.Fill(pipelineConfig)
			nestedPipelineExecutor.Setup(nestedPipelineConfig)
			err := nestedPipelineExecutor.Run(&currentFlow,nestedPipelineConfig)
			if err != nil {
				execStatus = false
				nestedPipelineExecutor.Log(err.Error())
			}
			pipeConfig := nestedPipelineExecutor.GetPipelineOutput()
			err = p.contextManager.SaveState(toolKey,pipeConfig)
		}
		if execStatus {
			if vertex.Children.Size() > 0 {
				// RunScript those children
				for _ , child := range vertex.Children.Values() {
					childFlow := child.(*dag.Vertex)
					p.waitGroup.Add(1)
					p.executeSingleVertex(b,config,childFlow)
				}

			}
		}
	case SHOULD_QUEUE:
		if p.memory.AddToMemory(vertex){
			//Spawn the current step until all other dependencies are run successfully
			fmt.Println(fmt.Sprintf("Spawning Tool (%s) until dependencies finish execution....",currentFlow.Name))
			p.waitGroup.Add(1)
		}

	case DONT_RUN:
		fallthrough
	default:
		return
	}
}
func (p *PipelineExecutor) runLocally(b *pipelines.BioPipeline, config models.FlowConfig) error {
	fmt.Println(fmt.Sprintf("Running Pipeline (%s) Locally....",b.Name))
	//Create a Directed Acyclic Graph of the current pipeline
	graph , err := pipelines.CreateGraph(b)
	if err != nil {
		fmt.Println(fmt.Sprintf("Received Error : %s",err.Error()))
		return nil
	}
	parents := graph.SourceVertices()
	for _ , parent := range parents{
		//RunScript each parent individually.
		p.waitGroup.Add(1)
		go p.executeSingleVertex(b,config,parent)
	}
	p.waitGroup.Wait()
	p.stopChan <- nil
	return nil
}
func (p *PipelineExecutor) prepareConfig(b *pipelines.BioPipeline,config models.FlowConfig) models.FlowConfig {
	tempConfig := models.FlowConfig{}
	for k , v := range config{
		tempConfig[k] = v
	}
	pipelineKey := resolver.ResolvePipelineKey(b.ID)
	pipelineConfig , err := p.GetContext().GetStateManager().GetPipelineState(pipelineKey)
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to fetch Pipeline Configuration for %s",pipelineKey))
		return tempConfig
	}
	tempConfig.Fill(pipelineConfig)
	return tempConfig
}

func (p *PipelineExecutor) runOnCluster(b *pipelines.BioPipeline, config models.FlowConfig) error {
	return p.runLocally(b,config)
}

func (p *PipelineExecutor) AddTransform(transformCall TransformCall) {
	p.transformations = append(p.transformations,transformCall)
}
func (p *PipelineExecutor) ClearTransformations() bool {
	// Clear all Transformations
	p.transformations = p.transformations[:0]
	return true
}

func (p *PipelineExecutor) Setup(config models.FlowConfig) error {
	p.waitGroup = sync.WaitGroup{}
	p.memory = &ExecutorMemory{}
	p.memory.SetUp()
	p.mutex = &sync.Mutex{}
	p.ticker = time.NewTicker(5 * time.Second)
	p.stopChan = make(chan interface{},1)
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
func (p *PipelineExecutor) createLogFile(config models.FlowConfig) error {
	workflowOutputFile := strings.Join([]string{
		fmt.Sprintf("%v",config[config2.WF_INSTANCE_OUTDIR]),
		"workflow.logs",
	},"/")
	p.logger = &log.Logger{}
	p.logger.SetPrefix(fmt.Sprintf("%v: ",config2.BIOFLOWS_DISPLAY_NAME))
	file,  err := os.Create(workflowOutputFile)
	if err != nil {
		return err
	}
	p.logger.SetOutput(file)
	return nil
}

func (p *PipelineExecutor) Log(logs ...interface{}) {
	p.logger.Println(logs...)
	fmt.Println(logs...)
}


