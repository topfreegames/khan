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

	"github.com/Pallinder/go-randomdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/khan/api"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
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
		"cooldownAfterDelete":           40,
	}
}

var _ = Describe("Game API Handler", func() {
	var testDb, db models.DB
	var a *api.App

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())

		a = GetDefaultTestApp()
		db = a.Db(nil)
		fixtures.ConfigureAndStartGoWorkers()
	})

	Describe("Create Game Handler", func() {
		It("Should create game", func() {
			payload := getGamePayload("", "")
			status, body := PostJSON(a, "/games", payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(payload["publicID"].(string)))

			dbGame, err := models.GetGameByPublicID(db, payload["publicID"].(string))
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
			Expect(dbGame.PlayerUpdateMetadataFieldsHookTriggerWhitelist).To(BeEmpty())
			Expect(dbGame.ClanUpdateMetadataFieldsHookTriggerWhitelist).To(BeEmpty())
			Expect(dbGame.CooldownBeforeInvite).To(Equal(0))
			Expect(dbGame.CooldownBeforeApply).To(Equal(3600))
			Expect(dbGame.MaxPendingInvites).To(Equal(-1))
		})

		It("Should create game with custom optional params", func() {
			payload := getGamePayload("", "")
			payload["maxPendingInvites"] = 27
			payload["cooldownBeforeApply"] = 2874
			payload["cooldownBeforeInvite"] = 2384
			payload["playerHookFieldsWhitelist"] = "a,b"
			payload["clanHookFieldsWhitelist"] = "c,d"
			status, body := PostJSON(a, "/games", payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(payload["publicID"].(string)))

			dbGame, err := models.GetGameByPublicID(db, payload["publicID"].(string))
			Expect(err).NotTo(HaveOccurred())
			Expect(dbGame.CooldownBeforeInvite).To(Equal(2384))
			Expect(dbGame.CooldownBeforeApply).To(Equal(2874))
			Expect(dbGame.MaxPendingInvites).To(Equal(27))
			Expect(dbGame.PlayerUpdateMetadataFieldsHookTriggerWhitelist).To(Equal(payload["playerHookFieldsWhitelist"]))
			Expect(dbGame.ClanUpdateMetadataFieldsHookTriggerWhitelist).To(Equal(payload["clanHookFieldsWhitelist"]))
		})

		It("Should not create game if missing parameters", func() {
			payload := getGamePayload("", "")
			delete(payload, "maxMembers")
			status, body := PostJSON(a, "/games", payload)

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("maxMembers is required"))
		})

		It("Should not create game if bad payload", func() {
			payload := getGamePayload("", "")
			payload["minLevelToCreateInvitation"] = 0
			status, body := PostJSON(a, "/games", payload)

			Expect(status).To(Equal(400))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("minLevelToCreateInvitation should be greater or equal to minMembershipLevel"))
		})

		It("Should not create game if invalid payload", func() {
			status, body := Post(a, "/games", "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"].(string)).To(ContainSubstring(InvalidJSONError))
		})

		It("Should not create game and not panic if nil MembershipLevels", func() {
			payload := getGamePayload("", "")
			payload["membershipLevels"] = nil
			status, body := PostJSON(a, "/games", payload)

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"].(string)).To(Equal("membershipLevels is required"))
		})

		It("Should not create game and not panic if empty MembershipLevels", func() {
			payload := getGamePayload("", "")
			payload["membershipLevels"] = make(map[string]interface{}, 0)
			status, body := PostJSON(a, "/games", payload)

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"].(string)).To(Equal("membershipLevels is required"))
		})

		It("Should not create game if invalid data", func() {
			payload := getGamePayload("game-id-is-too-large-for-this-field-should-be-less-than-36-chars", "")
			status, body := PostJSON(a, "/games", payload)

			Expect(status).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("pq: value too long for type character varying(36)"))
		})
	})

	Describe("Update Game Handler", func() {
		It("Should update game", func() {
			game := fixtures.GameFactory.MustCreate().(*models.Game)
			err := db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			metadata := map[string]interface{}{"y": "10"}
			payload := getGamePayload(game.PublicID, game.Name)
			payload["metadata"] = metadata
			payload["cooldownAfterDeny"] = 0

			route := fmt.Sprintf("/games/%s", game.PublicID)
			status, body := PutJSON(a, route, payload)
			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbGame, err := models.GetGameByPublicID(db, game.PublicID)
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
			gameID := uuid.NewV4().String()
			payload := getGamePayload(gameID, gameID)

			route := fmt.Sprintf("/games/%s", gameID)
			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbGame, err := models.GetGameByPublicID(db, gameID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbGame.Metadata).To(Equal(payload["metadata"]))
			Expect(dbGame.PublicID).To(Equal(gameID))
			Expect(dbGame.Name).To(Equal(payload["name"]))
		})

		It("Should not update game if missing parameters", func() {
			game := fixtures.GameFactory.MustCreate().(*models.Game)
			err := db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			metadata := map[string]interface{}{"y": "10"}
			payload := getGamePayload(game.PublicID, game.Name)
			payload["metadata"] = metadata
			delete(payload, "name")
			delete(payload, "minLevelOffsetToPromoteMember")
			delete(payload, "maxMembers")

			route := fmt.Sprintf("/games/%s", game.PublicID)
			status, body := PutJSON(a, route, payload)
			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("name is required, minLevelOffsetToPromoteMember is required, maxMembers is required"))
		})

		It("Should not update game if bad payload", func() {
			game := fixtures.GameFactory.MustCreate().(*models.Game)
			err := db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			payload := getGamePayload(game.PublicID, game.Name)
			payload["minLevelToCreateInvitation"] = 0

			route := fmt.Sprintf("/games/%s", game.PublicID)
			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(400))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("minLevelToCreateInvitation should be greater or equal to minMembershipLevel"))
		})

		It("Should not update game and not panic if nil MembershipLevels", func() {
			game := fixtures.GameFactory.MustCreate().(*models.Game)
			err := db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			payload := getGamePayload(game.PublicID, game.Name)
			payload["membershipLevels"] = nil

			route := fmt.Sprintf("/games/%s", game.PublicID)
			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(400))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("membershipLevels is required"))
		})

		It("Should not update game and not panic if empty MembershipLevels", func() {
			game := fixtures.GameFactory.MustCreate().(*models.Game)
			err := db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			payload := getGamePayload(game.PublicID, game.Name)
			payload["membershipLevels"] = make(map[string]interface{}, 0)

			route := fmt.Sprintf("/games/%s", game.PublicID)
			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(400))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("membershipLevels is required"))
		})

		It("Should not update game if invalid payload", func() {
			status, body := Put(a, "/games/game-id", "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"].(string)).To(ContainSubstring(InvalidJSONError))
		})

		It("Should not update game if invalid data", func() {
			game := fixtures.GameFactory.MustCreate().(*models.Game)
			err := db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			payload := getGamePayload(game.PublicID, strings.Repeat("a", 256))

			route := fmt.Sprintf("/games/%s", game.PublicID)
			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("pq: value too long for type character varying(255)"))
		})
	})

	Describe("Game Hooks", func() {
		Describe("Update Game Hook", func() {
			It("Should call update game hook", func() {
				hooks, err := fixtures.GetHooksForRoutes(testDb, []string{
					"http://localhost:52525/update",
				}, models.GameUpdatedHook)
				Expect(err).NotTo(HaveOccurred())
				responses := startRouteHandler([]string{"/update"}, 52525)

				gameID := hooks[0].GameID

				payload := getGamePayload(gameID, uuid.NewV4().String())

				route := fmt.Sprintf("/games/%s", gameID)
				status, body := PutJSON(a, route, payload)
				Expect(status).To(Equal(http.StatusOK))
				var result map[string]interface{}
				json.Unmarshal([]byte(body), &result)
				Expect(result["success"]).To(BeTrue())

				Eventually(func() int {
					return len(*responses)
				}).Should(Equal(1))
			})
		})
	})
})
