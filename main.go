package main

import (
	"github.com/topfreegames/khan/api"
	"github.com/topfreegames/khan/handlers"
)

func main() {
	app := api.GetDefaultApp()

	app.AddHandlers(api.URL{
		Method:  "GET",
		Path:    "/healthcheck",
		Handler: handlers.HealthcheckHandler,
	})

	app.Start()
}
