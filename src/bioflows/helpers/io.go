package helpers

import (
	"fmt"
	"os"
)


type FileDetails struct {
	Base string
	FileName string
	Scheme string
	Local bool
}
func (f FileDetails) String() string {
	return fmt.Sprintf("Base : %s\nFileName: %s\nScheme: %s\nLocal: %v",f.Base,
		f.FileName,f.Scheme,f.Local)
}

func WriteOrAppend(filename string , data []byte , perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}


func PathExists(filepath string) bool {
	_ , err := os.Stat(filepath)
	return !os.IsNotExist(err)
}
