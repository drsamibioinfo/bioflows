package logs

import (
	config2 "github.com/bioflows/src/bioflows/config"
	"github.com/bioflows/src/bioflows/logs/receivers"
	"github.com/bioflows/src/bioflows/models"
	"github.com/mbndr/logo"
	"os"
	"strings"
)

const (
	LOGS_SECTION_NAME = "logs"
	LOGS_OUTPUT_DIR = "output_dir"
)
func getReceiver(receiver models.LoggerReceiver,config models.FlowConfig) *logo.Receiver {
	var rec *logo.Receiver
	switch(receiver.Type) {
	case "kafka":
		rec = logo.NewReceiver(os.Stdout,config2.BIOFLOWS_NAME)
	case "es":
		fallthrough
	case "elasticsearch":
		fallthrough
	case "elastic":
		es := &receivers.ESReceiver{}
		es.SetConfig(receiver.ToMap())
		rec = logo.NewReceiver(es,config2.BIOFLOWS_NAME)
	default:
		rec = logo.NewReceiver(os.Stdout,config2.BIOFLOWS_NAME)
	}
	return rec
}


func NewLogger(config models.FlowConfig) *logo.Logger {
	logger := logo.NewLogger()
	logger.Receivers = append(logger.Receivers,logo.NewReceiver(os.Stdout,config2.BIOFLOWS_DISPLAY_NAME))

	if data , ok := config["logging"] ; ok {
		logging := data.(map[string]interface{})
		if receivers , ok := logging["receivers"]; ok {
			for _,receiver := range receivers.([]models.LoggerReceiver){
				rec := getReceiver(receiver,config)
				rec.Active = true
				rec.Color = true
				switch strings.ToLower(receiver.Level){
				case "debug":
					rec.Level = logo.DEBUG
				case "error":
				case "erro":
				case "err":
					rec.Level = logo.ERROR
				case "info":
					rec.Level = logo.INFO
				case "warn":
				case "warning":
				case "warnings":
					rec.Level = logo.WARN
				case "fatal":
					rec.Level = logo.FATAL
				}
				logger.Receivers = append(logger.Receivers,rec)
			}
		}
	}
	return logger
}