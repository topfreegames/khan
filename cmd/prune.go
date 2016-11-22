// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/topfreegames/khan/log"
	"github.com/uber-go/zap"
)

var pruneDebug bool
var pruneQuiet bool

func pruneStaleData() (models.PruneStats, error) {
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
		zap.String("source", "pruneCmd"),
		zap.String("operation", "Run"),
		zap.Bool("debug", pruneDebug),
	)

	log.D(cmdL, "Pruning stale data...")
	//stats, err := models.PruneStaleData(
	//ConfigFile,
	//debug,
	//l,
	//)
	if err != nil {
		return nil, err
	}
	log.D(cmdL, "Stale data pruned successfully.")
	return nil, nil
}

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prunes khan's old data",
	Long: `this command prunes old data from Khan's Data store. It is VERY advisable
to run this command frequently as it is idempotent.`,
	Run: func(cmd *cobra.Command, args []string) {
		stats, err := pruneStaleData()
		if err != nil {
			fmt.Printf("Stale data pruning failed with: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Stale data pruned successfully:\n%s.\n", stats.GetStats())
	},
}

func init() {
	RootCmd.AddCommand(pruneCmd)

	pruneCmd.Flags().BoolVarP(&pruneDebug, "debug", "d", false, "Debug mode")
	pruneCmd.Flags().BoolVarP(&pruneQuiet, "quiet", "q", false, "Quiet mode (log level error)")
}
