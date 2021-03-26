package pipelines

import (
	"bioflows/models"
	"encoding/json"
	"fmt"
	"strings"
)

type BioPipeline struct {
	Type         string               `json:"type,omitempty" yaml:"type,omitempty"`
	Depends      string               `json:"depends,omitempty" yaml:"depends,omitempty"`
	ImageId      string               `json:"imageId,omitempty" yaml:"imageId,omitempty"`
	ID           string               `json:"id,omitempty" yaml:"id,omitempty"`
	Order        int                  `json:"order,omitempty" yaml:"order,omitempty"`
	BioflowId    string               `json:"bioflowId,omitempty" yaml:"bioflowId,omitempty"`
	URL          string               `json:"url,omitempty" yaml:"url,omitempty"`
	Name         string               `json:"name" yaml:"name"`
	Description  string               `json:"description,omitempty" yaml:"description,omitempty"`
	Discussions  []string             `json:"discussions,omitempty" yaml:"discussions,omitempty"`
	Website      string               `json:"website,omitempty" yaml:"website,omitempty"`
	Version      string               `json:"version,omitempty" yaml:"version,omitempty"`
	Icon         string               `json:"icon,omitempty" yaml:"icon,omitempty"`
	Shadow       bool                 `json:"shadow,omitempty" yaml:"shadow,omitempty"`
	Loop bool 	`json:"loop,omitempty" yaml:"loop,omitempty"`
	LoopVar string `json:"loop_var,omitempty" yaml:"loop_var,omitempty"`
	Maintainer   *models.Maintainer   `json:"maintainer,omitempty" yaml:"maintainer,omitempty"`
	References   []models.Reference   `json:"references,omitempty" yaml:"references,omitempty"`
	Inputs       []models.Parameter   `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Config       []models.Parameter   `json:"config,omitempty" yaml:"config,omitempty"`
	Outputs      []models.Parameter   `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Scripts      []models.Script      `json:"scripts,omitempty" yaml:"scripts,omitempty"`
	Command      models.Scriptable    `json:"command" yaml:"command"`
	Dependencies []string             `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Deprecated   bool                 `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Conditions   []models.Scriptable  `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Steps        []BioPipeline        `json:"steps,omitempty" yaml:"steps,omitempty"`
	Notification *models.Notification `json:"notification,omitempty" yaml:"notification,omitempty"`
	Caps         *models.Capabilities `json:"caps,omitempty" yaml:"caps,omitempty"`
	ContainerConfig *models.ContainerConfig `json:"container,omitempty" yaml:"container,omitempty"`
}

func (instance *BioPipeline) GetIdentifier() string {
	return fmt.Sprintf("%s-%s", instance.Name, instance.ID)
}

func (instance *BioPipeline) Prepare() {
	//if the tool name is not set, then use the tool ID
	if len(instance.Name) <= 0 {
		instance.Name = instance.ID
	}
	//If the tool name is set , use that as the tool instance name
	instance.Name = instance.Name
}

func (p BioPipeline) ToTool() *models.Tool {
	t := &models.Tool{}
	t.Type = p.Type
	t.Depends = p.Depends
	t.URL = p.URL
	t.ImageId = p.ImageId
	t.Notification = p.Notification
	t.ID = p.ID
	t.Order = p.Order
	t.Caps = p.Caps
	t.BioflowId = p.BioflowId
	t.Name = p.BioflowId
	t.Description = p.Description
	t.Discussions = make([]string, len(p.Discussions))
	copy(t.Discussions, p.Discussions)
	t.Website = p.Website
	t.Version = p.Version
	t.Icon = p.Icon
	t.Shadow = p.Shadow
	t.Maintainer = p.Maintainer
	t.Loop = p.Loop
	t.LoopVar = p.LoopVar
	t.Scripts = make([]models.Script,len(p.Scripts))
	copy(t.Scripts,p.Scripts)
	t.References = make([]models.Reference, len(p.References))
	copy(t.References, p.References)
	t.Inputs = make([]models.Parameter, len(p.Inputs))
	copy(t.Inputs, p.Inputs)
	t.Config = make([]models.Parameter, len(p.Config))
	copy(t.Config, p.Config)
	t.Outputs = make([]models.Parameter, len(p.Outputs))
	copy(t.Outputs, p.Outputs)
	t.Command = p.Command
	t.Dependencies = make([]string, len(p.Dependencies))
	copy(t.Dependencies, p.Dependencies)
	t.Deprecated = p.Deprecated
	t.Conditions = make([]models.Scriptable, len(p.Conditions))
	copy(t.Conditions,p.Conditions)
	t.ContainerConfig = p.ContainerConfig
	return t
}

func (p BioPipeline) IsTool() bool {
	if len(p.Type) <= 0 {
		return true
	}
	if strings.ToLower(p.Type) == "pipeline" || strings.ToLower(p.Type) == "workflow" {
		return false
	}
	return true
}

func (p BioPipeline) IsPipeline() bool {
	return !p.IsTool()
}
func (p BioPipeline) IsLoop() bool {
	return p.Loop
}

func (p *BioPipeline) ToJson() string {
	bytes, err := json.Marshal(p)
	if err != nil {
		panic(fmt.Errorf("Unable to Marshal current BioPipeline into JSON"))
	}
	return string(bytes)
}
