package main

import (
	"github.com/bioflows/src/bioflows/kv"
	"fmt"
	"github.com/hashicorp/consul/api"
)

func main(){
	kvManager := &kv.ConsulKVStoreManager{}
	kvManager.Setup(kv.Credentials{
		Address:"localhost",
		Port:8500,
	})

	writeData , err := kvManager.Put(&api.KVPair{
		Key:"nodes/bioflows-agent",
		Value:[]byte("Hello Mohamed Fawzy"),
	},nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(fmt.Sprintf("Data Has been written at : %v" , writeData.RequestTime))
	fmt.Println("Now Fetching the data Key : nodes")
	kps , _ , err := kvManager.List("nodes/",nil)
	if err != nil {
		fmt.Println("Received Error : ",err.Error())
		return
	}
	if kps == nil {
		fmt.Println("No Keys Found that match the prefix.")
		return
	}
	for _ , kp := range kps {
		fmt.Println(fmt.Sprintf("Key : %s , Value : %s",kp.Key,string(kp.Value)))
	}

	fmt.Println("Finished.")

}
