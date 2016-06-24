// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"fmt"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/util"
)

// GameFactory is responsible for constructing test game instances
var GameFactory = factory.NewFactory(
	&Game{
		MinLevelToAcceptApplication:   2,
		MinLevelToCreateInvitation:    2,
		MinLevelToRemoveMember:        2,
		MinLevelOffsetToRemoveMember:  1,
		MinLevelOffsetToPromoteMember: 2,
		MinLevelOffsetToDemoteMember:  1,
		MaxMembers:                    100,
	},
).Attr("PublicID", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
}).Attr("Name", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
}).Attr("Metadata", func(args factory.Args) (interface{}, error) {
	return util.JSON{}, nil
}).Attr("MembershipLevels", func(args factory.Args) (interface{}, error) {
	return util.JSON{"Member": 1, "Elder": 2, "CoLeader": 3}, nil
})

// HookFactory is responsible for constructing event hook instances
var HookFactory = factory.NewFactory(
	&Hook{EventType: GameUpdatedHook, URL: "http://test/game-created"},
)

// CreateHookFactory is responsible for creating a test hook instance with the associated game
func CreateHookFactory(db DB, gameID string, eventType int, url string) (*Hook, error) {
	if gameID == "" {
		gameID = uuid.NewV4().String()
	}
	game := GameFactory.MustCreateWithOption(util.JSON{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}

	hook := HookFactory.MustCreateWithOption(util.JSON{
		"GameID":    gameID,
		"PublicID":  uuid.NewV4().String(),
		"EventType": eventType,
		"URL":       url,
	}).(*Hook)
	err = db.Insert(hook)
	if err != nil {
		return nil, err
	}
	return hook, nil
}

// GetHooksForRoutes gets hooks for all the specified routes
func GetHooksForRoutes(db DB, routes []string, eventType int) ([]*Hook, error) {
	var hooks []*Hook

	gameID := uuid.NewV4().String()

	game := GameFactory.MustCreateWithOption(util.JSON{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}

	for _, route := range routes {
		hook := HookFactory.MustCreateWithOption(util.JSON{
			"GameID":    gameID,
			"PublicID":  uuid.NewV4().String(),
			"EventType": eventType,
			"URL":       route,
		}).(*Hook)
		err := db.Insert(hook)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}

	return hooks, nil
}

func configureFactory(fct *factory.Factory) *factory.Factory {
	return fct.Attr("PublicID", func(args factory.Args) (interface{}, error) {
		return uuid.NewV4().String(), nil
	}).Attr("Name", func(args factory.Args) (interface{}, error) {
		return randomdata.FullName(randomdata.RandomGender), nil
	}).Attr("Metadata", func(args factory.Args) (interface{}, error) {
		return util.JSON{}, nil
	})
}

// PlayerFactory is responsible for constructing test player instances
var PlayerFactory = configureFactory(factory.NewFactory(
	&Player{},
))

// ClanFactory is responsible for constructing test clan instances
var ClanFactory = configureFactory(factory.NewFactory(
	&Clan{},
))

// CreatePlayerFactory is responsible for creating a test player instance with the associated game
func CreatePlayerFactory(db DB, gameID string) (*Game, *Player, error) {
	if gameID == "" {
		gameID = uuid.NewV4().String()
	}
	game := GameFactory.MustCreateWithOption(util.JSON{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, err
	}

	player := PlayerFactory.MustCreateWithOption(util.JSON{
		"GameID": gameID,
	}).(*Player)
	err = db.Insert(player)
	if err != nil {
		return nil, nil, err
	}
	return game, player, nil
}

// MembershipFactory is responsible for constructing test membership instances
var MembershipFactory = factory.NewFactory(
	&Membership{},
).SeqInt("GameID", func(n int) (interface{}, error) {
	return fmt.Sprintf("game-%d", n), nil
})

// GetClanWithMemberships returns a clan filled with the number of memberships specified
func GetClanWithMemberships(
	db DB, numberOfMemberships int, gameID string, clanPublicID string,
) (*Game, *Clan, *Player, []*Player, []*Membership, error) {
	if gameID == "" {
		gameID = uuid.NewV4().String()
	}
	if clanPublicID == "" {
		clanPublicID = uuid.NewV4().String()
	}

	game := GameFactory.MustCreateWithOption(util.JSON{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	owner := PlayerFactory.MustCreateWithOption(util.JSON{
		"GameID": gameID,
	}).(*Player)
	err = db.Insert(owner)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var players []*Player

	for i := 0; i < numberOfMemberships; i++ {
		player := PlayerFactory.MustCreateWithOption(util.JSON{
			"GameID": owner.GameID,
		}).(*Player)
		err = db.Insert(player)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		players = append(players, player)
	}

	clan := ClanFactory.MustCreateWithOption(util.JSON{
		"GameID":   owner.GameID,
		"PublicID": clanPublicID,
		"OwnerID":  owner.ID,
		"Metadata": util.JSON{"x": "a"},
	}).(*Clan)
	err = db.Insert(clan)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var memberships []*Membership

	for i := 0; i < numberOfMemberships; i++ {
		membership := MembershipFactory.MustCreateWithOption(util.JSON{
			"GameID":      owner.GameID,
			"PlayerID":    players[i].ID,
			"ClanID":      clan.ID,
			"RequestorID": owner.ID,
			"Metadata":    util.JSON{"x": "a"},
			"Level":       "Member",
		}).(*Membership)

		err = db.Insert(membership)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		memberships = append(memberships, membership)
	}

	return game, clan, owner, players, memberships, nil
}

// GetClanReachedMaxMemberships returns a clan with one approved membership, one unapproved membership and game MaxMembers=1
func GetClanReachedMaxMemberships(db DB) (*Game, *Clan, *Player, []*Player, []*Membership, error) {
	gameID := uuid.NewV4().String()
	clanPublicID := uuid.NewV4().String()

	game := GameFactory.MustCreateWithOption(util.JSON{
		"PublicID":   gameID,
		"MaxMembers": 1,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	owner := PlayerFactory.MustCreateWithOption(util.JSON{
		"GameID": gameID,
	}).(*Player)
	err = db.Insert(owner)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var players []*Player

	for i := 0; i < 2; i++ {
		player := PlayerFactory.MustCreateWithOption(util.JSON{
			"GameID": owner.GameID,
		}).(*Player)
		err = db.Insert(player)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		players = append(players, player)
	}

	clan := ClanFactory.MustCreateWithOption(util.JSON{
		"GameID":   owner.GameID,
		"PublicID": clanPublicID,
		"OwnerID":  owner.ID,
		"Metadata": util.JSON{"x": "a"},
	}).(*Clan)
	err = db.Insert(clan)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var memberships []*Membership

	membership := MembershipFactory.MustCreateWithOption(util.JSON{
		"GameID":      owner.GameID,
		"PlayerID":    players[0].ID,
		"ClanID":      clan.ID,
		"RequestorID": owner.ID,
		"Metadata":    util.JSON{"x": "a"},
		"Approved":    true,
		"Level":       "Member",
	}).(*Membership)
	err = db.Insert(membership)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	memberships = append(memberships, membership)

	membership = MembershipFactory.MustCreateWithOption(util.JSON{
		"GameID":      owner.GameID,
		"PlayerID":    players[1].ID,
		"ClanID":      clan.ID,
		"RequestorID": owner.ID,
		"Metadata":    util.JSON{"x": "a"},
		"Approved":    false,
		"Level":       "Member",
	}).(*Membership)
	err = db.Insert(membership)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	memberships = append(memberships, membership)

	return game, clan, owner, players, memberships, nil
}

// GetTestClans returns a list of clans for tests
func GetTestClans(db DB, gameID string, publicIDTemplate string, numberOfClans int) (*Player, []*Clan, error) {
	if gameID == "" {
		gameID = uuid.NewV4().String()
	}
	game := GameFactory.MustCreateWithOption(util.JSON{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, err
	}

	player := PlayerFactory.MustCreateWithOption(util.JSON{
		"GameID": gameID,
	}).(*Player)
	err = db.Insert(player)
	if err != nil {
		return nil, nil, err
	}

	if publicIDTemplate == "" {
		publicIDTemplate = uuid.NewV4().String()
	}

	var clans []*Clan
	for i := 0; i < numberOfClans; i++ {
		clan := ClanFactory.MustCreateWithOption(util.JSON{
			"GameID":   player.GameID,
			"PublicID": fmt.Sprintf("%s-%d", publicIDTemplate, i),
			"Name":     fmt.Sprintf("%s-%d", publicIDTemplate, i),
			"OwnerID":  player.ID,
		}).(*Clan)
		err = db.Insert(clan)
		if err != nil {
			return nil, nil, err
		}

		clans = append(clans, clan)
	}

	return player, clans, nil
}

// GetTestHooks return a fixed number of hooks for each event available
func GetTestHooks(db DB, gameID string, numberOfHooks int) ([]*Hook, error) {
	game := GameFactory.MustCreateWithOption(util.JSON{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}

	var hooks []*Hook
	for i := 0; i < 2; i++ {
		for j := 0; j < numberOfHooks; j++ {
			hook := HookFactory.MustCreateWithOption(util.JSON{
				"GameID":    gameID,
				"PublicID":  uuid.NewV4().String(),
				"EventType": i,
				"URL":       fmt.Sprintf("http://test/event-%d-%d", i, j),
			}).(*Hook)
			err = db.Insert(hook)
			if err != nil {
				return nil, err
			}
			hooks = append(hooks, hook)
		}
	}

	return hooks, nil
}

//GetTestPlayerWithMemberships returns a player with approved, rejected and banned memberships
func GetTestPlayerWithMemberships(db DB, gameID string, approvedMemberships, rejectedMemberships, bannedMemberships, pendingMemberships int) (*Player, error) {
	game := GameFactory.MustCreateWithOption(util.JSON{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}

	owner := PlayerFactory.MustCreateWithOption(util.JSON{
		"GameID": gameID,
	}).(*Player)
	err = db.Insert(owner)
	if err != nil {
		return nil, err
	}

	player := PlayerFactory.MustCreateWithOption(util.JSON{
		"GameID": owner.GameID,
	}).(*Player)
	err = db.Insert(player)
	if err != nil {
		return nil, err
	}

	createClan := func() (*Clan, error) {
		clan := ClanFactory.MustCreateWithOption(util.JSON{
			"GameID":   owner.GameID,
			"PublicID": uuid.NewV4().String(),
			"OwnerID":  owner.ID,
			"Metadata": util.JSON{"x": "a"},
		}).(*Clan)
		err = db.Insert(clan)
		if err != nil {
			return nil, err
		}

		return clan, nil
	}

	createMembership := func(approved, denied, banned bool) (*Membership, error) {
		clan, err := createClan()
		if err != nil {
			return nil, err
		}

		payload := util.JSON{
			"GameID":      owner.GameID,
			"PlayerID":    player.ID,
			"ClanID":      clan.ID,
			"RequestorID": owner.ID,
			"Metadata":    util.JSON{"x": "a"},
			"Approved":    approved,
			"Denied":      denied,
			"Banned":      banned,
		}
		if banned {
			payload["DeletedAt"] = time.Now().UnixNano()
			payload["DeletedBy"] = owner.ID
		}

		membership := MembershipFactory.MustCreateWithOption(payload).(*Membership)

		err = db.Insert(membership)
		if err != nil {
			return nil, err
		}

		return membership, nil
	}

	for i := 0; i < approvedMemberships; i++ {
		_, err := createMembership(true, false, false)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < rejectedMemberships; i++ {
		_, err := createMembership(false, true, false)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < bannedMemberships; i++ {
		_, err := createMembership(false, false, true)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < pendingMemberships; i++ {
		_, err := createMembership(false, false, false)
		if err != nil {
			return nil, err
		}
	}

	return player, nil
}
