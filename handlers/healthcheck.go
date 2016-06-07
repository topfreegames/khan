package handlers

import "github.com/plimble/ace"

//HealthcheckHandler is the handler responsible for validating that the app is still up
func HealthcheckHandler(c *ace.C) {
	app := GetApp(c)

	rows, err := app.Db.Raw("SELECT 1").Rows()
	defer rows.Close()
	if err != nil {
		c.String(500, "Failed to connect to postgres...")
		return
	}

	c.String(200, app.Config.GetString("healthcheck.workingText"))
}
