// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
)

var membershipResult *http.Response

func BenchmarkApplyForMembership(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	game, clan, _, _, _, err := fixtures.GetClanWithMemberships(db, 20, 20, 20, 20, "", "")
	if err != nil {
		panic(err.Error())
	}
	game.MaxMembers = clan.MembershipCount + b.N
	_, err = db.Update(game)
	if err != nil {
		panic(err.Error())
	}

	clan.AllowApplication = true
	_, err = db.Update(clan)
	if err != nil {
		panic(err.Error())
	}

	var players []*models.Player
	for i := 0; i < b.N; i++ {
		player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
			"GameID": game.PublicID,
		}).(*models.Player)
		err = db.Insert(player)
		if err != nil {
			panic(err.Error())
		}
		players = append(players, player)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := map[string]interface{}{
			"level":          "Member",
			"playerPublicID": players[i].PublicID,
		}
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s/memberships/application", game.PublicID, clan.PublicID))
		res, err := postTo(route, payload)
		validateResp(res, err)
		res.Body.Close()

		membershipResult = res
	}
}

func BenchmarkInviteForMembership(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(db, 20, 20, 20, 20, "", "")
	if err != nil {
		panic(err.Error())
	}
	game.MaxMembers = clan.MembershipCount + b.N
	_, err = db.Update(game)
	if err != nil {
		panic(err.Error())
	}

	var players []*models.Player
	for i := 0; i < b.N; i++ {
		player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
			"GameID": game.PublicID,
		}).(*models.Player)
		err = db.Insert(player)
		if err != nil {
			panic(err.Error())
		}
		players = append(players, player)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := map[string]interface{}{
			"level":             "Member",
			"playerPublicID":    players[i].PublicID,
			"requestorPublicID": owner.PublicID,
		}
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s/memberships/invitation", game.PublicID, clan.PublicID))
		res, err := postTo(route, payload)
		validateResp(res, err)
		res.Body.Close()

		membershipResult = res
	}
}

func BenchmarkApproveMembershipApplication(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(db, 0, 0, 0, b.N, "", "")
	if err != nil {
		panic(err.Error())
	}
	game.MaxMembers = clan.MembershipCount + b.N
	_, err = db.Update(game)
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < b.N; i++ {
		memberships[i].RequestorID = memberships[i].PlayerID
		_, err = db.Update(memberships[i])
		if err != nil {
			panic(err.Error())
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := map[string]interface{}{
			"playerPublicID":    players[i].PublicID,
			"requestorPublicID": owner.PublicID,
		}
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s/memberships/application/approve", game.PublicID, clan.PublicID))
		res, err := postTo(route, payload)
		validateResp(res, err)
		res.Body.Close()

		membershipResult = res
	}
}

func BenchmarkApproveMembershipInvitation(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	game, clan, _, players, _, err := fixtures.GetClanWithMemberships(db, 0, 0, 0, b.N, "", "")
	if err != nil {
		panic(err.Error())
	}
	game.MaxMembers = clan.MembershipCount + b.N
	_, err = db.Update(game)
	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := map[string]interface{}{
			"playerPublicID":    players[i].PublicID,
			"requestorPublicID": players[i].PublicID,
		}
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s/memberships/invitation/approve", game.PublicID, clan.PublicID))
		res, err := postTo(route, payload)
		validateResp(res, err)
		res.Body.Close()

		membershipResult = res
	}
}

func BenchmarkDeleteMembership(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	game, clan, owner, players, _, err := fixtures.GetClanWithMemberships(db, b.N, 0, 0, 0, "", "")
	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := map[string]interface{}{
			"playerPublicID":    players[i].PublicID,
			"requestorPublicID": owner.PublicID,
		}
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s/memberships/delete", game.PublicID, clan.PublicID))
		res, err := postTo(route, payload)
		validateResp(res, err)
		res.Body.Close()

		membershipResult = res
	}
}

func BenchmarkPromoteMembership(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	game, clan, owner, players, _, err := fixtures.GetClanWithMemberships(db, b.N, 0, 0, 0, "", "")
	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := map[string]interface{}{
			"playerPublicID":    players[i].PublicID,
			"requestorPublicID": owner.PublicID,
		}
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s/memberships/promote", game.PublicID, clan.PublicID))
		res, err := postTo(route, payload)
		validateResp(res, err)
		res.Body.Close()

		membershipResult = res
	}
}

func BenchmarkDemoteMembership(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(db, b.N, 0, 0, 0, "", "")
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < b.N; i++ {
		memberships[i].Level = "Elder"
		_, err = db.Update(memberships[i])
		if err != nil {
			panic(err.Error())
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := map[string]interface{}{
			"playerPublicID":    players[i].PublicID,
			"requestorPublicID": owner.PublicID,
		}
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s/memberships/demote", game.PublicID, clan.PublicID))
		res, err := postTo(route, payload)
		validateResp(res, err)
		res.Body.Close()

		membershipResult = res
	}
}
