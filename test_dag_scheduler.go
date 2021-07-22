package main

import (
	"github.com/bioflows/src/bioflows/executors"
	"github.com/bioflows/src/bioflows/models/pipelines"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func main(){
	pipeline := &pipelines.BioPipeline{}
	tool_in, err := os.Open("/home/snouto/projects/bioflows/scripts/nested.yaml")

	if err != nil {
		fmt.Printf("There was an error opening the tool file, %v\n",err)
		os.Exit(1)
	}
	mytool_content, err := ioutil.ReadAll(tool_in)
	if err != nil {
		fmt.Printf("Error reading the contents of the tool , %v\n",err)
		os.Exit(1)
	}
	err = yaml.Unmarshal([]byte(mytool_content),pipeline)
	if err != nil {
		//fmt.Println("There was a problem unmarshaling the current tool")
		fmt.Println(err.Error())
		return
	}
	graph , err := pipelines.CreateGraph(pipeline)
	scheduler := executors.DagScheduler{}
	rankedList , err := scheduler.Rank(pipeline,graph)
	for index , sublist := range rankedList {
		fmt.Printf("Slot: %d\n",index)
		for _, node := range sublist {
			if node == nil {
				continue
			}
			fmt.Printf("Slot: %d , Node: %s\n",index,node.ID)
		}
		fmt.Println("#################################################")
	}


}
