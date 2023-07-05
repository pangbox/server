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

package gamepacket

import (
	"github.com/pangbox/server/common"
	gamemodel "github.com/pangbox/server/game/model"
	"github.com/pangbox/server/pangya"
)

var ClientMessageTable = common.NewMessageTable(map[uint16]ClientMessage{
	0x0002: &ClientAuth{},
	0x0003: &ClientMessageSend{},
	0x0004: &ClientJoinChannel{},
	0x0006: &ClientGameEnd{},
	0x0007: &ClientGetUserOnlineStatus{},
	0x0008: &ClientRoomCreate{},
	0x0009: &ClientRoomJoin{},
	0x000A: &ClientRoomEdit{},
	0x000B: &ClientTutorialStart{},
	0x000C: &ClientRoomUserEquipmentChange{},
	0x000D: &ClientPlayerReady{},
	0x000E: &ClientPlayerStartGame{},
	0x000F: &ClientRoomLeave{},
	0x0011: &ClientReadyStartHole{},
	0x0012: &ClientShotCommit{},
	0x0013: &ClientShotRotate{},
	0x0014: &ClientShotMeterInput{},
	0x0015: &ClientShotPower{},
	0x0016: &ClientShotClubChange{},
	0x0017: &ClientShotItemUse{},
	0x0018: &ClientUserTypingIndicator{},
	0x0019: &ClientShotCometRelief{},
	0x001A: &ClientHoleInfo{},
	0x001B: &ClientShotSync{},
	0x001C: &ClientRoomSync{},
	0x001D: &ClientBuyItem{},
	0x0020: &ClientEquipmentUpdate{},
	0x0022: &ClientShotActiveUserAcknowledge{},
	0x0026: &ClientRoomKick{},
	0x002D: &ClientRoomInfo{},
	0x002F: &ClientGetPlayerData{},
	0x0030: &ClientPauseGame{},
	0x0031: &ClientHoleEnd{},
	0x0032: &ClientSetIdleStatus{},
	0x0033: &ClientException{},
	0x0034: &ClientFirstShotReady{},
	0x0037: &ClientLastPlayerLeaveGame{},
	0x0042: &ClientShotArrow{},
	0x0043: &ClientRequestServerList{},
	0x0048: &ClientLoadProgress{},
	0x004F: &Client004F{},
	0x0063: &ClientRoomLoungeAction{},
	0x0069: &ClientUserMacrosSet{},
	0x0081: &ClientMultiplayerJoin{},
	0x0082: &ClientMultiplayerLeave{},
	0x0088: &Client0088{},
	0x008B: &ClientRequestMessengerList{},
	0x0098: &ClientRareShopOpen{},
	0x009C: &ClientRequestPlayerHistory{},
	0x00AE: &ClientTutorialClear{},
	0x00B5: &ClientEnterMyRoom{},
	0x00B7: &ClientRequestInventory{},
	0x00C1: &Client00C1{},
	0x00CC: &ClientLockerCombinationAttempt{},
	0x00D3: &ClientLockerInventoryRequest{},
	0x00FE: &Client00FE{},
	0x0108: &ClientGuildListRequest{},
	0x012A: &ClientScratchyMenuOpen{},
	0x0140: &ClientShopJoin{},
	0x0143: &ClientRequestInboxList{},
	0x0144: &ClientRequestInboxMessage{},
	0x014B: &ClientBlackPapelPlay{},
	0x0151: &ClientQuestStatusRequest{},
	0x0157: &ClientAchievementStatusRequest{},
	0x016E: &ClientRequestDailyReward{},
	0x0176: &ClientEventLobbyJoin{},
	0x0177: &ClientEventLobbyLeave{},
	0x0184: &ClientAssistModeToggle{},
	0x0186: &ClientBigPapelPlay{},
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

// ClientRareShopOpen notifies the server if a user opens the rare shop menu.
type ClientRareShopOpen struct {
	ClientMessage_
}

// ClientRoomEdit is sent when the client changes room settings.
type ClientRoomEdit struct {
	ClientMessage_
	Unknown    uint16
	NumChanges uint8 `struct:"sizeof=Changes"`
	Changes    []gamemodel.RoomSettingsChange
}

// ClientRoomCreate is sent by the client when creating a room.
type ClientRoomCreate struct {
	ClientMessage_
	Unknown         byte
	ShotTimerMS     uint32
	GameTimerMS     uint32
	MaxUsers        uint8
	RoomType        byte
	NumHoles        byte
	Course          byte
	HoleProgression byte
	Unknown2        [4]byte
	RoomName        common.PString
	Password        common.PString
	Unknown3        [4]byte
}

// ClientRoomJoin is sent by the client when joining a room.
type ClientRoomJoin struct {
	ClientMessage_
	RoomNumber   int16
	RoomPassword common.PString
}

// ClientJoinChannel is a message sent when the client joins a channel.
type ClientJoinChannel struct {
	ClientMessage_
	ChannelID byte
}

// ClientGameEnd contains information after the end of a game.
type ClientGameEnd struct {
	ClientMessage_
	// TODO
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

// ClientPlayerStartGame
type ClientPlayerStartGame struct {
	ClientMessage_
	Unknown uint32
}

// ClientRoomLeave is sent by the client when leaving a room back to lobby
type ClientRoomLeave struct {
	ClientMessage_
	Unknown    byte
	RoomNumber int16
	Unknown2   uint32
	Unknown3   uint32
	Unknown4   uint32
	Unknown5   uint32
}

type ClientReadyStartHole struct {
	ClientMessage_
}

type ClientShotCommit struct {
	ClientMessage_
	UnknownFlag      bool    `struct:"uint16"`
	Unknown          [9]byte `struct-if:"UnknownFlag"`
	ShotStrength     float32
	ShotAccuracy     float32
	ShotEnglishCurve float32
	ShotEnglishSpin  float32
	Unknown2         [30]byte
	Unknown3         [4]float32
}

type ClientShotRotate struct {
	ClientMessage_
	Angle float32
}

type ClientShotMeterInput struct {
	ClientMessage_
	Sequence uint8
	Value    float32
}

type ClientShotPower struct {
	ClientMessage_
	Level uint8
}

type ClientShotClubChange struct {
	ClientMessage_
	Club uint8
}

type ClientShotItemUse struct {
	ClientMessage_
	ItemTypeID uint32
}

type ClientUserTypingIndicator struct {
	ClientMessage_
	Status int16 // 1 = started, -1 = stopped
}

type ClientShotCometRelief struct {
	ClientMessage_
	X, Y, Z float32
}

type ClientHoleInfo struct {
	ClientMessage_
	Num        uint8
	Unknown1   uint32
	Unknown2   uint32
	Par        uint8
	TeeX, TeeZ float32
	PinX, PinZ float32
}

type SyncEntry struct {
	Unknown1 uint8
	Unknown2 uint32
}

type ClientShotSync struct {
	ClientMessage_
	Data    gamemodel.ShotSyncData
	Unknown [16]byte
}

type ClientRoomSync struct {
	ClientMessage_
	Unknown1   uint8
	EntryCount uint8
	Entries    []SyncEntry
}

type PurchaseItem struct {
	Unknown       uint32
	ItemTypeID    uint32
	Unknown2      uint16
	Unknown3      uint16
	Quantity      uint32
	ItemCostPang  uint32
	ItemCostPoint uint32
}

// ClientBuyItem is sent by the client to buy an item from the shop.
type ClientBuyItem struct {
	ClientMessage_
	Unknown1 byte
	NumItems uint16 `struct:"sizeof=Items"`
	Items    []PurchaseItem
}

type UpdateCaddie struct {
	CaddieID uint32
}

type UpdateConsumables struct {
	ItemTypeID [10]uint32
}

type UpdateComet struct {
	ItemTypeID uint32
}

type UpdateDecoration struct {
	BackgroundTypeID uint32
	FrameTypeID      uint32
	StickerTypeID    uint32
	SlotTypeID       uint32
	CutInTypeID      uint32
	TitleTypeID      uint32
}

type UpdateCharacter struct {
	CharacterID uint32
}

type UpdateUnknown1 struct {
	Unknown uint32
}

type UpdateUnknown2 struct {
	CharacterID uint32
	Unknown     [4]uint32
}

// ClientEquipmentUpdate updates the user's equipment.
type ClientEquipmentUpdate struct {
	ClientMessage_
	Type        uint8
	CharParts   *pangya.PlayerCharacterData `struct-if:"Type == 0"`
	Caddie      *UpdateCaddie               `struct-if:"Type == 1"`
	Consumables *UpdateConsumables          `struct-if:"Type == 2"`
	Comet       *UpdateComet                `struct-if:"Type == 3"`
	Decoration  *UpdateDecoration           `struct-if:"Type == 4"`
	Character   *UpdateCharacter            `struct-if:"Type == 5"`
	Unknown1    *UpdateUnknown1             `struct-if:"Type == 8"`
	Unknown2    *UpdateUnknown2             `struct-if:"Type == 9"`
}

type ClientShotActiveUserAcknowledge struct {
	ClientMessage_
}

type ClientRoomKick struct {
	ClientMessage_
	ConnID uint32
}

type ClientRoomInfo struct {
	ClientMessage_
	RoomNumber uint16
}

type ClientShotArrow struct {
	ClientMessage_
	// TODO
}

// ClientRequestServerList is a message sent to request the current
// list of game servers.
type ClientRequestServerList struct {
	ClientMessage_
}

type ClientLoadProgress struct {
	ClientMessage_
	Progress uint8
}

// Client004F is sent when the client gags you from chatting due to typing too much or too many obscenities
type Client004F struct {
	ClientMessage_
}

// ClientRoomLoungeAction
type ClientRoomLoungeAction struct {
	ClientMessage_
	gamemodel.RoomAction
}

// ClientRequestMessengerList is a message sent to request the current
// list of message servers.
type ClientRequestMessengerList struct {
	ClientMessage_
}

// ClientAchievementStatusRequest requests Achievement Status for a user.
type ClientAchievementStatusRequest struct {
	ClientMessage_
	UserID uint32
}

// ClientGetPlayerData is a message sent by the client to request
// the client state.
type ClientGetPlayerData struct {
	ClientMessage_
	UserID  uint32
	Request byte
}

type ClientPauseGame struct {
	ClientMessage_
	Pause bool `struct:"byte"`
}

type ClientHoleEnd struct {
	ClientMessage_
	Stats pangya.PlayerStats
}

// ClientSetIdleStatus sets whether or not the client is idle in a room.
type ClientSetIdleStatus struct {
	ClientMessage_
	Idle bool `struct:"byte"`
}

// ClientException is a message sent when the client encounters an
// error.
type ClientException struct {
	ClientMessage_
	Empty   byte
	Message common.PString
}

type ClientFirstShotReady struct {
	ClientMessage_
}

// ClientLastPlayerLeaveGame is sent when the last player leaves a room.
// We can't rely on it for much; it just tells us the client thinks it shouldn't
// be punished for leaving the room since everyone else has already left.
// In the future it may need to be rejected in some cases.
type ClientLastPlayerLeaveGame struct {
	ClientMessage_
}

// ClientRequestPlayerHistory is an unknown message.
type ClientRequestPlayerHistory struct {
	ClientMessage_
}

// ClientTutorialClear is an unknown message.
type ClientTutorialClear struct {
	ClientMessage_
}

type ClientEnterMyRoom struct {
	ClientMessage_
	UserID     uint32
	RoomUserID uint32
}

type ClientRequestInventory struct {
	ClientMessage_
	UserID  uint32
	Unknown byte
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

type Client00C1 struct {
	ClientMessage_
	Unknown byte
}

type ClientLockerCombinationAttempt struct {
	ClientMessage_
	Combination common.PString
}

type ClientLockerInventoryRequest struct {
	ClientMessage_
}

type Client00FE struct {
	ClientMessage_
}

type ClientGuildListRequest struct {
	ClientMessage_
	Page uint32
}

type ClientScratchyMenuOpen struct {
	ClientMessage_
}

type ClientBlackPapelPlay struct {
	ClientMessage_
}

type ClientQuestStatusRequest struct {
	ClientMessage_
}

type ClientBigPapelPlay struct {
	ClientMessage_
}
