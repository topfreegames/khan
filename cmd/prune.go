// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

var pruneDebug bool
var pruneQuiet bool

//PruneStaleData prunes old data from the DB
func PruneStaleData(debug, quiet bool) (*models.PruneStats, error) {
	return executePruning(debug, quiet)
}

func executePruning(debug, quiet bool) (*models.PruneStats, error) {
	InitConfig()
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

	host := viper.GetString("postgres.host")
	user := viper.GetString("postgres.user")
	dbName := viper.GetString("postgres.dbname")
	password := viper.GetString("postgres.password")
	port := viper.GetInt("postgres.port")
	sslMode := viper.GetString("postgres.sslMode")

	db, err := models.GetDB(host, user, port, sslMode, dbName, password)
	if err != nil {
		log.E(cmdL, "Failed to connect to DB.", func(cm log.CM) {
			cm.Write(
				zap.Error(err),
				zap.String("host", host),
				zap.String("user", user),
				zap.Int("port", port),
				zap.String("sslMode", sslMode),
				zap.String("dbName", dbName),
			)
		})

		return nil, err
	}

	log.D(cmdL, "Loading games...")
	games, err := models.GetAllGames(db)
	if err != nil {
		log.E(cmdL, "Failed to load games.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, err
	}

	totals := &models.PruneStats{}

	for _, game := range games {
		log.D(cmdL, "Processing stale data for game.", func(cm log.CM) {
			cm.Write(zap.String("GameID", game.PublicID))
		})
		log.D(cmdL, "Pruning stale data...")

		pendingApplicationsExpiration := game.Metadata["pendingApplicationsExpiration"]
		pendingInvitesExpiration := game.Metadata["pendingInvitesExpiration"]
		deniedMembershipsExpiration := game.Metadata["deniedMembershipsExpiration"]
		deletedMembershipsExpiration := game.Metadata["deletedMembershipsExpiration"]

		if pendingApplicationsExpiration == nil ||
			pendingInvitesExpiration == nil ||
			deniedMembershipsExpiration == nil ||
			deletedMembershipsExpiration == nil {
			log.W(cmdL, "Game does not have stale expiration configuration.", func(cm log.CM) {
				cm.Write(
					zap.String("GameID", game.PublicID),
					zap.Object("pendingApplicationsExpiration", pendingApplicationsExpiration),
					zap.Object("pendingInvitesExpiration", pendingInvitesExpiration),
					zap.Object("deniedMembershipsExpiration", deniedMembershipsExpiration),
					zap.Object("deletedMembershipsExpiration", deletedMembershipsExpiration),
				)
			})
			continue
		}

		options := &models.PruneOptions{
			GameID:                        game.PublicID,
			PendingApplicationsExpiration: int(pendingApplicationsExpiration.(float64)),
			PendingInvitesExpiration:      int(pendingInvitesExpiration.(float64)),
			DeniedMembershipsExpiration:   int(deniedMembershipsExpiration.(float64)),
			DeletedMembershipsExpiration:  int(deletedMembershipsExpiration.(float64)),
		}

		stats, err := models.PruneStaleData(
			options,
			db,
			l,
		)
		if err != nil {
			log.E(cmdL, "Failed to prune stale data for game.", func(cm log.CM) {
				cm.Write(zap.Error(err), zap.String("gameID", game.PublicID))
			})

			return nil, err
		}
		log.I(cmdL, "Stale data for game pruned successfully.", func(cm log.CM) {
			cm.Write(
				zap.Int("PendingApplicationsPruned", stats.PendingApplicationsPruned),
				zap.Int("PendingInvitesPruned", stats.PendingInvitesPruned),
				zap.Int("DeniedMembershipsPruned", stats.DeniedMembershipsPruned),
				zap.Int("DeletedMembershipsPruned", stats.DeletedMembershipsPruned),
				zap.String("GameID", game.PublicID),
			)
		})

		totals.PendingApplicationsPruned += stats.PendingApplicationsPruned
		totals.PendingInvitesPruned += stats.PendingInvitesPruned
		totals.DeletedMembershipsPruned += stats.DeletedMembershipsPruned
		totals.DeniedMembershipsPruned += stats.DeniedMembershipsPruned
	}
	log.I(cmdL, "Stale data pruned successfully.", func(cm log.CM) {
		cm.Write(
			zap.Int("PendingApplicationsPruned", totals.PendingApplicationsPruned),
			zap.Int("PendingInvitesPruned", totals.PendingInvitesPruned),
			zap.Int("DeniedMembershipsPruned", totals.DeniedMembershipsPruned),
			zap.Int("DeletedMembershipsPruned", totals.DeletedMembershipsPruned),
		)
	})
	return totals, nil
}

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prunes khan's old data",
	Long: `This command prunes old data from Khan's Data store.

It is VERY advisable to run this command frequently as it is idempotent.

*WARNING*:
	This command deletes data from Khan's database and the data CANNOT be recovered.
	Please ensure that you have frequent backups before running this command continuously.
`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := PruneStaleData(pruneDebug, pruneQuiet)
		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(pruneCmd)

	pruneCmd.Flags().BoolVarP(&pruneDebug, "debug", "d", false, "Debug mode")
	pruneCmd.Flags().BoolVarP(&pruneQuiet, "quiet", "q", false, "Quiet mode (log level error)")
}
