package main

import (
	"fmt"
	"github.com/mbndr/logo"
	"os"
)
type mywriter struct{}

func (self *mywriter) Write(p []byte) (bytes int, err error){
	fmt.Printf("MyWriter Message: %s",string(p))
	return len(p) , nil
}

func main(){
	cliRec := logo.NewReceiver(os.Stderr, "prefix ")
	cliRec.Color = true
	cliRec.Level = logo.DEBUG
	log := logo.NewLogger(cliRec,logo.NewReceiver(os.Stdout,"Prefix"),
		logo.NewReceiver(&mywriter{},"My Writer"))
	log.Receivers = append(log.Receivers,logo.NewReceiver(os.Stdout,"Hello"))
	log.Error("This is an error message")
}
