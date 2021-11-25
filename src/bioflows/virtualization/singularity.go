package virtualization

import (
	"github.com/bioflows/src/bioflows/models"
	"bytes"
)

type SingularityVirtualizationManager struct {

}

func (s *SingularityVirtualizationManager) RunToolInstance(instance models.ToolInstance,imageName string) (*bytes.Buffer,*bytes.Buffer,error){

	return nil , nil , nil
}

func (s *SingularityVirtualizationManager) ListImages() []BioFlowImage{
	return nil
}

func (s *SingularityVirtualizationManager) PrepareImage(imageURL string , workflow models.BioWorkflow) error{
	return nil
}

func (s *SingularityVirtualizationManager) ListContainers() []BioflowContainer{
	return nil
}

func (s *SingularityVirtualizationManager)  StartContainer(containerName string , imageName string,commands []string) (*bytes.Buffer,*bytes.Buffer,error) {
	return nil,nil,nil
}
func (s *SingularityVirtualizationManager) StopContainer (containerID string) error{

	return nil
}

func (s *SingularityVirtualizationManager) PullImage(imageURL string) (*bytes.Buffer , error){
	return nil , nil
}
