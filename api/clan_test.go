// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"gopkg.in/olivere/elastic.v5"

	"github.com/Pallinder/go-randomdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/khan/api"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/testing"
)

var _ = Describe("Clan API Handler", func() {
	var testDb, db models.DB
	var a *api.App

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())

		a = GetDefaultTestApp()
		db = a.Db(nil)
		a.NonblockingStartWorkers()
	})

	AfterEach(func() {
		DestroyTestES()
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
				"metadata":         map[string]interface{}{"x": "a"},
				"allowApplication": true,
				"autoJoin":         true,
			}
			status, body := PostJSON(a, GetGameRoute(player.GameID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(clanPublicID))

			dbClan, err := models.GetClanByPublicID(db, player.GameID, clanPublicID)
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

		It("Should create clan into mongodb if its configured", func() {
			mongo, err := GetTestMongo()
			Expect(err).NotTo(HaveOccurred())
			_, player, err := models.CreatePlayerFactory(testDb, "")
			Expect(err).NotTo(HaveOccurred())

			metadata := map[string]interface{}{"x": "a"}

			clanPublicID := randomdata.FullName(randomdata.RandomGender)
			payload := map[string]interface{}{
				"publicID":         clanPublicID,
				"name":             randomdata.FullName(randomdata.RandomGender),
				"ownerPublicID":    player.PublicID,
				"metadata":         metadata,
				"allowApplication": true,
				"autoJoin":         true,
			}
			status, body := PostJSON(a, GetGameRoute(player.GameID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(clanPublicID))

			colName := fmt.Sprintf("clans_%s", player.GameID)

			col, sess := mongo.C(colName)
			defer sess.Close()

			var res *models.Clan
			Eventually(func() *models.Clan {
				err = col.FindId(clanPublicID).One(&res)
				return res
			}, 5).Should(Not(BeNil()))

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.PublicID).To(Equal(clanPublicID))
			Expect(res.GameID).To(Equal(player.GameID))
			Expect(res.Metadata).To(BeEquivalentTo(metadata))
		})

		// TODO: fix this when hardcoded boomforce is removed
		XIt("Should index clan into ES when created", func() {
			es := GetTestES()
			_, player, err := models.CreatePlayerFactory(testDb, "")
			Expect(err).NotTo(HaveOccurred())

			clanPublicID := randomdata.FullName(randomdata.RandomGender)
			payload := map[string]interface{}{
				"publicID":         clanPublicID,
				"name":             randomdata.FullName(randomdata.RandomGender),
				"ownerPublicID":    player.PublicID,
				"metadata":         map[string]interface{}{"x": "a"},
				"allowApplication": true,
				"autoJoin":         true,
			}
			status, body := PostJSON(a, GetGameRoute(player.GameID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(clanPublicID))

			indexName := "khan-test"

			var res *elastic.GetResult
			err = testing.WaitForFunc(10, func() error {
				var err error
				res, err = es.Client.Get().Index(indexName).Type("clan").Id(clanPublicID).Do(context.TODO())
				return err
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.Index).To(Equal(indexName))
			Expect(res.Type).To(Equal("clan"))
			Expect(res.Id).To(Equal(clanPublicID))
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
			status, body := PostJSON(a, GetGameRoute(player.GameID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("name is required, ownerPublicID is required")))
		})

		It("Should not create clan if invalid payload", func() {
			gameID := "gameID"
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
				"metadata":         map[string]interface{}{"x": "a"},
				"allowApplication": true,
				"autoJoin":         true,
			}
			status, body := PostJSON(a, GetGameRoute(game.PublicID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusNotFound))
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
				"metadata":         map[string]interface{}{"x": "a"},
				"allowApplication": true,
				"autoJoin":         true,
			}
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

			dbClan, err := models.GetClanByPublicID(db, clan.GameID, clan.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbClan.OwnerID).To(Equal(memberships[0].PlayerID))
		})

		It("Should not leave a clan if invalid clan", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/leave", "random-id"))
			status, body := Post(a, route, "")

			Expect(status).To(Equal(http.StatusNotFound))
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
			status, body := PostJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbClan, err := models.GetClanByPublicID(db, clan.GameID, clan.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbClan.OwnerID).To(Equal(players[0].ID))
		})

		It("Should not transfer a clan ownership if missing parameters", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/transfer-ownership", "public-id"))
			status, body := PostJSON(a, route, map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("playerPublicID is required"))

		})

		It("Should not transfer a clan ownership if invalid payload", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/transfer-ownership", "random-id"))
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
			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbClan, err := models.GetClanByPublicID(db, gameID, publicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbClan.GameID).To(Equal(gameID))
			Expect(dbClan.PublicID).To(Equal(publicID))
			Expect(dbClan.Name).To(Equal(clanName))
			Expect(dbClan.OwnerID).To(Equal(owner.ID))
			Expect(dbClan.Metadata).To(Equal(metadata))
			Expect(dbClan.AllowApplication).To(Equal(!clan.AllowApplication))
			Expect(dbClan.AutoJoin).To(Equal(!clan.AutoJoin))
		})

		It("Should update Mongo if update clan", func() {
			mongo, err := GetTestMongo()
			Expect(err).NotTo(HaveOccurred())
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			publicID := clan.PublicID
			colName := fmt.Sprintf("clans_%s", gameID)

			col, sess := mongo.C(colName)
			defer sess.Close()

			var res *models.Clan
			Eventually(func() *models.Clan {
				err = col.FindId(publicID).One(&res)
				return res
			}, 5).Should(Not(BeNil()))

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.Name).To(Equal(clan.Name))
			Expect(res.OwnerID).To(Equal(owner.ID))
			Expect(res.PublicID).To(Equal(publicID))
			Expect(res.AutoJoin).To(Equal(clan.AutoJoin))

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
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", clan.PublicID))
			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Eventually(func() *models.Clan {
				err = col.FindId(publicID).One(&res)
				return res
			}).Should(SatisfyAll(
				Not(BeNil()),
				WithTransform(func(res *models.Clan) string {
					return res.Name
				}, Equal(clanName)),
				WithTransform(func(res *models.Clan) map[string]interface{} {
					return res.Metadata
				}, BeEquivalentTo(metadata)),
			))
		})

		// TODO: fix this when hardcoded boomforce is removed
		XIt("Should update ES if update clan", func() {
			es := GetTestES()

			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := owner.PublicID
			metadata := map[string]interface{}{"new": "metadata"}

			indexName := "khan-test"
			Eventually(func() *elastic.GetResult {
				res, _ := es.Client.Get().Index(indexName).Type("clan").Id(publicID).Do(context.TODO())
				return res
			}).Should(Not(BeNil()))

			payload := map[string]interface{}{
				"name":             clanName,
				"ownerPublicID":    ownerPublicID,
				"metadata":         metadata,
				"allowApplication": !clan.AllowApplication,
				"autoJoin":         !clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() *elastic.GetResult {
				res, _ := es.Client.Get().Index(indexName).Type("clan").Id(publicID).Do(context.TODO())
				return res
			}).Should(SatisfyAll(
				Not(BeNil()),
				WithTransform(func(res *elastic.GetResult) string {
					return res.Index
				}, Equal(indexName)),
				WithTransform(func(res *elastic.GetResult) string {
					return res.Type
				}, Equal("clan")),
				WithTransform(func(res *elastic.GetResult) string {
					return res.Id
				}, Equal(publicID)),
				WithTransform(func(res *elastic.GetResult) *models.Clan {
					updClan, _ := models.GetClanFromJSON(*res.Source)
					return updClan
				}, SatisfyAll(
					WithTransform(func(updClan *models.Clan) interface{} {
						return updClan.Metadata["new"]
					}, BeEquivalentTo(metadata["new"])),
					WithTransform(func(updClan *models.Clan) interface{} {
						return updClan.Metadata
					}, HaveKeyWithValue("x", "a")),
					WithTransform(func(updClan *models.Clan) bool {
						return updClan.AllowApplication
					}, BeEquivalentTo(!clan.AllowApplication)),
					WithTransform(func(updClan *models.Clan) bool {
						return updClan.AutoJoin
					}, BeEquivalentTo(!clan.AutoJoin)),
				)),
			))
		})

		It("Should not update clan if missing parameters", func() {
			route := GetGameRoute("gameID", fmt.Sprintf("/clans/%s", "publicID"))
			status, body := PutJSON(a, route, map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("name is required, ownerPublicID is required"))
		})

		It("Should not update clan if invalid payload", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("/clans/%s", "random-id"))

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
			metadata := map[string]interface{}{"x": "a"}

			payload := map[string]interface{}{
				"name":             clan.Name,
				"ownerPublicID":    randomdata.FullName(randomdata.RandomGender),
				"metadata":         metadata,
				"allowApplication": clan.AllowApplication,
				"autoJoin":         clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))

			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Player was not found with id: %s", payload["ownerPublicID"])))
		})

		It("Should not update clan if invalid data", func() {
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			publicID := clan.PublicID

			payload := map[string]interface{}{
				"name":             strings.Repeat("s", 256),
				"ownerPublicID":    owner.PublicID,
				"metadata":         map[string]interface{}{"x": "a"},
				"allowApplication": clan.AllowApplication,
				"autoJoin":         clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))

			status, body := PutJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("pq: value too long for type character varying(255)"))
		})
	})

	Describe("List All Clans Handler", func() {
		It("Should get all clans", func() {
			player, expectedClans, err := models.GetTestClans(testDb, "", "", 10)
			Expect(err).NotTo(HaveOccurred())
			sort.Sort(models.ClanByName(expectedClans))

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

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())

			Expect(result["name"]).To(Equal(clan.Name))
			Expect(result["publicID"]).To(Equal(clan.PublicID))
			resultMetadata := result["metadata"].(map[string]interface{})
			for k, v := range resultMetadata {
				Expect(v).To(Equal(clan.Metadata[k]))
			}
		})

		It("Should get details for clan with short publicID", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?shortID=true", clan.PublicID[0:8])))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())

			Expect(result["name"]).To(Equal(clan.Name))
			Expect(result["publicID"]).To(Equal(clan.PublicID))
			resultMetadata := result["metadata"].(map[string]interface{})
			for k, v := range resultMetadata {
				Expect(v).To(Equal(clan.Metadata[k]))
			}
		})

		It("Should get clan members", func() {
			gameID := uuid.NewV4().String()
			_, clan, _, _, _, err := models.GetClanWithMemberships(
				testDb, 10, 0, 0, 0, gameID, "clan-details-api-clan",
			)
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())

			Expect(result["roster"] == nil).To(BeFalse())
		})

		It("Should fail with 400 if maxPendingApplications cannot be parsed as uint", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?maxPendingApplications=xablau", clan.PublicID)))

			Expect(status).To(Equal(http.StatusBadRequest))
			Expect(body).To(ContainSubstring("invalid syntax"))
		})

		It("Should fail with 400 if maxPendingApplications is above allowed", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?maxPendingApplications=101", clan.PublicID)))

			Expect(status).To(Equal(http.StatusBadRequest))
			Expect(body).To(ContainSubstring("above allowed"))
		})

		It("Should fail with 400 if maxPendingInvites cannot be parsed as uint", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?maxPendingInvites=xablau", clan.PublicID)))

			Expect(status).To(Equal(http.StatusBadRequest))
			Expect(body).To(ContainSubstring("invalid syntax"))
		})

		It("Should fail with 400 if maxPendingInvites is above allowed", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?maxPendingInvites=101", clan.PublicID)))

			Expect(status).To(Equal(http.StatusBadRequest))
			Expect(body).To(ContainSubstring("above allowed"))
		})

		It("Should fail with 400 if pendingApplicationsOrder is not a valid order string", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?pendingApplicationsOrder=xablau", clan.PublicID)))

			Expect(status).To(Equal(http.StatusBadRequest))
			Expect(body).To(ContainSubstring("order is invalid"))
		})

		It("Should fail with 400 if pendingInvitesOrder is not a valid order string", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?pendingInvitesOrder=xablau", clan.PublicID)))

			Expect(status).To(Equal(http.StatusBadRequest))
			Expect(body).To(ContainSubstring("order is invalid"))
		})

		type pendingMembershipPayload struct {
			Player struct {
				Metadata struct {
					ID int `json:"id"`
				} `json:"metadata"`
			} `json:"player"`
		}
		type retrieveClanPayload struct {
			Success     bool `json:"success"`
			Memberships struct {
				PendingApplications []pendingMembershipPayload `json:"pendingApplications"`
				PendingInvites      []pendingMembershipPayload `json:"pendingInvites"`
			} `json:"memberships"`
		}

		It("should get pending applications even if max amount is not set", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 10, 0, 0, 10, "", "", false, false, true)
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)))
			Expect(status).To(Equal(http.StatusOK))

			var result retrieveClanPayload
			json.Unmarshal([]byte(body), &result)

			Expect(len(result.Memberships.PendingApplications)).To(BeNumerically(">", 0))
		})

		It("should get pending invites even if max amount is not set", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 10, 0, 0, 10, "", "", false, true, true)
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)))
			Expect(status).To(Equal(http.StatusOK))

			var result retrieveClanPayload
			json.Unmarshal([]byte(body), &result)

			Expect(len(result.Memberships.PendingInvites)).To(BeNumerically(">", 0))
		})

		validateRetrieveClanResponse := func(memberships []*models.Membership, body, pendingsOrder string, pendingsLength int, pendingsAreInvites bool) {
			Expect(pendingsLength).To(BeNumerically("<", len(memberships)))

			var result retrieveClanPayload
			json.Unmarshal([]byte(body), &result)

			Expect(result.Success).To(BeTrue())

			pendings := result.Memberships.PendingApplications
			if pendingsAreInvites {
				pendings = result.Memberships.PendingInvites
			}
			Expect(len(pendings)).To(Equal(pendingsLength))

			lastOrFirstFromAllPendingMemberships := memberships[len(memberships)-1].ID
			if pendingsOrder == ">" {
				lastOrFirstFromAllPendingMemberships = memberships[0].ID
			}
			for _, pending := range pendings {
				Expect(pending.Player.Metadata.ID).To(BeNumerically(pendingsOrder, lastOrFirstFromAllPendingMemberships))
			}
		}

		It("Should get newest pending applications if order is not set", func() {
			maxPending := 7
			_, clan, _, _, memberships, err := models.GetClanWithMemberships(testDb, 10, 0, 0, 10, "", "", false, false, true)
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?maxPendingApplications=%v", clan.PublicID, maxPending)))
			Expect(status).To(Equal(http.StatusOK))
			validateRetrieveClanResponse(memberships, body, ">", maxPending, false)
		})

		It("Should get newest pending applications", func() {
			maxPending := 7
			_, clan, _, _, memberships, err := models.GetClanWithMemberships(testDb, 10, 0, 0, 10, "", "", false, false, true)
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?maxPendingApplications=%v&pendingApplicationsOrder=newest", clan.PublicID, maxPending)))
			Expect(status).To(Equal(http.StatusOK))
			validateRetrieveClanResponse(memberships, body, ">", maxPending, false)
		})

		It("Should get oldest pending applications", func() {
			maxPending := 7
			_, clan, _, _, memberships, err := models.GetClanWithMemberships(testDb, 10, 0, 0, 10, "", "", false, false, true)
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?maxPendingApplications=%v&pendingApplicationsOrder=oldest", clan.PublicID, maxPending)))
			Expect(status).To(Equal(http.StatusOK))
			validateRetrieveClanResponse(memberships, body, "<", maxPending, false)
		})

		It("Should get newest pending invites if order is not set", func() {
			maxPending := 7
			_, clan, _, _, memberships, err := models.GetClanWithMemberships(testDb, 10, 0, 0, 10, "", "", false, true, true)
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?maxPendingInvites=%v", clan.PublicID, maxPending)))
			Expect(status).To(Equal(http.StatusOK))
			validateRetrieveClanResponse(memberships, body, ">", maxPending, true)
		})

		It("Should get newest pending invites", func() {
			maxPending := 7
			_, clan, _, _, memberships, err := models.GetClanWithMemberships(testDb, 10, 0, 0, 10, "", "", false, true, true)
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?maxPendingInvites=%v&pendingInvitesOrder=newest", clan.PublicID, maxPending)))
			Expect(status).To(Equal(http.StatusOK))
			validateRetrieveClanResponse(memberships, body, ">", maxPending, true)
		})

		It("Should get oldest pending invites", func() {
			maxPending := 7
			_, clan, _, _, memberships, err := models.GetClanWithMemberships(testDb, 10, 0, 0, 10, "", "", false, true, true)
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s?maxPendingInvites=%v&pendingInvitesOrder=oldest", clan.PublicID, maxPending)))
			Expect(status).To(Equal(http.StatusOK))
			validateRetrieveClanResponse(memberships, body, "<", maxPending, true)
		})
	})

	Describe("Retrieve Clan Members Handler", func() {
		It("Should get clans player ids", func() {
			gameID := uuid.NewV4().String()
			_, clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 10, 0, 0, 0, gameID, "clan1")
			Expect(err).NotTo(HaveOccurred())
			status, body := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s/members", clan.PublicID)))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())

			Expect(len(result["members"].([]interface{}))).To(Equal(clan.MembershipCount))
			Expect(result["members"].([]interface{})).To(ContainElement(owner.PublicID))

			for _, p := range players {
				Expect(result["members"].([]interface{})).To(ContainElement(p.PublicID))
			}
		})
	})

	Describe("Retrieve Clans Handler", func() {
		It("Should get details for clans", func() {
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

			Expect(result["success"]).To(BeTrue())
			Expect(result["missingClans"] == nil).To(BeFalse())
			missingClans := result["missingClans"].([]interface{})
			Expect(len(missingClans)).To(Equal(1))
			Expect(status).To(Equal(http.StatusOK))
		})

		It("Should fail with 400 if empty query string", func() {
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

		It("Should not fail if some clans do not exists", func() {
			gameID := uuid.NewV4().String()
			_, clan1, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID, "clan1")
			Expect(err).NotTo(HaveOccurred())
			_, clan2, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID, "clan2", true)
			Expect(err).NotTo(HaveOccurred())
			_, clan3, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID, "clan3", true)
			Expect(err).NotTo(HaveOccurred())

			clanIDs := []string{clan1.PublicID, clan2.PublicID, clan3.PublicID, "unexistent_clan", "unexistent_clan2"}

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

			Expect(result["missingClans"] == nil).To(BeFalse())
			missingClans := result["missingClans"].([]interface{})
			Expect(len(missingClans)).To(Equal(2))

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
	})

	Describe("Retrieve Clan Summary Handler", func() {
		It("Should get details for clan", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

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
			status, body := Get(a, GetGameRoute("game-id", "/clans/dont-exist/summary"))
			Expect(status).To(Equal(http.StatusNotFound))
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

			err = testing.CreateClanNameTextIndexInMongo(GetTestMongo, gameID)
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, GetGameRoute(player.GameID, "clans/search?term=APISEARCH"))

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

		It("Should search for a clan by publicID", func() {
			gameID := uuid.NewV4().String()
			player, expectedClans, err := models.GetTestClans(
				testDb, gameID, "clan-apisearch-clan", 10,
			)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(1000 * time.Millisecond)
			status, body := Get(a, GetGameRoute(
				player.GameID, fmt.Sprintf("clans/search?term=%s", expectedClans[3].PublicID),
			))

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())

			clans := result["clans"].([]interface{})
			Expect(len(clans)).To(Equal(1))

			Expect(clans[0].(map[string]interface{})["publicID"].(string)).
				To(Equal(expectedClans[3].PublicID))
		})

		It("Should unicode search for a clan", func() {
			gameID := uuid.NewV4().String()
			player, expectedClans, err := models.GetTestClans(
				testDb, gameID, "clan-apisearch-clan", 10,
			)
			Expect(err).NotTo(HaveOccurred())

			err = testing.CreateClanNameTextIndexInMongo(GetTestMongo, gameID)
			Expect(err).NotTo(HaveOccurred())

			url := "clans/search?term=💩clán-clan-APISEARCH"
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

		It("Should search for a clan with punctuation symbols in name", func() {
			gameID := uuid.NewV4().String()
			player, expectedClans, err := models.GetTestClans(
				testDb, gameID, "$#+-", 10,
			)
			Expect(err).NotTo(HaveOccurred())

			err = testing.CreateClanNameRegularIndexInMongo(GetTestMongo, gameID)
			Expect(err).NotTo(HaveOccurred())

			escapedTerm := url.QueryEscape("💩clán-$")
			url := fmt.Sprintf("clans/search?term=%s&useRegexSearch=true", escapedTerm)
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
				"metadata":         map[string]interface{}{"x": "a"},
				"allowApplication": true,
				"autoJoin":         true,
			}
			status, body := PostJSON(a, GetGameRoute(player.GameID, "/clans"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(clanPublicID))

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

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
				It("Should not call update clan hook", func() {
					hooks, err := models.GetHooksForRoutes(testDb, []string{
						"http://localhost:52525/clanupdated",
					}, models.ClanUpdatedHook)
					Expect(err).NotTo(HaveOccurred())
					responses := startRouteHandler([]string{"/clanupdated"}, 52525)

					_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
					Expect(err).NotTo(HaveOccurred())

					gameID := clan.GameID
					publicID := clan.PublicID
					clanName := clan.Name
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
					status, body := PutJSON(a, route, payload)

					Expect(status).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(body), &result)
					Expect(result["success"]).To(BeTrue())

					Consistently(func() int {
						return len(*responses)
					}, 100*time.Millisecond, time.Millisecond).Should(Equal(0))
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
					status, body := PutJSON(a, route, payload)

					Expect(status).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(body), &result)
					Expect(result["success"]).To(BeTrue())

					Eventually(func() int {
						return len(*responses)
					}).Should(Equal(1))

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
					status, body := PutJSON(a, route, payload)

					Expect(status).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(body), &result)
					Expect(result["success"]).To(BeTrue())

					Eventually(func() int {
						return len(*responses)
					}).Should(Equal(1))

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
					status, body := PutJSON(a, route, payload)

					Expect(status).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(body), &result)
					Expect(result["success"]).To(BeTrue())

					Consistently(func() int {
						return len(*responses)
					}, 100*time.Millisecond, time.Millisecond).Should(Equal(0))
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
						status, body := PutJSON(a, route, payload)

						Expect(status).To(Equal(http.StatusOK))
						var result map[string]interface{}
						json.Unmarshal([]byte(body), &result)
						Expect(result["success"]).To(BeTrue())

						Eventually(func() int {
							return len(*responses)
						}).Should(Equal(1))
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
						status, body := PutJSON(a, route, payload)

						Expect(status).To(Equal(http.StatusOK))
						var result map[string]interface{}
						json.Unmarshal([]byte(body), &result)
						Expect(result["success"]).To(BeTrue())

						Eventually(func() int {
							return len(*responses)
						}).Should(Equal(1))
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
			status, body := PostJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

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
			status, body := PostJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["isDeleted"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

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
			status, body := PostJSON(a, route, payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["newOwner"]).NotTo(BeNil())
			Expect(result["previousOwner"]).NotTo(BeNil())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

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
