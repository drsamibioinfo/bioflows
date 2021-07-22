package logs

import (
	"log"
	"github.com/bioflows/src/bioflows/config"
	"os"
	"strings"
	"fmt"
)
const (
	LOGS_SECTION_NAME = "logs"
	LOGS_OUTPUT_DIR = "output_dir"
)

var logger *log.Logger

func init() {

	logger = &log.Logger{}
	logger.SetPrefix(config.BIOFLOWS_DISPLAY_NAME)
	if  result , _ := config.HasKey(LOGS_SECTION_NAME,LOGS_OUTPUT_DIR); result{

		output_dir , _ := config.GetKeyAsString(LOGS_SECTION_NAME,LOGS_OUTPUT_DIR)
		output_file , err := os.Create(strings.Join([]string{output_dir,config.BIOFLOWS_DISPLAY_NAME},"/"))
		if err != nil {
			//fmt.Println("Received Error while initializing the logs : ")
			//fmt.Println(err.Error())
			return
		}
		logger.SetOutput(output_file)
	}else{
		logger.SetOutput(os.Stdout)
	}
}

func NewLogger(name string , prefix string , location string) (*log.Logger , error) {
	newLogger := &log.Logger{}
	newLogger.SetPrefix(prefix)
	output_file , err := os.Create(strings.Join([]string{location,name},"/"))
	if err != nil {
		return nil , err
	}
	newLogger.SetOutput(output_file)
	return newLogger , nil
}

func WriteLog(log string)  {
	logger.Println(log)
	fmt.Println(log)
}