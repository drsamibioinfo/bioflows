package main

import (
	"github.com/bioflows/src/bioflows/managers"
	"fmt"
)

func main(){
	cluster := managers.ClusterStateManager{}
	config := make(map[string]interface{})
	m := make(map[string]string)
	m["address"] = "127.0.0.1"
	m["port"] = "8500"
	m["Scheme"] = "http"
	config["cluster"] = m

	if err := cluster.Setup(config); err != nil {
		fmt.Println(err.Error())
		return
	}
	// define data
	data := make(map[string]interface{})
	data["First"] = "First"
	data["Second"] = "Second"
	data["Third"] = "Third"

	if err := cluster.SetStateByID("nodes/FirstNode",data); err != nil{
		fmt.Println(err.Error())
		return
	}
	fmt.Println("State was persisted successfully.")
	anotherData , err := cluster.GetStateByID("nodes/FirstNode")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	converted_data := anotherData.(map[string]interface{})
	fmt.Println(converted_data)



}
