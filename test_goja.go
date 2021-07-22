package main

import (
	"fmt"
	"github.com/bioflows/src/bioflows/scripts/io"
	"github.com/dop251/goja"

)


func main(){
	vm := goja.New()
//	js := `
//	var files = io.SelectMultiple('/home/snouto/Downloads',".tar.gz");
//	for(var i = 0 ; i < files.length;i++){
//	io.Print(files[i]);
//}
//`
	js2 := `
	var contents = io.ReadFile('/home/snouto/projects/bioflows/requirements.txt')
	io.Print(contents)
`

	io := &io.IO{
		VM : vm,
	}
	vm.Set("io",io)

	_ , err := vm.RunString(js2)
	if err != nil {
		fmt.Printf("Error : %s",err.Error())
	}
}
