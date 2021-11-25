package engine

import (
	"github.com/bioflows/src/bioflows/config"
	"github.com/bioflows/src/bioflows/kv"
	"github.com/bioflows/src/bioflows/services"
	"strconv"
)


type BioFlowOrchestrator struct{

	orchestrator services.Orchestrator
}



func (o *BioFlowOrchestrator) Setup() error{
	orchestrator , err := services.GetDefaultOrchestrator()
	if err != nil {
		return err
	}
	address , err := config.GetKeyAsString(services.SERVICES_SECTION_NAME,"address")
	if err != nil {
		return  err
	}
	portString , err := config.GetKeyAsString(services.SERVICES_SECTION_NAME,"port")
	if err != nil {
		return err
	}
	port , err := strconv.Atoi(portString)
	if err != nil {
		return err
	}
	username , _ := config.GetKeyAsString(services.SERVICES_SECTION_NAME,"username")
	password , _ := config.GetKeyAsString(services.SERVICES_SECTION_NAME,"password")
	creds := kv.Credentials{
		Address:address,
		Port:int64(port),
		Username:username,
		Password:password,
	}
	orchestrator.Setup(creds)
	o.orchestrator = orchestrator
	return nil
}


