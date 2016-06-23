// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/util"
)

func getGamePayload(publicID, name string) util.JSON {
	if publicID == "" {
		publicID = randomdata.FullName(randomdata.RandomGender)
	}
	if name == "" {
		name = randomdata.FullName(randomdata.RandomGender)
	}
	return util.JSON{
		"publicID":                      publicID,
		"name":                          name,
		"membershipLevels":              util.JSON{"Member": 1, "Elder": 2, "CoLeader": 3},
		"metadata":                      util.JSON{"x": "a"},
		"minMembershipLevel":            1,
		"maxMembershipLevel":            10,
		"minLevelToAcceptApplication":   1,
		"minLevelToCreateInvitation":    1,
		"minLevelToRemoveMember":        1,
		"minLevelOffsetToRemoveMember":  1,
		"minLevelOffsetToPromoteMember": 1,
		"minLevelOffsetToDemoteMember":  1,
		"maxMembers":                    100,
	}
}

func TestGameHandler(t *testing.T) {
	t.Parallel()
	g := Goblin(t)

	testDb, err := models.GetTestDB()
	g.Assert(err == nil).IsTrue()

	g.Describe("Create Game Handler", func() {
		g.It("Should create game", func() {
			a := GetDefaultTestApp()

			payload := getGamePayload("", "")
			res := PostJSON(a, "/games", t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()
			g.Assert(result["publicID"]).Equal(payload["publicID"].(string))

			dbGame, err := models.GetGameByPublicID(a.Db, payload["publicID"].(string))
			AssertNotError(g, err)
			g.Assert(dbGame.PublicID).Equal(payload["publicID"])
			g.Assert(dbGame.Name).Equal(payload["name"])
			membershipLevels := payload["membershipLevels"].(util.JSON)
			for k, v := range membershipLevels {
				g.Assert(v.(int)).Equal(int(dbGame.MembershipLevels[k].(float64)))
			}
			g.Assert(dbGame.Metadata).Equal(payload["metadata"])
			g.Assert(dbGame.MinMembershipLevel).Equal(payload["minMembershipLevel"])
			g.Assert(dbGame.MaxMembershipLevel).Equal(payload["maxMembershipLevel"])
			g.Assert(dbGame.MinLevelToAcceptApplication).Equal(payload["minLevelToAcceptApplication"])
			g.Assert(dbGame.MinLevelToCreateInvitation).Equal(payload["minLevelToCreateInvitation"])
			g.Assert(dbGame.MinLevelOffsetToPromoteMember).Equal(payload["minLevelOffsetToPromoteMember"])
			g.Assert(dbGame.MinLevelOffsetToDemoteMember).Equal(payload["minLevelOffsetToDemoteMember"])
			g.Assert(dbGame.MaxMembers).Equal(payload["maxMembers"])
		})

		g.It("Should not create game if missing parameters", func() {
			a := GetDefaultTestApp()
			payload := getGamePayload("", "")
			delete(payload, "maxMembers")
			res := PostJSON(a, "/games", t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("maxMembers is required")
		})

		g.It("Should not create game if bad payload", func() {
			a := GetDefaultTestApp()
			payload := getGamePayload("", "")
			payload["minMembershipLevel"] = payload["maxMembershipLevel"].(int) + 1
			res := PostJSON(a, "/games", t, payload)

			g.Assert(res.Raw().StatusCode).Equal(422)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("maxMembershipLevel should be greater or equal to minMembershipLevel, minLevelToAcceptApplication should be greater or equal to minMembershipLevel, minLevelToCreateInvitation should be greater or equal to minMembershipLevel, minLevelToRemoveMember should be greater or equal to minMembershipLevel")
		})

		g.It("Should not create game if invalid payload", func() {
			a := GetDefaultTestApp()
			res := PostBody(a, "/games", t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not create game if invalid data", func() {
			a := GetDefaultTestApp()
			payload := getGamePayload("game-id-is-too-large-for-this-field-should-be-less-than-36-chars", "")
			res := PostJSON(a, "/games", t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: value too long for type character varying(36)")
		})
	})

	g.Describe("Update Game Handler", func() {
		g.It("Should update game", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			AssertNotError(g, err)

			metadata := util.JSON{"y": "10"}
			payload := getGamePayload(game.PublicID, game.Name)
			payload["metadata"] = metadata

			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, t, payload)
			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbGame, err := models.GetGameByPublicID(a.Db, game.PublicID)
			AssertNotError(g, err)
			g.Assert(dbGame.Metadata).Equal(metadata)
			membershipLevels := payload["membershipLevels"].(util.JSON)
			for k, v := range membershipLevels {
				g.Assert(v.(int)).Equal(int(dbGame.MembershipLevels[k].(float64)))
			}
			g.Assert(dbGame.PublicID).Equal(game.PublicID)
			g.Assert(dbGame.Name).Equal(game.Name)
			g.Assert(dbGame.MinMembershipLevel).Equal(payload["minMembershipLevel"])
			g.Assert(dbGame.MaxMembershipLevel).Equal(payload["maxMembershipLevel"])
			g.Assert(dbGame.MinLevelToAcceptApplication).Equal(payload["minLevelToAcceptApplication"])
			g.Assert(dbGame.MinLevelToCreateInvitation).Equal(payload["minLevelToCreateInvitation"])
			g.Assert(dbGame.MinLevelOffsetToPromoteMember).Equal(payload["minLevelOffsetToPromoteMember"])
			g.Assert(dbGame.MinLevelOffsetToDemoteMember).Equal(payload["minLevelOffsetToDemoteMember"])
			g.Assert(dbGame.MaxMembers).Equal(payload["maxMembers"])
		})

		g.It("Should insert if game does not exist", func() {
			a := GetDefaultTestApp()

			gameID := uuid.NewV4().String()
			payload := getGamePayload(gameID, gameID)

			route := fmt.Sprintf("/games/%s", gameID)
			res := PutJSON(a, route, t, payload)
			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbGame, err := models.GetGameByPublicID(a.Db, gameID)
			AssertNotError(g, err)
			g.Assert(dbGame.Metadata).Equal(payload["metadata"])
			g.Assert(dbGame.PublicID).Equal(gameID)
			g.Assert(dbGame.Name).Equal(payload["name"])
		})

		g.It("Should not update game if missing parameters", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			AssertNotError(g, err)

			metadata := util.JSON{"y": "10"}
			payload := getGamePayload(game.PublicID, game.Name)
			payload["metadata"] = metadata
			delete(payload, "name")
			delete(payload, "minLevelOffsetToPromoteMember")
			delete(payload, "maxMembers")

			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, t, payload)
			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("name is required, minLevelOffsetToPromoteMember is required, maxMembers is required")
		})

		g.It("Should not update game if bad payload", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			AssertNotError(g, err)

			payload := getGamePayload(game.PublicID, game.Name)
			payload["minMembershipLevel"] = payload["maxMembershipLevel"].(int) + 1

			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(422)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("maxMembershipLevel should be greater or equal to minMembershipLevel, minLevelToAcceptApplication should be greater or equal to minMembershipLevel, minLevelToCreateInvitation should be greater or equal to minMembershipLevel, minLevelToRemoveMember should be greater or equal to minMembershipLevel")
		})

		g.It("Should not update game if invalid payload", func() {
			a := GetDefaultTestApp()
			res := PutBody(a, "/games/game-id", t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not update game if invalid data", func() {
			a := GetDefaultTestApp()

			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			AssertNotError(g, err)

			payload := getGamePayload(game.PublicID, strings.Repeat("a", 256))

			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: value too long for type character varying(255)")
		})
	})

	g.Describe("Game Hooks", func() {
		g.Describe("Update Game Hook", func() {
			g.It("Should call update game hook", func() {
				hooks, err := models.GetHooksForRoutes(testDb, []string{
					"http://localhost:52525/update",
				}, models.GameUpdatedHook)
				g.Assert(err == nil).IsTrue()
				responses := startRouteHandler([]string{"/update"}, 52525)

				app := GetDefaultTestApp()
				time.Sleep(time.Second)

				gameID := hooks[0].GameID

				payload := getGamePayload(gameID, uuid.NewV4().String())

				route := fmt.Sprintf("/games/%s", gameID)
				res := PutJSON(app, route, t, payload)
				g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
				var result util.JSON
				json.Unmarshal([]byte(res.Body().Raw()), &result)
				g.Assert(result["success"]).IsTrue()

				g.Assert(len(*responses)).Equal(1)
			})
		})
	})
}
