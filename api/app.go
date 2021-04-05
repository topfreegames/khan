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
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/jrallison/go-workers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	newrelic "github.com/newrelic/go-agent"
	gocache "github.com/patrickmn/go-cache"
	"github.com/rcrowley/go-metrics"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	eecho "github.com/topfreegames/extensions/v9/echo"
	extechomiddleware "github.com/topfreegames/extensions/v9/echo/middleware"
	gorp "github.com/topfreegames/extensions/v9/gorp/interfaces"
	"github.com/topfreegames/extensions/v9/jaeger"
	extnethttpmiddleware "github.com/topfreegames/extensions/v9/middleware"
	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	extworkermiddleware "github.com/topfreegames/extensions/v9/worker/middleware"
	"github.com/topfreegames/khan/caches"
	"github.com/topfreegames/khan/es"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/mongo"
	"github.com/topfreegames/khan/queues"
	"github.com/topfreegames/khan/util"
	"github.com/uber-go/zap"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// App is a struct that represents a Khan API Application
type App struct {
	ID             string
	Test           bool
	Debug          bool
	Port           int
	Host           string
	ConfigPath     string
	Errors         metrics.EWMA
	App            *eecho.Echo
	Engine         engine.Server
	Config         *viper.Viper
	Dispatcher     *Dispatcher
	ESWorker       *models.ESWorker
	MongoWorker    *models.MongoWorker
	Logger         zap.Logger
	ESClient       *es.Client
	MongoDB        interfaces.MongoDB
	ReadBufferSize int
	Fast           bool
	NewRelic       newrelic.Application
	DDStatsD       *extnethttpmiddleware.DogStatsD
	EncryptionKey  []byte

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
	app.configureSentry()
	app.configureNewRelic()
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

func (app *App) configureSentry() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := app.Config.GetString("sentry.url")
	log.D(l, fmt.Sprintf("Configuring sentry with URL %s", sentryURL))
	raven.SetDSN(sentryURL)
	raven.SetRelease(util.VERSION)
}

func (app *App) configureStatsD() error {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureStatsD"),
	)

	ddstatsd, err := extnethttpmiddleware.NewDogStatsD(app.Config)
	if err != nil {
		log.E(l, "Failed to initialize DogStatsD.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}
	app.DDStatsD = ddstatsd
	l.Info("Initialized DogStatsD successfully.")

	return nil
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

func (app *App) configureJaeger() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureJaeger"),
	)

	opts := jaeger.Options{
		Disabled:    app.Config.GetBool("jaeger.disabled"),
		Probability: app.Config.GetFloat64("jaeger.samplingProbability"),
		ServiceName: app.Config.GetString("jaeger.serviceName"),
	}

	_, err := jaeger.Configure(opts)
	if err != nil {
		l.Error("Failed to initialize Jaeger.")
	}
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
	l := app.Logger.With(
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
	app.Config.SetDefault("jaeger.disabled", true)
	app.Config.SetDefault("jaeger.samplingProbability", 0.001)
	app.Config.SetDefault("security.encryptionKey", "00000000000000000000000000000000")

	app.setHandlersConfigurationDefaults()

	log.D(l, "Configuration defaults set.")
}

func (app *App) setHandlersConfigurationDefaults() {
	app.setRetrieveClanHandlerConfigurationDefaults()
}

func (app *App) setRetrieveClanHandlerConfigurationDefaults() {
	SetRetrieveClanHandlerConfigurationDefaults(app.Config)
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

	app.EncryptionKey = []byte(app.Config.GetString("security.encryptionKey"))
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

	//NewRelicMiddleware has to stand out from all others
	a.Use(NewNewRelicMiddleware(app, app.Logger).Serve)

	a.Use(NewRecoveryMiddleware(app.onErrorHandler).Serve)
	a.Use(extechomiddleware.NewResponseTimeMetricsMiddleware(app.DDStatsD).Serve)
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

	// pprof
	pprofHandlers := map[string]func(http.ResponseWriter, *http.Request){
		"/debug/pprof":         pprof.Index,
		"/debug/pprof/profile": pprof.Profile,
		"/debug/pprof/trace":   pprof.Trace,
	}

	for k, v := range pprofHandlers {
		a.Get(k, fasthttp.WrapHandler(
			fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(v)),
		))
	}

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
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "GetHooks"),
	)

	start := time.Now()
	log.D(l, "Retrieving hooks...")
	dbHooks, err := models.GetAllHooks(app.Db(ctx))
	if err != nil {
		log.E(l, "Retrieve hooks failed.", func(cm log.CM) {
			cm.Write(zap.String("error", err.Error()))
		})
		return nil
	}
	log.D(l, "Hooks retrieved successfully.", func(cm log.CM) {
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
	l := app.Logger.With(
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
	log.D(l, "Retrieving game...")

	game, err := models.GetGameByPublicID(app.Db(ctx), gameID)
	if err != nil {
		log.E(l, "Retrieve game failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, err
	}

	log.D(l, "Game retrieved succesfully.", func(cm log.CM) {
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

	workers.Middleware.Append(extworkermiddleware.NewResponseTimeMetricsMiddleware(app.DDStatsD))
	workers.Process(queues.KhanQueue, app.Dispatcher.PerformDispatchHook, workerCount)
	workers.Process(queues.KhanESQueue, app.ESWorker.PerformUpdateES, workerCount)
	workers.Process(queues.KhanMongoQueue, app.MongoWorker.PerformUpdateMongo, workerCount)
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

func (app *App) initMongoWorker() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "initMongoWorker"),
	)

	log.D(l, "Initializing mongo worker...")
	mongoWorker := models.NewMongoWorker(app.Logger, app.Config)
	log.I(l, "Mongo Worker initialized successfully")
	app.MongoWorker = mongoWorker
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
	log.D(l, "Hook dispatched successfully.", func(cm log.CM) {
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
	app.db.Close()
	log.I(l, "DB connection closed succesfully.")
}

//BeginTrans in the current Db connection
func (app *App) BeginTrans(ctx context.Context, l zap.Logger) (gorp.Transaction, error) {
	log.D(l, "Beginning DB tx...")
	tx, err := app.Db(ctx).Begin()
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
func (app *App) Rollback(tx gorp.Transaction, msg string, c echo.Context, l zap.Logger, err error) error {
	txErr := tx.Rollback()
	if txErr != nil {
		log.E(l, fmt.Sprintf("%s and failed to rollback transaction.", msg), func(cm log.CM) {
			cm.Write(zap.Error(txErr), zap.String("originalError", err.Error()))
		})
		return txErr
	}
	return nil
}

//Commit transaction
func (app *App) Commit(tx gorp.Transaction, msg string, c echo.Context, l zap.Logger) error {
	txErr := tx.Commit()
	if txErr != nil {
		log.E(l, fmt.Sprintf("%s failed to commit transaction.", msg), func(cm log.CM) {
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
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "Start"),
	)

	defer app.finalizeApp()
	log.I(l, "app started", func(cm log.CM) {
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
		log.I(l, "shutting down", func(cm log.CM) {
			cm.Write(zap.String("signal", fmt.Sprintf("%v", s)),
				zap.Int("graceperiod", graceperiod))
		})
		time.Sleep(time.Duration(graceperiod) * time.Millisecond)
	}
	log.I(l, "app stopped")
}
