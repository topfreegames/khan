package cmd

import (
	"github.com/spf13/cobra"
	"github.com/topfreegames/khan/loadtest"
	"github.com/topfreegames/khan/log"
	"github.com/uber-go/zap"
)

var sharedClansFile string
var nGoroutines int

var loadtestCmd = &cobra.Command{
	Use:   "loadtest",
	Short: "runs a load test against a remote Khan API",
	Long: `Runs a load test against a remote Khan API with the specified arguments.
You can use environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		loadtest.LoadRandomWords()
		logger := zap.New(zap.NewJSONEncoder(), zap.InfoLevel)

		exitChannel := make(chan bool)
		routine := func() {
			logger := logger.With(
				zap.String("source", "cmd/loadtest.go"),
				zap.String("operation", "loadtestCmd.Run/goroutine"),
			)

			app := loadtest.GetApp(ConfigFile, sharedClansFile, logger)
			if err := app.Run(); err != nil {
				log.E(logger, "Goroutine exited with error. Restarting...", func(cm log.CM) {
					cm.Write(zap.String("error", err.Error()))
				})
				exitChannel <- false
			} else {
				log.I(logger, "Goroutine exited without errors.")
				exitChannel <- true
			}
		}
		for i := 0; i < nGoroutines; i++ {
			go routine()
		}
		for i := 0; i < nGoroutines; {
			if ok := <-exitChannel; ok {
				i++
			} else {
				go routine()
			}
		}

		logger = logger.With(
			zap.String("source", "cmd/loadtest.go"),
			zap.String("operation", "loadtestCmd.Run"),
		)
		log.I(logger, "Application exited.")
	},
}

func init() {
	RootCmd.AddCommand(loadtestCmd)

	loadtestCmd.Flags().StringVar(
		&sharedClansFile,
		"clans",
		"./config/loadTestSharedClans.yaml",
		"shared clans list for load test",
	)

	loadtestCmd.Flags().IntVar(
		&nGoroutines,
		"goroutines",
		1,
		"number of goroutines to spawn for concurrent load tests",
	)
}
