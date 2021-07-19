package executors

import (
	"github.com/bioflows/src/bioflows/models"
	"github.com/bioflows/src/bioflows/models/pipelines"
)

var (
	DEFAULT_CHAINERS = make([]TransformCall,0)
)

type TransformCall func (b *pipelines.BioPipeline,config models.FlowConfig) error

func init(){
	DEFAULT_CHAINERS = append(DEFAULT_CHAINERS,[]TransformCall{UseUrl,}...)
}

func PreprocessPipeline(b *pipelines.BioPipeline,config models.FlowConfig, transforms ...TransformCall)  {
	if transforms != nil && len(transforms) > 0 && transforms[0] != nil {
		DEFAULT_CHAINERS = append(DEFAULT_CHAINERS,transforms...)
	}
	if len(DEFAULT_CHAINERS) <= 0{
		return
	}
	var err error
	for _ , transform := range DEFAULT_CHAINERS {
		err = transform(b,config)
		if err != nil {
			panic(err)
		}
	}
}
