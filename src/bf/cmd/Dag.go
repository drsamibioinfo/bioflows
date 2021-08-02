package cmd

import (
	"github.com/bioflows/src/bioflows/cli"
	"fmt"
	"github.com/spf13/cobra"
)


var DagCmd = &cobra.Command {
	Use:"Dag",
	Short: "This command enables creation of GraphViz graph for BioFlows Pipeline",
	Long: `This command enables the creation of GraphViz Graph for BioFlows Pipeline`,
	RunE: func(cmd *cobra.Command,args []string) error{
		if len(args) <= 0 {
			return fmt.Errorf("BioFlows Dag requires two parameters")
		}
		PipelineFile := args[0]
		if len(PipelineFile) <= 0 {
			return fmt.Errorf("BioFlows Dag requires two parameters")
		}
		dotString,  err := cli.RenderGraphViz(PipelineFile,bfLogger)
		if err != nil {
			return err
		}
		fmt.Print(dotString)
		return nil
	},
}

func init(){

	rootCmd.AddCommand(DagCmd)
}
