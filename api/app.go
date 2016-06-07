package api

import (
	"fmt"

	"github.com/plimble/ace"
	"github.com/topfreegames/khan/handlers"
)

//App is a struct that represents a Khan API Application
type App struct {
	Port int
	Host string
	App  *ace.Ace
}

//GetDefaultApp returns a new Khan API Application bound to 0.0.0.0:8888
func GetDefaultApp() *App {
	return GetApp("0.0.0.0", 8888)
}

//GetApp returns a new Khan API Application
func GetApp(host string, port int) *App {
	app := &App{
		Host: host,
		Port: port,
	}
	app.Configure()
	return app
}

//Configure instantiates the required dependencies for Khan Api Application
func (app *App) Configure() {
	app.configureApplication()
}

func (app *App) configureApplication() {
	app.App = ace.New()
	app.App.GET("/healthcheck", handlers.HealthcheckHandler)
}

//Start starts listening for web requests at specified host and port
func (app *App) Start() {
	app.App.Run(fmt.Sprintf("%s:%d", app.Host, app.Port))
}
