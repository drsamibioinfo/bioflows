package main

import (
	"github.com/bioflows/src/bioflows/models"
	json2 "encoding/json"
	"fmt"
)

func main(){
	p := models.FlowConfig{}
	cf := models.FlowConfig{}
	cf["remote"] = true
	cf["output_dir"] = "/home/snouto"
	cf["data_dir"] = "/home/snouto/projects"
	p.Fill(cf)
	p["cf"] = cf
	//json , err := p.ToJson()
	//if err != nil {
	//	fmt.Println("There was an error during json marshalling: ",err.Error())
	//	return
	//}
	//fmt.Println(json)
	bytes , _ := p.ToBytes()
	task := models.Task{
		TaskId: " snouto",
		Task:   []byte("Hello Mohamed"),
		Config: bytes,
	}
	json , err := json2.Marshal(task)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	anotherTask := models.Task{}
	anotherTask.FromJson(json)
	jsonData , _ := anotherTask.ToJson()
	fmt.Println("Another Task: ", jsonData)

}
