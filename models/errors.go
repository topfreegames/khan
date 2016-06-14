// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import "fmt"

//ModelNotFoundError identifies that a given model was not found in the Database with the given ID
type ModelNotFoundError struct {
	Type string
	ID   interface{}
}

func (e *ModelNotFoundError) Error() string {
	return fmt.Sprintf("%s was not found with id: %v", e.Type, e.ID)
}

//EmptyGameIDError identifies that a request was made for a model without the proper game id
type EmptyGameIDError struct {
	Type string
}

func (e *EmptyGameIDError) Error() string {
	return fmt.Sprintf("Game ID is required to retrieve %s!", e.Type)
}

//PlayerCannotCreateMembershipError identifies that a given player is not allowed to create a membership
type PlayerCannotCreateMembershipError struct {
	PlayerID interface{}
	ClanID   interface{}
}

func (e *PlayerCannotCreateMembershipError) Error() string {
	return fmt.Sprintf("Player %v cannot create membership for clan %v", e.PlayerID, e.ClanID)
}

//PlayerCannotApproveOrDenyMembershipError identifies that a given player is not allowed to accept/refuse a membership
type PlayerCannotApproveOrDenyMembershipError struct {
	Action      string
	PlayerID    interface{}
	ClanID      interface{}
	RequestorID interface{}
}

func (e *PlayerCannotApproveOrDenyMembershipError) Error() string {
	return fmt.Sprintf("Player %v cannot %s membership for player %v and clan %v", e.RequestorID, e.Action, e.PlayerID, e.ClanID)
}

//CannotApproveOrDenyMembershipAlreadyProcessedError identifies that a given player is not allowed to accept/refuse a membership
type CannotApproveOrDenyMembershipAlreadyProcessedError struct {
	Action string
}

func (e *CannotApproveOrDenyMembershipAlreadyProcessedError) Error() string {
	return fmt.Sprintf("Cannot %s membership that was already approved or denied", e.Action)
}

//InvalidMembershipActionError identifies that a given action is not valid
type InvalidMembershipActionError struct {
	Action string
}

func (e *InvalidMembershipActionError) Error() string {
	return fmt.Sprintf("%s a membership is not a valid action.", e.Action)
}
