package loadtest

func (app *App) setPlayerConfigurationDefaults() {
	app.setOperationProbabilityConfigDefault("createPlayer", 1)
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

			createdPublicID, err := app.client.CreatePlayer(nil, playerPublicID, getRandomPlayerName(), getMetadataWithRandomScore())
			if err != nil {
				return err
			}
			if createdPublicID != playerPublicID {
				return &GenericError{"WrongPublicIDError", "Operation createPlayer returned no error with public ID different from requested."}
			}

			return app.cache.createPlayer(playerPublicID)
		},
	}
}
