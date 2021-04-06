// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"strings"

	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

func validateUpdatePlayerDispatch(game *models.Game, sourcePlayer *models.Player, player *models.Player, metadata map[string]interface{}, logger zap.Logger) bool {
	cl := logger.With(
		zap.String("playerUpdateMetadataFieldsHookTriggerWhitelist", game.PlayerUpdateMetadataFieldsHookTriggerWhitelist),
	)

	if sourcePlayer == nil {
		log.D(cl, "Player did not exist before. Dispatching event...")
		return true
	}

	changedName := player.Name != sourcePlayer.Name
	if changedName {
		log.D(cl, "Player name changed")
		return true
	}

	if game.PlayerUpdateMetadataFieldsHookTriggerWhitelist == "" {
		log.D(cl, "Player has no metadata whitelist for update hook")
		return false
	}

	log.D(cl, "Verifying fields for player update hook dispatch...")
	fields := strings.Split(game.PlayerUpdateMetadataFieldsHookTriggerWhitelist, ",")
	for _, field := range fields {
		oldVal, existsOld := sourcePlayer.Metadata[field]
		newVal, existsNew := metadata[field]
		log.D(logger, "Verifying field for change...", func(cm log.CM) {
			cm.Write(
				zap.Bool("existsOld", existsOld),
				zap.Bool("existsNew", existsNew),
				zap.Object("oldVal", oldVal),
				zap.Object("newVal", newVal),
				zap.String("field", field),
			)
		})
		//fmt.Println("field", field, "existsOld", existsOld, "oldVal", oldVal, "existsNew", existsNew, "newVal", newVal)

		if existsOld != existsNew {
			log.D(logger, "Found difference in field. Dispatching hook...", func(cm log.CM) {
				cm.Write(zap.String("field", field))
			})
			return true
		}

		if existsOld && oldVal != newVal {
			log.D(logger, "Found difference in field. Dispatching hook...", func(cm log.CM) {
				cm.Write(zap.String("field", field))
			})
			return true
		}
	}

	return false
}
