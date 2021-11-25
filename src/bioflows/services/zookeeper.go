package services

import (
	"github.com/bioflows/src/bioflows/kv"
	"github.com/hashicorp/consul/api"
)

type ZooKeeperOrchestrator struct {

	kvStore *kv.ZookeeperKVStoreManager
}

func (o *ZooKeeperOrchestrator) Services() (map[string]*api.AgentService,error){
	return nil , nil
}

func (o *ZooKeeperOrchestrator) Setup(credentials kv.Credentials) error{
	return nil
}

func(o *ZooKeeperOrchestrator) FindService(serviceName , tag string , passingOnly bool) ([]*api.ServiceEntry, *api.QueryMeta, error){
	return nil , nil , nil
}

func(o *ZooKeeperOrchestrator) Deregister(id string) error{
	return nil
}

func(o *ZooKeeperOrchestrator) Register(name string , address string, port int) error{
	return nil
}