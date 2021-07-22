package main

import (
	"github.com/bioflows/src/bioflows/helpers/profiling"
	"fmt"
)

func main(){
	fmt.Println(profiling.GetCPUProfile())
}
