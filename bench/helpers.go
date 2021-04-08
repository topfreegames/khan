// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
	"github.com/topfreegames/khan/mongo"
)

func getRoute(url string) string {
	return fmt.Sprintf("http://localhost:8888%s", url)
}

func get(url string) (*http.Response, error) {
	return sendTo("GET", url, nil)
}

func postTo(url string, payload map[string]interface{}) (*http.Response, error) {
	return sendTo("POST", url, payload)
}

func putTo(url string, payload map[string]interface{}) (*http.Response, error) {
	return sendTo("PUT", url, payload)
}

func sendTo(method, url string, payload map[string]interface{}) (*http.Response, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(payloadJSON))
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getClanPayload(ownerID, clanPublicID string) map[string]interface{} {
	return map[string]interface{}{
		"publicID":         clanPublicID,
		"name":             clanPublicID,
		"ownerPublicID":    ownerID,
		"metadata":         map[string]interface{}{"x": "a"},
		"allowApplication": true,
		"autoJoin":         true,
	}
}

func getPlayerPayload(playerPublicID string) map[string]interface{} {
	return map[string]interface{}{
		"publicID": playerPublicID,
		"name":     playerPublicID,
		"metadata": map[string]interface{}{"x": "a"},
	}
}

func getGameAndPlayer(db models.DB, mongoDB interfaces.MongoDB) (*models.Game, *models.Player, error) {
	game := fixtures.GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID":          uuid.NewV4().String(),
		"MaxClansPerPlayer": 999999,
	}).(*models.Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, err
	}
	player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":   game.PublicID,
		"PublicID": uuid.NewV4().String(),
	}).(*models.Player)
	err = db.Insert(player)
	if err != nil {
		return nil, nil, err
	}

	err = mongoDB.Run(mongo.GetClanNameTextIndexCommand(game.PublicID, false), nil)
	if err != nil {
		return nil, nil, err
	}

	return game, player, nil
}

func validateResp(res *http.Response, err error) {
	if err != nil {
		panic(err)
	}
	if res.StatusCode != 200 {
		bts, _ := ioutil.ReadAll(res.Body)
		fmt.Printf("Request failed with status code %d\n", res.StatusCode)
		panic(string(bts))
	}
}

func createClans(db models.DB, game *models.Game, owner *models.Player, numberOfClans int) ([]*models.Clan, error) {
	var clans []*models.Clan
	for i := 0; i < numberOfClans; i++ {
		clan := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
			"GameID":   game.PublicID,
			"PublicID": uuid.NewV4().String(),
			"OwnerID":  owner.ID,
		}).(*models.Clan)

		err := db.Insert(clan)
		if err != nil {
			return nil, err
		}
		clans = append(clans, clan)
	}

	return clans, nil
}
