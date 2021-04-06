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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
)

var _ = Describe("Hook API Handler", func() {
	var testDb models.DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create Hook Handler", func() {
		It("Should create hook", func() {
			a := GetDefaultTestApp()
			db := a.Db(nil)
			game := fixtures.GameFactory.MustCreate().(*models.Game)
			err := db.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			payload := map[string]interface{}{
				"type":    models.GameUpdatedHook,
				"hookURL": "http://test/create",
			}
			status, body := PostJSON(a, GetGameRoute(game.PublicID, "/hooks"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).NotTo(BeEquivalentTo(""))

			dbHook, err := models.GetHookByPublicID(
				db, game.PublicID, result["publicID"].(string),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbHook.GameID).To(Equal(game.PublicID))
			Expect(dbHook.PublicID).To(Equal(result["publicID"]))
			Expect(dbHook.EventType).To(Equal(payload["type"]))
			Expect(dbHook.URL).To(Equal(payload["hookURL"]))
		})

		It("Should not create hook if missing parameters", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/hooks")
			status, body := PostJSON(a, route, map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("hookURL is required"))
		})

		It("Should not create hook if invalid payload", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/hooks")
			status, body := Post(a, route, "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"].(string)).To(ContainSubstring(InvalidJSONError))
		})
	})

	Describe("Delete Hook Handler", func() {
		It("Should delete hook", func() {
			a := GetDefaultTestApp()

			hook, err := fixtures.CreateHookFactory(testDb, "", models.GameUpdatedHook, "http://test/update")
			Expect(err).NotTo(HaveOccurred())

			status, body := Delete(a, GetGameRoute(hook.GameID, fmt.Sprintf("/hooks/%s", hook.PublicID)))

			Expect(status).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			number, err := testDb.SelectInt("select count(*) from hooks where id=$1", hook.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(number == 0).To(BeTrue())
		})
	})
})
