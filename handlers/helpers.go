package handlers

import (
	"github.com/plimble/ace"
	"github.com/topfreegames/khan/api"
)

//GetApp returns an app based on a web framework context
func GetApp(c *ace.C) *api.App {
	return c.Get("app").(*api.App)
}
