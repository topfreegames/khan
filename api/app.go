// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gopkg.in/gorp.v1"

	"github.com/golang/glog"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/spf13/viper"
	"github.com/topfreegames/khan/models"
	"github.com/valyala/fasthttp"
)

// App is a struct that represents a Khan API Application
type App struct {
	Debug      bool
	Port       int
	Host       string
	ConfigPath string
	App        *iris.Framework
	Db         models.DB
	Config     *viper.Viper
	Hooks      map[string]map[int][]*models.Hook
}

// GetApp returns a new Khan API Application
func GetApp(host string, port int, configPath string, debug bool) *App {
	app := &App{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
	}
	app.Configure()
	return app
}

// Configure instantiates the required dependencies for Khan Api Application
func (app *App) Configure() {
	app.setConfigurationDefaults()
	app.loadConfiguration()
	app.connectDatabase()
	app.configureApplication()
	app.loadHooks()
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("postgres.host", "localhost")
	app.Config.SetDefault("postgres.user", "khan")
	app.Config.SetDefault("postgres.dbName", "khan")
	app.Config.SetDefault("postgres.port", 5432)
	app.Config.SetDefault("postgres.sslMode", "disable")
	app.Config.SetDefault("webhooks.timeout", 2)
}

func (app *App) loadConfiguration() {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("khan")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		glog.Infof("Using config file: %s", app.Config.ConfigFileUsed())
	} else {
		panic(fmt.Sprintf("Could not load configuration file from: %s", app.ConfigPath))
	}
}

func (app *App) connectDatabase() {
	host := app.Config.GetString("postgres.host")
	user := app.Config.GetString("postgres.user")
	dbName := app.Config.GetString("postgres.dbname")
	password := app.Config.GetString("postgres.password")
	port := app.Config.GetInt("postgres.port")
	sslMode := app.Config.GetString("postgres.sslMode")

	db, err := models.GetDB(host, user, port, sslMode, dbName, password)

	if err != nil {
		glog.Errorf(
			"Could not connect to Postgres at %s:%d with user %s and db %s with password %s (%s)\n",
			host, port, user, dbName, password, err,
		)
		panic(err)
	}
	app.Db = db
}

func (app *App) configureApplication() {
	app.App = iris.New()
	a := app.App

	if app.Debug {
		a.Use(logger.New(iris.Logger))
	}
	// a.Use(recovery.New(os.Stderr))
	a.Use(&TransactionMiddleware{App: app})

	a.Get("/healthcheck", HealthCheckHandler(app))

	// Game Routes
	a.Post("/games", CreateGameHandler(app))
	a.Put("/games/:gameID", UpdateGameHandler(app))

	// Hook Routes
	a.Post("/games/:gameID/hooks", CreateHookHandler(app))
	a.Delete("/games/:gameID/hooks/:publicID", RemoveHookHandler(app))

	// Player Routes
	a.Post("/games/:gameID/players", CreatePlayerHandler(app))
	a.Put("/games/:gameID/players/:playerPublicID", UpdatePlayerHandler(app))

	// Clan Routes
	a.Get("/games/:gameID/clan-search", SearchClansHandler(app))
	a.Get("/games/:gameID/clans", ListClansHandler(app))
	a.Post("/games/:gameID/clans", CreateClanHandler(app))
	a.Get("/games/:gameID/clans/:clanPublicID", RetrieveClanHandler(app))
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
}

func (app *App) loadHooks() {
	app.Hooks = make(map[string]map[int][]*models.Hook)

	go (func(a *App) {
		hooks, err := models.GetAllHooks(a.Db)
		if err != nil {
			glog.Fatalf(
				"Failed to retrieve hooks: %s", err.Error(),
			)
		}
		for _, hook := range hooks {
			if app.Hooks[hook.GameID] == nil {
				app.Hooks[hook.GameID] = make(map[int][]*models.Hook)
			}
			app.Hooks[hook.GameID][hook.EventType] = append(
				app.Hooks[hook.GameID][hook.EventType],
				hook,
			)
		}

		time.Sleep(time.Minute)
	})(app)
}

//HookNotFoundError means that hooks for the specified game and event type were not found
type HookNotFoundError struct {
	GameID    string
	EventType int
}

func (hook *HookNotFoundError) Error() string {
	return fmt.Sprintf("Hook not found for game %s with event type %d...", hook.GameID, hook.EventType)
}

// DispatchHooks dispatches web hooks for a specific game and event type
func (app *App) DispatchHooks(gameID string, eventType int, payload map[string]interface{}) error {
	if _, ok := app.Hooks[gameID]; !ok {
		return &HookNotFoundError{GameID: gameID, EventType: eventType}
	}
	if _, ok := app.Hooks[gameID][eventType]; !ok {
		return &HookNotFoundError{GameID: gameID, EventType: eventType}
	}

	timeout := time.Duration(app.Config.GetInt("webhooks.timeout")) * time.Second
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var errors []error
	for _, hook := range app.Hooks[gameID][eventType] {
		glog.Infof("Sending webhook to %s", hook.URL)

		client := fasthttp.Client{
			Name: fmt.Sprintf("khan-%s", VERSION),
		}

		req := fasthttp.AcquireRequest()
		req.Header.SetMethod("POST")
		req.SetRequestURI(hook.URL)
		req.AppendBody(payloadJSON)
		resp := fasthttp.AcquireResponse()
		err := client.DoTimeout(req, resp, timeout)
		if err != nil {
			glog.Error(fmt.Sprintf("Could not request webhook %s: %s", hook.URL, err.Error()))
			errors = append(errors, err)
			continue
		}
		if resp.StatusCode() > 399 {
			glog.Error(fmt.Sprintf(
				"Could not request webhook %s (status code: %d): %s",
				hook.URL,
				resp.StatusCode(),
				resp.Body(),
			))
			continue
		}
	}

	return nil
}

func (app *App) finalizeApp() {
	app.Db.(*gorp.DbMap).Db.Close()
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() {
	defer app.finalizeApp()
	app.App.Listen(fmt.Sprintf("%s:%d", app.Host, app.Port))
}
