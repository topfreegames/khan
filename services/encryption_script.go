package services

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
	gorp "github.com/topfreegames/extensions/v9/gorp/interfaces"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

// EncryptionScript is a struct that represents a Khan API Application
type EncryptionScript struct {
	Debug         bool
	ConfigPath    string
	Config        *viper.Viper
	Logger        zap.Logger
	EncryptionKey []byte
	db            gorp.Database
}

// GetEncryptionScript returns a new Khan API Application
func GetEncryptionScript(configPath string, debug bool, logger zap.Logger) *EncryptionScript {
	app := &EncryptionScript{
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
		Logger:     logger,
	}

	app.Configure()
	return app
}

// Configure instantiates the required dependencies for Khan Api Application
func (app *EncryptionScript) Configure() {
	app.setConfigurationDefaults()
	app.loadConfiguration()
	app.connectDatabase()
}

func (app *EncryptionScript) setConfigurationDefaults() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "setConfigurationDefaults"),
	)
	app.Config.SetDefault("graceperiod.ms", 500)
	app.Config.SetDefault("postgres.host", "localhost")
	app.Config.SetDefault("postgres.user", "khan")
	app.Config.SetDefault("postgres.dbName", "khan")
	app.Config.SetDefault("postgres.port", 5432)
	app.Config.SetDefault("postgres.sslMode", "disable")
	app.Config.SetDefault("security.encryptionKey", "00000000000000000000000000000000")
	app.Config.SetDefault("script.tick", "1s")
	app.Config.SetDefault("script.playerAmount", "500")

	log.D(logger, "Configuration defaults set.")
}

func (app *EncryptionScript) loadConfiguration() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "loadConfiguration"),
		zap.String("configPath", app.ConfigPath),
	)

	app.Config.SetConfigType("yaml")
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("khan")
	app.Config.AddConfigPath(".")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	log.D(logger, "Loading configuration file...")
	if err := app.Config.ReadInConfig(); err == nil {
		log.I(logger, "Loaded config file successfully.")
	} else {
		log.P(logger, "Config file failed to load.")
	}

	app.EncryptionKey = []byte(app.Config.GetString("security.encryptionKey"))
}

func (app *EncryptionScript) connectDatabase() {
	host := app.Config.GetString("postgres.host")
	user := app.Config.GetString("postgres.user")
	dbName := app.Config.GetString("postgres.dbname")
	password := app.Config.GetString("postgres.password")
	port := app.Config.GetInt("postgres.port")
	sslMode := app.Config.GetString("postgres.sslMode")

	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "connectDatabase"),
		zap.String("host", host),
		zap.String("user", user),
		zap.String("dbName", dbName),
		zap.Int("port", port),
		zap.String("sslMode", sslMode),
	)

	log.D(logger, "Connecting to database...")

	db, err := models.GetDB(host, user, port, sslMode, dbName, password)

	if err != nil {
		log.P(logger, "Could not connect to postgres...", func(cm log.CM) {
			cm.Write(zap.String("error", err.Error()))
		})
	}

	_, err = db.SelectInt("select count(*) from games")
	if err != nil {
		log.P(logger, "Could not connect to postgres...", func(cm log.CM) {
			cm.Write(zap.String("error", err.Error()))
		})
	}

	log.I(logger, "Connected to database successfully.")
	app.db = db
}

// Start starts listening for web requests at specified host and port
func (app *EncryptionScript) Start() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "Start"),
	)

	sg := make(chan os.Signal)
	stopScript := make(chan bool, 1)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	go app.executeScript(stopScript)

	// stop server
	select {
	case s := <-sg:
		graceperiod := app.Config.GetInt("graceperiod.ms")
		log.I(logger, "shutting down", func(cm log.CM) {
			cm.Write(zap.String("signal", fmt.Sprintf("%v", s)),
				zap.Int("graceperiod", graceperiod))
		})
		stopScript <- true
		time.Sleep(time.Duration(graceperiod) * time.Millisecond)
	}
	log.I(logger, "app stopped")
}

func (app *EncryptionScript) executeScript(stopChan chan bool) {
	ticker := time.NewTicker(app.Config.GetDuration("script.tick"))
	for {
		select {
		case <-stopChan:
			log.I(app.Logger, "Finishing script")
		case <-ticker.C:
			app.encryptPlayers()
		}
	}
}

func (app *EncryptionScript) encryptPlayers() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "executeScript"),
	)

	amount := app.Config.GetInt("script.playerAmount")

	initTime := time.Now()

	players, err := models.GetPlayersToEncrypt(app.db, app.EncryptionKey, amount)
	if err != nil {
		log.E(logger, "error on get players to encrypt", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return
	}

	if len(players) == 0 {
		logger.Warn("FINISHED, there is no player to encrypt")
		return
	}

	err = models.ApplySecurityChanges(app.db, app.EncryptionKey, players)
	if err != nil {
		log.E(logger, "error on update players", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return
	}

	app.Logger.Debug("encryption done", zap.String("spent time", time.Since(initTime).String()))
}
