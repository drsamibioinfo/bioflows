package managers

import (
	"github.com/bioflows/src/bioflows/helpers/profiling"
	"github.com/bioflows/src/bioflows/models"
	"github.com/bioflows/src/bioflows/resolver"
	"fmt"
	"github.com/hashicorp/consul/api"
	"net"
	"net/http"
	"runtime"
	"time"
)

const (
	RENEW_INTERVAL = "30s"
)

type ClusterServiceManager struct {
	config models.FlowConfig
	client *api.Client
	sessionId string
	isLeader bool
	doneChan chan struct{}
}

func (c *ClusterServiceManager) IsLeader() bool {
	return c.isLeader
}

func (c *ClusterServiceManager) Services() (map[string]*api.AgentService,error){
	return c.client.Agent().Services()

}
func (c *ClusterServiceManager) FindService(serviceName , tag string , passingOnly bool) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	addrs , meta , err := c.client.Health().Service(serviceName,tag,passingOnly,nil)
	if len(addrs) == 0 && err == nil {
		return nil,nil, fmt.Errorf("service ( %s ) was not found", serviceName)
	}
	if err != nil {
		return nil , nil , err
	}
	return addrs , meta , nil
}

func (c *ClusterServiceManager) Deregister(id string) error {
	return c.client.Agent().ServiceDeregister(id)
}

func (c *ClusterServiceManager) createSession(sessionEntry *api.SessionEntry) error {
	sessionId, _ , err := c.client.Session().Create(sessionEntry,nil)
	if err != nil {
		return err
	}
	c.sessionId = sessionId
	return nil
}

func (c *ClusterServiceManager) ReleaseSession() error {
	_ , err := c.client.Session().Destroy(c.sessionId,nil)
	return err
}
func (c *ClusterServiceManager) SelfElect() bool {
	c.Release()
	err := c.createSession(&api.SessionEntry{
		Name: resolver.ResolveLeaderKey(),
		Behavior: "delete",
		TTL:RENEW_INTERVAL,
	})
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	jsonData, _ := profiling.GetCPUProfile().ToJson()
	c.isLeader,_ , err = c.client.KV().Acquire(&api.KVPair{
		Key:resolver.ResolveLeaderKey(),
		Value: jsonData,
		Session:c.sessionId,
	},nil)
	if err != nil {
		fmt.Println(fmt.Sprintf("Acquiring Leader: %s",err.Error()))
		return false
	}
	// That means we have successfully acquired Leadership
	if c.isLeader{
		c.doneChan = make(chan struct{})
		go func(){
			c.client.Session().RenewPeriodic(
				RENEW_INTERVAL,
				c.sessionId,
				nil,c.doneChan)
		}()

	}
	return true
}
func (c *ClusterServiceManager) Release() error {
	//Close the done Channel
	close(c.doneChan)
	return c.ReleaseSession()

}



func (c *ClusterServiceManager) Register(name, address string , port int) error {
	serviceEntry := &api.AgentServiceRegistration{
		Name : name,
		Port: port,
		Address:address,
		Check: &api.AgentServiceCheck{
			Interval:"10m",
			Name:name,
			TCP:fmt.Sprintf("%s:%s",address,port),
		},
	}
	return c.client.Agent().ServiceRegister(serviceEntry)
}

func (c *ClusterServiceManager) Setup(config models.FlowConfig) error {
	c.config = config
	c.doneChan = make(chan struct{})
	cluster, ok := config["cluster"]
	if !ok {
		return fmt.Errorf("Cluster Section in Configuration settings doesn't exist")
	}
	var FQDN , Scheme string

	if section , ok := cluster.(map[interface{}]interface{});ok {
		address, _ := section["address"]
		port , _ := section["port"]
		Scheme = fmt.Sprintf("%v",section["scheme"])
		FQDN = fmt.Sprintf("%s:%s",address,port)
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


