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

var _ = Describe("Player API Handler", func() {
	var testDb models.DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create Player Handler", func() {
		It("Should create player", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			payload := map[string]interface{}{
				"publicID": randomdata.FullName(randomdata.RandomGender),
				"name":     randomdata.FullName(randomdata.RandomGender),
				"metadata": map[string]interface{}{"x": 1},
			}
			res := PostJSON(a, GetGameRoute(game.PublicID, "/players"), payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(payload["publicID"].(string)))

			dbPlayer, err := models.GetPlayerByPublicID(
				a.Db, game.PublicID, payload["publicID"].(string),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbPlayer.GameID).To(Equal(game.PublicID))
			Expect(dbPlayer.PublicID).To(Equal(payload["publicID"]))
			Expect(dbPlayer.Name).To(Equal(payload["name"]))
			Expect(dbPlayer.Metadata["x"]).To(BeEquivalentTo(payload["metadata"].(map[string]interface{})["x"]))
		})

		It("Should not create player if missing parameters", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/players")
			res := PostJSON(a, route, map[string]interface{}{})

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("publicID is required, name is required, metadata is required"))
		})

		It("Should not create player if invalid payload", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/players")
			res := PostBody(a, route, "invalid")

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(strings.Contains(result["reason"].(string), "While trying to read JSON")).To(BeTrue())
		})

		It("Should not create player if invalid data", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			payload := map[string]interface{}{
				"publicID": strings.Repeat("s", 256),
				"name":     randomdata.FullName(randomdata.RandomGender),
				"metadata": map[string]interface{}{"x": 1},
			}
			res := PostJSON(a, GetGameRoute(game.PublicID, "/players"), payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("pq: value too long for type character varying(255)"))
		})
	})

	Describe("Update Player Handler", func() {
		It("Should update player", func() {
			a := GetDefaultTestApp()
			_, player, err := models.CreatePlayerFactory(a.Db, "")
			Expect(err).NotTo(HaveOccurred())

			metadata := map[string]interface{}{"y": 10}
			payload := map[string]interface{}{
				"name":     player.Name,
				"metadata": metadata,
			}

			route := GetGameRoute(player.GameID, fmt.Sprintf("/players/%s", player.PublicID))
			res := PutJSON(a, route, payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			dbPlayer, err := models.GetPlayerByPublicID(a.Db, player.GameID, player.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbPlayer.GameID).To(Equal(player.GameID))
			Expect(dbPlayer.PublicID).To(Equal(player.PublicID))
			Expect(dbPlayer.Name).To(Equal(player.Name))
			Expect(dbPlayer.Metadata["y"]).To(BeEquivalentTo(metadata["y"]))
		})

		It("Should not update player if missing parameters", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/players/player-id")
			res := PutJSON(a, route, map[string]interface{}{})

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("name is required, metadata is required"))
		})

		It("Should not update player if invalid payload", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/players/fake")
			res := PutBody(a, route, "invalid")

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(strings.Contains(result["reason"].(string), "While trying to read JSON")).To(BeTrue())
		})

		It("Should not update player if invalid data", func() {
			a := GetDefaultTestApp()
			_, player, err := models.CreatePlayerFactory(a.Db, "")
			Expect(err).NotTo(HaveOccurred())

			payload := map[string]interface{}{
				"publicID": player.PublicID,
				"name":     strings.Repeat("s", 256),
				"metadata": map[string]interface{}{},
			}
			route := GetGameRoute(player.GameID, fmt.Sprintf("/players/%s", player.PublicID))
			res := PutJSON(a, route, payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusInternalServerError))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("pq: value too long for type character varying(255)"))
		})
	})

	Describe("Retrieve Player", func() {
		It("Should retrieve player", func() {
			a := GetDefaultTestApp()
			gameID := uuid.NewV4().String()
			player, err := models.GetTestPlayerWithMemberships(testDb, gameID, 5, 2, 3, 8)
			Expect(err).NotTo(HaveOccurred())

			route := GetGameRoute(player.GameID, fmt.Sprintf("/players/%s", player.PublicID))
			res := Get(a, route)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var playerDetails map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &playerDetails)
			Expect(playerDetails["success"]).To(BeTrue())

			// Player Details
			Expect(playerDetails["publicID"]).To(Equal(player.PublicID))
			Expect(playerDetails["name"]).To(Equal(player.Name))
			Expect(playerDetails["metadata"]).NotTo(BeEquivalentTo(nil))

			//Memberships
			Expect(len(playerDetails["memberships"].([]interface{}))).To(Equal(18))

			clans := playerDetails["clans"].(map[string]interface{}) // can't be map[string]interface{}
			approved := clans["approved"].([]interface{})
			denied := clans["denied"].([]interface{})
			banned := clans["banned"].([]interface{})
			pendingApplications := clans["pendingApplications"].([]interface{})
			pendingInvites := clans["pendingInvites"].([]interface{})

			Expect(len(approved)).To(Equal(5))
			Expect(len(denied)).To(Equal(2))
			Expect(len(banned)).To(Equal(3))
			Expect(len(pendingApplications)).To(Equal(0))
			Expect(len(pendingInvites)).To(Equal(8))
		})
		It("Should return 404 for invalid player", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("some-game", "/players/invalid-player")
			res := Get(a, route)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusNotFound))

			var playerDetails map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &playerDetails)
			Expect(playerDetails["success"]).To(BeFalse())
			Expect(playerDetails["reason"]).To(Equal("Player was not found with id: invalid-player"))
		})
	})

	Describe("Player Hooks", func() {
		It("Should call create player hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/playercreated",
			}, models.PlayerCreatedHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/playercreated"}, 52525)

			app := GetDefaultTestApp()
			time.Sleep(time.Second)

			gameID := hooks[0].GameID
			payload := map[string]interface{}{
				"publicID": randomdata.FullName(randomdata.RandomGender),
				"name":     randomdata.FullName(randomdata.RandomGender),
				"metadata": map[string]interface{}{"x": "a"},
			}
			res := PostJSON(app, GetGameRoute(gameID, "/players"), payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal(payload["publicID"].(string)))

			app.Dispatcher.Wait()

			Expect(len(*responses)).To(Equal(1))

			player := (*responses)[0]["payload"].(map[string]interface{})
			Expect(player["gameID"]).To(Equal(gameID))
			Expect(player["publicID"]).To(Equal(payload["publicID"]))
			Expect(player["name"]).To(Equal(payload["name"]))
			Expect(str(player["membershipCount"])).To(Equal("0"))
			Expect(str(player["ownershipCount"])).To(Equal("0"))
			playerMetadata := player["metadata"].(map[string]interface{})
			metadata := payload["metadata"].(map[string]interface{})
			for k, v := range playerMetadata {
				Expect(v).To(Equal(metadata[k]))
			}
		})

		Describe("Update Player Hook", func() {
			Describe("Without Whitelist", func() {
				It("Should call update player hook", func() {
					hooks, err := models.GetHooksForRoutes(testDb, []string{
						"http://localhost:52525/updated",
					}, models.PlayerUpdatedHook)
					Expect(err).NotTo(HaveOccurred())
					responses := startRouteHandler([]string{"/updated"}, 52525)

					player := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{"GameID": hooks[0].GameID}).(*models.Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					app := GetDefaultTestApp()
					time.Sleep(time.Second)

					gameID := hooks[0].GameID
					payload := map[string]interface{}{
						"publicID": player.PublicID,
						"name":     player.Name,
						"metadata": player.Metadata,
					}
					res := PutJSON(app, GetGameRoute(gameID, fmt.Sprintf("/players/%s", player.PublicID)), payload)

					Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(res.Body().Raw()), &result)
					Expect(result["success"]).To(BeTrue())

					app.Dispatcher.Wait()
					Expect(len(*responses)).To(Equal(1))

					playerPayload := (*responses)[0]["payload"].(map[string]interface{})
					Expect(playerPayload["gameID"]).To(Equal(gameID))
					Expect(playerPayload["publicID"]).To(Equal(payload["publicID"]))
					Expect(playerPayload["name"]).To(Equal(payload["name"]))
					Expect(str(playerPayload["membershipCount"])).To(Equal("0"))
					Expect(str(playerPayload["ownershipCount"])).To(Equal("0"))
					playerMetadata := playerPayload["metadata"].(map[string]interface{})
					metadata := payload["metadata"].(map[string]interface{})
					for k, v := range playerMetadata {
						Expect(v).To(Equal(metadata[k]))
					}
				})
			})
			Describe("With Whitelist", func() {
				It("Should call update player hook if whitelisted", func() {
					hooks, err := models.GetHooksForRoutes(testDb, []string{
						"http://localhost:52525/updated_whitelist",
					}, models.PlayerUpdatedHook)
					Expect(err).NotTo(HaveOccurred())
					responses := startRouteHandler([]string{"/updated_whitelist"}, 52525)

					sqlRes, err := testDb.Exec(
						"UPDATE games SET player_metadata_fields_whitelist='something,new' WHERE public_id=$1",
						hooks[0].GameID,
					)
					Expect(err).NotTo(HaveOccurred())
					count, err := sqlRes.RowsAffected()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(BeEquivalentTo(1))

					player := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": hooks[0].GameID,
						"Metadata": map[string]interface{}{
							"new": "something",
						},
					}).(*models.Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					app := GetDefaultTestApp()
					time.Sleep(time.Second)

					gameID := hooks[0].GameID
					payload := map[string]interface{}{
						"publicID": player.PublicID,
						"name":     player.Name,
						"metadata": map[string]interface{}{
							"new": "metadata",
						},
					}
					res := PutJSON(app, GetGameRoute(gameID, fmt.Sprintf("/players/%s", player.PublicID)), payload)

					Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(res.Body().Raw()), &result)
					Expect(result["success"]).To(BeTrue())

					app.Dispatcher.Wait()
					Expect(len(*responses)).To(Equal(1))

					playerPayload := (*responses)[0]["payload"].(map[string]interface{})
					Expect(playerPayload["gameID"]).To(Equal(gameID))
					Expect(playerPayload["publicID"]).To(Equal(payload["publicID"]))
					Expect(playerPayload["name"]).To(Equal(payload["name"]))
					Expect(str(playerPayload["membershipCount"])).To(Equal("0"))
					Expect(str(playerPayload["ownershipCount"])).To(Equal("0"))
					playerMetadata := playerPayload["metadata"].(map[string]interface{})
					metadata := payload["metadata"].(map[string]interface{})
					for k, v := range playerMetadata {
						Expect(v).To(Equal(metadata[k]))
					}
				})

				It("Should call update player hook if whitelisted and field is new", func() {
					hooks, err := models.GetHooksForRoutes(testDb, []string{
						"http://localhost:52525/updated_whitelist_3",
					}, models.PlayerUpdatedHook)
					Expect(err).NotTo(HaveOccurred())
					responses := startRouteHandler([]string{"/updated_whitelist_3"}, 52525)

					sqlRes, err := testDb.Exec(
						"UPDATE games SET player_metadata_fields_whitelist='something,new' WHERE public_id=$1",
						hooks[0].GameID,
					)
					Expect(err).NotTo(HaveOccurred())
					count, err := sqlRes.RowsAffected()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(BeEquivalentTo(1))

					player := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":   hooks[0].GameID,
						"Metadata": map[string]interface{}{},
					}).(*models.Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					app := GetDefaultTestApp()
					time.Sleep(time.Second)

					gameID := hooks[0].GameID
					payload := map[string]interface{}{
						"publicID": player.PublicID,
						"name":     player.Name,
						"metadata": map[string]interface{}{
							"new": "metadata",
						},
					}
					res := PutJSON(app, GetGameRoute(gameID, fmt.Sprintf("/players/%s", player.PublicID)), payload)

					Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(res.Body().Raw()), &result)
					Expect(result["success"]).To(BeTrue())

					app.Dispatcher.Wait()
					Expect(len(*responses)).To(Equal(1))

					playerPayload := (*responses)[0]["payload"].(map[string]interface{})
					Expect(playerPayload["gameID"]).To(Equal(gameID))
					Expect(playerPayload["publicID"]).To(Equal(payload["publicID"]))
					Expect(playerPayload["name"]).To(Equal(payload["name"]))
					Expect(str(playerPayload["membershipCount"])).To(Equal("0"))
					Expect(str(playerPayload["ownershipCount"])).To(Equal("0"))
					playerMetadata := playerPayload["metadata"].(map[string]interface{})
					metadata := payload["metadata"].(map[string]interface{})
					for k, v := range playerMetadata {
						Expect(v).To(Equal(metadata[k]))
					}
				})

				It("Should not call update player hook if not whitelisted", func() {
					hooks, err := models.GetHooksForRoutes(testDb, []string{
						"http://localhost:52525/updated_whitelist_2",
					}, models.PlayerUpdatedHook)
					Expect(err).NotTo(HaveOccurred())
					responses := startRouteHandler([]string{"/updated_whitelist_2"}, 52525)

					sqlRes, err := testDb.Exec(
						"UPDATE games SET player_metadata_fields_whitelist='something,new' WHERE public_id=$1",
						hooks[0].GameID,
					)
					Expect(err).NotTo(HaveOccurred())
					count, err := sqlRes.RowsAffected()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(BeEquivalentTo(1))

					player := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": hooks[0].GameID,
						"Metadata": map[string]interface{}{
							"else": "something",
						},
					}).(*models.Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					app := GetDefaultTestApp()
					time.Sleep(time.Second)

					gameID := hooks[0].GameID
					payload := map[string]interface{}{
						"publicID": player.PublicID,
						"name":     player.Name,
						"metadata": map[string]interface{}{
							"else": "metadata",
						},
					}
					res := PutJSON(app, GetGameRoute(gameID, fmt.Sprintf("/players/%s", player.PublicID)), payload)

					Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
					var result map[string]interface{}
					json.Unmarshal([]byte(res.Body().Raw()), &result)
					Expect(result["success"]).To(BeTrue())

					app.Dispatcher.Wait()
					Expect(len(*responses)).To(Equal(0))
				})

				It("Should call update player hook if whitelisted and player does not exist", func() {
					hooks, err := models.GetHooksForRoutes(testDb, []string{
						"http://localhost:52525/updated_whitelist_not_exist",
					}, models.PlayerUpdatedHook)
					Expect(err).NotTo(HaveOccurred())
					responses := startRouteHandler([]string{"/updated_whitelist_not_exist"}, 52525)

					sqlRes, err := testDb.Exec(
						"UPDATE games SET player_metadata_fields_whitelist='something,new' WHERE public_id=$1",
						hooks[0].GameID,
					)
					Expect(err).NotTo(HaveOccurred())
					count, err := sqlRes.RowsAffected()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(BeEquivalentTo(1))

					app := GetDefaultTestApp()
					time.Sleep(time.Second)

					gameID := hooks[0].GameID
					payload := map[string]interface{}{
						"publicID": uuid.NewV4().String(),
						"name":     uuid.NewV4().String(),
						"metadata": map[string]interface{}{
							"new": "metadata",
						},
					}
					res := PutJSON(app, GetGameRoute(gameID, fmt.Sprintf("/players/%s", payload["publicID"])), payload)

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
})
