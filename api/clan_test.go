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
	"net/http/httptest"
	"sort"
	"strings"

	"github.com/Pallinder/go-randomdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/api"
	"github.com/topfreegames/khan/models"
)

var _ = Describe("Clan API Handler", func() {
	var testDb models.DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create Clan Handler", func() {
		It("Should create clan", func() {
			_, player, err := models.CreatePlayerFactory(testDb, "")
			Expect(err).NotTo(HaveOccurred())

			clanPublicID := randomdata.FullName(randomdata.RandomGender)
			payload := map[string]interface{}{
				"publicID":         clanPublicID,
				"name":             randomdata.FullName(randomdata.RandomGender),
				"ownerPublicID":    player.PublicID,
				"metadata":         map[string]interface{}{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			a := GetDefaultTestApp()
			status, body := PostJSON(a, GetGameRoute(player.GameID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(clanPublicID))

			dbClan, err := models.GetClanByPublicID(a.Db, player.GameID, clanPublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbClan.GameID).To(Equal(player.GameID))
			Expect(dbClan.OwnerID).To(Equal(player.ID))
			Expect(dbClan.PublicID).To(Equal(payload["publicID"]))
			Expect(dbClan.Name).To(Equal(payload["name"]))
			Expect(dbClan.Metadata["x"]).To(
				BeEquivalentTo(payload["metadata"].(map[string]interface{})["x"]),
			)
			Expect(dbClan.AllowApplication).To(Equal(payload["allowApplication"]))
			Expect(dbClan.AutoJoin).To(Equal(payload["autoJoin"]))
		})

		It("Should not create clan if missing parameters", func() {
			_, player, err := models.CreatePlayerFactory(testDb, "")
			Expect(err).NotTo(HaveOccurred())

			clanPublicID := randomdata.FullName(randomdata.RandomGender)
			payload := map[string]interface{}{
				"publicID":         clanPublicID,
				"allowApplication": true,
				"autoJoin":         true,
			}
			a := GetDefaultTestApp()
			status, body := PostJSON(a, GetGameRoute(player.GameID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("name is required, ownerPublicID is required")))
		})

		It("Should not create clan if invalid payload", func() {
			gameID := "gameID"
			a := GetDefaultTestApp()
			status, body := Post(a, GetGameRoute(gameID, "/clans"), "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring(InvalidJSONError))
		})

		It("Should not create clan if owner does not exist", func() {
			game, _, err := models.CreatePlayerFactory(testDb, "")
			Expect(err).NotTo(HaveOccurred())

			payload := map[string]interface{}{
				"publicID":         randomdata.FullName(randomdata.RandomGender),
				"name":             randomdata.FullName(randomdata.RandomGender),
				"ownerPublicID":    randomdata.FullName(randomdata.RandomGender),
				"metadata":         map[string]interface{}{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			a := GetDefaultTestApp()
			status, body := PostJSON(a, GetGameRoute(game.PublicID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Player was not found with id: %s", payload["ownerPublicID"])))
		})

		It("Should not create clan if invalid data", func() {
			_, player, err := models.CreatePlayerFactory(testDb, "")
			Expect(err).NotTo(HaveOccurred())

			payload := map[string]interface{}{
				"publicID":         randomdata.FullName(randomdata.RandomGender),
				"name":             strings.Repeat("a", 256),
				"ownerPublicID":    player.PublicID,
				"metadata":         map[string]interface{}{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			a := GetDefaultTestApp()
			status, body := PostJSON(a, GetGameRoute(player.GameID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("pq: value too long for type character varying(255)"))
		})
	})

	Describe("Leave Clan Handler", func() {
		It("Should leave a clan and transfer ownership", func() {
			_, clan, owner, players, memberships, err := models.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/leave", clan.PublicID))
			a := GetDefaultTestApp()
			status, body := PostJSON(a, route, map[string]interface{}{})

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["previousOwner"]).NotTo(BeNil())
			Expect(result["newOwner"]).NotTo(BeNil())

			prevOwner := result["previousOwner"].(map[string]interface{})
			Expect(prevOwner["publicID"]).To(BeEquivalentTo(owner.PublicID))

			newOwner := result["newOwner"].(map[string]interface{})
			Expect(newOwner["publicID"]).To(BeEquivalentTo(players[0].PublicID))

			dbClan, err := models.GetClanByPublicID(a.Db, clan.GameID, clan.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbClan.OwnerID).To(Equal(memberships[0].PlayerID))
		})

		It("Should not leave a clan if invalid clan", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/leave", "random-id"))
			a := GetDefaultTestApp()
			status, body := Post(a, route, "")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(strings.Contains(result["reason"].(string), "Clan was not found with id: random-id")).To(BeTrue())
		})
	})

	Describe("Transfer Clan Ownership Handler", func() {
		It("Should transfer a clan ownership", func() {
			_, clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())
			ownerPublicID := owner.PublicID
			playerPublicID := players[0].PublicID

			payload := map[string]interface{}{
				"ownerPublicID":  ownerPublicID,
				"playerPublicID": playerPublicID,
			}
			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/transfer-ownership", clan.PublicID))
			a := GetDefaultTestApp()
			status, body := PostJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbClan, err := models.GetClanByPublicID(a.Db, clan.GameID, clan.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbClan.OwnerID).To(Equal(players[0].ID))
		})

		It("Should not transfer a clan ownership if missing parameters", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/transfer-ownership", "public-id"))
			a := GetDefaultTestApp()
			status, body := PostJSON(a, route, map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("playerPublicID is required"))

		})

		It("Should not transfer a clan ownership if invalid payload", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/transfer-ownership", "random-id"))
			a := GetDefaultTestApp()
			status, body := Post(a, route, "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring(InvalidJSONError))
		})
	})

	Describe("Update Clan Handler", func() {
		It("Should update clan", func() {
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := owner.PublicID
			metadata := map[string]interface{}{"new": "metadata"}

			payload := map[string]interface{}{
				"name":             clanName,
				"ownerPublicID":    ownerPublicID,
				"metadata":         metadata,
				"allowApplication": !clan.AllowApplication,
				"autoJoin":         !clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			a := GetDefaultTestApp()
			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbClan, err := models.GetClanByPublicID(a.Db, gameID, publicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbClan.GameID).To(Equal(gameID))
			Expect(dbClan.PublicID).To(Equal(publicID))
			Expect(dbClan.Name).To(Equal(clanName))
			Expect(dbClan.OwnerID).To(Equal(owner.ID))
			Expect(dbClan.Metadata).To(Equal(metadata))
			Expect(dbClan.AllowApplication).To(Equal(!clan.AllowApplication))
			Expect(dbClan.AutoJoin).To(Equal(!clan.AutoJoin))
		})

		It("Should not update clan if missing parameters", func() {
			route := GetGameRoute("gameID", fmt.Sprintf("/clans/%s", "publicID"))
			a := GetDefaultTestApp()
			status, body := PutJSON(a, route, map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("name is required, ownerPublicID is required"))
		})

		It("Should not update clan if invalid payload", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("/clans/%s", "random-id"))
			a := GetDefaultTestApp()

			status, body := Put(a, route, "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring(InvalidJSONError))
		})

		It("Should not update clan if player is not the owner", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			publicID := clan.PublicID
			metadata := map[string]interface{}{"x": 1}

			payload := map[string]interface{}{
				"name":             clan.Name,
				"ownerPublicID":    randomdata.FullName(randomdata.RandomGender),
				"metadata":         metadata,
				"allowApplication": clan.AllowApplication,
				"autoJoin":         clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			a := GetDefaultTestApp()

			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID)))
		})

		It("Should not update clan if invalid data", func() {
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			publicID := clan.PublicID

			payload := map[string]interface{}{
				"name":             strings.Repeat("s", 256),
				"ownerPublicID":    owner.PublicID,
				"metadata":         map[string]interface{}{"x": 1},
				"allowApplication": clan.AllowApplication,
				"autoJoin":         clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			a := GetDefaultTestApp()

			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("pq: value too long for type character varying(255)"))
		})

		Describe("Measure", func() {
			hasRun := false
			var ts *httptest.Server
			var app *api.App
			var route string
			var payloadJSON string

			BeforeEach(func() {
				if !hasRun {
					_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					gameID := clan.GameID
					publicID := clan.PublicID
					clanName := randomdata.FullName(randomdata.RandomGender)
					ownerPublicID := owner.PublicID
					metadata := map[string]interface{}{"new": "metadata"}

					payload := map[string]interface{}{
						"name":             clanName,
						"ownerPublicID":    ownerPublicID,
						"metadata":         metadata,
						"allowApplication": !clan.AllowApplication,
						"autoJoin":         !clan.AutoJoin,
					}
					payloadJSONB, err := json.Marshal(payload)
					payloadJSON = string(payloadJSONB)
					Expect(err).NotTo(HaveOccurred())
					route = GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))

					app = GetDefaultTestApp()
					ts = InitializeTestServer(app)

					hasRun = true
				}
			})
			Measure("Should update a Clan with UpdateClan", func(b Benchmarker) {
				req := GetRequest(app, ts, "PUT", route, payloadJSON)
				runtime := b.Time("runtime", func() {
					status, _ := PerformRequest(ts, req)
					Expect(status).To(Equal(http.StatusOK))
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 1), "Operation shouldn't take this long")
			}, 200)
		})
	})

	Describe("List All Clans Handler", func() {
		It("Should get all clans", func() {
			player, expectedClans, err := models.GetTestClans(testDb, "", "", 10)
			Expect(err).NotTo(HaveOccurred())
			sort.Sort(models.ClanByName(expectedClans))

			a := GetDefaultTestApp()

			status, body := Get(a, GetGameRoute(player.GameID, "/clans"))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())
			for index, clanObj := range result["clans"].([]interface{}) {
				clan := clanObj.(map[string]interface{}) // Can't be map[string]interface{}
				Expect(clan["name"]).To(Equal(expectedClans[index].Name))
				clanMetadata := clan["metadata"].(map[string]interface{})
				for k, v := range clanMetadata {
					Expect(v).To(Equal(expectedClans[index].Metadata[k]))
				}
				Expect(clan["publicID"]).To(Equal(expectedClans[index].PublicID))
				Expect(clan["ID"]).To(BeNil())
				Expect(clan["autoJoin"]).To(Equal(expectedClans[index].AutoJoin))
				Expect(clan["allowApplication"]).To(Equal(expectedClans[index].AllowApplication))
			}
		})

		It("Should return empty list if invalid game query", func() {
			a := GetDefaultTestApp()

			status, body := Get(a, GetGameRoute("invalid-query-game-id", "/clans"))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())
			Expect(len(result["clans"].([]interface{}))).To(Equal(0))
		})
	})

	Describe("Retrieve Clan Handler", func() {
		It("Should get details for clan", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			a := GetDefaultTestApp()

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())

			Expect(result["name"]).To(Equal(clan.Name))
			resultMetadata := result["metadata"].(map[string]interface{})
			for k, v := range resultMetadata {
				Expect(v).To(Equal(clan.Metadata[k]))
			}
			Expect(result["publicID"]).To(BeNil())
		})

		It("Should get clan members", func() {
			gameID := uuid.NewV4().String()
			_, clan, _, _, _, err := models.GetClanWithMemberships(
				testDb, 10, 0, 0, 0, gameID, "clan-details-api-clan",
			)
			Expect(err).NotTo(HaveOccurred())

			a := GetDefaultTestApp()

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())

			Expect(result["roster"] == nil).To(BeFalse())
		})
	})

	Describe("Retrieve Clans Handler", func() {
		It("Should get details for clans", func() {
			a := GetDefaultTestApp()

			gameID := uuid.NewV4().String()
			_, clan1, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID, "clan1")
			Expect(err).NotTo(HaveOccurred())
			_, clan2, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID, "clan2", true)
			Expect(err).NotTo(HaveOccurred())
			_, clan3, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID, "clan3", true)
			Expect(err).NotTo(HaveOccurred())

			clanIDs := []string{clan1.PublicID, clan2.PublicID, clan3.PublicID}

			url := fmt.Sprintf(
				"%s?clanPublicIds=%s",
				GetGameRoute(clan1.GameID, "clans-summary"),
				strings.Join(clanIDs, ","),
			)
			status, body := Get(a, url)
			Expect(status).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			resultClans := result["clans"].([]interface{})
			Expect(len(resultClans)).To(Equal(3))

			for _, resultClan := range resultClans {
				resultClanMap := resultClan.(map[string]interface{})
				Expect(resultClanMap["membershipCount"] == nil).To(BeFalse())
				Expect(resultClanMap["publicID"] == nil).To(BeFalse())
				Expect(resultClanMap["metadata"] == nil).To(BeFalse())
				Expect(resultClanMap["name"] == nil).To(BeFalse())
				Expect(resultClanMap["allowApplication"] == nil).To(BeFalse())
				Expect(resultClanMap["autoJoin"] == nil).To(BeFalse())
				Expect(len(resultClanMap)).To(Equal(6))

				idExist := false
				// check if publicID is in clanIDs
				for _, clanID := range clanIDs {
					if resultClanMap["publicID"] == clanID {
						idExist = true
					}
				}
				Expect(idExist).To(BeTrue())
			}
		})

		It("Should not get details for clans for unexistent game", func() {
			a := GetDefaultTestApp()

			gameID := uuid.NewV4().String()
			_, clan, _, _, _, err := models.GetClanWithMemberships(
				testDb, 10, 0, 0, 0, gameID, uuid.NewV4().String(),
			)
			Expect(err).NotTo(HaveOccurred())

			clanIDs := []string{clan.PublicID}

			url := fmt.Sprintf(
				"%s?clanPublicIds=%s",
				GetGameRoute("unexistent_game", "clans-summary"),
				strings.Join(clanIDs, ","),
			)
			status, body := Get(a, url)

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeFalse())
			Expect(status).To(Equal(http.StatusNotFound))
		})

		It("Should fail with 400 if empty query string", func() {
			a := GetDefaultTestApp()

			gameID := uuid.NewV4().String()
			_, clan1, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID, uuid.NewV4().String())
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan1.GameID, "clans-summary"))
			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("No clanPublicIds provided"))
		})
	})

	Describe("Retrieve Clan Summary Handler", func() {
		It("Should get details for clan", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			a := GetDefaultTestApp()

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s/summary", clan.PublicID)))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())

			Expect(result["membershipCount"].(float64)).To(BeEquivalentTo(clan.MembershipCount))
			Expect(result["publicID"]).To(Equal(clan.PublicID))
			Expect(result["name"]).To(Equal(clan.Name))
			Expect(result["allowApplication"]).To(Equal(clan.AllowApplication))
			Expect(result["autoJoin"]).To(Equal(clan.AutoJoin))
			resultMetadata := result["metadata"].(map[string]interface{})
			for k, v := range resultMetadata {
				Expect(v).To(Equal(clan.Metadata[k]))
			}
		})

		It("Should not get details for clan that does not exist", func() {
			a := GetDefaultTestApp()

			status, body := Get(a, GetGameRoute("game-id", "/clans/dont-exist/summary"))
			Expect(status).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Clan was not found with id: dont-exist"))
		})
	})

	Describe("Search Clan Handler", func() {
		It("Should search for a clan", func() {
			gameID := uuid.NewV4().String()
			player, expectedClans, err := models.GetTestClans(
				testDb, gameID, "clan-apisearch-clan", 10,
			)
			Expect(err).NotTo(HaveOccurred())

			a := GetDefaultTestApp()

			status, body := Get(a, GetGameRoute(player.GameID, "clan-search?term=APISEARCH"))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())

			clans := result["clans"].([]interface{})
			Expect(len(clans)).To(Equal(10))

			clansDict := map[string]map[string]interface{}{}
			for _, cl := range clans {
				clan := cl.(map[string]interface{})
				clansDict[clan["publicID"].(string)] = clan
			}

			for _, expectedClan := range expectedClans {
				clan := clansDict[expectedClan.PublicID]
				Expect(clan["name"]).To(Equal(expectedClan.Name))
				Expect(clan["membershipCount"].(float64)).To(BeEquivalentTo(expectedClan.MembershipCount))
				Expect(clan["autoJoin"]).To(Equal(expectedClan.AutoJoin))
				Expect(clan["allowApplication"]).To(Equal(expectedClan.AllowApplication))
			}
		})

		It("Should unicode search for a clan", func() {
			gameID := uuid.NewV4().String()
			player, expectedClans, err := models.GetTestClans(
				testDb, gameID, "clan-apisearch-clan", 10,
			)
			Expect(err).NotTo(HaveOccurred())

			a := GetDefaultTestApp()

			url := "clan-search?term=💩clán-clan-APISEARCH"
			status, body := Get(a, GetGameRoute(player.GameID, url))
			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())

			clans := result["clans"].([]interface{})
			Expect(len(clans)).To(Equal(10))

			clansDict := map[string]map[string]interface{}{}
			for _, cl := range clans {
				clan := cl.(map[string]interface{})
				clansDict[clan["publicID"].(string)] = clan
			}

			for _, expectedClan := range expectedClans {
				clan := clansDict[expectedClan.PublicID]
				Expect(clan["name"]).To(Equal(expectedClan.Name))
				Expect(clan["membershipCount"].(float64)).To(BeEquivalentTo(expectedClan.MembershipCount))
				Expect(clan["autoJoin"]).To(Equal(expectedClan.AutoJoin))
				Expect(clan["allowApplication"]).To(Equal(expectedClan.AllowApplication))
			}
		})

	})

	Describe("Clan Hooks", func() {
		It("Should call create clan hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/clancreated",
			}, models.ClanCreatedHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/clancreated"}, 52525)

			_, player, err := models.CreatePlayerFactory(testDb, hooks[0].GameID, true)
			Expect(err).NotTo(HaveOccurred())

			clanPublicID := uuid.NewV4().String()
			payload := map[string]interface{}{
				"publicID":         clanPublicID,
				"name":             randomdata.FullName(randomdata.RandomGender),
				"ownerPublicID":    player.PublicID,
				"metadata":         map[string]interface{}{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			a := GetDefaultTestApp()
			status, body := PostJSON(a, GetGameRoute(player.GameID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(clanPublicID))

			a.Dispatcher.Wait()

			Expect(len(*responses)).To(Equal(1))

			hookRes := (*responses)[0]["payload"].(map[string]interface{})
			Expect(hookRes["gameID"]).To(Equal(player.GameID))
			clan := hookRes["clan"].(map[string]interface{})
			Expect(clan["publicID"]).To(Equal(payload["publicID"]))
			Expect(clan["name"]).To(Equal(payload["name"]))
			Expect(str(clan["membershipCount"])).To(Equal("1"))
			Expect(clan["allowApplication"]).To(Equal(payload["allowApplication"]))
			Expect(clan["autoJoin"]).To(Equal(payload["autoJoin"]))
			clanMetadata := clan["metadata"].(map[string]interface{})
			metadata := payload["metadata"].(map[string]interface{})
			for k, v := range clanMetadata {
				Expect(str(v)).To(Equal(str(metadata[k])))
			}
		})

		Describe("Update Clan Hook", func() {
			Describe("Without whitelist", func() {
				It("Should call update clan hook", func() {
					hooks, err := models.GetHooksForRoutes(testDb, []string{
						"http://localhost:52525/clanupdated",
					}, models.ClanUpdatedHook)
					Expect(err).NotTo(HaveOccurred())
					responses := startRouteHandler([]string{"/clanupdated"}, 52525)

					_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
					Expect(err).NotTo(HaveOccurred())

					gameID := clan.GameID
					publicID := clan.PublicID
					clanName := randomdata.FullName(randomdata.RandomGender)
					ownerPublicID := owner.PublicID
					metadata := map[string]interface{}{"new": "metadata"}

					payload := map[string]interface{}{
						"name":             clanName,
						"ownerPublicID":    ownerPublicID,
						"metadata":         metadata,
						"allowApplication": clan.AllowApplication,
						"autoJoin":         clan.AutoJoin,
					}
					route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
					a := GetDefaultTestApp()
					status, body := PutJSON(a, route, payload)

					Expect(status).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(body), &result)
					Expect(result["success"]).To(BeTrue())

					a.Dispatcher.Wait()

					Expect(len(*responses)).To(Equal(1))

					hookRes := (*responses)[0]["payload"].(map[string]interface{})
					Expect(hookRes["gameID"]).To(Equal(hooks[0].GameID))
					rClan := hookRes["clan"].(map[string]interface{})
					Expect(rClan["publicID"]).To(Equal(publicID))
					Expect(rClan["name"]).To(Equal(payload["name"]))
					Expect(str(rClan["membershipCount"])).To(Equal("1"))
					Expect(rClan["allowApplication"]).To(Equal(payload["allowApplication"]))
					Expect(rClan["autoJoin"]).To(Equal(payload["autoJoin"]))
					clanMetadata := rClan["metadata"].(map[string]interface{})
					metadata = payload["metadata"].(map[string]interface{})
					for k, v := range clanMetadata {
						Expect(str(v)).To(Equal(str(metadata[k])))
					}
				})
			})

			Describe("With whitelist", func() {
				It("Should call update clan hook if field in whitelist", func() {
					hooks, err := models.GetHooksForRoutes(testDb, []string{
						"http://localhost:52525/clanupdatedhookwhitelist",
					}, models.ClanUpdatedHook)
					Expect(err).NotTo(HaveOccurred())
					responses := startRouteHandler([]string{"/clanupdatedhookwhitelist"}, 52525)

					sqlRes, err := testDb.Exec(
						"UPDATE games SET clan_metadata_fields_whitelist='something,new' WHERE public_id=$1",
						hooks[0].GameID,
					)
					Expect(err).NotTo(HaveOccurred())
					count, err := sqlRes.RowsAffected()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(BeEquivalentTo(1))

					_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
					Expect(err).NotTo(HaveOccurred())

					gameID := clan.GameID
					publicID := clan.PublicID
					ownerPublicID := owner.PublicID
					metadata := map[string]interface{}{"new": "metadata"}

					payload := map[string]interface{}{
						"name":             clan.Name,
						"ownerPublicID":    ownerPublicID,
						"metadata":         metadata,
						"allowApplication": clan.AllowApplication,
						"autoJoin":         clan.AutoJoin,
					}
					route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
					a := GetDefaultTestApp()
					status, body := PutJSON(a, route, payload)

					Expect(status).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(body), &result)
					Expect(result["success"]).To(BeTrue())

					a.Dispatcher.Wait()

					Expect(len(*responses)).To(Equal(1))

					hookRes := (*responses)[0]["payload"].(map[string]interface{})
					Expect(hookRes["gameID"]).To(Equal(hooks[0].GameID))
					rClan := hookRes["clan"].(map[string]interface{})
					Expect(rClan["publicID"]).To(Equal(publicID))
					Expect(rClan["name"]).To(Equal(payload["name"]))
					Expect(str(rClan["membershipCount"])).To(Equal("1"))
					Expect(rClan["allowApplication"]).To(Equal(payload["allowApplication"]))
					Expect(rClan["autoJoin"]).To(Equal(payload["autoJoin"]))
					clanMetadata := rClan["metadata"].(map[string]interface{})
					metadata = payload["metadata"].(map[string]interface{})
					for k, v := range clanMetadata {
						Expect(str(v)).To(Equal(str(metadata[k])))
					}
				})

				It("Should call update clan hook if field in whitelist is new", func() {
					hooks, err := models.GetHooksForRoutes(testDb, []string{
						"http://localhost:52525/clanupdatedhookwhitelist3",
					}, models.ClanUpdatedHook)
					Expect(err).NotTo(HaveOccurred())
					responses := startRouteHandler([]string{"/clanupdatedhookwhitelist3"}, 52525)

					sqlRes, err := testDb.Exec(
						"UPDATE games SET clan_metadata_fields_whitelist='something,else' WHERE public_id=$1",
						hooks[0].GameID,
					)
					Expect(err).NotTo(HaveOccurred())
					count, err := sqlRes.RowsAffected()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(BeEquivalentTo(1))

					_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
					Expect(err).NotTo(HaveOccurred())

					gameID := clan.GameID
					publicID := clan.PublicID
					ownerPublicID := owner.PublicID
					metadata := map[string]interface{}{"else": "metadata"}

					payload := map[string]interface{}{
						"name":             clan.Name,
						"ownerPublicID":    ownerPublicID,
						"metadata":         metadata,
						"allowApplication": clan.AllowApplication,
						"autoJoin":         clan.AutoJoin,
					}
					route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
					a := GetDefaultTestApp()
					status, body := PutJSON(a, route, payload)

					Expect(status).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(body), &result)
					Expect(result["success"]).To(BeTrue())

					a.Dispatcher.Wait()

					Expect(len(*responses)).To(Equal(1))

					hookRes := (*responses)[0]["payload"].(map[string]interface{})
					Expect(hookRes["gameID"]).To(Equal(hooks[0].GameID))
					rClan := hookRes["clan"].(map[string]interface{})
					Expect(rClan["publicID"]).To(Equal(publicID))
					Expect(rClan["name"]).To(Equal(payload["name"]))
					Expect(str(rClan["membershipCount"])).To(Equal("1"))
					Expect(rClan["allowApplication"]).To(Equal(payload["allowApplication"]))
					Expect(rClan["autoJoin"]).To(Equal(payload["autoJoin"]))
					clanMetadata := rClan["metadata"].(map[string]interface{})
					metadata = payload["metadata"].(map[string]interface{})
					for k, v := range clanMetadata {
						Expect(str(v)).To(Equal(str(metadata[k])))
					}
				})

				It("Should not call update clan hook if field not in whitelist", func() {
					hooks, err := models.GetHooksForRoutes(testDb, []string{
						"http://localhost:52525/clanupdatedhookwhitelist2",
					}, models.ClanUpdatedHook)
					Expect(err).NotTo(HaveOccurred())
					responses := startRouteHandler([]string{"/clanupdatedhookwhitelist2"}, 52525)

					sqlRes, err := testDb.Exec(
						"UPDATE games SET clan_metadata_fields_whitelist='other,else' WHERE public_id=$1",
						hooks[0].GameID,
					)
					Expect(err).NotTo(HaveOccurred())
					count, err := sqlRes.RowsAffected()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(BeEquivalentTo(1))

					_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
					Expect(err).NotTo(HaveOccurred())

					gameID := clan.GameID
					publicID := clan.PublicID
					ownerPublicID := owner.PublicID
					metadata := map[string]interface{}{"new": "metadata"}

					payload := map[string]interface{}{
						"name":             clan.Name,
						"ownerPublicID":    ownerPublicID,
						"metadata":         metadata,
						"allowApplication": clan.AllowApplication,
						"autoJoin":         clan.AutoJoin,
					}
					route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
					a := GetDefaultTestApp()
					status, body := PutJSON(a, route, payload)

					Expect(status).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(body), &result)
					Expect(result["success"]).To(BeTrue())

					a.Dispatcher.Wait()

					Expect(len(*responses)).To(Equal(0))
				})

				Describe("Should call update clan hook if clan details changed", func() {
					It("by name", func() {
						hooks, err := models.GetHooksForRoutes(testDb, []string{
							"http://localhost:52525/clanupdatedhookwhitelist4",
						}, models.ClanUpdatedHook)
						Expect(err).NotTo(HaveOccurred())
						responses := startRouteHandler([]string{"/clanupdatedhookwhitelist4"}, 52525)

						sqlRes, err := testDb.Exec(
							"UPDATE games SET clan_metadata_fields_whitelist='something,else' WHERE public_id=$1",
							hooks[0].GameID,
						)
						Expect(err).NotTo(HaveOccurred())
						count, err := sqlRes.RowsAffected()
						Expect(err).NotTo(HaveOccurred())
						Expect(count).To(BeEquivalentTo(1))

						_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
						Expect(err).NotTo(HaveOccurred())

						gameID := clan.GameID
						publicID := clan.PublicID
						ownerPublicID := owner.PublicID
						metadata := map[string]interface{}{}

						payload := map[string]interface{}{
							"name":             uuid.NewV4().String(),
							"ownerPublicID":    ownerPublicID,
							"metadata":         metadata,
							"allowApplication": clan.AllowApplication,
							"autoJoin":         clan.AutoJoin,
						}
						route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
						a := GetDefaultTestApp()
						status, body := PutJSON(a, route, payload)

						Expect(status).To(Equal(http.StatusOK))
						var result map[string]interface{}
						json.Unmarshal([]byte(body), &result)
						Expect(result["success"]).To(BeTrue())

						a.Dispatcher.Wait()

						Expect(len(*responses)).To(Equal(1))
					})

					It("by AutoJoin", func() {
						hooks, err := models.GetHooksForRoutes(testDb, []string{
							"http://localhost:52525/clanupdatedhookwhitelist6",
						}, models.ClanUpdatedHook)
						Expect(err).NotTo(HaveOccurred())
						responses := startRouteHandler([]string{"/clanupdatedhookwhitelist6"}, 52525)

						sqlRes, err := testDb.Exec(
							"UPDATE games SET clan_metadata_fields_whitelist='something,else' WHERE public_id=$1",
							hooks[0].GameID,
						)
						Expect(err).NotTo(HaveOccurred())
						count, err := sqlRes.RowsAffected()
						Expect(err).NotTo(HaveOccurred())
						Expect(count).To(BeEquivalentTo(1))

						_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
						Expect(err).NotTo(HaveOccurred())

						gameID := clan.GameID
						publicID := clan.PublicID
						ownerPublicID := owner.PublicID
						metadata := map[string]interface{}{}

						payload := map[string]interface{}{
							"name":             clan.Name,
							"ownerPublicID":    ownerPublicID,
							"metadata":         metadata,
							"allowApplication": clan.AllowApplication,
							"autoJoin":         !clan.AutoJoin,
						}
						route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
						a := GetDefaultTestApp()
						status, body := PutJSON(a, route, payload)

						Expect(status).To(Equal(http.StatusOK))
						var result map[string]interface{}
						json.Unmarshal([]byte(body), &result)
						Expect(result["success"]).To(BeTrue())

						a.Dispatcher.Wait()

						Expect(len(*responses)).To(Equal(1))
					})

					It("by AllowApplication", func() {
						hooks, err := models.GetHooksForRoutes(testDb, []string{
							"http://localhost:52525/clanupdatedhookwhitelist5",
						}, models.ClanUpdatedHook)
						Expect(err).NotTo(HaveOccurred())
						responses := startRouteHandler([]string{"/clanupdatedhookwhitelist5"}, 52525)

						sqlRes, err := testDb.Exec(
							"UPDATE games SET clan_metadata_fields_whitelist='something,else' WHERE public_id=$1",
							hooks[0].GameID,
						)
						Expect(err).NotTo(HaveOccurred())
						count, err := sqlRes.RowsAffected()
						Expect(err).NotTo(HaveOccurred())
						Expect(count).To(BeEquivalentTo(1))

						_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
						Expect(err).NotTo(HaveOccurred())

						gameID := clan.GameID
						publicID := clan.PublicID
						ownerPublicID := owner.PublicID
						metadata := map[string]interface{}{}

						payload := map[string]interface{}{
							"name":             clan.Name,
							"ownerPublicID":    ownerPublicID,
							"metadata":         metadata,
							"allowApplication": !clan.AllowApplication,
							"autoJoin":         clan.AutoJoin,
						}
						route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
						a := GetDefaultTestApp()
						status, body := PutJSON(a, route, payload)

						Expect(status).To(Equal(http.StatusOK))
						var result map[string]interface{}
						json.Unmarshal([]byte(body), &result)
						Expect(result["success"]).To(BeTrue())

						a.Dispatcher.Wait()

						Expect(len(*responses)).To(Equal(1))
					})
				})
			})
		})

		It("Should call leave clan hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/clanleave",
			}, models.ClanLeftHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/clanleave"}, 52525)

			_, clan, _, players, _, err := models.GetClanWithMemberships(testDb, 1, 0, 0, 0, hooks[0].GameID, "", true)
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			publicID := clan.PublicID

			payload := map[string]interface{}{}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s/leave", publicID))
			a := GetDefaultTestApp()
			status, body := PostJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			a.Dispatcher.Wait()

			Expect(len(*responses)).To(Equal(1))

			rClan := (*responses)[0]["payload"].(map[string]interface{})
			Expect(rClan["gameID"]).To(Equal(hooks[0].GameID))
			Expect(rClan["type"].(float64)).To(BeEquivalentTo(5))
			Expect(rClan["id"]).NotTo(BeEquivalentTo(nil))
			_, err = uuid.FromString(rClan["id"].(string))
			Expect(err).NotTo(HaveOccurred())
			Expect(rClan["timestamp"]).NotTo(BeEquivalentTo(nil))

			clanDetails := rClan["clan"].(map[string]interface{})
			Expect(clanDetails["publicID"]).To(Equal(clan.PublicID))
			Expect(clanDetails["name"]).To(Equal(clan.Name))
			Expect(clanDetails["membershipCount"]).To(BeEquivalentTo(1))
			Expect(clanDetails["allowApplication"]).To(Equal(clan.AllowApplication))
			Expect(clanDetails["autoJoin"]).To(Equal(clan.AutoJoin))

			newOwner := players[0]
			ownerDetails := rClan["newOwner"].(map[string]interface{})
			Expect(ownerDetails["publicID"]).To(Equal(newOwner.PublicID))
			Expect(ownerDetails["name"]).To(Equal(newOwner.Name))
			Expect(ownerDetails["membershipCount"]).To(BeEquivalentTo(0))
			Expect(ownerDetails["ownershipCount"]).To(BeEquivalentTo(1))
		})

		It("Should call leave clan hook when last member", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/clanleave2",
			}, models.ClanLeftHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/clanleave2"}, 52525)

			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			publicID := clan.PublicID

			payload := map[string]interface{}{}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s/leave", publicID))
			a := GetDefaultTestApp()
			status, body := PostJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["isDeleted"]).To(BeTrue())

			a.Dispatcher.Wait()

			Expect(len(*responses)).To(Equal(1))

			rClan := (*responses)[0]["payload"].(map[string]interface{})
			Expect(rClan["gameID"]).To(Equal(hooks[0].GameID))
			Expect(rClan["type"].(float64)).To(BeEquivalentTo(5))
			Expect(rClan["id"]).NotTo(BeEquivalentTo(nil))
			_, err = uuid.FromString(rClan["id"].(string))
			Expect(err).NotTo(HaveOccurred())
			Expect(rClan["timestamp"]).NotTo(BeEquivalentTo(nil))

			clanDetails := rClan["clan"].(map[string]interface{})
			Expect(clanDetails["publicID"]).To(Equal(clan.PublicID))
			Expect(clanDetails["name"]).To(Equal(clan.Name))
			Expect(clanDetails["membershipCount"]).To(BeEquivalentTo(1))
			Expect(clanDetails["allowApplication"]).To(Equal(clan.AllowApplication))
			Expect(clanDetails["autoJoin"]).To(Equal(clan.AutoJoin))

			Expect(rClan["newOwner"]).To(BeNil())
		})

		It("Should call transfer ownership hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/clantransfer",
			}, models.ClanOwnershipTransferredHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/clantransfer"}, 52525)

			_, clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 1, 0, 0, 0, hooks[0].GameID, "", true)
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			publicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID": players[0].PublicID,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s/transfer-ownership", publicID))
			a := GetDefaultTestApp()
			status, body := PostJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["newOwner"]).NotTo(BeNil())
			Expect(result["previousOwner"]).NotTo(BeNil())

			a.Dispatcher.Wait()

			Expect(len(*responses)).To(Equal(1))

			rClan := (*responses)[0]["payload"].(map[string]interface{})
			Expect(rClan["gameID"]).To(Equal(hooks[0].GameID))

			clanDetails := rClan["clan"].(map[string]interface{})
			Expect(clanDetails["publicID"]).To(Equal(clan.PublicID))
			Expect(clanDetails["name"]).To(Equal(clan.Name))
			Expect(clanDetails["membershipCount"]).To(BeEquivalentTo(2))
			Expect(clanDetails["allowApplication"]).To(Equal(clan.AllowApplication))
			Expect(clanDetails["autoJoin"]).To(Equal(clan.AutoJoin))

			newOwner := players[0]
			ownerDetails := rClan["newOwner"].(map[string]interface{})
			Expect(ownerDetails["publicID"]).To(Equal(newOwner.PublicID))
			Expect(ownerDetails["name"]).To(Equal(newOwner.Name))
			Expect(ownerDetails["membershipCount"]).To(BeEquivalentTo(0))
			Expect(ownerDetails["ownershipCount"]).To(BeEquivalentTo(1))

			previousOwnerDetails := rClan["previousOwner"].(map[string]interface{})
			Expect(previousOwnerDetails["publicID"]).To(Equal(owner.PublicID))
			Expect(previousOwnerDetails["name"]).To(Equal(owner.Name))
			Expect(previousOwnerDetails["membershipCount"]).To(BeEquivalentTo(1))
			Expect(previousOwnerDetails["ownershipCount"]).To(BeEquivalentTo(0))
		})
	})
})
