package loadtest

func (app *App) configurePlayerOperations() {
	app.appendOperation(app.getCreatePlayerOperation())
}

func (app *App) getCreatePlayerOperation() operation {
	operationKey := "createPlayer"
	app.setOperationProbabilityConfigDefault(operationKey, 1)
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
