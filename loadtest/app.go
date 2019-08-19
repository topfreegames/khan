package loadtest

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/dogstatsd"
	"github.com/topfreegames/khan/lib"
	"github.com/topfreegames/khan/log"
	"github.com/uber-go/zap"
)

// App represents a load test application
type App struct {
	config        *viper.Viper
	logger        zap.Logger
	cache         cache
	client        lib.KhanInterface
	datadog       *dogstatsd.DogStatsD
	poolQueue     chan operation
	poolExitQueue chan bool
	operations    []operation
}

// GetApp returns a new app
func GetApp(configFile, sharedClansFile string, logger zap.Logger) *App {
	app := &App{
		config:        viper.New(),
		logger:        logger,
		poolQueue:     make(chan operation),
		poolExitQueue: make(chan bool),
	}
	app.configure(configFile, sharedClansFile)
	return app
}

func (app *App) configure(configFile, sharedClansFile string) {
	app.setConfigurationDefaults()
	app.loadConfiguration(configFile)
	app.configureOperations()
	app.configureCache(sharedClansFile)
	app.configureClient()
	app.configureDatadog()
}

func (app *App) setConfigurationDefaults() {
	app.setClientConfigurationDefaults()
	app.setDatadogConfigurationDefaults()
}

func (app *App) setClientConfigurationDefaults() {
	app.config.SetDefault("loadtest.client.timeout", 500*time.Millisecond)
	app.config.SetDefault("loadtest.client.maxIdleConns", 100)
	app.config.SetDefault("loadtest.client.maxIdleConnsPerHost", http.DefaultMaxIdleConnsPerHost)
}

func (app *App) setDatadogConfigurationDefaults() {
	app.config.SetDefault("loadtest.datadog.host", "localhost:8125")
	app.config.SetDefault("loadtest.datadog.prefix", "khan_loadtest.")
}

func (app *App) loadConfiguration(configFile string) {
	l := app.logger.With(
		zap.String("source", "loadtest/app"),
		zap.String("operation", "loadConfiguration"),
		zap.String("configFile", configFile),
	)

	app.config.SetConfigType("yaml")
	app.config.SetConfigFile(configFile)
	app.config.SetEnvPrefix("khan")
	app.config.AddConfigPath(".")
	app.config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.config.AutomaticEnv()

	if err := app.config.ReadInConfig(); err == nil {
		log.I(l, "Loaded config file successfully.")
	} else {
		log.P(l, "Config file failed to load.")
	}
}

func (app *App) configureOperations() {
	app.configurePlayerOperations()
	app.configureClanOperations()
	app.configureMembershipOperations()
}

func (app *App) configureCache(sharedClansFile string) {
	gameMaxMembers := app.config.GetInt("loadtest.game.maxMembers")

	var err error
	app.cache, err = newCacheImpl(gameMaxMembers, sharedClansFile)
	if err != nil {
		l := app.logger.With(
			zap.String("source", "loadtest/app"),
			zap.String("operation", "configureCache"),
			zap.Int("gameMaxMembers", gameMaxMembers),
			zap.String("sharedClansFile", sharedClansFile),
			zap.String("error", err.Error()),
		)
		log.P(l, "Error configuring cache.")
	}
}

func (app *App) configureClient() {
	app.client = lib.NewKhanWithParams(&lib.KhanParams{
		Timeout:             app.config.GetDuration("loadtest.client.timeout"),
		MaxIdleConns:        app.config.GetInt("loadtest.client.maxIdleConns"),
		MaxIdleConnsPerHost: app.config.GetInt("loadtest.client.maxIdleConnsPerHost"),
		URL:                 app.config.GetString("loadtest.client.url"),
		User:                app.config.GetString("loadtest.client.user"),
		Pass:                app.config.GetString("loadtest.client.pass"),
		GameID:              app.config.GetString("loadtest.client.gameid"),
	})
}

func (app *App) configureDatadog() {
	host := app.config.GetString("loadtest.datadog.host")
	prefix := app.config.GetString("loadtest.datadog.prefix")

	var err error
	app.datadog, err = dogstatsd.New(host, prefix)
	if err != nil {
		app.datadog = nil

		l := app.logger.With(
			zap.String("source", "loadtest/app"),
			zap.String("operation", "configureDatadog"),
			zap.String("host", host),
			zap.String("prefix", prefix),
			zap.String("error", err.Error()),
		)
		log.E(l, "Error configuring datadog.")
	}
}

// Run executes the load test suite
func (app *App) Run() error {
	if err := app.cache.loadInitialData(app.client); err != nil {
		return err
	}

	app.startThreadPool()

	nOperations := app.config.GetInt("loadtest.operations.amount")
	periodMs := app.config.GetInt("loadtest.operations.period.ms")
	getPercent := func(cur int) int {
		return int((100 * int64(cur)) / int64(nOperations))
	}

	l := app.logger.With(
		zap.String("source", "loadtest/app"),
		zap.String("operation", "Run"),
		zap.Int("nOperations", nOperations),
		zap.Int("periodMs", periodMs),
	)

	startTime := time.Now()
	lastMeasureTime := startTime
	nOperationsSinceLastMeasure := 0

	var err error
	for i := 0; i < nOperations; i++ {
		err = app.performOperation()
		nOperationsSinceLastMeasure++
		if err != nil {
			break
		}

		if i < nOperations-1 {
			time.Sleep(time.Duration(periodMs) * time.Millisecond)
		}

		if time.Since(lastMeasureTime) > time.Second {
			throughput := float64(nOperationsSinceLastMeasure) / time.Since(lastMeasureTime).Seconds()
			app.datadog.Histogram("throughput", throughput, []string{}, 1)
			lastMeasureTime, nOperationsSinceLastMeasure = time.Now(), 0
		}

		if getPercent(i+1)/10 > getPercent(i)/10 {
			log.I(l, fmt.Sprintf("Goroutine completed %v%%.", getPercent(i+1)))
		}
	}

	app.stopThreadPool()

	if err != nil {
		return err
	}

	if app.datadog != nil {
		avg := float64(nOperations) / time.Since(startTime).Seconds()
		app.datadog.Histogram("average_throughput", avg, []string{}, 1)
	}

	return nil
}

const threadPoolSizeConfKey string = "loadtest.threadPool.size"

func (app *App) startThreadPool() {
	app.config.SetDefault(threadPoolSizeConfKey, 1)
	poolSize := app.config.GetInt(threadPoolSizeConfKey)
	for i := 0; i < poolSize; i++ {
		go func() {
			for {
				op := <-app.poolQueue
				if op.probability == 0 {
					app.poolExitQueue <- true
					return
				}
				if err := op.execute(); err != nil {
					l := app.logger.With(
						zap.String("source", "loadtest/app"),
						zap.String("operation", "threadPool/executeOperation"),
						zap.String("executedOperationKey", op.key),
						zap.String("error", err.Error()),
					)
					log.E(l, "Async operation returned error.")
				}
			}
		}()
	}
}

func (app *App) stopThreadPool() {
	poolSize := app.config.GetInt(threadPoolSizeConfKey)
	for i := 0; i < poolSize; i++ {
		app.poolQueue <- operation{}
	}
	for i := 0; i < poolSize; i++ {
		<-app.poolExitQueue
	}
}

func (app *App) performOperation() error {
	operation, err := app.getRandomOperation()
	if err != nil {
		return err
	}
	err = app.executeOperation(operation)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) getRandomOperation() (operation, error) {
	sampleSpace, err := app.getOperationSampleSpace()
	if err != nil {
		return operation{}, err
	}
	return getRandomOperationFromSampleSpace(sampleSpace, -1), nil
}

func (app *App) getOperationSampleSpace() ([]operation, error) {
	var pSum float64
	var sampleSpace []operation
	for _, operation := range app.operations {
		ok, err := operation.canExecute()
		if err != nil {
			return nil, err
		}
		if ok && operation.probability > 0 {
			pSum += operation.probability
			sampleSpace = append(sampleSpace, operation)
		}
	}
	if pSum <= 0 {
		return nil, &GenericError{"NoOperationsAvailableError", "No operations can be executed anymore."}
	}
	for i := range sampleSpace {
		sampleSpace[i].probability /= pSum
	}
	return sampleSpace, nil
}

func getRandomOperationFromSampleSpace(sampleSpace []operation, dice float64) operation {
	if dice < 0 {
		dice = rand.Float64()
	}
	var pSum float64
	for _, operation := range sampleSpace {
		pSum += operation.probability
		if dice <= pSum {
			return operation
		}
	}
	return sampleSpace[0]
}

func (app *App) executeOperation(op operation) error {
	if op.wontUpdateCache { // then run async
		app.poolQueue <- op
		return nil
	}
	return op.execute()
}

func (app *App) appendOperation(op operation) {
	app.operations = append(app.operations, op)
}

func (app *App) getOperationProbabilityConfigKey(operation string) string {
	return fmt.Sprintf("loadtest.operations.%s.probability", operation)
}

func (app *App) getOperationProbabilityConfig(operation string) float64 {
	return app.config.GetFloat64(app.getOperationProbabilityConfigKey(operation))
}

func (app *App) setOperationProbabilityConfigDefault(operation string, probability float64) {
	app.config.SetDefault(app.getOperationProbabilityConfigKey(operation), probability)
}
