package loadtest

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/topfreegames/khan/lib"
	"github.com/topfreegames/khan/log"
	"github.com/uber-go/zap"
)

// App represents a load test application
type App struct {
	config     *viper.Viper
	logger     zap.Logger
	cache      cache
	client     lib.KhanInterface
	operations []operation
}

// GetApp returns a new app
func GetApp(configFile, sharedClansFile string, logger zap.Logger) *App {
	app := &App{
		config: viper.New(),
		logger: logger,
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
}

func (app *App) setConfigurationDefaults() {
	app.setClientConfigurationDefaults()
}

func (app *App) setClientConfigurationDefaults() {
	app.config.SetDefault("loadtest.client.timeout", 500*time.Millisecond)
	app.config.SetDefault("loadtest.client.maxIdleConns", 100)
	app.config.SetDefault("loadtest.client.maxIdleConnsPerHost", http.DefaultMaxIdleConnsPerHost)
}

func (app *App) loadConfiguration(configFile string) {
	logger := app.logger.With(
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
		log.I(logger, "Loaded config file successfully.")
	} else {
		log.P(logger, "Config file failed to load.")
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
		logger := app.logger.With(
			zap.String("source", "loadtest/app"),
			zap.String("operation", "configureCache"),
			zap.Int("gameMaxMembers", gameMaxMembers),
			zap.String("sharedClansFile", sharedClansFile),
			zap.String("error", err.Error()),
		)
		log.P(logger, "Error configuring cache.")
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

// Run executes the load test suite
func (app *App) Run() error {
	if err := app.cache.loadInitialData(app.client); err != nil {
		return err
	}

	nOperations := app.config.GetInt("loadtest.operations.amount")
	intervalDuration := app.config.GetDuration("loadtest.operations.interval.duration")
	getPercent := func(cur int) int {
		return int((100 * int64(cur)) / int64(nOperations))
	}

	logger := app.logger.With(
		zap.String("source", "loadtest/app"),
		zap.String("operation", "Run"),
		zap.Int("nOperations", nOperations),
		zap.Duration("intervalDuration", intervalDuration),
	)

	for i := 0; i < nOperations; i++ {
		err := app.performOperation()
		if err != nil {
			return err
		}

		if i < nOperations-1 {
			time.Sleep(intervalDuration)
		}

		if getPercent(i+1)/10 > getPercent(i)/10 {
			log.I(logger, fmt.Sprintf("Goroutine completed %v%%.", getPercent(i+1)))
		}
	}

	return nil
}

func (app *App) performOperation() error {
	operation, err := app.getRandomOperation()
	if err != nil {
		return err
	}
	err = operation.execute()
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
