package cmd

import (
	"fmt"
	"github.com/bioflows/src/bioflows/cli"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

var ValidateCmd = &cobra.Command{
	Use: "validate [Tool/Pipeline Path]",
	Short: `validates a given BioFlows tool or pipeline definition file. It checks whether the file is valid and well-formatted or not.
    	The file path could be a Local File System Path or a remote URL.
`,
	Long: `validates a given BioFlows tool or pipeline definition file. It checks whether the file is valid and well-formatted or not.
    	The file path could be a Local File System Path or a remote URL.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		bfLogger.SetPrefix("ValidateTool")
		if len(args) <= 0 {
			return cmd.Usage()
		}
		filePath := args[0]
		valid, err := cli.ValidateYAML(filePath)
		if err != nil {
			bfLogger.Error(fmt.Sprintf("%s", err.Error()))
			return err
		}
		if valid {
			bfLogger.Info("Validate Tool: The Tool is valid.")
		} else {
			bfLogger.Error("Validate Tool: The tool is not valid.")
		}
		table , err := cli.GetRequirementsTableFor(filePath)
		if err != nil {
			bfLogger.Error(fmt.Sprintf("%s",err.Error()))
			return err
		}
		bfLogger.Info("")
		bfLogger.Info(chalk.Underline.TextStyle("BioFlows Pipeline Input Requirements"))
		fmt.Println(table.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ValidateCmd)
}
