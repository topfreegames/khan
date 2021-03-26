// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/services"
	"github.com/uber-go/zap"
)

var encryptionScriptDebug bool
var encryptionScriptQuiet bool

// encryptionScriptCmd represents the encryption script
var encryptionScriptCmd = &cobra.Command{
	Use:   "encryption-script",
	Short: "start the khan encryption-scription script",
	Long: `Starts khan encryption-scription script that encrypt player names.
You can use environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		ll := zap.InfoLevel
		if encryptionScriptDebug {
			ll = zap.DebugLevel
		}
		if encryptionScriptQuiet {
			ll = zap.ErrorLevel
		}
		logger := zap.New(
			zap.NewJSONEncoder(), // drop timestamps in tests
			ll,
		)

		cmdL := logger.With(
			zap.String("source", "encryptionScriptCmd"),
			zap.String("operation", "Run"),
			zap.Bool("debug", encryptionScriptDebug),
		)

		log.D(cmdL, "Creating application...")
		script := services.GetEncryptionScript(
			ConfigFile,
			encryptionScriptDebug,
			logger,
		)
		log.D(cmdL, "Application created successfully.")

		log.D(cmdL, "Starting script...")
		script.Start()
	},
}

func init() {
	RootCmd.AddCommand(encryptionScriptCmd)

	encryptionScriptCmd.Flags().BoolVarP(&encryptionScriptDebug, "debug", "d", false, "Debug mode")
	encryptionScriptCmd.Flags().BoolVarP(&encryptionScriptQuiet, "quiet", "q", false, "Quiet mode (log level error)")
}
