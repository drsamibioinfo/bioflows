package container

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bioflows/src/bioflows/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/mbndr/logo"
	"io"
)

const (
	DOCKER_REPOSITORY = "docker.io"
)

type DockerManager struct {
	client *client.Client
	logger *logo.Logger
	DockerConfig *container.Config
	HostConfig *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

func (d *DockerManager) AddAttachableVolume(volumePath string) {
	if d.HostConfig == nil {
		d.HostConfig = &container.HostConfig{}
	}
	d.HostConfig.Binds = append(d.HostConfig.Binds,fmt.Sprintf("%s:%s",volumePath))
}

func (d *DockerManager) SetLogger(logger *logo.Logger) {
	d.logger = logger
}
func (d *DockerManager) Log(logs ...interface{}) {
	d.logger.Info(logs...)
}


func (d *DockerManager) init() error{
	if d.client != nil {
		return nil
	}
	cli , err := client.NewClientWithOpts(client.FromEnv,client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	d.client = cli
	return nil
}

func (d *DockerManager) PullImage(imageURL string,containerConfig *models.ContainerConfig) (string,error) {
	d.init()
	buffer := &bytes.Buffer{}
	options := types.ImagePullOptions{}
	if containerConfig != nil {
		options.RegistryAuth , _ = containerConfig.GetAuth()
	}
	reader,err := d.client.ImagePull(context.Background(),imageURL,options)
	if err != nil {
		return "" , err
	}
	io.Copy(buffer,reader)
	return buffer.String() , nil
}

func (d *DockerManager) RunContainer(containerName string , ImageId string, commands []string,keep bool) (Out *bytes.Buffer, Err *bytes.Buffer,err error) {
	d.init()
	Out = &bytes.Buffer{}
	Err = &bytes.Buffer{}
	err = nil
	//randomContainerName := fmt.Sprintf("%s%d",containerName,rand.Int())
	resp , err := d.client.ContainerCreate(context.Background(),&container.Config{
		Image: ImageId,
		Cmd:  commands,
		Tty:  true,
	},
	d.HostConfig,
	d.NetworkingConfig,
	nil,"")
	if err != nil {
		d.Log(fmt.Sprintf("Error Creating Container : %s",err.Error()))
		return nil , nil , err
	}
	if !keep{
		defer func(){
			d.Log(fmt.Sprintf("Stopping Container : %s",resp.ID))
			stopErr := d.StopContainer(resp.ID)
			if stopErr != nil {
				d.Log(fmt.Sprintf("ContainerError: %s",stopErr.Error()))
			}
			d.Log(fmt.Sprintf("Deleting Container: %s",resp.ID))
			delErr := d.DeleteContainer(resp.ID)
			if delErr != nil{
				d.Log(fmt.Sprintf("ContainerError: %s",delErr.Error()))
			}
		}()

	}else{
		defer d.Log(fmt.Sprintf("Keeping Container: %s",resp.ID))
	}
	err = d.client.ContainerStart(context.Background(),resp.ID,types.ContainerStartOptions{})
	if err != nil {
		d.Log((fmt.Sprintf("Container: %s",err.Error())))
		return nil , nil , err
	}
	statusCh , errCh := d.client.ContainerWait(context.Background(),resp.ID,container.WaitConditionNotRunning)
	select {
		case err := <- errCh:
			if err != nil {
				panic(err)
			}
			case <- statusCh:

	}
	out , err := d.client.ContainerLogs(context.Background(),resp.ID,types.ContainerLogsOptions{ShowStderr: true,ShowStdout: true})
	if err != nil {
		return nil , nil , err
	}
	io.Copy(Out,out)
	io.Copy(Err,out)
	err = nil
	return Out , Err , err
}

func (d *DockerManager) StopContainer(containerId string) error {
	d.init()
	return d.client.ContainerStop(context.Background(),containerId,nil)
}

func (d *DockerManager) DeleteContainer(containerId string) error {
	d.init()
	return d.client.ContainerRemove(context.Background(),containerId,types.ContainerRemoveOptions{Force: true})
}


