package services

import (
	"github.com/bioflows/src/bioflows/kv"
	"fmt"
	"github.com/hashicorp/consul/api"
)

type ConsulOrchestrator struct{

	kvStore *kv.ConsulKVStoreManager
}


func (o *ConsulOrchestrator) Services() (map[string]*api.AgentService,error){
	client := o.kvStore.GetClient().(*api.Client)
	return client.Agent().Services()
}

func (o *ConsulOrchestrator) Setup(credentials kv.Credentials) error{
	o.kvStore = &kv.ConsulKVStoreManager{}
	return o.kvStore.Setup(credentials)
}

func(o *ConsulOrchestrator) FindService(serviceName , tag string , passingOnly bool) ([]*api.ServiceEntry, *api.QueryMeta, error){
	client := o.kvStore.GetClient().(*api.Client)
	addrs, meta, err := client.Health().Service(serviceName, tag, passingOnly, nil)

	if len(addrs) == 0 && err == nil {
		return nil,nil, fmt.Errorf("service ( %s ) was not found", serviceName)
	}
	if err != nil {
		return nil,nil, err
	}
	return addrs, meta, nil

}

func(o *ConsulOrchestrator) Deregister(id string) error{
	client := o.kvStore.GetClient().(*api.Client)
	return client.Agent().ServiceDeregister(id)
}

func(o *ConsulOrchestrator) Register(name string , address string, port int) error{
	client := o.kvStore.GetClient().(*api.Client)
	serviceEntry := &api.AgentServiceRegistration{
		Name:name,
		Port:port,
		Address:address,
		Check:&api.AgentServiceCheck{
			Interval:"10m",
			Name:name,
			TCP:fmt.Sprintf("%s:%d",address,port),
		},
	}
	return client.Agent().ServiceRegister(serviceEntry)
}

