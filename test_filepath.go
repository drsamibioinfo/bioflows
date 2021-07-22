package main

import (
	"github.com/bioflows/src/bioflows/helpers"
	"fmt"
	"path/filepath"
)

func main(){
	//r := "/home/snouto/projects/bioflows/scripts/looppip.yaml"
	r2 := "file:///env/looppip.yaml"
	details := helpers.FileDetails{}
	helpers.GetFileDetails(&details,r2)
	fmt.Println(details)
	fmt.Println(filepath.Join("http://www.google.com",details.Base))
	//u , _ := url.Parse(r)
	//dir := filepath.Dir(u.Path)
	//baseURL := fmt.Sprintf("%s://%s",u.Scheme,u.Host)
	//fmt.Println(baseURL)
	//fmt.Println(strings.Replace(r,filepath.Base(r),"",1))
	//fmt.Println(filepath.Base(r))
}

