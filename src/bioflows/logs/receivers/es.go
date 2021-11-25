package receivers

import (
	"context"
	"encoding/json"
	"fmt"
	config2 "github.com/bioflows/src/bioflows/config"
	es7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"strings"
	"time"
)

type ESReceiver struct{
	Config map[string]interface{}
	client *es7.Client

}

func (es *ESReceiver) IsReady() bool {
	if es.client == nil {
		return false
	}
	return true
}

func (es *ESReceiver) SetConfig(config map[string]interface{}) {
	es.Config = config
}

func (es *ESReceiver) Setup() error {
	if es.client != nil {
		return nil
	}
	var addresses []string = make([]string,0)
	var username string = ""
	var password string = ""

	if data, ok := es.Config["fields"]; !ok {
		return fmt.Errorf("Elastic Search logger requires configuration fields to be specific")
	}else{
		fields := data.(map[string]interface{})
		if addrs , ok := fields["addresses"]; ok {
			for _ , addr := range addrs.([]interface{}) {
				addresses = append(addresses,fmt.Sprintf("%v",addr))
			}
		}
		if un, ok := fields["username"] ; ok {
			username = un.(string)
		}
		if pass, ok := fields["password"]; ok {
			password = pass.(string)
		}
		cfg := es7.Config{
			Addresses:             addresses,
			Username:              username,
			Password:             password,
			CloudID:               "",
			APIKey:                "",
			ServiceToken:          "",
			Header:                nil,
			CACert:                nil,
			RetryOnStatus:         nil,
			DisableRetry:          false,
			EnableRetryOnTimeout:  false,
			MaxRetries:            0,
			DiscoverNodesOnStart:  false,
			DiscoverNodesInterval: 0,
			EnableMetrics:         false,
			EnableDebugLogger:     false,
			DisableMetaHeader:     false,
			RetryBackoff:          nil,
			Transport:             nil,
			Logger:                nil,
			Selector:              nil,
			ConnectionPoolFunc:    nil,
		}
		client , err := es7.NewClient(cfg)
		if err != nil {
			return err
		}
		es.client = client
		return nil
	}

}

func (es *ESReceiver) Write(p []byte) (int,error){
	if !es.IsReady() {
		return 0 , fmt.Errorf("ElasticSearch Logging Receiver is not initialized yet.")
	}
	Prefix := config2.BIOFLOWS_DISPLAY_NAME
	var level string = "INFO"
	if data , ok := es.Config["level"] ; ok {
		level = data.(string)
	}
	msg := LogMessage{
		Message: string(p),
		Time:    time.Now(),
		Prefix:  Prefix,
		Level:   level,
	}
	//Now send that struct to Elastic Search
	var index string = fmt.Sprintf("%sLogs",config2.BIOFLOWS_DISPLAY_NAME)
	if data , ok := es.Config["fields"]; ok {
		indexdata := data.(map[string]interface{})
		if index_name, ok := indexdata["index"]; ok {
			index = index_name.(string)
		}
	}
	doc , err := json.Marshal(msg)
	if err != nil {
		return 0 , err
	}
	req := esapi.IndexRequest{
		Index:               index,
		Body:                strings.NewReader(string(doc)),
		Refresh:             "true",
	}
	ctx , Cancel := context.WithCancel(context.Background())
	resp , err := req.Do(ctx,es.client)
	if err != nil {
		Cancel()
		return 0 , err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return 0 , fmt.Errorf("[%s] Error indexing current Log to ElasticSearch",resp.Status())
	}

	return len(p) , nil
}

