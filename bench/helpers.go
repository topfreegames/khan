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

	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/util"
)

func getRoute(url string) string {
	return fmt.Sprintf("http://localhost:8888%s", url)
}

func get(url string) (*http.Response, error) {
	return sendTo("GET", url, nil)
}

func postTo(url string, payload util.JSON) (*http.Response, error) {
	return sendTo("POST", url, payload)
}

func putTo(url string, payload util.JSON) (*http.Response, error) {
	return sendTo("PUT", url, payload)
}

func sendTo(method, url string, payload util.JSON) (*http.Response, error) {
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

func getClanPayload(ownerID, clanPublicID string) util.JSON {
	return util.JSON{
		"publicID":         clanPublicID,
		"name":             clanPublicID,
		"ownerPublicID":    ownerID,
		"metadata":         util.JSON{"x": 1},
		"allowApplication": true,
		"autoJoin":         true,
	}
}

func getGameAndPlayer(db models.DB) (*models.Game, *models.Player, error) {
	game := models.GameFactory.MustCreateWithOption(util.JSON{
		"PublicID": uuid.NewV4().String(),
	}).(*models.Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, err
	}
	player := models.PlayerFactory.MustCreateWithOption(util.JSON{
		"GameID":   game.PublicID,
		"PublicID": uuid.NewV4().String(),
	}).(*models.Player)
	err = db.Insert(player)
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
		fmt.Printf("Request failed with status code", res.StatusCode)
		panic(string(bts))
	}
}

func createClans(db models.DB, game *models.Game, owner *models.Player, numberOfClans int) ([]*models.Clan, error) {
	var clans []*models.Clan
	for i := 0; i < numberOfClans; i++ {
		clan := models.ClanFactory.MustCreateWithOption(util.JSON{
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
