// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"database/sql"
	"encoding/json"
)

type clanDetailsDAO struct {
	// Clan general information
	GameID               string
	ClanPublicID         string
	ClanName             string
	ClanMetadata         map[string]interface{}
	ClanAllowApplication bool
	ClanAutoJoin         bool
	ClanMembershipCount  int

	//Membership Information
	MembershipLevel     sql.NullString
	MembershipApproved  sql.NullBool
	MembershipDenied    sql.NullBool
	MembershipBanned    sql.NullBool
	MembershipCreatedAt sql.NullInt64
	MembershipUpdatedAt sql.NullInt64

	// Clan Owner Information
	OwnerPublicID string
	OwnerName     string
	OwnerMetadata map[string]interface{}

	// Member Information
	PlayerPublicID   sql.NullString
	PlayerName       sql.NullString
	DBPlayerMetadata sql.NullString
	PlayerMetadata   map[string]interface{}
	MembershipCount  int
	OwnershipCount   int

	// Requestor Information
	RequestorPublicID sql.NullString
	RequestorName     sql.NullString
}

func (member *clanDetailsDAO) Serialize(includeMembershipLevel bool) map[string]interface{} {
	result := map[string]interface{}{
		"player": map[string]interface{}{
			"publicID": nullOrString(member.PlayerPublicID),
			"name":     nullOrString(member.PlayerName),
		},
	}
	if member.DBPlayerMetadata.Valid {
		json.Unmarshal([]byte(nullOrString(member.DBPlayerMetadata)), &member.PlayerMetadata)
	} else {
		member.PlayerMetadata = map[string]interface{}{}
	}
	if includeMembershipLevel {
		result["level"] = nullOrString(member.MembershipLevel)
	}
	result["player"].(map[string]interface{})["metadata"] = member.PlayerMetadata
	return result
}

type playerDetailsDAO struct {
	// Player Details
	PlayerID        int
	PlayerName      string
	PlayerMetadata  map[string]interface{}
	PlayerPublicID  string
	PlayerCreatedAt int64
	PlayerUpdatedAt int64

	// Membership Details
	MembershipLevel     sql.NullString
	MembershipApproved  sql.NullBool
	MembershipDenied    sql.NullBool
	MembershipBanned    sql.NullBool
	MembershipCreatedAt sql.NullInt64
	MembershipUpdatedAt sql.NullInt64
	MembershipDeletedAt sql.NullInt64

	// Clan Details
	ClanPublicID   sql.NullString
	ClanName       sql.NullString
	DBClanMetadata sql.NullString
	ClanMetadata   map[string]interface{}
	ClanOwnerID    sql.NullInt64

	// Membership Requestor Details
	RequestorName       sql.NullString
	RequestorPublicID   sql.NullString
	DBRequestorMetadata sql.NullString
	RequestorMetadata   map[string]interface{}

	// Deleted by Details
	DeletedByName     sql.NullString
	DeletedByPublicID sql.NullString
}

func (p *playerDetailsDAO) Serialize() map[string]interface{} {
	result := map[string]interface{}{
		"level":     nullOrString(p.MembershipLevel),
		"approved":  nullOrBool(p.MembershipApproved),
		"denied":    nullOrBool(p.MembershipDenied),
		"banned":    nullOrBool(p.MembershipBanned),
		"createdAt": nullOrInt(p.MembershipCreatedAt),
		"updatedAt": nullOrInt(p.MembershipUpdatedAt),
		"deletedAt": nullOrInt(p.MembershipDeletedAt),
		"clan": map[string]interface{}{
			"publicID": nullOrString(p.ClanPublicID),
			"name":     nullOrString(p.ClanName),
		},
		"requestor": map[string]interface{}{
			"publicID": nullOrString(p.RequestorPublicID),
			"name":     nullOrString(p.RequestorName),
		},
	}

	if p.DBClanMetadata.Valid {
		json.Unmarshal([]byte(nullOrString(p.DBClanMetadata)), &p.ClanMetadata)
	} else {
		p.ClanMetadata = map[string]interface{}{}
	}
	result["clan"].(map[string]interface{})["metadata"] = p.ClanMetadata

	if p.DBRequestorMetadata.Valid {
		json.Unmarshal([]byte(nullOrString(p.DBRequestorMetadata)), &p.RequestorMetadata)
	} else {
		p.RequestorMetadata = map[string]interface{}{}
	}
	result["requestor"].(map[string]interface{})["metadata"] = p.RequestorMetadata

	if p.DeletedByPublicID.Valid {
		result["deletedBy"] = map[string]interface{}{
			"publicID": nullOrString(p.DeletedByPublicID),
			"name":     nullOrString(p.DeletedByName),
		}
	}
	return result
}
