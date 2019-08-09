package loadtest

func (app *App) setMembershipConfigurationDefaults() {
}

func (app *App) configureMembershipOperations() {
	app.appendOperation(app.getApplyForMembershipOperation())
}

func (app *App) getApplyForMembershipOperation() operation {
	operationKey := "applyForMembership"
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
