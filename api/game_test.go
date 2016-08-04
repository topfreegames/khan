// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
)

func getGamePayload(publicID, name string) map[string]interface{} {
	if publicID == "" {
		publicID = randomdata.FullName(randomdata.RandomGender)
	}
	if name == "" {
		name = randomdata.FullName(randomdata.RandomGender)
	}
	return map[string]interface{}{
		"publicID":                      publicID,
		"name":                          name,
		"membershipLevels":              map[string]interface{}{"Member": 1, "Elder": 2, "CoLeader": 3},
		"metadata":                      map[string]interface{}{"x": "a"},
		"minLevelToAcceptApplication":   1,
		"minLevelToCreateInvitation":    2,
		"minLevelToRemoveMember":        3,
		"minLevelOffsetToRemoveMember":  1,
		"minLevelOffsetToPromoteMember": 1,
		"minLevelOffsetToDemoteMember":  1,
		"maxMembers":                    100,
		"maxClansPerPlayer":             1,
		"cooldownAfterDeny":             30,
		"cooldownAfterDelete":           30,
	}
}

var _ = Describe("Game API Handler", func() {
	var testDb models.DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create Game Handler", func() {
		It("Should create game", func() {
			a := GetDefaultTestApp()

			payload := getGamePayload("", "")
			res := PostJSON(a, "/games", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(payload["publicID"].(string)))

			dbGame, err := models.GetGameByPublicID(a.Db, payload["publicID"].(string))
			Expect(err).NotTo(HaveOccurred())
			Expect(dbGame.PublicID).To(Equal(payload["publicID"]))
			Expect(dbGame.Name).To(Equal(payload["name"]))
			membershipLevels := payload["membershipLevels"].(map[string]interface{})
			for k, v := range membershipLevels {
				Expect(v.(int)).To(Equal(int(dbGame.MembershipLevels[k].(float64))))
			}
			Expect(dbGame.Metadata).To(Equal(payload["metadata"]))
			Expect(dbGame.MinMembershipLevel).To(Equal(1))
			Expect(dbGame.MaxMembershipLevel).To(Equal(3))
			Expect(dbGame.MinLevelToAcceptApplication).To(Equal(payload["minLevelToAcceptApplication"]))
			Expect(dbGame.MinLevelToCreateInvitation).To(Equal(payload["minLevelToCreateInvitation"]))
			Expect(dbGame.MinLevelOffsetToPromoteMember).To(Equal(payload["minLevelOffsetToPromoteMember"]))
			Expect(dbGame.MinLevelOffsetToDemoteMember).To(Equal(payload["minLevelOffsetToDemoteMember"]))
			Expect(dbGame.MaxMembers).To(Equal(payload["maxMembers"]))
			Expect(dbGame.MaxClansPerPlayer).To(Equal(payload["maxClansPerPlayer"]))
			Expect(dbGame.CooldownAfterDeny).To(Equal(payload["cooldownAfterDeny"]))
			Expect(dbGame.CooldownAfterDelete).To(Equal(payload["cooldownAfterDelete"]))
			Expect(dbGame.CooldownBeforeInvite).To(Equal(0))
			Expect(dbGame.CooldownBeforeApply).To(Equal(3600))
			Expect(dbGame.MaxPendingInvites).To(Equal(-1))
		})

		It("Should create game with custom optional params", func() {
			a := GetDefaultTestApp()

			payload := getGamePayload("", "")
			payload["maxPendingInvites"] = 27
			payload["cooldownBeforeApply"] = 2874
			payload["cooldownBeforeInvite"] = 2384
			res := PostJSON(a, "/games", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(payload["publicID"].(string)))

			dbGame, err := models.GetGameByPublicID(a.Db, payload["publicID"].(string))
			Expect(err).NotTo(HaveOccurred())
			Expect(dbGame.CooldownBeforeInvite).To(Equal(2384))
			Expect(dbGame.CooldownBeforeApply).To(Equal(2874))
			Expect(dbGame.MaxPendingInvites).To(Equal(27))
		})

		It("Should not create game if missing parameters", func() {
			a := GetDefaultTestApp()
			payload := getGamePayload("", "")
			delete(payload, "maxMembers")
			res := PostJSON(a, "/games", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("maxMembers is required"))
		})

		It("Should not create game if bad payload", func() {
			a := GetDefaultTestApp()
			payload := getGamePayload("", "")
			payload["minLevelToCreateInvitation"] = 0
			res := PostJSON(a, "/games", payload)

			Expect(res.Raw().StatusCode).To(Equal(422))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("minLevelToCreateInvitation should be greater or equal to minMembershipLevel"))
		})

		It("Should not create game if invalid payload", func() {
			a := GetDefaultTestApp()
			res := PostBody(a, "/games", "invalid")

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(strings.Contains(result["reason"].(string), "While trying to read JSON")).To(BeTrue())
		})

		It("Should not create game if invalid data", func() {
			a := GetDefaultTestApp()
			payload := getGamePayload("game-id-is-too-large-for-this-field-should-be-less-than-36-chars", "")
			res := PostJSON(a, "/games", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("pq: value too long for type character varying(36)"))
		})
	})

	Describe("Update Game Handler", func() {
		It("Should update game", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			metadata := map[string]interface{}{"y": "10"}
			payload := getGamePayload(game.PublicID, game.Name)
			payload["metadata"] = metadata
			payload["cooldownAfterDeny"] = 0

			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			dbGame, err := models.GetGameByPublicID(a.Db, game.PublicID)
			Expect(err).NotTo(HaveOccurred())

			Expect(dbGame.Metadata).To(Equal(metadata))
			membershipLevels := payload["membershipLevels"].(map[string]interface{})
			for k, v := range membershipLevels {
				Expect(v.(int)).To(Equal(int(dbGame.MembershipLevels[k].(float64))))
			}
			Expect(dbGame.PublicID).To(Equal(game.PublicID))
			Expect(dbGame.Name).To(Equal(game.Name))
			Expect(dbGame.MinMembershipLevel).To(Equal(1))
			Expect(dbGame.MaxMembershipLevel).To(Equal(3))
			Expect(dbGame.MinLevelToAcceptApplication).To(Equal(payload["minLevelToAcceptApplication"]))
			Expect(dbGame.MinLevelToCreateInvitation).To(Equal(payload["minLevelToCreateInvitation"]))
			Expect(dbGame.MinLevelOffsetToPromoteMember).To(Equal(payload["minLevelOffsetToPromoteMember"]))
			Expect(dbGame.MinLevelOffsetToDemoteMember).To(Equal(payload["minLevelOffsetToDemoteMember"]))
			Expect(dbGame.MaxMembers).To(Equal(payload["maxMembers"]))
			Expect(dbGame.MaxClansPerPlayer).To(Equal(payload["maxClansPerPlayer"]))
			Expect(dbGame.CooldownAfterDeny).To(Equal(payload["cooldownAfterDeny"]))
			Expect(dbGame.CooldownAfterDelete).To(Equal(payload["cooldownAfterDelete"]))
		})

		It("Should insert if game does not exist", func() {
			a := GetDefaultTestApp()

			gameID := uuid.NewV4().String()
			payload := getGamePayload(gameID, gameID)

			route := fmt.Sprintf("/games/%s", gameID)
			res := PutJSON(a, route, payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			dbGame, err := models.GetGameByPublicID(a.Db, gameID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbGame.Metadata).To(Equal(payload["metadata"]))
			Expect(dbGame.PublicID).To(Equal(gameID))
			Expect(dbGame.Name).To(Equal(payload["name"]))
		})

		It("Should not update game if missing parameters", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			metadata := map[string]interface{}{"y": "10"}
			payload := getGamePayload(game.PublicID, game.Name)
			payload["metadata"] = metadata
			delete(payload, "name")
			delete(payload, "minLevelOffsetToPromoteMember")
			delete(payload, "maxMembers")

			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("name is required, minLevelOffsetToPromoteMember is required, maxMembers is required"))
		})

		It("Should not update game if bad payload", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			payload := getGamePayload(game.PublicID, game.Name)
			payload["minLevelToCreateInvitation"] = 0

			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, payload)

			Expect(res.Raw().StatusCode).To(Equal(422))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("minLevelToCreateInvitation should be greater or equal to minMembershipLevel"))
		})

		It("Should not update game if invalid payload", func() {
			a := GetDefaultTestApp()
			res := PutBody(a, "/games/game-id", "invalid")

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(strings.Contains(result["reason"].(string), "While trying to read JSON")).To(BeTrue())
		})

		It("Should not update game if invalid data", func() {
			a := GetDefaultTestApp()

			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			payload := getGamePayload(game.PublicID, strings.Repeat("a", 256))

			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("pq: value too long for type character varying(255)"))
		})
	})

	Describe("Game Hooks", func() {
		Describe("Update Game Hook", func() {
			It("Should call update game hook", func() {
				hooks, err := models.GetHooksForRoutes(testDb, []string{
					"http://localhost:52525/update",
				}, models.GameUpdatedHook)
				Expect(err).NotTo(HaveOccurred())
				responses := startRouteHandler([]string{"/update"}, 52525)

				app := GetDefaultTestApp()
				time.Sleep(time.Second)

				gameID := hooks[0].GameID

				payload := getGamePayload(gameID, uuid.NewV4().String())

				route := fmt.Sprintf("/games/%s", gameID)
				res := PutJSON(app, route, payload)
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
				var result map[string]interface{}
				json.Unmarshal([]byte(res.Body().Raw()), &result)
				Expect(result["success"]).To(BeTrue())

				app.Dispatcher.Wait()

				Expect(len(*responses)).To(Equal(1))
			})
		})
	})
})
