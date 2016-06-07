package handlers

import (
	"github.com/plimble/ace"
	"github.com/topfreegames/khan/api"
)

func GetApp(c *ace.C) *api.App {
	return c.Get("app").(*api.App)
}
