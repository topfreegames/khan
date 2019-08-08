package loadtest

import (
	"fmt"
	"math/rand"
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
	operations []Operation
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
	app.config.SetDefault("loadtest.requests.amount", 0)
	app.config.SetDefault("loadtest.requests.period.ms", 0)
	app.setClanConfigurationDefaults()
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
	app.configureClanOperations()
}

func (app *App) configureCache(sharedClansFile string) {
	var err error
	app.cache, err = getCacheImpl(app.config, sharedClansFile)
	if err != nil {
		l := app.logger.With(
			zap.String("source", "loadtest/app"),
			zap.String("operation", "configureCache"),
			zap.String("sharedClansFile", sharedClansFile),
		)
		log.P(l, "Error reading shared clans config.", func(cm log.CM) {
			cm.Write(zap.String("error", err.Error()))
		})
	}
}

func (app *App) configureClient() {
	app.client = lib.NewKhan(app.config)
}

// Run executes the load test suite
func (app *App) Run() error {
	nRequests := app.config.GetInt("loadtest.requests.amount")
	periodMs := app.config.GetInt("loadtest.requests.period.ms")
	for i := 0; i < nRequests; i++ {
		err := app.performOperation()
		if err != nil {
			return err
		}
		if i+1 < nRequests {
			time.Sleep(time.Duration(periodMs) * time.Millisecond)
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

func (app *App) getRandomOperation() (Operation, error) {
	sampleSpace, err := app.getOperationSampleSpace()
	if err != nil {
		return Operation{}, err
	}
	return getRandomOperationFromSampleSpace(sampleSpace, -1)
}

func (app *App) getOperationSampleSpace() ([]Operation, error) {
	var pSum float64
	var sampleSpace []Operation
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
	normalizeSampleSpace(sampleSpace, pSum)
	return sampleSpace, nil
}

func normalizeSampleSpace(sampleSpace []Operation, pSum float64) {
	for i := range sampleSpace {
		sampleSpace[i].probability /= pSum
	}
}

func getRandomOperationFromSampleSpace(sampleSpace []Operation, dice float64) (Operation, error) {
	if dice < 0 {
		dice = rand.Float64()
	}
	var pSum float64
	for _, operation := range sampleSpace {
		pSum += operation.probability
		if dice <= pSum {
			return operation, nil
		}
	}
	return Operation{}, &GenericError{"SampleSpaceSumBelowOneError", "Sum of all probabilities is less than one."}
}

func (app *App) appendOperation(operation Operation) {
	app.operations = append(app.operations, operation)
}

func (app *App) getOperationProbabilityConfig(operation string) float64 {
	key := fmt.Sprintf("loadtest.operations.%s.probability", operation)
	return app.config.GetFloat64(key)
}
