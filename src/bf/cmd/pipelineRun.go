package cmd

import (
	"github.com/bioflows/src/bioflows/cli"
	"github.com/bioflows/src/bioflows/models"
	"errors"
	"github.com/spf13/cobra"
	"strings"
)

var (
	clean          bool
	positionalArgs models.FlowConfig
)

var workflowRunCmd = &cobra.Command{
	Use:"run [pipeline file .bp]",
	Short: "",
	Long:"",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1{
			return errors.New("Please provide a pipeline file to run.")
		}
		if len(DataDir) < 1{
			return errors.New("Data Directory Flag is required.")
		}
		if len(OutputDir) < 1 {
			return errors.New("Output Directory Flag is required.")
		}
		toolPath := args[0]
		return cli.RunPipeline(cfgFile,toolPath,OutputDir,DataDir, initialsConfig,clean, positionalArgs)
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1{
			return cmd.Help()
		}
		positionalArgs, _ = parseArgs(args[1:])
		return nil
	},

}

func parseArgs(args []string) (models.FlowConfig,error){
	config := models.FlowConfig{}
	for _ , item := range args {
		if !strings.Contains(item,"="){
			continue
		}
		splitted := strings.Split(item,"=")
		config[splitted[0]] = splitted[1]
	}
	return config, nil
}

func init(){
	workflowRunCmd.PersistentFlags().BoolVar(&clean,"clean",false,"This command cleans all metadata associated with this pipeline from the distributed in-memory Key/Value Store, " +
		"in case you are running in a distributed mode. this command has no effect if you are running in a local mode.")

	workflowRunCmd.MarkFlagRequired(OutputDir)
	workflowRunCmd.MarkFlagRequired(DataDir)
	WorkflowCmd.AddCommand(workflowRunCmd)
}