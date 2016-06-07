package api

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //This is required to use postgres with gorm
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recovery"
	"github.com/spf13/viper"
)

//App is a struct that represents a Khan API Application
type App struct {
	Debug      bool
	Port       int
	Host       string
	ConfigPath string
	App        *iris.Iris
	Db         *gorm.DB
	Config     *viper.Viper
}

//GetDefaultApp returns a new Khan API Application bound to 0.0.0.0:8888
func GetDefaultApp() *App {
	return GetApp("0.0.0.0", 8888, "./config/local.yaml", true)
}

//GetApp returns a new Khan API Application
func GetApp(host string, port int, configPath string, debug bool) *App {
	app := &App{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
	}
	app.Configure()
	return app
}

//Configure instantiates the required dependencies for Khan Api Application
func (app *App) Configure() {
	app.setConfigurationDefaults()
	app.loadConfiguration()
	app.connectDatabase()
	app.configureApplication()
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("postgres.host", "localhost")
	app.Config.SetDefault("postgres.user", "khan")
	app.Config.SetDefault("postgres.dbName", "khan")
	app.Config.SetDefault("postgres.port", 5432)
	app.Config.SetDefault("postgres.sslMode", "disable")
}

func (app *App) loadConfiguration() {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("khan")
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", app.Config.ConfigFileUsed())
	}
}

func (app *App) connectDatabase() {
	host := app.Config.GetString("postgres.host")
	user := app.Config.GetString("postgres.user")
	dbName := app.Config.GetString("postgres.dbname")
	password := app.Config.GetString("postgres.password")
	port := app.Config.GetInt("postgres.port")
	sslMode := app.Config.GetString("postgres.sslMode")

	connStr :=
		fmt.Sprintf("host=%s user=%s port=%d sslmode=%s dbname=%s", host, user, port, sslMode, dbName)

	if password != "" {
		connStr += fmt.Sprintf(" password=%s", password)
	}

	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		fmt.Printf(
			"Could not connect to Postgres at %s:%d with user %s and db %s with password %s (%s)\n",
			host, port, user, dbName, password, err,
		)
		os.Exit(1)
	}
	app.Db = db
}

func (app *App) configureApplication() {
	app.App = iris.New()
	a := app.App

	if app.Debug {
		iris.Use(logger.New(iris.Logger()))
	}
	iris.Use(recovery.New(os.Stderr))

	a.Get("/healthcheck", HealthCheckHandler(app))
}

func (app *App) finalizeApp() {
	app.Db.Close()
}

//Start starts listening for web requests at specified host and port
func (app *App) Start() {
	defer app.finalizeApp()
	app.App.Listen(fmt.Sprintf("%s:%d", app.Host, app.Port))
}
