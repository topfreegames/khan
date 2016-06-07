package api

import (
	"fmt"

	"github.com/plimble/ace"
	"github.com/spf13/viper"
)

//App is a struct that represents a Khan API Application
type App struct {
	Port       int
	Host       string
	ConfigPath string
	App        *ace.Ace
	Config     *viper.Viper
}

//GetDefaultApp returns a new Khan API Application bound to 0.0.0.0:8888
func GetDefaultApp() *App {
	return GetApp("0.0.0.0", 8888, "./config/local.yaml")
}

//GetApp returns a new Khan API Application
func GetApp(host string, port int, configPath string) *App {
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
	app.configureApplication()
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
}

func (app *App) loadConfiguration() {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", app.Config.ConfigFileUsed())
	}
}

func (app *App) configureApplication() {
	app.App = ace.New()

	app.App.Use(func(c *ace.C) {
		c.Set("app", app)
		c.Next()
	})
}

//URL specifies a triple of method, path and request handler
type URL struct {
	Method  string
	Path    string
	Handler ace.HandlerFunc
}

//AddHandlers adds the specified handlers to the route
func (app *App) AddHandlers(urls ...URL) {
	for _, currURL := range urls {
		urls := []ace.HandlerFunc{currURL.Handler}
		app.App.Handle(currURL.Method, currURL.Path, urls)
	}
}

//Start starts listening for web requests at specified host and port
func (app *App) Start() {
	app.App.Run(fmt.Sprintf("%s:%d", app.Host, app.Port))
}
