// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"strings"

	"github.com/kataras/iris"
)

// HealthCheckHandler is the handler responsible for validating that the app is still up
func HealthCheckHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		c.Set("route", "Healthcheck")
		db, err := app.GetCtxDB(c)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		workingString := app.Config.GetString("healthcheck.workingText")
		_, err = db.SelectInt("select count(*) from games")
		if err != nil {
			c.Write(fmt.Sprintf("Error connecting to database: %s", err))
			c.SetStatusCode(500)
			return
		}

		c.SetStatusCode(iris.StatusOK)
		workingString = strings.TrimSpace(workingString)
		c.Write(workingString)
	}
}
