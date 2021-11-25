package virtualization

import (
	"github.com/bioflows/src/bioflows/config"
	"github.com/bioflows/src/bioflows/models"
	"bytes"
	"strings"
)

const (
	VIRTUALIZATION_SECTION_NAME="virtualization"
	VIRTUALIZATION_KEY_NAME="manager_name"

	DOCKER_VIRTUALIZATION = "docker"
	SINGULARITY_VIRTUALIZATION="singularity"
)

type BioflowContainer struct {
	Names []string
	ID string
	Image string
	ImageID string
}

type BioFlowImage struct {
	ID string `json:"id"`
	Name string `json:"name,omitempty"`
	ParentID string `json:"parentId,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
	Size int64 `json:"size,omitempty"`
}

type VirtualizationManager interface {

	//Returns list of container names that are running on the local node
	ListContainers() []BioflowContainer
	StartContainer(containerName string ,imageName string,commands []string) (*bytes.Buffer , *bytes.Buffer , error)
	StopContainer (containerID string) error
	PullImage(imageURL string )  (*bytes.Buffer , error)
	RunToolInstance(instance models.ToolInstance,imageName string) (*bytes.Buffer,*bytes.Buffer,error)
	ListImages() []BioFlowImage
	PrepareImage(imageURL string , workflow models.BioWorkflow) error

}

var activeManager VirtualizationManager

func init(){

	result, err := config.HasKey(VIRTUALIZATION_SECTION_NAME,VIRTUALIZATION_KEY_NAME)
	if !result && err != nil {

		//logs.WriteLog(err.Error())
		return
	}

	val , err := config.GetKeyAsString(VIRTUALIZATION_SECTION_NAME,VIRTUALIZATION_KEY_NAME)
	if err != nil {
		//logs.WriteLog(err.Error())
		return
	}

	switch(strings.ToLower(val)){

	case SINGULARITY_VIRTUALIZATION:
		activeManager = &SingularityVirtualizationManager{}
		break;

	case DOCKER_VIRTUALIZATION:
		fallthrough
	default:
		activeManager = &DockerVirtualizationManager{}
		break
	}

}

func GetDefaultVirtualizationManager() VirtualizationManager{
	return activeManager
}

func NewVirtualManager() VirtualizationManager{


	var newManager VirtualizationManager

	result, err := config.HasKey(VIRTUALIZATION_SECTION_NAME,VIRTUALIZATION_KEY_NAME)
	if !result && err != nil {

		//logs.WriteLog(err.Error())
		return nil
	}

	val , err := config.GetKeyAsString(VIRTUALIZATION_SECTION_NAME,VIRTUALIZATION_KEY_NAME)
	if err != nil {
		//logs.WriteLog(err.Error())
		return nil
	}

	switch(strings.ToLower(val)){

	case SINGULARITY_VIRTUALIZATION:
		newManager = &SingularityVirtualizationManager{}
		break;

	case DOCKER_VIRTUALIZATION:
		fallthrough
	default:
		newManager = &DockerVirtualizationManager{}
		break
	}

	return newManager

}
