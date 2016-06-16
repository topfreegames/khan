// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"fmt"
	"path"
	"runtime"

	"gopkg.in/gorp.v1"

	"bitbucket.org/liamstask/goose/lib/goose"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/topfreegames/khan/models"
)

var migrateDown bool
var migrationVersion int64

func getDatabase() (*gorp.DbMap, error) {
	host := viper.GetString("postgres.host")
	user := viper.GetString("postgres.user")
	dbName := viper.GetString("postgres.dbname")
	password := viper.GetString("postgres.password")
	port := viper.GetInt("postgres.port")
	sslMode := viper.GetString("postgres.sslMode")

	db, err := models.GetDB(host, user, port, sslMode, dbName, password)
	return db.(*gorp.DbMap), err
}

func getGooseConf() *goose.DBConf {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("Could not find configuration file...")
		return nil
	}
	migrationsPath := path.Join(path.Dir(filename), "../db/migrations/")

	return &goose.DBConf{
		MigrationsDir: migrationsPath,
		Env:           "production",
		Driver: goose.DBDriver{
			Name:    "postgres",
			OpenStr: "",
			Import:  "github.com/jinzhu/gorm/dialects/postgres",
			Dialect: &goose.PostgresDialect{},
		},
	}
}

func runMigrations(migrateDown bool, migrationVersion int64) {
	conf := getGooseConf()
	db, err := getDatabase()
	if err != nil {
		panic(fmt.Sprintf("Could not connect to database: %s", err.Error()))
		return
	}

	targetVersion := migrationVersion
	if targetVersion == -1 {
		// Get the latest possible migration
		latest, err := goose.GetMostRecentDBVersion(conf.MigrationsDir)
		if err != nil {
			panic(fmt.Sprintf("Could not get migrations at %s: %s", conf.MigrationsDir, err.Error()))
			return
		}
		targetVersion = latest
	}

	// Migrate up to the latest version
	err = goose.RunMigrationsOnDb(conf, conf.MigrationsDir, targetVersion, db.Db)
	if err != nil {
		panic(fmt.Sprintf("Could not run migrations to %d: %s", targetVersion, err.Error()))
		return
	}
	fmt.Printf("Migrated database successfully to version %d.\n", targetVersion)
}

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrates the database up or down",
	Long:  `Migrate the database specified in the configuration file to the given version (or latest if none provided)`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		runMigrations(migrateDown, migrationVersion)
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().Int64VarP(&migrationVersion, "target", "t", -1, "Version to run up to or down to")
}
