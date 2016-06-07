package handlers

import "github.com/plimble/ace"

//HealthcheckHandler is the handler responsible for validating that the app is still up
func HealthcheckHandler(c *ace.C) {
	c.String(200, "WORKING")
}
