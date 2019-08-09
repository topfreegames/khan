package loadtest

func (app *App) setPlayerConfigurationDefaults() {
}

func (app *App) configurePlayerOperations() {
	app.appendOperation(app.getCreatePlayerOperation())
}

func (app *App) getCreatePlayerOperation() operation {
	operationKey := "createPlayer"
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			return true, nil
		},
		execute: func() error {
			playerPublicID := getRandomPublicID()
			_, err := app.client.CreatePlayer(nil, playerPublicID, getRandomPlayerName(), getMetadataWithRandomScore())
			if err != nil {
				return err
			}
			return app.cache.addFreePlayer(playerPublicID)
		},
	}
}
