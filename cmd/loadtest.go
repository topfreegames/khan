package cmd

import (
	"github.com/spf13/cobra"
	"github.com/topfreegames/khan/loadtest"
	"github.com/topfreegames/khan/log"
	"github.com/uber-go/zap"
)

var sharedClansFile string

var loadtestCmd = &cobra.Command{
	Use:   "loadtest",
	Short: "runs a load test against a remote Khan API",
	Long: `Runs a load test against a remote Khan API with the specified arguments.
You can use environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := zap.New(zap.NewJSONEncoder(), zap.InfoLevel)

		l := logger.With(
			zap.String("source", "loadtestCmd"),
			zap.String("operation", "Run"),
		)

		app := loadtest.GetApp(ConfigFile, sharedClansFile, logger)
		if err := app.Run(); err != nil {
			log.P(l, "Application exited with error.", func(cm log.CM) {
				cm.Write(zap.String("error", err.Error()))
			})
		} else {
			log.I(l, "Application exited without errors.")
		}
	},
}

func init() {
	RootCmd.AddCommand(loadtestCmd)

	loadtestCmd.Flags().StringVar(
		&sharedClansFile,
		"clans",
		"./config/loadTestSharedClans.yaml",
		"shared clans list for load test (default is ./config/loadTestSharedClans.yaml)",
	)
}
