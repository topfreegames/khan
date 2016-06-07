package handlers

import "github.com/plimble/ace"

//HealthcheckHandler is the handler responsible for validating that the app is still up
func HealthcheckHandler(c *ace.C) {
	app := GetApp(c)
	c.String(200, app.Config.GetString("healthcheck.workingText"))
}
