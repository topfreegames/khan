// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/gorp"
	"github.com/topfreegames/goose/lib/goose"
	"github.com/topfreegames/khan/db"
	"github.com/topfreegames/khan/models"
)

var migrationVersion int64

func createTempDbDir() (string, error) {
	dir, err := ioutil.TempDir("", "migrations")
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	fmt.Printf("Created temporary directory %s.\n", dir)
	assetNames := db.AssetNames()
	for _, assetName := range assetNames {
		asset, err := db.Asset(assetName)
		if err != nil {
			return "", err
		}
		fileName := strings.SplitN(assetName, "/", 2)[1] // remove migrations folder from fileName
		err = ioutil.WriteFile(filepath.Join(dir, fileName), asset, 0777)
		if err != nil {
			return "", err
		}
		fmt.Printf("Wrote migration file %s.\n", fileName)
	}
	return dir, nil
}

func getDatabase() (*gorp.Database, error) {
	viper.SetEnvPrefix("khan")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	host := viper.GetString("postgres.host")
	user := viper.GetString("postgres.user")
	dbName := viper.GetString("postgres.dbname")
	password := viper.GetString("postgres.password")
	port := viper.GetInt("postgres.port")
	sslMode := viper.GetString("postgres.sslMode")

	fmt.Printf(
		"\nConnecting to %s:%d as %s using sslMode=%s to db %s...\n\n",
		host, port, user, sslMode, dbName,
	)
	db, err := models.GetDB(host, user, port, sslMode, dbName, password)
	return db.(*gorp.Database), err
}

func getGooseConf() *goose.DBConf {
	migrationsDir, err := createTempDbDir()

	if err != nil {
		panic("Could not create migration files...")
	}

	return &goose.DBConf{
		MigrationsDir: migrationsDir,
		Env:           "production",
		Driver: goose.DBDriver{
			Name:    "postgres",
			OpenStr: "",
			Dialect: &goose.PostgresDialect{},
		},
	}
}

// MigrationError identified rigrations running error
type MigrationError struct {
	Message string
}

func (err *MigrationError) Error() string {
	return fmt.Sprintf("Could not run migrations: %s", err.Message)
}

//RunMigrations in selected DB
func RunMigrations(migrationVersion int64) error {
	conf := getGooseConf()
	defer os.RemoveAll(conf.MigrationsDir)
	db, err := getDatabase()
	if err != nil {
		return &MigrationError{fmt.Sprintf("could not connect to database: %s", err.Error())}
	}

	targetVersion := migrationVersion
	if targetVersion == -1 {
		// Get the latest possible migration
		latest, err := goose.GetMostRecentDBVersion(conf.MigrationsDir)
		if err != nil {
			return &MigrationError{fmt.Sprintf("could not get migrations at %s: %s", conf.MigrationsDir, err.Error())}
		}
		targetVersion = latest
	}

	// Migrate up to the latest version
	err = goose.RunMigrationsOnDb(conf, conf.MigrationsDir, targetVersion, db.Inner().Db)
	if err != nil {
		return &MigrationError{fmt.Sprintf("could not run migrations to %d: %s", targetVersion, err.Error())}
	}
	fmt.Printf("Migrated database successfully to version %d.\n", targetVersion)
	return nil
}

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrates the database up or down",
	Long:  `Migrate the database specified in the configuration file to the given version (or latest if none provided)`,
	Run: func(cmd *cobra.Command, args []string) {
		InitConfig()
		err := RunMigrations(migrationVersion)
		if err != nil {
			panic(err.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().Int64VarP(&migrationVersion, "target", "t", -1, "Version to run up to or down to")
}
