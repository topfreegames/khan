package api

import (
	"fmt"

	"github.com/kataras/iris"
)

//HealthCheckHandler is the handler responsible for validating that the app is still up
func HealthCheckHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		workingString := app.Config.GetString("healthcheck.workingText")
		num, err := app.Db.SelectInt("select 1")
		if num != 1 || err != nil {
			c.Write(fmt.Sprintf("Error connecting to database: %s", err))
			c.SetStatusCode(500)
			return
		}

		c.SetStatusCode(iris.StatusOK)
		c.Write(workingString)
	}
}
