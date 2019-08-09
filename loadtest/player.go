package loadtest

func (app *App) setPlayerConfigurationDefaults() {
}

func (app *App) configurePlayerOperations() {
	app.appendOperation(app.getUpdatePlayerOperation())
}

func (app *App) getUpdatePlayerOperation() operation {
	operationKey := "updatePlayer"
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			return true, nil
		},
		execute: func() error {
			return nil
		},
	}
}
