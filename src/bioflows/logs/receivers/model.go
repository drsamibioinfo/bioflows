package receivers


import "time"

type LogMessage struct {
	Message string `json:"message,omitempty"`
	Time time.Time	`json:"time,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Level string `json:"level,omitempty"`
}


