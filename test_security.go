package main

import (
	"fmt"
	"github.com/bioflows/src/bioflows/security"
)

func main(){
	privKey , err := security.ReadPrivKeyFile("/home/snouto/.ssh/bfsecurity","")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(privKey.Size())
}
