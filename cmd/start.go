// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/topfreegames/khan/api"
)

var host string
var port int
var debug bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts khan server",
	Long: `Starts khan server with the specified arguments. You can use
environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		app := api.GetApp(
			host,
			port,
			cfgFile,
			debug,
		)

		//app.AddHandlers(api.URL{
		//Method:  "GET",
		//Path:    "/healthcheck",
		//Handler: handlers.HealthcheckHandler,
		//})

		app.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind khan to")
	startCmd.Flags().IntVarP(&port, "port", "p", 8888, "Port to bind khan to")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
}
