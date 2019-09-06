package loadtest

import (
	"github.com/topfreegames/khan/lib"
)

func (app *App) configureClanOperations() {
	app.appendOperation(app.getUpdateSharedClanScoreOperation())
	app.appendOperation(app.getCreateClanOperation())
	app.appendOperation(app.getRetrieveClanOperation())
	app.appendOperation(app.getLeaveClanOperation())
	app.appendOperation(app.getTransferClanOwnershipOperation())
	app.appendOperation(app.getSearchClansOperation())
}

func (app *App) getUpdateSharedClanScoreOperation() operation {
	operationKey := "updateSharedClanScore"
	app.setOperationProbabilityConfigDefault(operationKey, 1)
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			count, err := app.cache.getSharedClansCount()
			if err != nil {
				return false, err
			}
			return count > 0, nil
		},
		execute: func() error {
			clanPublicID, playerPublicID, err := app.cache.chooseRandomSharedClanAndPlayer()
			if err != nil {
				return err
			}

			// updatePlayer
			result, err := app.client.UpdatePlayer(nil, playerPublicID, getRandomPlayerName(), getMetadataWithRandomScore())
			if err != nil {
				return err
			}
			if result == nil {
				return &GenericError{"NilPayloadError", "Operation updatePlayer returned no error with nil payload."}
			}
			if !result.Success {
				return &GenericError{"FailurePayloadError", "Operation updatePlayer returned no error with failure payload."}
			}

			// getClan
			clan, err := app.client.RetrieveClan(nil, clanPublicID)
			if err != nil {
				return err
			}
			if clan == nil {
				return &GenericError{"NilPayloadError", "Operation retrieveClan returned no error with nil payload."}
			}
			if clan.PublicID != clanPublicID {
				return &GenericError{"WrongPublicIDError", "Operation retrieveClan returned no error with public ID different from requested."}
			}

			// updateClan
			clanScore := getScoreFromMetadata(clan.Owner.Metadata)
			for _, member := range clan.Roster {
				clanScore += getScoreFromMetadata(member.Player.Metadata)
			}
			clanMetadata := map[string]interface{}{
				"score": clanScore,
			}
			result, err = app.client.UpdateClan(nil, &lib.ClanPayload{
				PublicID:         clan.PublicID,
				Name:             clan.Name,
				OwnerPublicID:    clan.Owner.PublicID,
				Metadata:         clanMetadata,
				AllowApplication: clan.AllowApplication,
				AutoJoin:         clan.AutoJoin,
			})
			if err != nil {
				return err
			}
			if result == nil {
				return &GenericError{"NilPayloadError", "Operation updateClan returned no error with nil payload."}
			}
			if !result.Success {
				return &GenericError{"FailurePayloadError", "Operation updateclan returned no error with failure payload."}
			}

			return nil
		},
	}
}

func (app *App) getCreateClanOperation() operation {
	operationKey := "createClan"
	app.setOperationProbabilityConfigDefault(operationKey, 1)
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			count, err := app.cache.getFreePlayersCount()
			if err != nil {
				return false, err
			}
			return count > 0, nil
		},
		execute: func() error {
			playerPublicID, err := app.cache.chooseRandomFreePlayer()
			if err != nil {
				return err
			}

			clanPublicID := getRandomPublicID()
			createdPublicID, err := app.client.CreateClan(nil, &lib.ClanPayload{
				PublicID:         clanPublicID,
				Name:             getRandomClanName(10),
				OwnerPublicID:    playerPublicID,
				Metadata:         getMetadataWithRandomScore(),
				AllowApplication: true,
				AutoJoin:         true,
			})
			if err != nil {
				return err
			}
			if createdPublicID != clanPublicID {
				return &GenericError{"WrongPublicIDError", "Operation createClan returned no error with public ID different from requested."}
			}

			return app.cache.createClan(clanPublicID, playerPublicID)
		},
	}
}

func (app *App) getRetrieveClanOperation() operation {
	operationKey := "retrieveClan"
	app.setOperationProbabilityConfigDefault(operationKey, 1)
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			count, err := app.cache.getOwnerPlayersCount()
			if err != nil {
				return false, err
			}
			return count > 0, nil
		},
		execute: func() error {
			clanPublicID, err := app.cache.chooseRandomClan()
			if err != nil {
				return err
			}

			clan, err := app.client.RetrieveClan(nil, clanPublicID)
			if err != nil {
				return err
			}
			if clan == nil {
				return &GenericError{"NilPayloadError", "Operation retrieveClan returned no error with nil payload."}
			}
			if clan.PublicID != clanPublicID {
				return &GenericError{"WrongPublicIDError", "Operation retrieveClan returned no error with public ID different from requested."}
			}

			return nil
		},
	}
}

func (app *App) getLeaveClanOperation() operation {
	operationKey := "leaveClan"
	app.setOperationProbabilityConfigDefault(operationKey, 1)
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			count, err := app.cache.getOwnerPlayersCount()
			if err != nil {
				return false, err
			}
			return count > 0, nil
		},
		execute: func() error {
			clanPublicID, err := app.cache.chooseRandomClan()
			if err != nil {
				return err
			}

			leaveClanResult, err := app.client.LeaveClan(nil, clanPublicID)
			if err != nil {
				return err
			}
			if leaveClanResult == nil {
				return &GenericError{"NilPayloadError", "Operation leaveClan returned no error with nil payload."}
			}

			oldOwnerPublicID := leaveClanResult.PreviousOwner.PublicID
			newOwnerPublicID := ""
			if leaveClanResult.NewOwner != nil {
				newOwnerPublicID = leaveClanResult.NewOwner.PublicID
			}
			return app.cache.leaveClan(clanPublicID, oldOwnerPublicID, newOwnerPublicID)
		},
	}
}

func (app *App) getTransferClanOwnershipOperation() operation {
	operationKey := "transferClanOwnership"
	app.setOperationProbabilityConfigDefault(operationKey, 1)
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			count, err := app.cache.getMemberPlayersCount()
			if err != nil {
				return false, nil
			}
			return count > 0, nil
		},
		execute: func() error {
			newOwnerPublicID, clanPublicID, err := app.cache.chooseRandomMemberPlayerAndClan()
			if err != nil {
				return err
			}

			transferOwnershipResult, err := app.client.TransferOwnership(nil, newOwnerPublicID, clanPublicID)
			if err != nil {
				return err
			}
			if transferOwnershipResult == nil {
				return &GenericError{"NilPayloadError", "Operation transferClanOwnership returned no error with nil payload."}
			}

			oldOwnerPublicID := transferOwnershipResult.PreviousOwner.PublicID
			return app.cache.transferClanOwnership(clanPublicID, oldOwnerPublicID, newOwnerPublicID)
		},
	}
}

func (app *App) getSearchClansOperation() operation {
	operationKey := "searchClans"
	app.setOperationProbabilityConfigDefault(operationKey, 1)
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			return true, nil
		},
		execute: func() error {
			searchClansResult, err := app.client.SearchClans(nil, getRandomClanName(10))
			if err != nil {
				return err
			}
			if searchClansResult == nil {
				return &GenericError{"NilPayloadError", "Operation searchClans returned no error with nil payload."}
			}
			if !searchClansResult.Success {
				return &GenericError{"FailurePayloadError", "Operation searchClans returned no error with failure payload."}
			}
			return nil
		},
	}
}
