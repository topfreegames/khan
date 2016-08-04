// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"encoding/json"
	"fmt"

	"github.com/topfreegames/khan/util"

	"gopkg.in/gorp.v1"
)

// Player identifies uniquely one player in a given game
type Player struct {
	ID              int                    `db:"id"`
	GameID          string                 `db:"game_id"`
	PublicID        string                 `db:"public_id"`
	Name            string                 `db:"name"`
	Metadata        map[string]interface{} `db:"metadata"`
	MembershipCount int                    `db:"membership_count"`
	OwnershipCount  int                    `db:"ownership_count"`
	CreatedAt       int64                  `db:"created_at"`
	UpdatedAt       int64                  `db:"updated_at"`
}

// PreInsert populates fields before inserting a new player
func (p *Player) PreInsert(s gorp.SqlExecutor) error {
	p.CreatedAt = util.NowMilli()
	p.UpdatedAt = p.CreatedAt
	return nil
}

// PreUpdate populates fields before updating a player
func (p *Player) PreUpdate(s gorp.SqlExecutor) error {
	p.UpdatedAt = util.NowMilli()
	return nil
}

//Serialize the player information to JSON
func (p *Player) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"gameID":          p.GameID,
		"publicID":        p.PublicID,
		"name":            p.Name,
		"metadata":        p.Metadata,
		"membershipCount": p.MembershipCount,
		"ownershipCount":  p.OwnershipCount,
	}
}

// IncrementPlayerMembershipCount increments the player membership count
func IncrementPlayerMembershipCount(db DB, id, by int) error {
	query := `
	UPDATE players SET membership_count=membership_count+$1
	WHERE players.id=$2
	`
	res, err := db.Exec(query, by, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return &ModelNotFoundError{"Player", id}
	}
	return nil
}

// IncrementPlayerOwnershipCount increments the player ownership count
func IncrementPlayerOwnershipCount(db DB, id, by int) error {
	query := `
	UPDATE players SET ownership_count=ownership_count+$1
	WHERE players.id=$2
	`
	res, err := db.Exec(query, by, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return &ModelNotFoundError{"Player", id}
	}
	return nil
}

// GetPlayerByID returns a player by id
func GetPlayerByID(db DB, id int) (*Player, error) {
	obj, err := db.Get(Player{}, id)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, &ModelNotFoundError{"Player", id}
	}

	player := obj.(*Player)
	return player, nil
}

// GetPlayerByPublicID returns a player by their public id
func GetPlayerByPublicID(db DB, gameID string, publicID string) (*Player, error) {
	var players []*Player
	_, err := db.Select(&players, "SELECT * FROM players WHERE game_id=$1 AND public_id=$2", gameID, publicID)
	if err != nil {
		return nil, err
	}
	if players == nil || len(players) < 1 {
		return nil, &ModelNotFoundError{"Player", publicID}
	}
	return players[0], nil
}

// CreatePlayer creates a new player
func CreatePlayer(db DB, gameID, publicID, name string, metadata map[string]interface{}, upsert bool) (*Player, error) {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	query := `
			INSERT INTO players(game_id, public_id, name, metadata, created_at, updated_at)
						VALUES($1, $2, $3, $4, $5, $5)%s`
	onConflict := ` ON CONFLICT (game_id, public_id)
			DO UPDATE set name=$3, metadata=$4, updated_at=$5
			WHERE players.game_id=$1 and players.public_id=$2`

	if upsert {
		query = fmt.Sprintf(query, onConflict)
	} else {
		query = fmt.Sprintf(query, "")
	}
	_, err = db.Exec(query,
		gameID, publicID, name, metadataJSON, util.NowMilli())
	if err != nil {
		return nil, err
	}
	return GetPlayerByPublicID(db, gameID, publicID)
}

// UpdatePlayer updates an existing player
func UpdatePlayer(db DB, gameID, publicID, name string, metadata map[string]interface{}) (*Player, error) {
	return CreatePlayer(db, gameID, publicID, name, metadata, true)
}

// GetPlayerOwnershipDetails returns detailed information about a player owned clans
func GetPlayerOwnershipDetails(db DB, gameID, publicID string) (map[string]interface{}, error) {
	query := `
	SELECT c.*
	FROM clans c
		INNER JOIN players p on p.id=c.owner_id
	WHERE
		c.game_id=$1 and p.public_id=$2`

	var clans []Clan
	_, err := db.Select(&clans, query, gameID, publicID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	memberships := []map[string]interface{}{}
	owned := []map[string]interface{}{}

	if len(clans) > 0 {
		clanFromDetail := func(clan Clan) map[string]interface{} {
			return map[string]interface{}{
				"publicID": clan.PublicID,
				"name":     clan.Name,
			}
		}

		membershipFromClan := func(clan Clan) map[string]interface{} {
			return map[string]interface{}{
				"level":    "owner",
				"approved": true,
				"denied":   false,
				"banned":   false,
				"clan": map[string]interface{}{
					"metadata":        clan.Metadata,
					"name":            clan.Name,
					"publicID":        clan.PublicID,
					"membershipCount": clan.MembershipCount,
				},
				"createdAt":  clan.CreatedAt,
				"updatedAt":  clan.CreatedAt,
				"approvedAt": clan.CreatedAt,
				"deletedAt":  0,
			}
		}

		for _, clan := range clans {
			m := membershipFromClan(clan)
			memberships = append(memberships, m)

			clanDetail := clanFromDetail(clan)
			owned = append(owned, clanDetail)
		}
	}

	result["memberships"] = memberships
	result["clans"] = owned
	return result, nil
}

// GetPlayerMembershipDetails returns detailed information about a player and their memberships
func GetPlayerMembershipDetails(db DB, gameID, publicID string) (map[string]interface{}, error) {
	query := `
	SELECT
		p.id PlayerID, p.name PlayerName, p.metadata PlayerMetadata, p.public_id PlayerPublicID,
		p.created_at PlayerCreatedAt, p.updated_at PlayerUpdatedAt,
		m.membership_level MembershipLevel,
		m.approved MembershipApproved, m.denied MembershipDenied, m.banned MembershipBanned,
		c.public_id ClanPublicID, c.name ClanName, c.metadata DBClanMetadata, c.owner_id ClanOwnerID,
		c.membership_count ClanMembershipCount,
		r.name RequestorName, r.public_id RequestorPublicID, r.metadata DBRequestorMetadata,
		w.membership_level RequestorMembershipLevel,
		a.name ApproverName, a.public_id ApproverPublicID, a.metadata DBApproverMetadata,
		y.name DenierName, y.public_id DenierPublicID, y.metadata DBDenierMetadata,
		m.created_at MembershipCreatedAt,
		m.updated_at MembershipUpdatedAt,
		m.deleted_at MembershipDeletedAt,
		m.approved_at MembershipApprovedAt, m.denied_at MembershipDeniedAt,
		m.message MembershipMessage,
		d.name DeletedByName, d.public_id DeletedByPublicID
	FROM players p
		LEFT OUTER JOIN memberships m on m.player_id = p.id
		LEFT OUTER JOIN clans c on c.id=m.clan_id
		LEFT OUTER JOIN players d on d.id=m.deleted_by
		LEFT OUTER JOIN players r on r.id=m.requestor_id
		LEFT OUTER JOIN players a on a.id=m.approver_id
		LEFT OUTER JOIN players y on y.id=m.denier_id
		LEFT OUTER JOIN memberships w on w.player_id = r.id
	WHERE
		p.game_id=$1 and p.public_id=$2`

	var details []playerDetailsDAO
	_, err := db.Select(&details, query, gameID, publicID)
	if err != nil {
		return nil, err
	}

	if len(details) == 0 {
		return nil, &ModelNotFoundError{"Player", publicID}
	}

	result := make(map[string]interface{})

	result["name"] = details[0].PlayerName
	result["metadata"] = details[0].PlayerMetadata
	result["publicID"] = details[0].PlayerPublicID
	result["createdAt"] = details[0].PlayerCreatedAt
	result["updatedAt"] = details[0].PlayerUpdatedAt

	if details[0].MembershipLevel.Valid {
		// Player has memberships
		memberships := []map[string]interface{}{}

		approved := []map[string]interface{}{}
		denied := []map[string]interface{}{}
		banned := []map[string]interface{}{}
		pendingApplications := []map[string]interface{}{}
		pendingInvites := []map[string]interface{}{}

		clanFromDetail := func(clanDetail playerDetailsDAO) map[string]interface{} {
			return map[string]interface{}{
				"publicID": nullOrString(clanDetail.ClanPublicID),
				"name":     nullOrString(clanDetail.ClanName),
			}
		}

		for _, detail := range details {
			ma := nullOrBool(detail.MembershipApproved)
			md := nullOrBool(detail.MembershipDenied)
			mb := nullOrBool(detail.MembershipBanned)
			mdel := !mb && detail.MembershipDeletedAt.Valid && detail.MembershipDeletedAt.Int64 > 0

			if !mdel {
				m := detail.Serialize()
				memberships = append(memberships, m)

				clanDetail := clanFromDetail(detail)
				switch {
				case !ma && !md && !mb:
					if detail.RequestorPublicID.Valid && detail.RequestorPublicID.String == detail.PlayerPublicID {
						pendingApplications = append(pendingApplications, clanDetail)
					} else {
						pendingInvites = append(pendingInvites, clanDetail)
					}
				case ma:
					approved = append(approved, clanDetail)
				case md:
					denied = append(denied, clanDetail)
				case mb:
					banned = append(banned, clanDetail)
				}
			}
		}

		result["memberships"] = memberships
		result["clans"] = map[string]interface{}{
			"approved":            approved,
			"denied":              denied,
			"banned":              banned,
			"pendingApplications": pendingApplications,
			"pendingInvites":      pendingInvites,
		}

	} else {
		result["memberships"] = []map[string]interface{}{}
		result["clans"] = map[string]interface{}{
			"approved":            []map[string]interface{}{},
			"denied":              []map[string]interface{}{},
			"banned":              []map[string]interface{}{},
			"pendingApplications": []map[string]interface{}{},
			"pendingInvites":      []map[string]interface{}{},
		}
	}

	return result, nil
}

// GetPlayerDetails returns detailed information about a player and their memberships
func GetPlayerDetails(db DB, gameID, publicID string) (map[string]interface{}, error) {
	result, err := GetPlayerMembershipDetails(db, gameID, publicID)
	if err != nil {
		return nil, err
	}
	ownerships, err := GetPlayerOwnershipDetails(db, gameID, publicID)
	if err != nil {
		return nil, err
	}
	result["clans"].(map[string]interface{})["owned"] = ownerships["clans"]
	result["memberships"] = append(result["memberships"].([]map[string]interface{}), ownerships["memberships"].([]map[string]interface{})...)
	return result, nil
}
