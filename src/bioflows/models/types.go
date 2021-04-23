package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
)

type FlowConfig map[string]interface{}

func (f *FlowConfig) Fill(config map[string]interface{})  {
	for k , v := range config {
		(*f)[k] = v
	}

}

func (f *FlowConfig) Add(name string , val interface{}) {
	(*f)[name] = val
}

func (f *FlowConfig) GetAsMap() map[string]interface{}{
	newMap := make(map[string]interface{})
	for k , v := range *f {
		newMap[k] = v
	}
	return newMap
}
func (f *FlowConfig) ToBytes() ([]byte,error){
	data := f.GetAsMap()
	bytes , err := json.Marshal(data)
	if err != nil {
		return nil , err
	}
	return bytes, nil
}
func (f *FlowConfig) ToJson() (string,error) {
	data , err := f.ToBytes()
	if err != nil {
		return "" , err
	}
	return string(data) , nil
}
func (f *FlowConfig) FromJson(jsonData []byte) error {
	return json.Unmarshal(jsonData,f)
}

type Parameter struct {
	DisplayName string      `json:"displayname,omitempty" yaml:"displayname,omitempty"`
	Name        string      `json:"name" yaml:"name"`
	Type        string      `json:"type,omitempty" yaml:"type,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Value       interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	Attachable *bool 		`json:"attach,omitempty" yaml:"attach,omitempty"`
}
func (p *Parameter) IsAttachable() bool {
	return *p.Attachable
}

func (p *Parameter) GetParamValue() string {
	return fmt.Sprintf("%v", p.Value)
}

func (p *Parameter) Fill(t *Parameter) error {
	if t.Value != nil {
		p.Value = t.Value
	}
	if t.Description != "" {
		p.Description = t.Description
	}
	if t.DisplayName != "" {
		p.DisplayName = t.DisplayName
	}
	if t.Type != "" {
		p.Type = t.Type
	}
	if t.Attachable != nil {
		*p.Attachable = *t.Attachable
	}
	return nil
}

func (p *Parameter) GetDescription() string{
	if len(p.Description) > 0{
		return p.Description
	}else{
		return p.DisplayName
	}

}

type Reference struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Website     string `json:"website,omitempty" yaml:"website,omitempty"`
}

type Maintainer struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	FullName string `json:"fullname,omitempty" yaml:"fullname,omitempty"`
	Email    string `json:"email,omitempty" yaml:"email,omitempty"`
}

type Notification struct {
	To    string `json:"to" yaml:"to"`
	CC    string `json:"cc,omitempty" yaml:"cc,omitempty"`
	Title string `json:"title" yaml:"title"`
	Body  string `json:"body" yaml:"body"`
}

type Capabilities struct {
	CPU int `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory int `json:"memory,omitempty" yaml:"memory,omitempty"`
}

type ContainerConfig struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
}

func (c *ContainerConfig) GetAuth() (string,error) {
	auth := types.AuthConfig{
		Username:      c.Username,
		Password:      c.Password,
	}
	encodedJson , err := json.Marshal(auth)
	if err != nil {
		return "" , err
	}
	return base64.URLEncoding.EncodeToString(encodedJson) , nil
}

