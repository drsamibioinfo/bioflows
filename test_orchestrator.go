package main

import (
	"github.com/bioflows/src/bioflows/kv"
	"github.com/bioflows/src/bioflows/services"
	"fmt"
)

func main(){
	fmt.Println("Testing Consul Service Registration/Discovery")
	o := services.ConsulOrchestrator{}
	o.Setup(kv.Credentials{
		Address:"localhost",
		Port:8500,
	})
	//err := o.Register("services/agents/bioflow2","localhost",8088)
	//if err != nil {
	//	fmt.Println("Received Error : ",err.Error())
	//	return
	//}
	services , err := o.Services()
	if err != nil {
		fmt.Println("Received another Error : ",err.Error())
		return
	}
	for k , v := range services{
		fmt.Println(fmt.Sprintf("Key : %s , Service Name : %s , Service Address : %s",k,v.Service,v.Address))
	}
	entries , _ , err := o.FindService("services/agents/bioflow2","",false)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _ , entry := range entries{
		fmt.Println(fmt.Sprintf("Service : %s",entry.Service.Service))
	}
	fmt.Println("Finished")
}
