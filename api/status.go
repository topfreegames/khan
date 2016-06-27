// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"

	"github.com/kataras/iris"
	"github.com/topfreegames/khan/util"
)

// StatusHandler is the handler responsible for reporting khan status
func StatusHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		payload := util.JSON{
			"app": util.JSON{
				"errorRate": app.Errors.Rate(),
			},
			"dispatch": util.JSON{
				"pendingJobs": app.Dispatcher.Jobs,
			},
		}

		payloadJSON, _ := json.Marshal(payload)
		c.SetStatusCode(iris.StatusOK)
		c.Write(string(payloadJSON))
	}
}
