// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
)

//StatusHandler is the handler responsible for reporting khan status
func StatusHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "Status")
		payload := map[string]interface{}{
			"app": map[string]interface{}{
				"errorRate": app.Errors.Rate(),
			},
			"dispatch": map[string]interface{}{
				"pendingJobs": app.Dispatcher.Jobs,
			},
		}

		var payloadJSON []byte
		WithSegment("response-marshalling", c, func() error {
			payloadJSON, _ = json.Marshal(payload)
			return nil
		})
		return c.String(http.StatusOK, string(payloadJSON))
	}
}
