package io

import (
	"fmt"
	"github.com/dop251/goja"
	"io/ioutil"
	"os"
	"strings"
)

type IO struct{
	VM *goja.Runtime
}
//Print(...anything)
func (o *IO) Print(call goja.FunctionCall) goja.Value{
	text := make([]string,0)
	for _ , val := range call.Arguments{
		text = append(text,val.String())
	}
	return o.VM.ToValue(strings.Join(text," "))
}
//ReadFile(fullFilePath)
func(o *IO) ReadFile(call goja.FunctionCall) goja.Value {
	fileName := call.Arguments[0].String()
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return o.VM.ToValue(string(contents))
}
//SelectSingle(dir,handle)
func (o *IO) SelectSingle(call goja.FunctionCall) goja.Value{
	if len(call.Arguments) < 2{
		panic(fmt.Errorf("SelectSingle Function takes two arguments."))
	}
	dir := call.Arguments[0].String()
	handle := call.Arguments[1].String()
	filteredFiles := make([]string,0)
	foundFiles , err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _ , file := range foundFiles{
		if strings.Contains(file.Name(),handle){
			filteredFiles = append(filteredFiles,strings.Join([]string{dir,file.Name()},string(os.PathSeparator)))
		}
	}
	//Return only a single matched file
	return o.VM.ToValue(filteredFiles[0])
}

//SelectMultiple(dir,handles)
func (o *IO) SelectMultiple(call goja.FunctionCall) goja.Value {
	dir := call.Arguments[0].String()
	handles := make([]string,0)
	filteredFiles := make([]interface{},0)
	for _,  val := range call.Arguments[1:]{
		handles = append(handles,val.String())
	}
	foundFiles , err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _ , file := range foundFiles{
		for _,  handle := range handles{
			if strings.Contains(file.Name(),handle){
				filteredFiles = append(filteredFiles,file.Name())
			}
		}
	}
	arr := o.VM.NewArray(filteredFiles...)
	return o.VM.ToValue(arr.Export())
}
//ListDir(DirPath)
func (o *IO) ListDir(call goja.FunctionCall) goja.Value {
	dir := call.Arguments[0].String()
	files := make([]interface{},0)
	foundFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _ , file := range foundFiles {
		//fullFilePath := strings.Join([]string{dir,file.Name()},string(os.PathSeparator))
		files = append(files,file.Name())
	}
	arr := o.VM.NewArray(files...)
	return o.VM.ToValue(arr.Export())
}
