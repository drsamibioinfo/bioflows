package main

import (
	"fmt"
	"github.com/hoisie/mustache"
)

func main(){
	var config map[string]interface{} = make(map[string]interface{})
	config["input_dir"] = "/home/snouto"
	array := make([]string,1)
	array = append(array,[]string{"Mohamed","Ibrahim","Fawzy"}...)
	config["names"] = array
	config["loop_index"] = 1

	data := mustache.Render("{{names.index}}",config)
	fmt.Println(data)

}
