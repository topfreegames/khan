// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"time"

	"github.com/satori/go.uuid"

	"gopkg.in/gorp.v1"
)

const (
	//GameUpdatedHook happens when a game is updated
	GameUpdatedHook = iota

	//PlayerCreatedHook happens when a new player is created
	PlayerCreatedHook

	//PlayerUpdatedHook happens when a player is updated
	PlayerUpdatedHook

	//ClanCreatedHook happens when a clan is created
	ClanCreatedHook

	//ClanUpdatedHook happens when a clan is updated
	ClanUpdatedHook
)

// Hook identifies a webhook for a given event
type Hook struct {
	ID        int    `db:"id"`
	GameID    string `db:"game_id"`
	PublicID  string `db:"public_id"`
	EventType int    `db:"event_type"`
	URL       string `db:"url"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

// PreInsert populates fields before inserting a new hook
func (h *Hook) PreInsert(s gorp.SqlExecutor) error {
	h.CreatedAt = time.Now().UnixNano() / 1000000
	h.UpdatedAt = h.CreatedAt
	return nil
}

// PreUpdate populates fields before updating a hook
func (h *Hook) PreUpdate(s gorp.SqlExecutor) error {
	h.UpdatedAt = time.Now().UnixNano() / 1000000
	return nil
}

// GetHookByID returns a hook by id
func GetHookByID(db DB, id int) (*Hook, error) {
	obj, err := db.Get(Hook{}, id)
	if err != nil || obj == nil {
		return nil, &ModelNotFoundError{"Hook", id}
	}

	hook := obj.(*Hook)
	return hook, nil
}

// GetHookByPublicID returns a hook by game id and public id
func GetHookByPublicID(db DB, gameID string, publicID string) (*Hook, error) {
	var hook Hook
	err := db.SelectOne(&hook, "SELECT * FROM hooks WHERE game_id=$1 AND public_id=$2", gameID, publicID)
	if err != nil || &hook == nil {
		return nil, &ModelNotFoundError{"Hook", publicID}
	}
	return &hook, nil
}

// GetHookByDetails returns a hook by its details (GameID, EventType and Hook URL)
// If no hook is found returns nil.
func GetHookByDetails(db DB, gameID string, eventType int, hookURL string) *Hook {
	var hook Hook
	err := db.SelectOne(&hook, "SELECT * FROM hooks WHERE game_id=$1 AND event_type=$2 AND url=$3", gameID, eventType, hookURL)
	if err != nil || &hook == nil {
		return nil
	}
	return &hook
}

// CreateHook returns a newly created event hook
func CreateHook(db DB, gameID string, eventType int, url string) (*Hook, error) {
	hook := GetHookByDetails(db, gameID, eventType, url)

	if hook != nil {
		return hook, nil
	}

	publicID := uuid.NewV4().String()
	hook = &Hook{
		GameID:    gameID,
		PublicID:  publicID,
		EventType: eventType,
		URL:       url,
	}
	err := db.Insert(hook)
	if err != nil {
		return nil, err
	}
	return hook, nil
}

// RemoveHook removes a hook by public ID
func RemoveHook(db DB, gameID string, publicID string) error {
	hook, err := GetHookByPublicID(db, gameID, publicID)
	if err != nil {
		return err
	}
	_, err = db.Delete(hook)
	return err
}

// GetAllHooks returns all the available hooks
func GetAllHooks(db DB) ([]*Hook, error) {
	var hooks []*Hook
	_, err := db.Select(&hooks, "SELECT * FROM hooks")
	if err != nil {
		return nil, err
	}
	return hooks, nil
}
