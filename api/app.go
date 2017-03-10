// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/gorp.v1"

	"github.com/getsentry/raven-go"
	"github.com/jrallison/go-workers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rcrowley/go-metrics"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/khan/es"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/queues"
	"github.com/topfreegames/khan/util"
	"github.com/uber-go/zap"
)

// App is a struct that represents a Khan API Application
type App struct {
	ID             string
	Debug          bool
	Background     bool
	Port           int
	Host           string
	ConfigPath     string
	Errors         metrics.EWMA
	App            *echo.Echo
	Engine         engine.Server
	Db             models.DB
	Config         *viper.Viper
	Dispatcher     *Dispatcher
	ESWorker       *models.ESWorker
	Logger         zap.Logger
	ESClient       *es.Client
	ReadBufferSize int
	Fast           bool
	NewRelic       newrelic.Application
}

// GetApp returns a new Khan API Application
func GetApp(host string, port int, configPath string, debug bool, logger zap.Logger, fast bool) *App {
	app := &App{
		ID:             "default",
		Fast:           fast,
		Host:           host,
		Port:           port,
		ConfigPath:     configPath,
		Config:         viper.New(),
		Debug:          debug,
		Logger:         logger,
		ReadBufferSize: 30000,
	}

	app.Configure()
	return app
}

// Configure instantiates the required dependencies for Khan Api Application
func (app *App) Configure() {
	app.setConfigurationDefaults()
	app.loadConfiguration()
	app.configureSentry()
	app.configureNewRelic()
	app.connectDatabase()
	app.configureApplication()
	app.configureElasticsearch()
	app.initDispatcher()
	app.initESWorker()
	app.configureGoWorkers()
}

func (app *App) configureSentry() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := app.Config.GetString("sentry.url")
	log.I(l, fmt.Sprintf("Configuring sentry with URL %s", sentryURL))
	raven.SetDSN(sentryURL)
	raven.SetRelease(util.VERSION)
}

func (app *App) configureNewRelic() error {
	newRelicKey := app.Config.GetString("newrelic.key")
	appName := app.Config.GetString("newrelic.appName")
	if appName == "" {
		appName = "Khan"
	}

	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("appName", appName),
		zap.String("operation", "configureNewRelic"),
	)

	config := newrelic.NewConfig(appName, newRelicKey)
	if newRelicKey == "" {
		l.Info("New Relic is not enabled..")
		config.Enabled = false
	}
	nr, err := newrelic.NewApplication(config)
	if err != nil {
		l.Error("Failed to initialize New Relic.", zap.Error(err))
		return err
	}

	app.NewRelic = nr
	l.Info("Initialized New Relic successfully.")

	return nil
}

func (app *App) configureElasticsearch() {
	if app.Config.GetBool("elasticsearch.enabled") == true {
		app.ESClient = es.GetClient(
			app.Config.GetString("elasticsearch.host"),
			app.Config.GetInt("elasticsearch.port"),
			app.Config.GetString("elasticsearch.index"),
			app.Config.GetBool("elasticsearch.sniff"),
			app.Logger,
			app.Debug,
			app.NewRelic,
		)
	}
}

func (app *App) setConfigurationDefaults() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "setConfigurationDefaults"),
	)
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("postgres.host", "localhost")
	app.Config.SetDefault("postgres.user", "khan")
	app.Config.SetDefault("postgres.dbName", "khan")
	app.Config.SetDefault("postgres.port", 5432)
	app.Config.SetDefault("postgres.sslMode", "disable")
	app.Config.SetDefault("webhooks.timeout", 2)
	app.Config.SetDefault("elasticsearch.host", "localhost")
	app.Config.SetDefault("elasticsearch.port", 9234)
	app.Config.SetDefault("elasticsearch.sniff", true)
	app.Config.SetDefault("elasticsearch.index", "khan")
	app.Config.SetDefault("elasticsearch.enabled", false)
	app.Config.SetDefault("khan.maxPendingInvites", -1)
	app.Config.SetDefault("khan.defaultCooldownBeforeInvite", -1)
	app.Config.SetDefault("khan.defaultCooldownBeforeApply", -1)
	log.D(l, "Configuration defaults set.")
}

func (app *App) loadConfiguration() {
	l := app.Logger.With(
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

	log.D(l, "Loading configuration file...")
	if err := app.Config.ReadInConfig(); err == nil {
		log.I(l, "Loaded config file successfully.")
	} else {
		log.P(l, "Config file failed to load.")
	}
}

func (app *App) connectDatabase() {
	host := app.Config.GetString("postgres.host")
	user := app.Config.GetString("postgres.user")
	dbName := app.Config.GetString("postgres.dbname")
	password := app.Config.GetString("postgres.password")
	port := app.Config.GetInt("postgres.port")
	sslMode := app.Config.GetString("postgres.sslMode")

	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "connectDatabase"),
		zap.String("host", host),
		zap.String("user", user),
		zap.String("dbName", dbName),
		zap.Int("port", port),
		zap.String("sslMode", sslMode),
	)

	log.D(l, "Connecting to database...")

	db, err := models.GetDB(host, user, port, sslMode, dbName, password)

	if err != nil {
		log.P(l, "Could not connect to postgres...", func(cm log.CM) {
			cm.Write(zap.String("error", err.Error()))
		})
	}

	_, err = db.SelectInt("select count(*) from games")
	if err != nil {
		log.P(l, "Could not connect to postgres...", func(cm log.CM) {
			cm.Write(zap.String("error", err.Error()))
		})
	}

	log.I(l, "Connected to database successfully.")
	app.Db = db
}

func (app *App) onErrorHandler(err error, stack []byte) {
	log.E(app.Logger, "Panic occurred.", func(cm log.CM) {
		cm.Write(
			zap.String("source", "app"),
			zap.String("panicText", err.Error()),
			zap.String("stack", string(stack)),
		)
	})
	tags := map[string]string{
		"source": "app",
		"type":   "panic",
	}
	raven.CaptureError(err, tags)
}

func (app *App) configureApplication() {
	app.Engine = standard.New(fmt.Sprintf("%s:%d", app.Host, app.Port))
	if app.Fast {
		engine := fasthttp.New(fmt.Sprintf("%s:%d", app.Host, app.Port))
		engine.ReadBufferSize = app.ReadBufferSize
		app.Engine = engine
	}
	app.App = echo.New()
	a := app.App

	_, w, _ := os.Pipe()
	a.SetLogOutput(w)

	basicAuthUser := app.Config.GetString("basicauth.username")
	if basicAuthUser != "" {
		basicAuthPass := app.Config.GetString("basicauth.password")

		a.Use(middleware.BasicAuth(func(username, password string) bool {
			return username == basicAuthUser && password == basicAuthPass
		}))
	}

	//NewRelicMiddleware has to stand out from all others
	a.Use(NewNewRelicMiddleware(app, app.Logger).Serve)

	a.Use(NewRecoveryMiddleware(app.onErrorHandler).Serve)
	a.Use(NewVersionMiddleware().Serve)
	a.Use(NewSentryMiddleware(app).Serve)
	a.Use(NewLoggerMiddleware(app.Logger).Serve)
	a.Use(NewBodyExtractionMiddleware().Serve)

	a.Get("/healthcheck", HealthCheckHandler(app))
	a.Get("/status", StatusHandler(app))

	// Game Routes
	a.Post("/games", CreateGameHandler(app))
	a.Put("/games/:gameID", UpdateGameHandler(app))

	// Hook Routes
	a.Post("/games/:gameID/hooks", CreateHookHandler(app))
	a.Delete("/games/:gameID/hooks/:publicID", RemoveHookHandler(app))

	// Player Routes
	a.Post("/games/:gameID/players", CreatePlayerHandler(app))
	a.Put("/games/:gameID/players/:playerPublicID", UpdatePlayerHandler(app))
	a.Get("/games/:gameID/players/:playerPublicID", RetrievePlayerHandler(app))

	// Clan Routes
	a.Get("/games/:gameID/clan-search", SearchClansHandler(app))
	a.Get("/games/:gameID/clans", ListClansHandler(app))
	a.Post("/games/:gameID/clans", CreateClanHandler(app))
	a.Get("/games/:gameID/clans-summary", RetrieveClansSummariesHandler(app))
	a.Get("/games/:gameID/clans/:clanPublicID", RetrieveClanHandler(app))
	a.Get("/games/:gameID/clans/:clanPublicID/summary", RetrieveClanSummaryHandler(app))
	a.Put("/games/:gameID/clans/:clanPublicID", UpdateClanHandler(app))
	a.Post("/games/:gameID/clans/:clanPublicID/leave", LeaveClanHandler(app))
	a.Post("/games/:gameID/clans/:clanPublicID/transfer-ownership", TransferOwnershipHandler(app))

	//// Membership Routes
	a.Post("/games/:gameID/clans/:clanPublicID/memberships/application", ApplyForMembershipHandler(app))
	a.Post("/games/:gameID/clans/:clanPublicID/memberships/application/:action", ApproveOrDenyMembershipApplicationHandler(app))
	a.Post("/games/:gameID/clans/:clanPublicID/memberships/invitation", InviteForMembershipHandler(app))
	a.Post("/games/:gameID/clans/:clanPublicID/memberships/invitation/:action", ApproveOrDenyMembershipInvitationHandler(app))
	a.Post("/games/:gameID/clans/:clanPublicID/memberships/delete", DeleteMembershipHandler(app))
	a.Post("/games/:gameID/clans/:clanPublicID/memberships/promote", PromoteOrDemoteMembershipHandler(app, "promote"))
	a.Post("/games/:gameID/clans/:clanPublicID/memberships/demote", PromoteOrDemoteMembershipHandler(app, "demote"))

	app.Errors = metrics.NewEWMA15()

	go func() {
		app.Errors.Tick()
		time.Sleep(5 * time.Second)
	}()
}

func (app *App) addError() {
	app.Errors.Update(1)
}

//GetHooks returns all available hooks
func (app *App) GetHooks() map[string]map[int][]*models.Hook {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "GetHooks"),
	)

	start := time.Now()
	log.D(l, "Retrieving hooks...")
	dbHooks, err := models.GetAllHooks(app.Db)
	if err != nil {
		log.E(l, "Retrieve hooks failed.", func(cm log.CM) {
			cm.Write(zap.String("error", err.Error()))
		})
		return nil
	}
	log.I(l, "Hooks retrieved successfully.", func(cm log.CM) {
		cm.Write(zap.Duration("hookRetrievalDuration", time.Now().Sub(start)))
	})

	hooks := make(map[string]map[int][]*models.Hook)
	for _, hook := range dbHooks {
		if hooks[hook.GameID] == nil {
			hooks[hook.GameID] = make(map[int][]*models.Hook)
		}
		hooks[hook.GameID][hook.EventType] = append(
			hooks[hook.GameID][hook.EventType],
			hook,
		)
	}

	return hooks
}

//GetGame returns a game by Public ID
func (app *App) GetGame(gameID string) (*models.Game, error) {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "GetGame"),
		zap.String("gameID", gameID),
	)

	start := time.Now()
	log.D(l, "Retrieving game...")

	game, err := models.GetGameByPublicID(app.Db, gameID)
	if err != nil {
		log.E(l, "Retrieve game failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, err
	}

	log.I(l, "Game retrieved succesfully.", func(cm log.CM) {
		cm.Write(zap.Duration("gameRetrievalDuration", time.Now().Sub(start)))
	})
	return game, nil
}

func (app *App) configureGoWorkers() {
	redisHost := app.Config.GetString("redis.host")
	redisPort := app.Config.GetInt("redis.port")
	redisDatabase := app.Config.GetInt("redis.database")
	redisPool := app.Config.GetInt("redis.pool")
	workerCount := app.Config.GetInt("webhooks.workers")
	if redisPool == 0 {
		redisPool = 30
	}

	if workerCount == 0 {
		workerCount = 5
	}

	l := app.Logger.With(
		zap.String("source", "dispatcher"),
		zap.String("operation", "Configure"),
		zap.Int("workerCount", workerCount),
		zap.String("redisHost", redisHost),
		zap.Int("redisPort", redisPort),
		zap.Int("redisDatabase", redisDatabase),
		zap.Int("redisPool", redisPool),
	)

	opts := map[string]string{
		// location of redis instance
		"server": fmt.Sprintf("%s:%d", redisHost, redisPort),
		// instance of the database
		"database": strconv.Itoa(redisDatabase),
		// number of connections to keep open with redis
		"pool": strconv.Itoa(redisPool),
		// unique process id
		"process": uuid.NewV4().String(),
	}
	redisPass := app.Config.GetString("redis.password")
	if redisPass != "" {
		opts["password"] = redisPass
	}
	l.Debug("Configuring workers...")
	workers.Configure(opts)

	workers.Process(queues.KhanQueue, app.Dispatcher.PerformDispatchHook, workerCount)
	workers.Process(queues.KhanESQueue, app.ESWorker.PerformUpdateES, workerCount)
	l.Info("Worker configured.")
}

//StartWorkers "starts" the dispatcher
func (app *App) StartWorkers() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "StartWorkers"),
	)

	log.D(l, "Starting workers...")
	if app.Config.GetBool("webhooks.runStats") {
		jobsStatsPort := app.Config.GetInt("webhooks.statsPort")
		go workers.StatsServer(jobsStatsPort)
	}

	workers.Run()
}

//NonblockingStartWorkers non-blocking
func (app *App) NonblockingStartWorkers() {
	workers.Start()
}

func (app *App) initESWorker() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "initESWorker"),
	)

	log.D(l, "Initializing es worker...")
	esWorker := models.NewESWorker(app.Logger)
	log.I(l, "ES Worker initialized successfully")
	app.ESWorker = esWorker
}

func (app *App) initDispatcher() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "initDispatcher"),
	)

	log.D(l, "Initializing dispatcher...")

	disp, err := NewDispatcher(app)
	if err != nil {
		log.P(l, "Dispatcher failed to initialize.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return
	}
	log.I(l, "Dispatcher initialized successfully")

	app.Dispatcher = disp
}

// DispatchHooks dispatches web hooks for a specific game and event type
func (app *App) DispatchHooks(gameID string, eventType int, payload map[string]interface{}) error {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "DispatchHooks"),
		zap.String("gameID", gameID),
		zap.Int("eventType", eventType),
	)

	start := time.Now()
	log.D(l, "Dispatching hook...")
	app.Dispatcher.DispatchHook(gameID, eventType, payload)
	log.I(l, "Hook dispatched successfully.", func(cm log.CM) {
		cm.Write(zap.Duration("hookDispatchDuration", time.Now().Sub(start)))
	})
	return nil
}

func (app *App) finalizeApp() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "finalizeApp"),
	)

	log.D(l, "Closing DB connection...")
	app.Db.(*gorp.DbMap).Db.Close()
	log.I(l, "DB connection closed succesfully.")
}

//BeginTrans in the current Db connection
func (app *App) BeginTrans(l zap.Logger) (*gorp.Transaction, error) {
	log.D(l, "Beginning DB tx...")
	tx, err := (app.Db).(*gorp.DbMap).Begin()
	if err != nil {
		log.E(l, "Failed to begin tx.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, err
	}
	log.D(l, "Tx begun successfuly.")
	return tx, nil
}

//Rollback transaction
func (app *App) Rollback(tx *gorp.Transaction, msg string, c echo.Context, l zap.Logger, err error) error {
	sErr := WithSegment("tx-rollback", c, func() error {
		txErr := tx.Rollback()
		if txErr != nil {
			log.E(l, fmt.Sprintf("%s and failed to rollback transaction.", msg), func(cm log.CM) {
				cm.Write(zap.Error(txErr), zap.String("originalError", err.Error()))
			})
			return txErr
		}
		return nil
	})
	return sErr
}

//Commit transaction
func (app *App) Commit(tx *gorp.Transaction, msg string, c echo.Context, l zap.Logger) error {
	err := WithSegment("tx-commit", c, func() error {
		txErr := tx.Commit()
		if txErr != nil {
			log.E(l, fmt.Sprintf("%s failed to commit transaction.", msg), func(cm log.CM) {
				cm.Write(zap.Error(txErr))
			})
			return txErr
		}

		return nil
	})
	return err
}

// GetCtxDB returns the proper database connection depending on the request context
func (app *App) GetCtxDB(ctx echo.Context) (models.DB, error) {
	val := ctx.Get("db")
	if val != nil {
		return val.(models.DB), nil
	}

	return app.Db, nil
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "Start"),
	)

	defer app.finalizeApp()
	log.D(l, "App started.", func(cm log.CM) {
		cm.Write(zap.String("host", app.Host), zap.Int("port", app.Port))
	})

	if app.Background {
		go func() {
			app.App.Run(app.Engine)
		}()
	} else {
		app.App.Run(app.Engine)
	}
}
