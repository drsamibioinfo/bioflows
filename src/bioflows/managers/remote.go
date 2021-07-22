package managers

import (
	"github.com/bioflows/src/bioflows/helpers"
	"github.com/bioflows/src/bioflows/models"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"net"
	"net/http"
	"runtime"
	"time"
)

var (
	ERR_CONSUL_CLIENT_NULL = fmt.Errorf("Consul Client is null....")
	ERR_KV_EMPTY           = fmt.Errorf("KV Store is Empty....")
	ERR_NOT_FOUND = fmt.Errorf("Not Found")
)

type ClusterStateManager struct {
	client *api.Client
}
func (c *ClusterStateManager) RemoveConfigByID(key string) bool {
	_ , err := c.client.KV().DeleteTree(key,nil)
	if err != nil {
		return false
	}
	return true
}
func (c *ClusterStateManager) GetPipelineState(pipelineKey string) (models.FlowConfig, error) {

	if c.client != nil {
		finalConfig := models.FlowConfig{}
		kv := c.client.KV()
		pairs , _ , err := kv.List(pipelineKey,nil)
		if err != nil {
			return nil , err
		}
		if len(pairs) <= 0 {
			return nil , ERR_KV_EMPTY
		}
		for _ , pair := range pairs {
			if pair == nil{
				continue
			}
			state := make(map[string]interface{})
			err = json.Unmarshal(pair.Value,&state)
			if err != nil {
				continue
			}
			finalConfig[helpers.GetToolIdFromKey(pair.Key)] = state
		}
		return finalConfig , nil
	}
	return nil , ERR_CONSUL_CLIENT_NULL
}
func (c *ClusterStateManager) GetStateByID(stepId string) (interface{},error){
	if c.client != nil{
		kv := c.client.KV()
		kpair , _ , err := kv.Get(stepId,nil)
		if err != nil {
			return nil , err
		}
		if kpair == nil {
			return nil , ERR_NOT_FOUND
		}
		state := make(map[string]interface{})
		err = json.Unmarshal(kpair.Value,&state)
		if err != nil {
			return nil , err
		}
		return state, nil
	}
	return nil , ERR_CONSUL_CLIENT_NULL
}
func (c *ClusterStateManager) SetStateByID(stepId string,config interface{}) error {
	if c.client != nil {
		kv := c.client.KV()
		data , err := json.Marshal(config)
		if err != nil {
			return err
		}
		kpair := &api.KVPair{
			Key: stepId,
			Value: data,
		}
		_ , err = kv.Put(kpair,nil)
		return err
	}
	return nil
}

func (c *ClusterStateManager) Setup(config map[string]interface{}) error {
	cluster, ok := config["cluster"]
	if !ok {
		return fmt.Errorf("Cluster Section in Configuration settings doesn't exist")
	}
	var FQDN , Scheme string

	if section , ok := cluster.(map[interface{}]interface{});ok {
		address, _ := section["address"]
		port , _ := section["port"]
		Scheme = fmt.Sprintf("%v",section["scheme"])
		FQDN = fmt.Sprintf("%s:%d",address,port)
	}
	agentConfig := &api.Config{
		Address : FQDN,
		Scheme: Scheme,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}
	client, err := api.NewClient(agentConfig)
	if err != nil {
		return err
	}
	c.client = client

	return nil
}