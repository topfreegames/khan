// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ConfigFile is the configuration file used for running a command
var ConfigFile string

// RootCmd is the root command for khan CLI application
var RootCmd = &cobra.Command{
	Use:   "khan",
	Short: "khan handles clans",
	Long:  `Use khan to handle clans for your game.`,
}

// Execute runs RootCmd to initialize khan CLI application
func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		glog.Fatal(err)
		os.Exit(-1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(
		&ConfigFile, "config", "c", "./config/local.yaml",
		"config file (default is ./config/local.yaml",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if ConfigFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(ConfigFile)
	}
	viper.SetEnvPrefix("khan")
	viper.SetConfigName(".khan") // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		glog.Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}
