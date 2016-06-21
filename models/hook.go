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
	//GameCreatedHook indicates an event that happens when a game is created
	GameCreatedHook = iota
	//GameUpdatedHook indicates an event that happens when a game is updated
	GameUpdatedHook
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
	h.CreatedAt = time.Now().UnixNano()
	h.UpdatedAt = h.CreatedAt
	return nil
}

// PreUpdate populates fields before updating a hook
func (h *Hook) PreUpdate(s gorp.SqlExecutor) error {
	h.UpdatedAt = time.Now().UnixNano()
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

// CreateHook returns a newly created event hook
func CreateHook(db DB, gameID string, eventType int, url string) (*Hook, error) {
	publicID := uuid.NewV4().String()
	hook := &Hook{
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
