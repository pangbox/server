// Copyright (C) 2018-2023, John Chadwick <john@jchw.io>
//
// Permission to use, copy, modify, and/or distribute this software for any purpose
// with or without fee is hereby granted, provided that the above copyright notice
// and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
// FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
// OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
// TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
// THIS SOFTWARE.
//
// SPDX-FileCopyrightText: Copyright (c) 2018-2023 John Chadwick
// SPDX-License-Identifier: ISC

package game

import (
	"github.com/pangbox/server/common"
	"github.com/pangbox/server/pangya"
)

var ClientMessageTable = common.NewMessageTable(map[uint16]ClientMessage{
	0x0002: &ClientAuth{},
	0x0003: &ClientMessageSend{},
	0x0004: &ClientJoinChannel{},
	0x0007: &ClientGetUserOnlineStatus{},
	0x0008: &ClientRoomCreate{},
	0x000A: &ClientRoomEdit{},
	0x000B: &ClientTutorialStart{},
	0x000C: &ClientRoomUserEquipmentChange{},
	0x000D: &ClientPlayerReady{},
	0x000F: &ClientRoomLeave{},
	0x001A: &Client001A{},
	0x001D: &ClientBuyItem{},
	0x0020: &ClientEquipmentUpdate{},
	0x002F: &ClientGetUserData{},
	0x0033: &ClientException{},
	0x0043: &ClientRequestServerList{},
	0x0048: &ClientUnknownCounter{},
	0x0063: &ClientRoomLoungeAction{},
	0x0069: &ClientUserMacrosSet{},
	0x0081: &ClientMultiplayerJoin{},
	0x0082: &ClientMultiplayerLeave{},
	0x0088: &Client0088{},
	0x008B: &ClientRequestMessengerList{},
	0x009C: &Client009C{},
	0x00AE: &ClientTutorialClear{},
	0x00FE: &Client00FE{},
	0x0140: &ClientShopJoin{},
	0x0143: &ClientRequestInboxList{},
	0x0144: &ClientRequestInboxMessage{},
	0x016E: &ClientRequestDailyReward{},
	0x0176: &ClientEventLobbyJoin{},
	0x0177: &ClientEventLobbyLeave{},
	0x0184: &ClientAssistModeToggle{},
})

// ClientAuth is a message sent to authenticate a session.
type ClientAuth struct {
	ClientMessage_
	Username common.PString
	Unknown1 uint32
	Unknown2 uint32
	Unknown3 uint16
	LoginKey common.PString
	Version  common.PString
}

// ClientMessageSend is sent when the client sends a public chat message.
type ClientMessageSend struct {
	ClientMessage_
	Nickname common.PString
	Message  common.PString
}

// ClientGetUserOnlineStatus is sent to get information of a user.
type ClientGetUserOnlineStatus struct {
	ClientMessage_
	Unknown  uint8
	Username common.PString
}

type SettingsChange struct {
	Type             byte
	RoomName         *common.PString `struct-if:"Type == 0"`
	Password         *common.PString `struct-if:"Type == 1"`
	RoomType         *byte           `struct-if:"Type == 2"`
	Course           *byte           `struct-if:"Type == 3"`
	NumHoles         *uint8          `struct-if:"Type == 4"`
	HoleProgression  *uint8          `struct-if:"Type == 5"`
	ShotTimerSeconds *uint8          `struct-if:"Type == 6"`
	MaxUsers         *uint8          `struct-if:"Type == 7"`
	GameTimerMinutes *uint8          `struct-if:"Type == 8"`
	ArtifactID       *uint32         `struct-if:"Type == 13"`
	NaturalWind      *uint32         `struct-if:"Type == 14"`
}

// ClientRoomEdit is sent when the client changes room settings.
type ClientRoomEdit struct {
	ClientMessage_
	Unknown    uint16
	NumChanges uint8 `struct:"sizeof=Changes"`
	Changes    []SettingsChange
}

// ClientRoomCreate is sent by the client when creating a room.
type ClientRoomCreate struct {
	ClientMessage_
	Unknown     byte
	ShotTimerMS uint32
	GameTimerMS uint32
	MaxUsers    uint8
	RoomType    byte
	NumHoles    byte
	Course      byte
	Unknown2    [5]byte
	RoomName    common.PString
	Password    common.PString
	Unknown3    [4]byte
}

// ClientJoinChannel is a message sent when the client joins a channel.
type ClientJoinChannel struct {
	ClientMessage_
	ChannelID byte
}

// ClientTutorialStart is sent when starting a tutorial.
type ClientTutorialStart struct {
	ClientMessage_
	// TODO
}

// ClientRoomUserEquipmentChange is sent when a user's equipment changes in a room.
type ClientRoomUserEquipmentChange struct {
	ClientMessage_
	// TODO
}

// ClientPlayerReady is sent by the client when they are ready/to start the game.
type ClientPlayerReady struct {
	ClientMessage_
	State byte
}

// ClientRoomLeave is sent by the client when leaving a room back to lobby
type ClientRoomLeave struct {
	ClientMessage_
	Unknown    byte
	RoomNumber uint16
	Unknown2   uint32
	Unknown3   uint32
	Unknown4   uint32
	Unknown5   uint32
}

type Client001A struct {
	ClientMessage_
}

type PurchaseItem struct {
	Unknown        uint32
	ItemID         uint32
	Unknown2       uint16
	Unknown3       uint16
	Quantity       uint32
	ItemCostPang   uint32
	ItemCostCookie uint32
}

// ClientBuyItem is sent by the client to buy an item from the shop.
type ClientBuyItem struct {
	ClientMessage_
	Unknown1 byte
	NumItems uint16 `struct:"sizeof=Items"`
	Items    []PurchaseItem
}

// ClientEquipmentUpdate
type ClientEquipmentUpdate struct {
	ClientMessage_
}

// ClientRequestServerList is a message sent to request the current
// list of game servers.
type ClientRequestServerList struct {
	ClientMessage_
}

type ClientUnknownCounter struct {
	ClientMessage_
	Unknown uint8
}

type LoungeActionRotation struct {
	Z float32
}

type LoungeActionPosition struct {
	X, Y, Z float32
}

type LoungeAction struct {
	ActionType  byte
	Rotation    *LoungeActionRotation `struct-if:"ActionType == 0"`
	PositionAbs *LoungeActionRotation `struct-if:"ActionType == 4"`
	PositionRel *LoungeActionRotation `struct-if:"ActionType == 6"`
	Emote       *common.PString       `struct-if:"ActionType == 7"`
	Departure   *uint32               `struct-if:"ActionType == 8"`
}

// ClientRoomLoungeAction
type ClientRoomLoungeAction struct {
	ClientMessage_
	LoungeAction
}

// ClientRequestMessengerList is a message sent to request the current
// list of message servers.
type ClientRequestMessengerList struct {
	ClientMessage_
}

// ClientGetUserData is a message sent by the client to request
// the client state.
type ClientGetUserData struct {
	ClientMessage_
	UserID  uint32
	Request byte
}

// ClientException is a message sent when the client encounters an
// error.
type ClientException struct {
	ClientMessage_
	Empty   byte
	Message common.PString
}

// Client009C is an unknown message.
type Client009C struct {
	ClientMessage_
}

// ClientTutorialClear is an unknown message.
type ClientTutorialClear struct {
	ClientMessage_
}

// ClientShopJoin is an unknown message.
type ClientShopJoin struct {
	ClientMessage_
}

// ClientUserMacrosSet is a message sent to set the user's macros.
type ClientUserMacrosSet struct {
	ClientMessage_
	MacroList pangya.MacroList
}

// ClientMultiplayerJoin is the message sent when joining multiplayer.
type ClientMultiplayerJoin struct {
	ClientMessage_
}

// ClientMultiplayerLeave is sent when the client exits multiplayer mode.
type ClientMultiplayerLeave struct {
	ClientMessage_
}

// ClientEventLobbyJoin is the message sent when joining the event lobby.
type ClientEventLobbyJoin struct {
	ClientMessage_
}

// ClientEventLobbyLeave is sent when the client exits the event lobby.
type ClientEventLobbyLeave struct {
	ClientMessage_
}

// ClientAssistModeToggle is sent when assist mode is toggled.
type ClientAssistModeToggle struct {
	ClientMessage_
}

// Client0088 is an unknown message.
type Client0088 struct {
	ClientMessage_
}

// ClientRequestInboxList is the message sent to request the inbox.
type ClientRequestInboxList struct {
	ClientMessage_
	PageNum uint32
}

// ClientRequestInboxMessage is sent by the client to retrieve a message from the inbox.
type ClientRequestInboxMessage struct {
	ClientMessage_
	MessageID uint32
}

// ClientRequestDailyReward is the message sent to request the daily
// reward.
type ClientRequestDailyReward struct {
	ClientMessage_
}

type Client00FE struct {
	ClientMessage_
}
