/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/bioflows/src/bioflows/cli"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [tool file .bt]",
	Short: "This command will run a specific BioFlows Tool Definition file (.bt) and wraps the single Tool in a new Workflow.",
	Long: `This command will run a specific BioFlows Tool Definition file (.bt) and wraps the single Tool in a new Workflow.`,
	RunE: func(cmd *cobra.Command, args []string)  error{
		if len(args) < 1{
			return fmt.Errorf("Please specify the location of Bioflows Tool")
		}
		if len(DataDir) < 1{
			return errors.New("Data Directory Flag is required.")
		}
		if len(OutputDir) < 1 {
			return errors.New("Output Directory Flag is required.")
		}
		toolPath := args[0]
		return cli.RunTool(cfgFile,toolPath,WorkflowId,WorkflowName,OutputDir,DataDir, initialsConfig,positionalArgs)
	},
	Args: func(cmd *cobra.Command, args []string) error {
		positionalArgs, _ = parseArgs(args[1:])
		return nil
	},
}

func init() {
	//define local flags for run subcommand
	runCmd.Flags().StringVarP(&WorkflowId,"workflowId","i","","Assign a unique Identifier for the workflow in order to distinguish the workflow later.")
	runCmd.Flags().StringVarP(&WorkflowName,"workflowName","n","myworkflow","Assign a human readable identifier for the current workflow.")
	runCmd.MarkFlagRequired("workflowId")
	runCmd.MarkFlagRequired("workflowName")

	ToolCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
