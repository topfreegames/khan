package loadtest

import "math/rand"

func (app *App) setClanConfigurationDefaults() {
}

func (app *App) configureClanOperations() {
	app.appendOperation(app.getRetrieveClanOperation())
}

func (app *App) getRetrieveClanOperation() Operation {
	operationKey := "retrieveClan"
	return Operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			count, err := app.cache.getSharedClansCount()
			if err != nil {
				return false, err
			}
			return count > 0, nil
		},
		execute: func() error {
			clanPublicID, err := app.chooseRandomSharedClanPublicID()
			if err != nil {
				return err
			}
			_, err = app.client.RetrieveClan(nil, clanPublicID)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func (app *App) chooseRandomSharedClanPublicID() (string, error) {
	count, err := app.cache.getSharedClansCount()
	if err != nil {
		return "", err
	}
	idx := int(rand.Float64() * float64(count))
	return app.cache.getSharedClanByPublicID(idx)
}
