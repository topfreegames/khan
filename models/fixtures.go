// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/satori/go.uuid"
)

//PlayerFactory is responsible for constructing test player instances
var PlayerFactory = factory.NewFactory(
	&Player{},
).Attr("GameID", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
}).Attr("PublicID", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
}).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.FullName(randomdata.RandomGender), nil
}).Attr("Metadata", func(args factory.Args) (interface{}, error) {
	return "{}", nil
})

//ClanFactory is responsible for constructing test clan instances
var ClanFactory = factory.NewFactory(
	&Clan{},
).SeqInt("GameID", func(n int) (interface{}, error) {
	return fmt.Sprintf("game-%d", n), nil
}).SeqInt("PublicID", func(n int) (interface{}, error) {
	return fmt.Sprintf("clan-%d", n), nil
}).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.FullName(randomdata.RandomGender), nil
}).Attr("Metadata", func(args factory.Args) (interface{}, error) {
	return "{}", nil
})

//MembershipFactory is responsible for constructing test membership instances
var MembershipFactory = factory.NewFactory(
	&Membership{},
).SeqInt("GameID", func(n int) (interface{}, error) {
	return fmt.Sprintf("game-%d", n), nil
})
