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
	"github.com/uber-go/zap"

	"github.com/go-gorp/gorp"
	egorp "github.com/topfreegames/extensions/v9/gorp/interfaces"
)

// Player identifies uniquely one player in a given game
type Player struct {
	ID              int64                  `db:"id"`
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
func (p *Player) Serialize(encryptionKey []byte) map[string]interface{} {
	return decryptPlayerName(map[string]interface{}{
		"gameID":          p.GameID,
		"publicID":        p.PublicID,
		"name":            p.Name,
		"metadata":        p.Metadata,
		"membershipCount": p.MembershipCount,
		"ownershipCount":  p.OwnershipCount,
	}, encryptionKey)
}

//SerializeClanParticipant the player information to JSON
func (p *Player) SerializeClanParticipant(encryptionKey []byte) map[string]interface{} {
	return decryptPlayerName(map[string]interface{}{
		"publicID": p.PublicID,
		"name":     p.Name,
		"metadata": p.Metadata,
	}, encryptionKey)
}

//SerializeClanActor the player information to JSON
func (p *Player) SerializeClanActor(encryptionKey []byte) map[string]interface{} {
	return decryptPlayerName(map[string]interface{}{
		"publicID": p.PublicID,
		"name":     p.Name,
	}, encryptionKey)
}

//SerializeWithLevel serialize player fields: PublicID and Name with MembershipCount passed by param
func (p *Player) SerializeWithLevel(encryptionKey []byte, level string) map[string]interface{} {
	return decryptPlayerName(map[string]interface{}{
		"publicID": p.PublicID,
		"name":     p.Name,
		"level":    level,
	}, encryptionKey)
}

// UpdatePlayerMembershipCount updates the player membership count
func UpdatePlayerMembershipCount(db DB, id int64) error {
	query := `
	UPDATE players SET membership_count=membership.count
	FROM (
		SELECT COUNT(*) as count
		FROM memberships m
		WHERE
			m.player_id = $1 AND m.deleted_at = 0 AND m.approved = true AND
			m.denied = false AND m.banned = false
	) as membership
	WHERE players.id=$1
	`
	res, err := db.Exec(query, id)
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

// UpdatePlayerOwnershipCount updates the player ownership count
func UpdatePlayerOwnershipCount(db DB, id int64) error {
	query := `
	UPDATE players SET ownership_count=ownership.count
	FROM (
		SELECT COUNT(*) as count
		FROM clans c
		WHERE c.owner_id = $1
	) as ownership
	WHERE players.id=$1
	`
	res, err := db.Exec(query, id)
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
func GetPlayerByID(db DB, encryptionKey []byte, id int64) (*Player, error) {
	playerInterface, err := db.Get(Player{}, id)
	if err != nil {
		return nil, err
	}
	if playerInterface == nil {
		return nil, &ModelNotFoundError{"Player", id}
	}

	player := playerInterface.(*Player)
	name, err := util.DecryptData(player.Name, encryptionKey)
	if err != nil {
		return player, nil
	}

	player.Name = name
	return player, nil

}

// GetPlayerByPublicID returns a player by their public id
func GetPlayerByPublicID(db DB, encryptionKey []byte, gameID string, publicID string) (*Player, error) {
	var players []*Player
	_, err := db.Select(&players, "SELECT * FROM players WHERE game_id=$1 AND public_id=$2", gameID, publicID)
	if err != nil {
		return nil, err
	}
	if players == nil || len(players) < 1 {
		return nil, &ModelNotFoundError{"Player", publicID}
	}

	player := players[0]
	name, err := util.DecryptData(player.Name, encryptionKey)
	if err != nil {
		return player, nil
	}

	player.Name = name
	return player, nil
}

// CreatePlayer creates a new player
func CreatePlayer(db DB, logger zap.Logger, encryptionKey []byte, gameID, publicID, name string, metadata map[string]interface{}) (*Player, error) {
	markAsEncrypted := true
	encryptedName, err := util.EncryptData(name, encryptionKey)
	if err != nil {
		encryptedName = name
		markAsEncrypted = false
	}

	player := &Player{
		GameID:   gameID,
		PublicID: publicID,
		Name:     encryptedName,
		Metadata: metadata,
	}
	err = db.Insert(player)
	if err != nil {
		return nil, err
	}

	if markAsEncrypted {
		err = db.Insert(&EncryptedPlayer{PlayerID: player.ID})
		if err != nil {
			logger.Error("Error on insert EncryptedPlayer", zap.Error(err))
		}
	}

	return GetPlayerByID(db, encryptionKey, player.ID)
}

// UpdatePlayer updates an existing player
func UpdatePlayer(db DB, logger zap.Logger, encryptionKey []byte, gameID, publicID, name string, metadata map[string]interface{}) (*Player, error) {
	markAsEncrypted := true
	encryptedName, err := util.EncryptData(name, encryptionKey)
	if err != nil {
		encryptedName = name
		markAsEncrypted = false
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO players(game_id, public_id, name, metadata, created_at, updated_at)
						VALUES($1, $2, $3, $4, $5, $5) ON CONFLICT (game_id, public_id)
						DO UPDATE set name=$3, metadata=$4, updated_at=$5
						WHERE players.game_id=$1 and players.public_id=$2
						RETURNING id`

	var lastID int64
	lastID, err = db.SelectInt(query,
		gameID, publicID, encryptedName, metadataJSON, util.NowMilli())
	if err != nil {
		return nil, err
	}

	if markAsEncrypted {
		queryEncrypt := `INSERT INTO encrypted_players (player_id) VALUES ($1) ON CONFLICT DO NOTHING`
		_, err = db.Exec(queryEncrypt, lastID)
		if err != nil {
			logger.Error("Error on insert EncryptedPlayer", zap.Error(err))
		}
	}

	return GetPlayerByID(db, encryptionKey, lastID)
}

// GetPlayerOwnershipDetails returns detailed information about a player owned clans
func GetPlayerOwnershipDetails(db DB, gameID, publicID string) (map[string]interface{}, error) {
	query := `
	SELECT c.*
	FROM players p
	INNER JOIN clans c ON c.owner_id=p.id
	WHERE p.game_id=$1 AND p.public_id=$2 
	`

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

// GetPlayerDetails returns detailed information about a player and their memberships
func GetPlayerDetails(db DB, encryptionKey []byte, gameID, publicID string) (map[string]interface{}, error) {
	result, err := getPlayerMembershipDetails(db, encryptionKey, gameID, publicID)
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

// GetPlayersToEncrypt get players that have plain text name
func GetPlayersToEncrypt(db DB, encryptionKey []byte, amount int) ([]*Player, error) {
	query := `SELECT p.*
	FROM players p
		LEFT JOIN encrypted_players ep ON p.id = ep.player_id
	WHERE ep.player_id IS NULL
	LIMIT $1`

	var players []*Player
	_, err := db.Select(&players, query, amount)
	if err != nil {
		return nil, err
	}

	return players, nil
}

// ApplySecurityChanges encrypt and update player
func ApplySecurityChanges(db egorp.Database, encryptionKey []byte, players []*Player) error {

	trx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, player := range players {
		encryptedName, err := util.EncryptData(player.Name, encryptionKey)
		if err != nil {
			return err
		}

		player.Name = encryptedName

		if err != nil {
			err = trx.Rollback()
			return err
		}

		_, err = trx.Update(player)
		if err != nil {
			err = trx.Rollback()
			return err
		}

		err = trx.Insert(&EncryptedPlayer{PlayerID: player.ID})
		if err != nil {
			err = trx.Rollback()
			return err
		}
	}

	err = trx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// getPlayerMembershipDetails returns detailed information about a player and their memberships
func getPlayerMembershipDetails(db DB, encryptionKey []byte, gameID, publicID string) (map[string]interface{}, error) {
	player, err := GetPlayerByPublicID(db, encryptionKey, gameID, publicID)
	if err != nil {
		return nil, err
	}
	//TODO: Include this again once membership level is in the membership table
	//w.membership_level RequestorMembershipLevel,
	query := `
	SELECT
		p.id PlayerID, p.name PlayerName, p.metadata PlayerMetadata, p.public_id PlayerPublicID,
		p.created_at PlayerCreatedAt, p.updated_at PlayerUpdatedAt,
		m.membership_level MembershipLevel,
		m.approved MembershipApproved, m.denied MembershipDenied, m.banned MembershipBanned,
		c.public_id ClanPublicID, c.name ClanName, c.metadata DBClanMetadata, c.owner_id ClanOwnerID,
		c.membership_count ClanMembershipCount,
		NULL RequestorMembershipLevel,
		r.name RequestorName, r.public_id RequestorPublicID, r.metadata DBRequestorMetadata,
		a.name ApproverName, a.public_id ApproverPublicID, a.metadata DBApproverMetadata,
		y.name DenierName, y.public_id DenierPublicID, y.metadata DBDenierMetadata,
		m.created_at MembershipCreatedAt,
		m.updated_at MembershipUpdatedAt,
		m.deleted_at MembershipDeletedAt,
		m.approved_at MembershipApprovedAt, m.denied_at MembershipDeniedAt,
		m.message MembershipMessage,
		d.name DeletedByName, d.public_id DeletedByPublicID
	FROM players p
		LEFT OUTER JOIN (
			SELECT * FROM memberships im WHERE im.player_id=$2 AND (im.approved=true OR im.denied=true OR im.banned=true)
			UNION
			(SELECT * FROM memberships im WHERE im.player_id=$2 AND im.deleted_at=0 AND im.approved=false AND im.denied=false AND im.banned=false ORDER BY updated_at DESC LIMIT $3)
		) m ON p.id = m.player_id
		LEFT OUTER JOIN clans c on c.id=m.clan_id
		LEFT OUTER JOIN players d on d.id=m.deleted_by
		LEFT OUTER JOIN players r on r.id=m.requestor_id
		LEFT OUTER JOIN players a on a.id=m.approver_id
		LEFT OUTER JOIN players y on y.id=m.denier_id
	WHERE
		p.game_id=$1 and p.id=$2`

	var details []playerDetailsDAO
	_, err = db.Select(&details, query, gameID, player.ID, 5)
	if err != nil {
		return nil, err
	}

	if len(details) == 0 {
		return nil, &ModelNotFoundError{"Player", publicID}
	}

	result := make(map[string]interface{})

	result["name"], err = util.DecryptData(details[0].PlayerName, encryptionKey)
	if err != nil {
		result["name"] = details[0].PlayerName
	}
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
			approvedMembership := nullOrBool(detail.MembershipApproved)
			deniedMembership := nullOrBool(detail.MembershipDenied)
			bannedMembership := nullOrBool(detail.MembershipBanned)
			deletedMembership := !bannedMembership && detail.MembershipDeletedAt.Valid && detail.MembershipDeletedAt.Int64 > 0

			if !deletedMembership {
				membership := detail.Serialize(encryptionKey)
				memberships = append(memberships, membership)

				clanDetail := clanFromDetail(detail)
				switch {
				case !approvedMembership && !deniedMembership && !bannedMembership:
					if detail.RequestorPublicID.Valid && detail.RequestorPublicID.String == detail.PlayerPublicID {
						pendingApplications = append(pendingApplications, clanDetail)
					} else {
						pendingInvites = append(pendingInvites, clanDetail)
					}
				case approvedMembership:
					approved = append(approved, clanDetail)
				case deniedMembership:
					denied = append(denied, clanDetail)
				case bannedMembership:
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

func decryptPlayerName(payload map[string]interface{}, encryptionKey []byte) map[string]interface{} {
	name, err := util.DecryptData(fmt.Sprint(payload["name"]), encryptionKey)
	if err != nil {
		return payload
	}

	payload["name"] = name
	return payload
}
