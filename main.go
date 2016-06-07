package main

import "github.com/topfreegames/khan/api"

func main() {
	app := api.GetDefaultApp()
	app.Start()
}
