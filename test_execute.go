package main

import (
	"github.com/bioflows/src/bioflows/process"
	"fmt"
	"os"
)

func main(){

	executor := &process.CommandExecutor{Command: "ls -ll /home/snouto"}
	_, err := executor.Run()
	if err != nil {
		fmt.Printf("Error : %v\n",err)
		os.Exit(1)
	}
	buffOut, buffErr := executor.GetOutput() , executor.GetError()
	fmt.Printf("Output: %s\n Error: %s",buffOut.String(),buffErr.String())

}
