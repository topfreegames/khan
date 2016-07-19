// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/gorp.v1"

	"github.com/getsentry/raven-go"
	"github.com/kataras/iris"
	"github.com/kataras/iris/config"
	"github.com/rcrowley/go-metrics"
	"github.com/spf13/viper"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

// App is a struct that represents a Khan API Application
type App struct {
	Debug      bool
	Port       int
	Host       string
	ConfigPath string
	Errors     metrics.EWMA
	App        *iris.Framework
	Db         models.DB
	Config     *viper.Viper
	Dispatcher *Dispatcher
	Logger     zap.Logger
}

// GetApp returns a new Khan API Application
func GetApp(host string, port int, configPath string, debug bool, logger zap.Logger) *App {
	app := &App{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
		Logger:     logger,
	}

	app.Configure()
	return app
}

// Configure instantiates the required dependencies for Khan Api Application
func (app *App) Configure() {
	app.setConfigurationDefaults()
	app.loadConfiguration()
	app.configureSentry()
	app.connectDatabase()
	app.configureApplication()
	app.initDispatcher()
}

func (app *App) configureSentry() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := app.Config.GetString("sentry.url")
	l.Info(fmt.Sprintf("Configuring sentry with URL %s", sentryURL))
	raven.SetDSN(sentryURL)
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
	l.Debug("Configuration defaults set.")
}

func (app *App) loadConfiguration() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "loadConfiguration"),
		zap.String("configPath", app.ConfigPath),
	)

	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("khan")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	l.Debug("Loading configuration file...")
	if err := app.Config.ReadInConfig(); err == nil {
		l.Info("Loaded config file successfully.")
	} else {
		l.Panic("Config file failed to load.")
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

	l.Debug("Connecting to database...")
	db, err := models.GetDB(host, user, port, sslMode, dbName, password)

	if err != nil {
		l.Panic(
			"Could not connect to postgres...",
			zap.String("error", err.Error()),
		)
	}

	_, err = db.SelectInt("select count(*) from games")
	if err != nil {
		l.Panic(
			"Could not connect to postgres...",
			zap.String("error", err.Error()),
		)
	}

	l.Info("Connected to database successfully.")
	app.Db = db
}

func (app *App) onErrorHandler(err error, stack []byte) {
	app.Logger.Error(
		"Panic occurred.",
		zap.String("source", "app"),
		zap.String("panicText", err.Error()),
		zap.String("stack", string(stack)),
	)
	tags := map[string]string{
		"source": "app",
		"type":   "panic",
	}
	raven.CaptureError(err, tags)
}

func (app *App) configureApplication() {
	c := config.Iris{
		DisableBanner: true,
	}

	app.App = iris.New(c)
	a := app.App

	a.Use(NewLoggerMiddleware(app.Logger))
	a.Use(&RecoveryMiddleware{OnError: app.onErrorHandler})
	a.Use(&TransactionMiddleware{App: app})
	a.Use(&VersionMiddleware{App: app})
	a.Use(&SentryMiddleware{App: app})

	a.OnError(iris.StatusInternalServerError, func(ctx *iris.Context) {
		app.Logger.Error(
			"Internal server error happened.",
			zap.String("error", string(ctx.Response.Body())),
			zap.String("source", "app"),
		)
		ctx.Write("INTERNAL SERVER ERROR")
	})

	a.OnError(iris.StatusNotFound, func(ctx *iris.Context) {
		app.Logger.Warn(
			"Route not found.", zap.String("url", ctx.Request.URI().String()),
			zap.String("source", "app"),
		)
		ctx.Write("Not Found")
	})

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
	a.Get("/games/:gameID/clans/:clanPublicID", RetrieveClanHandler(app))
	a.Get("/games/:gameID/clans/:clanPublicID/summary", RetrieveClanSummaryHandler(app))
	a.Put("/games/:gameID/clans/:clanPublicID", UpdateClanHandler(app))
	a.Post("/games/:gameID/clans/:clanPublicID/leave", LeaveClanHandler(app))
	a.Post("/games/:gameID/clans/:clanPublicID/transfer-ownership", TransferOwnershipHandler(app))

	// Membership Routes
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
	l.Debug("Retrieving hooks...")
	dbHooks, err := models.GetAllHooks(app.Db)
	if err != nil {
		l.Error(
			"Retrieve hooks failed.",
			zap.String("error", err.Error()),
		)
		return nil
	}
	l.Info("Hooks retrieved successfully.", zap.Duration("hookRetrievalDuration", time.Now().Sub(start)))

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
	l.Debug("Retrieving game...")

	game, err := models.GetGameByPublicID(app.Db, gameID)
	if err != nil {
		l.Error(
			"Retrieve game failed.",
			zap.Error(err),
		)
		return nil, err
	}

	l.Info(
		"Game retrieved succesfully.",
		zap.Duration("gameRetrievalDuration", time.Now().Sub(start)),
	)
	return game, nil
}

func (app *App) initDispatcher() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "initDispatcher"),
	)

	l.Debug("Initializing dispatcher...")
	disp, err := NewDispatcher(app, 5, 1000)
	if err != nil {
		l.Panic("Dispatcher failed to initialize.", zap.Error(err))
		return
	}
	l.Info("Dispatcher initialized successfully")

	l.Debug("Starting dispatcher...")
	app.Dispatcher = disp
	app.Dispatcher.Start()
	l.Info("Dispatcher started successfully.")
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
	l.Debug("Dispatching hook...")
	app.Dispatcher.DispatchHook(gameID, eventType, payload)
	l.Info(
		"Hook dispatched successfully.",
		zap.Duration("hookDispatchDuration", time.Now().Sub(start)),
	)
	return nil
}

func (app *App) finalizeApp() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "finalizeApp"),
	)

	l.Debug("Closing DB connection...")
	app.Db.(*gorp.DbMap).Db.Close()
	l.Info("DB connection closed succesfully.")
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "Start"),
	)

	defer app.finalizeApp()
	l.Debug("App started.", zap.String("host", app.Host), zap.Int("port", app.Port))
	app.App.Listen(fmt.Sprintf("%s:%d", app.Host, app.Port))
}
