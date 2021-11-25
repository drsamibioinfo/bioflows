package filesystem

import (
	"github.com/bioflows/src/bioflows/config"
	"strings"
)

const (
	FILESYSTEM_SECTION_NAME="filesystem"
	FILESYSTEM_MANAGER_KEY = "manager_name"
	NORMAL_FILESYSTEM = "normal"
	HADOOP_FILESYSTEM = "hadoop"
)

type FileSystemManager interface {
	EnumerateFolder(directory string) []string
}

var activeManager FileSystemManager

func init(){

	result , err := config.HasKey(FILESYSTEM_SECTION_NAME,FILESYSTEM_MANAGER_KEY)
	if !result && err != nil {
		//logs.WriteLog(err.Error())
		return
	}

	val , err := config.GetKeyAsString(FILESYSTEM_SECTION_NAME,FILESYSTEM_MANAGER_KEY)
	if err != nil {
		//logs.WriteLog(err.Error())
		return
	}

	switch(strings.ToLower(val)){
		case HADOOP_FILESYSTEM:
			activeManager = &HadoopFileSystemManager{}
			break;
		case NORMAL_FILESYSTEM:
			fallthrough
		default:
			activeManager = &NormalFileSystemManager{}
			break
	}
}

func GetDefaultFileSystemManager() FileSystemManager{
	return activeManager
}
