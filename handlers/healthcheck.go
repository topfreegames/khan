package handlers

import (
	"github.com/plimble/ace"
	"github.com/spf13/viper"
)

//HealthcheckHandler is the handler responsible for validating that the app is still up
func HealthcheckHandler(c *ace.C) {
	c.String(200, viper.GetString("healthcheck.workingText"))
}
