package models

import "github.com/jinzhu/gorm"

//Clan identifies uniquely one clan in a given game
type Clan struct {
	gorm.Model
	ClanID   string `gorm:"column:clan_id;size:255"`
	Name     string `gorm:"size:2000"`
	GameID   string `gorm:"column:game_id;size:10"`
	Metadata string `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`
}
