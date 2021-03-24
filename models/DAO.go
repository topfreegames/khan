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
	MembershipLevel      sql.NullString
	MembershipApproved   sql.NullBool
	MembershipDenied     sql.NullBool
	MembershipBanned     sql.NullBool
	MembershipCreatedAt  sql.NullInt64
	MembershipUpdatedAt  sql.NullInt64
	MembershipApprovedAt sql.NullInt64
	MembershipDeniedAt   sql.NullInt64
	MembershipMessage    sql.NullString

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

	// Approver Information
	ApproverPublicID sql.NullString
	ApproverName     sql.NullString

	// Denier Information
	DenierPublicID sql.NullString
	DenierName     sql.NullString
}

func (member *clanDetailsDAO) Serialize(encryptionKey []byte, includeMembershipLevel bool) map[string]interface{} {
	player := &Player{
		PublicID: nullOrString(member.PlayerPublicID),
		Name:     nullOrString(member.PlayerName),
	}

	if member.DBPlayerMetadata.Valid {
		json.Unmarshal([]byte(nullOrString(member.DBPlayerMetadata)), &member.PlayerMetadata)
	} else {
		member.PlayerMetadata = map[string]interface{}{}
	}
	player.Metadata = member.PlayerMetadata
	result := map[string]interface{}{
		"player": player.SerializeClanParticipant(encryptionKey),
	}

	if includeMembershipLevel {
		result["level"] = nullOrString(member.MembershipLevel)
	}

	if member.ApproverName.Valid {
		approver := &Player{
			PublicID: member.ApproverPublicID.String,
			Name:     member.ApproverName.String,
		}
		result["player"].(map[string]interface{})["approver"] = approver.SerializeClanActor(encryptionKey)
	} else if member.DenierName.Valid {
		denier := &Player{
			PublicID: member.DenierPublicID.String,
			Name:     member.DenierName.String,
		}
		result["player"].(map[string]interface{})["denier"] = denier.SerializeClanActor(encryptionKey)
	}
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
	MembershipLevel      sql.NullString
	MembershipApproved   sql.NullBool
	MembershipDenied     sql.NullBool
	MembershipBanned     sql.NullBool
	MembershipCreatedAt  sql.NullInt64
	MembershipUpdatedAt  sql.NullInt64
	MembershipDeletedAt  sql.NullInt64
	MembershipApprovedAt sql.NullInt64
	MembershipDeniedAt   sql.NullInt64
	MembershipMessage    sql.NullString

	// Clan Details
	ClanPublicID        sql.NullString
	ClanName            sql.NullString
	DBClanMetadata      sql.NullString
	ClanMetadata        map[string]interface{}
	ClanOwnerID         sql.NullInt64
	ClanMembershipCount sql.NullInt64

	// Membership Requestor Details
	RequestorName            sql.NullString
	RequestorPublicID        sql.NullString
	DBRequestorMetadata      sql.NullString
	RequestorMetadata        map[string]interface{}
	RequestorMembershipLevel sql.NullString

	// Membership Approver Details
	ApproverName       sql.NullString
	ApproverPublicID   sql.NullString
	DBApproverMetadata sql.NullString
	ApproverMetadata   map[string]interface{}

	// Membership Denier Details
	DenierName       sql.NullString
	DenierPublicID   sql.NullString
	DBDenierMetadata sql.NullString
	DenierMetadata   map[string]interface{}

	// Deleted by Details
	DeletedByName     sql.NullString
	DeletedByPublicID sql.NullString
}

func (p *playerDetailsDAO) Serialize(encryptionKey []byte) map[string]interface{} {
	requestor := &Player{
		PublicID: nullOrString(p.RequestorPublicID),
		Name:     nullOrString(p.RequestorName),
	}
	result := map[string]interface{}{
		"level":      nullOrString(p.MembershipLevel),
		"approved":   nullOrBool(p.MembershipApproved),
		"denied":     nullOrBool(p.MembershipDenied),
		"banned":     nullOrBool(p.MembershipBanned),
		"createdAt":  nullOrInt(p.MembershipCreatedAt),
		"updatedAt":  nullOrInt(p.MembershipUpdatedAt),
		"deletedAt":  nullOrInt(p.MembershipDeletedAt),
		"approvedAt": nullOrInt(p.MembershipApprovedAt),
		"deniedAt":   nullOrInt(p.MembershipDeniedAt),
		"message":    nullOrString(p.MembershipMessage),
		"clan": map[string]interface{}{
			"publicID":        nullOrString(p.ClanPublicID),
			"name":            nullOrString(p.ClanName),
			"membershipCount": nullOrInt(p.ClanMembershipCount),
		},
		"requestor": requestor.SerializeWithLevel(encryptionKey, nullOrString(p.RequestorMembershipLevel)),
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
		deleter := &Player{
			PublicID: nullOrString(p.DeletedByPublicID),
			Name:     nullOrString(p.DeletedByName),
		}
		result["deletedBy"] = deleter.SerializeClanActor(encryptionKey)
	}

	if p.ApproverPublicID.Valid {
		if p.DBApproverMetadata.Valid {
			json.Unmarshal([]byte(nullOrString(p.DBApproverMetadata)), &p.ApproverMetadata)
		} else {
			p.ApproverMetadata = map[string]interface{}{}
		}

		approver := &Player{
			PublicID: nullOrString(p.ApproverPublicID),
			Name:     nullOrString(p.ApproverName),
			Metadata: p.ApproverMetadata,
		}
		result["approver"] = approver.SerializeClanParticipant(encryptionKey)
	}

	if p.DenierPublicID.Valid {
		if p.DBDenierMetadata.Valid {
			json.Unmarshal([]byte(nullOrString(p.DBDenierMetadata)), &p.DenierMetadata)
		} else {
			p.DenierMetadata = map[string]interface{}{}
		}

		denier := &Player{
			PublicID: nullOrString(p.DenierPublicID),
			Name:     nullOrString(p.DenierName),
			Metadata: p.DenierMetadata,
		}
		result["denier"] = denier.SerializeClanParticipant(encryptionKey)
	}
	return result
}
