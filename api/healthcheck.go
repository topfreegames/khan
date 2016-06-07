package api

import (
	"github.com/kataras/iris"
)

//HealthCheckHandler is the handler responsible for validating that the app is still up
func HealthCheckHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		workingString := app.Config.GetString("healthcheck.workingText")
		c.SetStatusCode(iris.StatusOK)
		c.Write(workingString)
	}
}
