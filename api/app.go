// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jrallison/go-workers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	opentracing "github.com/opentracing/opentracing-go"
	gocache "github.com/patrickmn/go-cache"
	"github.com/rcrowley/go-metrics"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	eecho "github.com/topfreegames/extensions/v9/echo"
	extechomiddleware "github.com/topfreegames/extensions/v9/echo/middleware"
	gorp "github.com/topfreegames/extensions/v9/gorp/interfaces"
	extnethttpmiddleware "github.com/topfreegames/extensions/v9/middleware"
	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	extworkermiddleware "github.com/topfreegames/extensions/v9/worker/middleware"
	"github.com/topfreegames/khan/caches"
	"github.com/topfreegames/khan/es"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/mongo"
	"github.com/topfreegames/khan/queues"
	"github.com/uber-go/zap"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

// App is a struct that represents a Khan API Application
type App struct {
	ID                  string
	Test                bool
	Debug               bool
	Port                int
	Host                string
	ConfigPath          string
	Errors              metrics.EWMA
	App                 *eecho.Echo
	Engine              engine.Server
	Config              *viper.Viper
	Dispatcher          *Dispatcher
	ESWorker            *models.ESWorker
	MongoWorker         *models.MongoWorker
	Logger              zap.Logger
	ESClient            *es.Client
	MongoDB             interfaces.MongoDB
	ReadBufferSize      int
	Fast                bool
	DDStatsD            *extnethttpmiddleware.DogStatsD
	EncryptionKey       []byte
	getGameCache        *gocache.Cache
	clansSummariesCache *caches.ClansSummaries
	db                  gorp.Database
}

// GetApp returns a new Khan API Application
func GetApp(host string, port int, configPath string, debug bool, logger zap.Logger, fast, test bool) *App {
	app := &App{
		ID:             "default",
		Test:           test,
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
	app.configureStatsD()
	app.configureJaeger()
	app.connectDatabase()
	app.configureApplication()
	app.configureElasticsearch()
	app.configureMongoDB()
	app.initDispatcher()
	app.initESWorker()
	app.initMongoWorker()
	app.configureGoWorkers()
	app.configureCaches()
}

func (app *App) configureCaches() {
	app.configureGetGameCache()
	app.configureClansSummariesCache()
}

func (app *App) configureGetGameCache() {
	// TTL
	ttlKey := "caches.getGame.ttl"
	app.Config.SetDefault(ttlKey, time.Minute)
	ttl := app.Config.GetDuration(ttlKey)
	if ttl <= 0 {
		ttl = time.Minute
	}

	// cleanup
	cleanupIntervalKey := "caches.getGame.cleanupInterval"
	app.Config.SetDefault(cleanupIntervalKey, time.Minute)
	cleanupInterval := app.Config.GetDuration(cleanupIntervalKey)

	app.getGameCache = gocache.New(ttl, cleanupInterval)
}

func (app *App) configureClansSummariesCache() {
	// TTL
	ttlKey := "caches.clansSummaries.ttl"
	app.Config.SetDefault(ttlKey, time.Minute)
	ttl := app.Config.GetDuration(ttlKey)
	if ttl <= 0 {
		ttl = time.Minute
	}

	// cleanup
	cleanupIntervalKey := "caches.clansSummaries.cleanupInterval"
	app.Config.SetDefault(cleanupIntervalKey, time.Minute)
	cleanupInterval := app.Config.GetDuration(cleanupIntervalKey)

	app.clansSummariesCache = &caches.ClansSummaries{
		Cache: gocache.New(ttl, cleanupInterval),
	}
}

func (app *App) configureStatsD() error {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureStatsD"),
	)

	ddstatsd, err := extnethttpmiddleware.NewDogStatsD(app.Config)
	if err != nil {
		log.E(logger, "Failed to initialize DogStatsD.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}
	app.DDStatsD = ddstatsd
	logger.Info("Initialized DogStatsD successfully.")

	return nil
}

func (app *App) configureJaeger() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureJaeger"),
	)

	cfg, err := jaegercfg.FromEnv()
	if cfg.ServiceName == "" {
		logger.Error("Could not init jaeger tracer without ServiceName, either set environment JAEGER_SERVICE_NAME or cfg.ServiceName = \"my-api\"")
		return
	}
	if err != nil {
		logger.Error("Could not parse Jaeger env vars: %s", zap.Error(err))
		return
	}
	tracer, _, err := cfg.NewTracer()
	if err != nil {
		logger.Error("Could not initialize jaeger tracer: %s", zap.Error(err))
		return
	}
	opentracing.SetGlobalTracer(tracer)
	logger.Info("Tracer configured", zap.String("jaeger-agent", cfg.Reporter.LocalAgentHostPort))
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
		)
	}
}

func (app *App) configureMongoDB() {
	database := app.Config.GetString("mongodb.databaseName")
	app.Config.Set("mongodb.database", database)

	if app.Config.GetBool("mongodb.enabled") == true {
		var err error
		app.MongoDB, err = mongo.GetMongo(
			app.Logger,
			app.Config,
		)
		if err != nil {
			app.Logger.Error(err.Error())
		}
	}
}

func (app *App) setConfigurationDefaults() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "setConfigurationDefaults"),
	)
	app.Config.SetDefault("graceperiod.ms", 5000)
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("postgres.host", "localhost")
	app.Config.SetDefault("postgres.user", "khan")
	app.Config.SetDefault("postgres.dbName", "khan")
	app.Config.SetDefault("postgres.port", 5432)
	app.Config.SetDefault("postgres.sslMode", "disable")
	app.Config.SetDefault("webhooks.timeout", 500)
	app.Config.SetDefault("webhooks.maxIdleConnsPerHost", http.DefaultMaxIdleConnsPerHost)
	app.Config.SetDefault("webhooks.maxIdleConns", 100)
	app.Config.SetDefault("elasticsearch.host", "localhost")
	app.Config.SetDefault("elasticsearch.port", 9234)
	app.Config.SetDefault("elasticsearch.sniff", true)
	app.Config.SetDefault("elasticsearch.index", "khan")
	app.Config.SetDefault("elasticsearch.enabled", false)
	app.Config.SetDefault("khan.maxPendingInvites", -1)
	app.Config.SetDefault("khan.defaultCooldownBeforeInvite", -1)
	app.Config.SetDefault("khan.defaultCooldownBeforeApply", -1)
	app.Config.SetDefault("security.encryptionKey", "")

	app.setHandlersConfigurationDefaults()

	log.D(logger, "Configuration defaults set.")
}

func (app *App) setHandlersConfigurationDefaults() {
	app.setRetrieveClanHandlerConfigurationDefaults()
}

func (app *App) setRetrieveClanHandlerConfigurationDefaults() {
	SetRetrieveClanHandlerConfigurationDefaults(app.Config)
}

func (app *App) loadConfiguration() {
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

func (app *App) connectDatabase() {
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

func (app *App) onErrorHandler(err error, stack []byte) {
	log.E(app.Logger, "Panic occurred.", func(cm log.CM) {
		cm.Write(
			zap.String("source", "app"),
			zap.String("panicText", err.Error()),
			zap.String("stack", string(stack)),
		)
	})
}

func (app *App) configureApplication() {
	app.Engine = standard.New(fmt.Sprintf("%s:%d", app.Host, app.Port))
	if app.Fast {
		engine := fasthttp.New(fmt.Sprintf("%s:%d", app.Host, app.Port))
		engine.ReadBufferSize = app.ReadBufferSize
		app.Engine = engine
	}
	app.App = eecho.New()
	a := app.App

	_, w, _ := os.Pipe()
	a.SetLogOutput(w)

	basicAuthUser := app.Config.GetString("basicauth.username")
	if basicAuthUser != "" {
		basicAuthPass := app.Config.GetString("basicauth.password")

		a.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
			Skipper: func(c echo.Context) bool {
				return c.Path() == "/healthcheck"
			},
			Validator: func(username, password string) bool {
				return username == basicAuthUser && password == basicAuthPass
			},
		}))
	}

	a.Use(NewRecoveryMiddleware(app.onErrorHandler).Serve)
	a.Use(extechomiddleware.NewResponseTimeMetricsMiddleware(app.DDStatsD).Serve)
	a.Use(NewVersionMiddleware().Serve)
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
	a.Get("/games/:gameID/clans/search", SearchClansHandler(app))
	a.Get("/games/:gameID/clans", ListClansHandler(app))
	a.Post("/games/:gameID/clans", CreateClanHandler(app))
	a.Get("/games/:gameID/clans-summary", RetrieveClansSummariesHandler(app))
	a.Get("/games/:gameID/clans/:clanPublicID", RetrieveClanHandler(app))
	a.Get("/games/:gameID/clans/:clanPublicID/members", RetrieveClanMembersHandler(app))
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
func (app *App) GetHooks(ctx context.Context) map[string]map[int][]*models.Hook {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "GetHooks"),
	)

	start := time.Now()
	log.D(logger, "Retrieving hooks...")
	dbHooks, err := models.GetAllHooks(app.Db(ctx))
	if err != nil {
		log.E(logger, "Retrieve hooks failed.", func(cm log.CM) {
			cm.Write(zap.String("error", err.Error()))
		})
		return nil
	}
	log.D(logger, "Hooks retrieved successfully.", func(cm log.CM) {
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
func (app *App) GetGame(ctx context.Context, gameID string) (*models.Game, error) {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "GetGame"),
		zap.String("gameID", gameID),
	)

	key := gameID
	value, present := app.getGameCache.Get(key)
	if present {
		return value.(*models.Game), nil
	}

	start := time.Now()
	log.D(logger, "Retrieving game...")

	game, err := models.GetGameByPublicID(app.Db(ctx), gameID)
	if err != nil {
		log.E(logger, "Retrieve game failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, err
	}

	log.D(logger, "Game retrieved succesfully.", func(cm log.CM) {
		cm.Write(zap.Duration("gameRetrievalDuration", time.Now().Sub(start)))
	})
	app.getGameCache.Set(key, game, gocache.DefaultExpiration)
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

	logger := app.Logger.With(
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
	logger.Debug("Configuring workers...")
	workers.Configure(opts)

	workers.Middleware.Append(extworkermiddleware.NewResponseTimeMetricsMiddleware(app.DDStatsD))
	workers.Process(queues.KhanQueue, app.Dispatcher.PerformDispatchHook, workerCount)
	workers.Process(queues.KhanESQueue, app.ESWorker.PerformUpdateES, workerCount)
	workers.Process(queues.KhanMongoQueue, app.MongoWorker.PerformUpdateMongo, workerCount)
	logger.Info("Worker configured.")
}

//StartWorkers "starts" the dispatcher
func (app *App) StartWorkers() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "StartWorkers"),
	)

	log.D(logger, "Starting workers...")
	if app.Config.GetBool("webhooks.runStats") {
		jobsStatsPort := app.Config.GetInt("webhooks.statsPort")
		go workers.StatsServer(jobsStatsPort)
	}
	workers.Run()
}

func (app *App) initESWorker() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "initESWorker"),
	)

	log.D(logger, "Initializing es worker...")
	esWorker := models.NewESWorker(app.Logger)
	log.I(logger, "ES Worker initialized successfully")
	app.ESWorker = esWorker
}

func (app *App) initMongoWorker() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "initMongoWorker"),
	)

	log.D(logger, "Initializing mongo worker...")
	mongoWorker := models.NewMongoWorker(app.Logger, app.Config)
	log.I(logger, "Mongo Worker initialized successfully")
	app.MongoWorker = mongoWorker
}

func (app *App) initDispatcher() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "initDispatcher"),
	)

	log.D(logger, "Initializing dispatcher...")

	disp, err := NewDispatcher(app)
	if err != nil {
		log.P(logger, "Dispatcher failed to initialize.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return
	}
	log.I(logger, "Dispatcher initialized successfully")

	app.Dispatcher = disp
}

// DispatchHooks dispatches web hooks for a specific game and event type
func (app *App) DispatchHooks(gameID string, eventType int, payload map[string]interface{}) error {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "DispatchHooks"),
		zap.String("gameID", gameID),
		zap.Int("eventType", eventType),
	)

	start := time.Now()
	log.D(logger, "Dispatching hook...")
	app.Dispatcher.DispatchHook(gameID, eventType, payload)
	log.D(logger, "Hook dispatched successfully.", func(cm log.CM) {
		cm.Write(zap.Duration("hookDispatchDuration", time.Now().Sub(start)))
	})
	return nil
}

func (app *App) finalizeApp() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "finalizeApp"),
	)

	log.D(logger, "Closing DB connection...")
	app.db.Close()
	log.I(logger, "DB connection closed succesfully.")
}

//BeginTrans in the current Db connection
func (app *App) BeginTrans(ctx context.Context, logger zap.Logger) (gorp.Transaction, error) {
	log.D(logger, "Beginning DB tx...")
	tx, err := app.Db(ctx).Begin()
	if err != nil {
		log.E(logger, "Failed to begin tx.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, err
	}
	log.D(logger, "Tx begun successfuly.")
	return tx, nil
}

//Rollback transaction
func (app *App) Rollback(tx gorp.Transaction, msg string, c echo.Context, logger zap.Logger, err error) error {
	txErr := tx.Rollback()
	if txErr != nil {
		log.E(logger, fmt.Sprintf("%s and failed to rollback transaction.", msg), func(cm log.CM) {
			cm.Write(zap.Error(txErr), zap.String("originalError", err.Error()))
		})
		return txErr
	}
	return nil
}

//Commit transaction
func (app *App) Commit(tx gorp.Transaction, msg string, c echo.Context, logger zap.Logger) error {
	txErr := tx.Commit()
	if txErr != nil {
		log.E(logger, fmt.Sprintf("%s failed to commit transaction.", msg), func(cm log.CM) {
			cm.Write(zap.Error(txErr))
		})
		return txErr
	}
	return nil
}

// GetCtxDB returns the proper database connection depending on the request context
func (app *App) GetCtxDB(ctx echo.Context) (gorp.Database, error) {
	val := ctx.Get("db")
	if val != nil {
		return val.(gorp.Database), nil
	}

	return app.Db(ctx.StdContext()), nil
}

// Db returns a gorp database connection using the given context
func (app *App) Db(ctx context.Context) gorp.Database {
	if ctx == nil {
		ctx = context.Background()
	}
	return app.db.WithContext(ctx).(gorp.Database)
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() {
	logger := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "Start"),
	)

	defer app.finalizeApp()
	log.I(logger, "app started", func(cm log.CM) {
		cm.Write(zap.String("host", app.Host), zap.Int("port", app.Port))
	})

	go func() {
		app.App.Run(app.Engine)
	}()

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	// stop server
	select {
	case s := <-sg:
		graceperiod := app.Config.GetInt("graceperiod.ms")
		log.I(logger, "shutting down", func(cm log.CM) {
			cm.Write(zap.String("signal", fmt.Sprintf("%v", s)),
				zap.Int("graceperiod", graceperiod))
		})
		time.Sleep(time.Duration(graceperiod) * time.Millisecond)
	}
	log.I(logger, "app stopped")
}
