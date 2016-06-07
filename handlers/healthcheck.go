package handlers

import "github.com/plimble/ace"

func HealthcheckHandler(c *ace.C) {
	c.String(200, "Working")
}
