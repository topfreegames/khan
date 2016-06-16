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
)

//FailWith fails with the specified message
func FailWith(status int, message string, c *iris.Context) {
	result, _ := json.Marshal(map[string]interface{}{
		"success": false,
		"reason":  message,
	})
	c.SetStatusCode(status)
	c.Write(string(result))
}

//FailWithJSON fails with the specified json
func FailWithJSON(status int, payload map[string]interface{}, c *iris.Context) {
	payload["success"] = false
	result, _ := json.Marshal(payload)
	c.SetStatusCode(status)
	c.Write(string(result))
}

//SucceedWith sends payload to user with status 200
func SucceedWith(payload map[string]interface{}, c *iris.Context) {
	payload["success"] = true
	result, _ := json.Marshal(payload)
	c.SetStatusCode(200)
	c.Write(string(result))
}
