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
	"github.com/topfreegames/khan/log"
	"github.com/uber-go/zap"
)

var workerDebug bool
var workerQuiet bool

// workerCmd represents the start command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "starts the khan hook dispatching worker",
	Long: `Starts khan hook dispatching worker with the specified arguments. You can use
environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		ll := zap.InfoLevel
		if debug {
			ll = zap.DebugLevel
		}
		if quiet {
			ll = zap.ErrorLevel
		}
		l := zap.New(
			zap.NewJSONEncoder(), // drop timestamps in tests
			ll,
		)

		cmdL := l.With(
			zap.String("source", "workerCmd"),
			zap.String("operation", "Run"),
			zap.String("host", host),
			zap.Int("port", port),
			zap.Bool("debug", debug),
		)

		log.D(cmdL, "Creating application...")
		app := api.GetApp(
			host,
			port,
			ConfigFile,
			debug,
			l,
			false,
		)
		log.D(cmdL, "Application created successfully.")

		log.D(cmdL, "Starting dispatcher...")
		app.StartWorkers()
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)

	workerCmd.Flags().BoolVarP(&workerDebug, "debug", "d", false, "Debug mode")
	workerCmd.Flags().BoolVarP(&workerQuiet, "quiet", "q", false, "Quiet mode (log level error)")
}
