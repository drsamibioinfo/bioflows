package virtualization

import (
	"github.com/bioflows/src/bioflows/models"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"

)

type DockerVirtualizationManager struct {
	client *client.Client
	version string


}

func (d *DockerVirtualizationManager) init(){
	cli , err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic("Unable to start Bioflows Docker Virtualization Manager.")
	}
	d.client = cli
}

func (d *DockerVirtualizationManager) ListImages() []BioFlowImage {
	d.init()
	ctx := context.Background()

	images, err := d.client.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}
	bioImages := make([]BioFlowImage,0)
	for _, image := range images {
		bioImages = append(bioImages,BioFlowImage{
			ID : image.ID,
			Labels:image.Labels,
			ParentID:image.ParentID,
			Size:image.Size,
		})
	}
	return bioImages
}

func (d *DockerVirtualizationManager) PrepareImage(imageURL string , workflow models.BioWorkflow) error{

	images := d.ListImages()
	exists := false
	if images != nil && len(images) > 0 {
		for _ , bioimage := range images{
			if(bioimage.ID == workflow.GetIdentifier() || bioimage.Name == workflow.GetIdentifier()) {
				exists = true
				break
			}
		}
	}
	if exists{
		return fmt.Errorf("Image already exists")

	}
	_ , err := d.PullImage(imageURL)
	if err != nil {
		panic(err)

	}

	ctx := context.Background()
	createResp, err := d.client.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"bioflowinstall", workflow.ID},
	}, nil, nil, nil,"")
	if err != nil {
		panic(err)
	}

	if err := d.client.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := d.client.ContainerWait(ctx, createResp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	_, err = d.client.ContainerCommit(ctx, createResp.ID, types.ContainerCommitOptions{Reference: workflow.GetIdentifier()})
	if err != nil {
		panic(err)
	}

	return err
}

func (d *DockerVirtualizationManager) RunToolInstance(instance models.ToolInstance,imageName string) (*bytes.Buffer,*bytes.Buffer,error){
	return d.StartContainer(instance.GetContainerName(),imageName,instance.PrepareCommand())
}

func (d *DockerVirtualizationManager) ListContainers() []BioflowContainer{
	d.init()
	containers, err := d.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	bioflows_containers := make([]BioflowContainer,0)

	for _ , container := range containers{

		if container.ID == "" || container.Image == "" || len(container.ID) <= 0 || len(container.Image) <= 0{
			continue
		}
		bioflows_containers = append(bioflows_containers,BioflowContainer{
			Names:container.Names,
			ID : container.ID,
			Image : container.Image,
			ImageID:container.ImageID,
		})
	}

	return bioflows_containers
}

func (d *DockerVirtualizationManager) PullImage(imageURL string) (*bytes.Buffer,error){

	d.init()

	var buffer *bytes.Buffer = &bytes.Buffer{}

	reader, err := d.client.ImagePull(context.Background(), imageURL, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(buffer, reader)
	return buffer , err
}

func (d *DockerVirtualizationManager) StartContainer(containerName string , imageName string,commands []string) (*bytes.Buffer,*bytes.Buffer,error) {

	d.init()

	var StdOut *bytes.Buffer = &bytes.Buffer{}
	var StdErr *bytes.Buffer = &bytes.Buffer{}
	resp, err := d.client.ContainerCreate(context.Background(), &container.Config{
		Image: imageName,
		Cmd:   commands,
		Tty:   true,
	}, nil, nil,nil, containerName)
	if err != nil {
		panic(err)
	}

	if err := d.client.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := d.client.ContainerWait(context.Background(), resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := d.client.ContainerLogs(context.Background(), resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}
	io.Copy(StdOut,out)
	io.Copy(StdErr,out)

	return StdOut, StdErr , err

}
func (d *DockerVirtualizationManager) StopContainer (containerID string) error{

	d.init()
	if err := d.client.ContainerStop(context.Background(), containerID, nil); err != nil {
		return err
	}
	return nil
}
