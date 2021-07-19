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
	"fmt"
	"github.com/spf13/cobra"
	"os"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	OutputDir      string
	DataDir        string
	WorkflowId     string
	WorkflowName   string
	initialsConfig string
)

var cfgFile string

var Version string = "1.0.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bf",
	Short: "Distributed Containerized Bioinformatics and Computational Workflows/Pipeline engine built in Golang",
	Long: `BioFlows is a distributed pipeline framework for expressing , designing and running scalable reproducible and distributed computational bioinformatics workflows in cloud containers. BioFlows Framework consists of software tools and cloud microservices that communicate together to achieve a highly distributed , highly coordinated and fault tolerant environment to run parallel bioinformatics pipelines onto cloud containers and cloud servers. BioFlows also has BioFlows Description Language (BDL) which is an imperative and declarative standard for describing and expressing computational bioinformatics tools and pipelines, BDL is flexible , easy to use and a human readable language that enables researchers to design reproducible and scalable computational pipelines.The language is based entirely on Yet Another Markup Language (YAML).`,
	Version: Version,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	RunScript: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&OutputDir,"output_dir","","Output Directory where the running tool will save data.")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bf.yaml)")
	rootCmd.PersistentFlags().StringVar(&DataDir,"data_dir","","The directory which contains raw data.")
	rootCmd.PersistentFlags().StringVar(&initialsConfig,"initials","","A file which contains your Pipeline specific initial parameters' values. You can know the required parameters for your pipeline through reading its definition file or running bf validate command.")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".bf" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".bf")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		cfgFile = viper.ConfigFileUsed()
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
