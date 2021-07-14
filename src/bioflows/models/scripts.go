package models

type Scriptable string

func (s Scriptable) ToString() string {
	return string(s)
}

type Script struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	Code Scriptable `json:"code,omitempty" yaml:"code,omitempty"`
	CodeFile string `json:"file,omitempty" yaml:"file,omitempty"`
	Order int `json:"order,omitempty" yaml:"order,omitempty"`
	After bool `json:"after,omitempty" yaml:"after,omitempty"`
	Before bool `json:"before,omitempty" yaml:"before,omitempty"`
	InLoop bool `json:"inloop,omitempty" yaml:"inloop,omitempty"`

}

func (s Script) IsInLoop() bool {
	return s.InLoop
}

func (s Script) IsBefore() bool {
	return s.Before
}

func (s Script) IsAfter() bool {
	return s.After
}
