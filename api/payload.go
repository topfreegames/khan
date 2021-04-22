// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

//go:generate easyjson -all -no_std_marshalers $GOFILE

package api

import (
	"fmt"

	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/util"
	"github.com/uber-go/zap"
)

//Validatable indicates that a struct can be validated
type Validatable interface {
	Validate() []string
}

//ValidatePayload for any validatable payload
func ValidatePayload(payload Validatable) []string {
	return payload.Validate()
}

//NewValidation for validating structs
func NewValidation() *Validation {
	return &Validation{
		errors: []string{},
	}
}

//Validation struct
type Validation struct {
	errors []string
}

func (v *Validation) validateRequired(name string, value interface{}) {
	if value == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", name))
	}
}

func (v *Validation) validateRequiredString(name, value string) {
	if value == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", name))
	}
}

func (v *Validation) validateRequiredInt(name string, value int) {
	if value == 0 {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", name))
	}
}

func (v *Validation) validateRequiredMap(name string, value map[string]interface{}) {
	if value == nil || len(value) == 0 {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", name))
	}
}

func (v *Validation) validateCustom(name string, valFunc func() []string) {
	errors := valFunc()
	if len(errors) > 0 {
		v.errors = append(v.errors, errors...)
	}
}

//Errors in validation
func (v *Validation) Errors() []string {
	return v.errors
}

func logPayloadErrors(logger zap.Logger, errors []string) {
	var fields []zap.Field
	for _, err := range errors {
		fields = append(fields, zap.String("validationError", err))
	}
	log.W(logger, "Payload is not valid", func(cm log.CM) {
		cm.Write(fields...)
	})
}

//CreateClanPayload maps the payload for the Create Clan route
type CreateClanPayload struct {
	PublicID         string                 `json:"publicID"`
	Name             string                 `json:"name"`
	OwnerPublicID    string                 `json:"ownerPublicID"`
	Metadata         map[string]interface{} `json:"metadata"`
	AllowApplication bool                   `json:"allowApplication"`
	AutoJoin         bool                   `json:"autoJoin"`
}

//Validate all the required fields for creating a clan
func (ccp *CreateClanPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("publicID", ccp.PublicID)
	v.validateRequiredString("name", ccp.Name)
	v.validateRequiredString("ownerPublicID", ccp.OwnerPublicID)
	v.validateRequired("metadata", ccp.Metadata)
	return v.Errors()
}

//UpdateClanPayload maps the payload for the Update Clan route
type UpdateClanPayload struct {
	Name             string                 `json:"name"`
	OwnerPublicID    string                 `json:"ownerPublicID"`
	Metadata         map[string]interface{} `json:"metadata"`
	AllowApplication bool                   `json:"allowApplication"`
	AutoJoin         bool                   `json:"autoJoin"`
}

//Validate all the required fields for updating a clan
func (ucp *UpdateClanPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("name", ucp.Name)
	v.validateRequiredString("ownerPublicID", ucp.OwnerPublicID)
	v.validateRequired("metadata", ucp.Metadata)
	return v.Errors()
}

//TransferClanOwnershipPayload maps the payload for the Transfer Clan Ownership route
type TransferClanOwnershipPayload struct {
	PlayerPublicID string `json:"playerPublicID"`
}

//Validate all the required fields for transferring a clan ownership
func (tcop *TransferClanOwnershipPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("playerPublicID", tcop.PlayerPublicID)
	return v.Errors()
}

//CreatePlayerPayload maps the payload for the Create Player route
type CreatePlayerPayload struct {
	PublicID string                 `json:"publicID"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

//Validate all the required fields for creating a player
func (cpp *CreatePlayerPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("publicID", cpp.PublicID)
	v.validateRequiredString("name", cpp.Name)
	v.validateRequired("metadata", cpp.Metadata)
	return v.Errors()
}

//UpdatePlayerPayload maps the payload for the Update Player route
type UpdatePlayerPayload struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

//Validate all the required fields for updating a player
func (upp *UpdatePlayerPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("name", upp.Name)
	v.validateRequired("metadata", upp.Metadata)
	return v.Errors()
}

//UpdateGamePayload maps the payload required for the Update game route
type UpdateGamePayload struct {
	Name                          string                 `json:"name"`
	MembershipLevels              map[string]interface{} `json:"membershipLevels"`
	Metadata                      map[string]interface{} `json:"metadata"`
	MinLevelToAcceptApplication   int                    `json:"minLevelToAcceptApplication"`
	MinLevelToCreateInvitation    int                    `json:"minLevelToCreateInvitation"`
	MinLevelToRemoveMember        int                    `json:"minLevelToRemoveMember"`
	MinLevelOffsetToRemoveMember  int                    `json:"minLevelOffsetToRemoveMember"`
	MinLevelOffsetToPromoteMember int                    `json:"minLevelOffsetToPromoteMember"`
	MinLevelOffsetToDemoteMember  int                    `json:"minLevelOffsetToDemoteMember"`
	MaxMembers                    int                    `json:"maxMembers"`
	MaxClansPerPlayer             int                    `json:"maxClansPerPlayer"`
	CooldownAfterDeny             int                    `json:"cooldownAfterDeny"`
	CooldownAfterDelete           int                    `json:"cooldownAfterDelete"`
}

//Validate the update game payload
func (p *UpdateGamePayload) Validate() []string {
	v := NewValidation()

	var minMembershipLevel int

	v.validateRequiredMap("membershipLevels", p.MembershipLevels)

	if len(p.MembershipLevels) > 0 {
		sortedLevels := util.SortLevels(p.MembershipLevels)
		minMembershipLevel = sortedLevels[0].Value
	}

	v.validateRequiredString("name", p.Name)
	v.validateRequired("metadata", p.Metadata)
	v.validateRequiredInt("minLevelOffsetToRemoveMember", p.MinLevelOffsetToRemoveMember)
	v.validateRequiredInt("minLevelOffsetToPromoteMember", p.MinLevelOffsetToPromoteMember)
	v.validateRequiredInt("minLevelOffsetToDemoteMember", p.MinLevelOffsetToDemoteMember)
	v.validateRequiredInt("maxMembers", p.MaxMembers)
	v.validateRequiredInt("maxClansPerPlayer", p.MaxClansPerPlayer)

	v.validateCustom("minLevelToAcceptApplication", func() []string {
		if p.MinLevelToAcceptApplication < minMembershipLevel {
			return []string{"minLevelToAcceptApplication should be greater or equal to minMembershipLevel"}
		}
		return []string{}
	})
	v.validateCustom("minLevelToCreateInvitation", func() []string {
		if p.MinLevelToCreateInvitation < minMembershipLevel {
			return []string{"minLevelToCreateInvitation should be greater or equal to minMembershipLevel"}
		}
		return []string{}
	})
	v.validateCustom("minLevelToRemoveMember", func() []string {
		if p.MinLevelToRemoveMember < minMembershipLevel {
			return []string{"minLevelToRemoveMember should be greater or equal to minMembershipLevel"}
		}
		return []string{}
	})

	return v.Errors()
}

//CreateGamePayload maps the payload required for the Create game route
type CreateGamePayload struct {
	PublicID                      string                 `json:"publicID"`
	Name                          string                 `json:"name"`
	MembershipLevels              map[string]interface{} `json:"membershipLevels"`
	Metadata                      map[string]interface{} `json:"metadata"`
	MinLevelToAcceptApplication   int                    `json:"minLevelToAcceptApplication"`
	MinLevelToCreateInvitation    int                    `json:"minLevelToCreateInvitation"`
	MinLevelToRemoveMember        int                    `json:"minLevelToRemoveMember"`
	MinLevelOffsetToRemoveMember  int                    `json:"minLevelOffsetToRemoveMember"`
	MinLevelOffsetToPromoteMember int                    `json:"minLevelOffsetToPromoteMember"`
	MinLevelOffsetToDemoteMember  int                    `json:"minLevelOffsetToDemoteMember"`
	MaxMembers                    int                    `json:"maxMembers"`
	MaxClansPerPlayer             int                    `json:"maxClansPerPlayer"`
	CooldownAfterDeny             int                    `json:"cooldownAfterDeny"`
	CooldownAfterDelete           int                    `json:"cooldownAfterDelete"`
}

//Validate the create game payload
func (p *CreateGamePayload) Validate() []string {
	v := NewValidation()

	var minMembershipLevel int

	v.validateRequiredMap("membershipLevels", p.MembershipLevels)

	if len(p.MembershipLevels) > 0 {
		sortedLevels := util.SortLevels(p.MembershipLevels)
		minMembershipLevel = sortedLevels[0].Value
	}

	v.validateRequiredString("name", p.Name)
	v.validateRequired("metadata", p.Metadata)
	v.validateRequiredInt("minLevelOffsetToRemoveMember", p.MinLevelOffsetToRemoveMember)
	v.validateRequiredInt("minLevelOffsetToPromoteMember", p.MinLevelOffsetToPromoteMember)
	v.validateRequiredInt("minLevelOffsetToDemoteMember", p.MinLevelOffsetToDemoteMember)
	v.validateRequiredInt("maxMembers", p.MaxMembers)
	v.validateRequiredInt("maxClansPerPlayer", p.MaxClansPerPlayer)

	v.validateCustom("minLevelToAcceptApplication", func() []string {
		if p.MinLevelToAcceptApplication < minMembershipLevel {
			return []string{"minLevelToAcceptApplication should be greater or equal to minMembershipLevel"}
		}
		return []string{}
	})
	v.validateCustom("minLevelToCreateInvitation", func() []string {
		if p.MinLevelToCreateInvitation < minMembershipLevel {
			return []string{"minLevelToCreateInvitation should be greater or equal to minMembershipLevel"}
		}
		return []string{}
	})
	v.validateCustom("minLevelToRemoveMember", func() []string {
		if p.MinLevelToRemoveMember < minMembershipLevel {
			return []string{"minLevelToRemoveMember should be greater or equal to minMembershipLevel"}
		}
		return []string{}
	})

	return v.Errors()
}

//ApplyForMembershipPayload maps the payload required for the Apply for Membership route
type ApplyForMembershipPayload struct {
	Level          string `json:"level"`
	PlayerPublicID string `json:"playerPublicID"`
}

//Validate all the required fields
func (afmp *ApplyForMembershipPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("level", afmp.Level)
	v.validateRequiredString("playerPublicID", afmp.PlayerPublicID)
	return v.Errors()
}

//InviteForMembershipPayload maps the payload required for the Invite for Membership route
type InviteForMembershipPayload struct {
	Level             string `json:"level"`
	PlayerPublicID    string `json:"playerPublicID"`
	RequestorPublicID string `json:"requestorPublicID"`
}

//Validate all the required fields
func (ifmp *InviteForMembershipPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("level", ifmp.Level)
	v.validateRequiredString("playerPublicID", ifmp.PlayerPublicID)
	v.validateRequiredString("requestorPublicID", ifmp.RequestorPublicID)
	return v.Errors()
}

//BasePayloadWithRequestorAndPlayerPublicIDs maps the payload required for many routes
type BasePayloadWithRequestorAndPlayerPublicIDs struct {
	PlayerPublicID    string `json:"playerPublicID"`
	RequestorPublicID string `json:"requestorPublicID"`
}

//Validate all the required fields
func (base *BasePayloadWithRequestorAndPlayerPublicIDs) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("playerPublicID", base.PlayerPublicID)
	v.validateRequiredString("requestorPublicID", base.RequestorPublicID)
	return v.Errors()
}

//ApproveOrDenyMembershipInvitationPayload maps the payload required for Approving or Denying a membership
type ApproveOrDenyMembershipInvitationPayload struct {
	PlayerPublicID string `json:"playerPublicID"`
}

//Validate all the required fields
func (admip *ApproveOrDenyMembershipInvitationPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("playerPublicID", admip.PlayerPublicID)
	return v.Errors()
}

//HookPayload maps the payload required to create or update hooks
type HookPayload struct {
	Type    int    `json:"type"`
	HookURL string `json:"hookURL"`
}

//Validate all the required fields
func (hp *HookPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("hookURL", hp.HookURL)
	return v.Errors()
}
