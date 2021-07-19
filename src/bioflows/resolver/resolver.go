package resolver

import (
	"github.com/bioflows/src/bioflows/config"
	"strings"
)

func ResolveToolKey(toolId string , pipelineId string) string {
	// Tool Key: bioflows/pipelines/%pId/%tId
	return strings.Join([]string{pipelineId,toolId},"/")
}

func ResolveLeaderKey() string {
	//Leader Key: bioflows/nodes/leader
	return strings.Join([]string{config.BIOFLOWS_NAME, config.BIOFLOWS_NODES, config.BIOFLOWS_LEADER},"/")
}

func ResolvePipelineKey(pipelineId string) string {
	return strings.Join([]string{config.BIOFLOWS_NAME, config.BIOFLOWS_PIPELINES,pipelineId},"/")
}

