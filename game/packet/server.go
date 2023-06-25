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

type ServerConn = common.ServerConn[ClientMessage, ServerMessage]

var ServerMessageTable = common.NewMessageTable(map[uint16]ServerMessage{
	0x0040: &ServerEvent{},
	0x0044: &ServerPlayerData{},
	0x0046: &ServerUserCensus{},
	0x0047: &ServerRoomList{},
	0x0048: &ServerRoomCensus{},
	0x0049: &ServerRoomJoin{},
	0x004A: &ServerRoomStatus{},
	0x004B: &ServerRoomEquipmentData{},
	0x004C: &ServerRoomLeave{},
	0x004E: &Server004E{},
	0x0052: &ServerRoomGameData{},
	0x0053: &ServerRoomStartHole{},
	0x0055: &ServerRoomShotAnnounce{},
	0x0056: &ServerRoomShotRotateAnnounce{},
	0x0058: &ServerRoomShotPowerAnnounce{},
	0x0059: &ServerRoomClubChangeAnnounce{},
	0x005A: &ServerRoomItemUseAnnounce{},
	0x005B: &ServerRoomSetWind{},
	0x005D: &ServerRoomUserTypingAnnounce{},
	0x0060: &ServerRoomShotCometReliefAnnounce{},
	0x0063: &ServerRoomActiveUserAnnounce{},
	0x0064: &ServerRoomShotSync{},
	0x0065: &ServerRoomFinishHole{},
	0x0066: &ServerRoomFinishGame{},
	0x0068: &ServerPurchaseItemResponse{},
	0x006B: &ServerPlayerEquipmentUpdated{},
	0x0070: &ServerCharData{},
	0x0072: &ServerPlayerEquipment{},
	0x0073: &ServerPlayerInventory{},
	0x0076: &ServerGameInit{},
	0x0077: &Server0077{},
	0x0078: &ServerPlayerReady{},
	0x0086: &ServerRoomInfoResponse{},
	0x0090: &ServerPlayerFirstShotReady{},
	0x0092: &ServerOpponentQuit{},
	0x0095: &ServerMoneyUpdate{},
	0x0096: &ServerPointsBalance{},
	0x009E: &ServerRoomSetWeather{},
	0x009F: &ServerChannelList{},
	0x00A1: &ServerUserInfo{},
	0x00A3: &ServerPlayerLoadProgress{},
	0x00C4: &ServerRoomAction{},
	0x00C8: &ServerPangBalanceData{},
	0x00CC: &ServerRoomShotEnd{},
	0x00F1: &ServerMessageConnect{},
	0x00F5: &ServerMultiplayerJoined{},
	0x00F6: &ServerMultiplayerLeft{},
	0x010E: &ServerPlayerHistory{},
	0x011F: &ServerTutorialStatus{},
	0x012B: &ServerMyRoomEntered{},
	0x012D: &ServerMyRoomLayout{},
	0x0151: &Server0151{},
	0x0158: &ServerPlayerStats{},
	0x0168: &ServerPlayerInfo{},
	0x016A: &Server016A{},
	0x016C: &ServerLockerCombinationResponse{},
	0x0170: &ServerLockerInventoryResponse{},
	0x01F6: &Server01F6{},
	0x020E: &Server020E{},
	0x0210: &ServerInboxNotify{},
	0x0211: &ServerInboxList{},
	0x0212: &ServerMailMessage{},
	0x0216: &ServerUserStatusUpdate{},
	0x021D: &ServerAchievementProgress{},
	0x0230: &Server0230{},
	0x0231: &Server0231{},
	0x0248: &ServerLoginBonusStatus{},
	0x0250: &ServerEventLobbyJoined{},
	0x0251: &ServerEventLobbyLeft{},
	0x026A: &ServerAssistModeToggled{},
})

// ConnectMessage is the message sent upon connecting.
type ConnectMessage struct {
	Unknown [8]byte
	Key     byte
}

func (c *ConnectMessage) SetKey(key uint8) {
	c.Key = key
}

// EventType enumerates message data event types.
type EventType byte

const (
	ChatMessageEvent = 0
	GameEndEvent     = 16
)

// ChatMessage contains a global chat message
type ChatMessage struct {
	Nickname common.PString
	Message  common.PString
}

type GameEnd struct {
	Score   int32
	Pang    uint64
	Unknown uint8
}

// ServerEvent is a message that contains events.
type ServerEvent struct {
	ServerMessage_
	Type    byte
	Data    ChatMessage
	GameEnd *GameEnd `struct-if:"Type == 16"`
}

// ServerChannelList is a message that contains a list of all of the
// channels for a given server. Channels are isolated game zones within a region.
type ServerChannelList struct {
	ServerMessage_
	Count   byte `struct:"sizeof=Servers"`
	Servers []pangya.ServerEntry
}

// PlayerMainData contains the main player information, sent after logging in.
type PlayerMainData struct {
	ClientVersion common.PString
	ServerVersion common.PString
	Game          uint16
	PlayerData    pangya.PlayerData
	Unknown2      [321]byte
}

// ServerPlayerData contains important state information.
type ServerPlayerData struct {
	ServerMessage_
	SubType  byte
	MainData *PlayerMainData `struct-if:"SubType == 0"`
}

type CaddieUpdated struct {
	CaddieID uint32
}

type ConsumablesUpdated struct {
	ItemTypeID [10]uint32
}

type CometUpdated struct {
	ItemID     uint32
	ItemTypeID uint32
}

type DecorationUpdated struct {
	BackgroundTypeID uint32
	FrameTypeID      uint32
	StickerTypeID    uint32
	SlotTypeID       uint32
	CutInTypeID      uint32
	TitleTypeID      uint32
}

type CharacterUpdated struct {
	CharacterID uint32
}

type EquipmentUpdateType uint8

const (
	UpdatedCharParts   EquipmentUpdateType = 0
	UpdatedCaddie      EquipmentUpdateType = 1
	UpdatedConsumables EquipmentUpdateType = 2
	UpdatedComet       EquipmentUpdateType = 3
	UpdatedDecoration  EquipmentUpdateType = 4
	UpdatedCharacter   EquipmentUpdateType = 5
)

type ServerPlayerEquipmentUpdated struct {
	ServerMessage_
	Status      uint8
	Type        uint8
	CharParts   *pangya.PlayerCharacterData `struct-if:"Type == 0"`
	Caddie      *CaddieUpdated              `struct-if:"Type == 1"`
	Consumables *ConsumablesUpdated         `struct-if:"Type == 2"`
	Comet       *CometUpdated               `struct-if:"Type == 3"`
	Decoration  *DecorationUpdated          `struct-if:"Type == 4"`
	Character   *CharacterUpdated           `struct-if:"Type == 5"`
}

// ServerCharData contains the user's characters.
type ServerCharData struct {
	ServerMessage_
	Count1     uint16 `struct:"sizeof=Characters"`
	Count2     uint16 `struct:"sizeof=Characters"`
	Characters []pangya.PlayerCharacterData
}

type ServerPlayerEquipment struct {
	ServerMessage_
	Equipment pangya.PlayerEquipment
}

type InventoryItem struct {
	ItemID          uint32
	ItemTypeID      uint32
	Unknown         int32
	Quantity        uint32
	Unknown2        [7]byte
	Flags           byte
	RentalDateStart uint32
	Unknown3        uint32
	RentalDateEnd   uint32
	Unknown4        uint32
	Unknown5        [156]byte
}

type ServerPlayerInventory struct {
	ServerMessage_
	Remaining uint16
	Count     uint16 `struct:"sizeof=Inventory"`
	Inventory []InventoryItem
}

type GamePlayer struct {
	Number     uint16
	PlayerData pangya.PlayerData
	StartTime  pangya.SystemTime
	NumCards   uint8
}

type GameInitFull struct {
	NumPlayers byte `struct:"sizeof=Players"`
	Players    []GamePlayer
}

type GameInitMinimal struct {
	Unknown uint32
	Time    pangya.SystemTime
}

type GameInitType byte

const (
	GameInitTypeFull    = 0
	GameInitTypeMinimal = 4
)

type ServerGameInit struct {
	ServerMessage_
	SubType byte
	Full    *GameInitFull    `struct-if:"SubType == 0"`
	Minimal *GameInitMinimal `struct-if:"SubType == 4"`
}

type Server0077 struct {
	ServerMessage_
	Unknown uint32
}

// ServerUserInfo contains requested user information.
type ServerUserInfo struct {
	ServerMessage_
	ResponseCode uint8
	PlayerID     uint32
	UserInfo     pangya.PlayerInfo
}

type ServerPlayerLoadProgress struct {
	ServerMessage_
	ConnID   uint32
	Progress uint8
}

type ServerRoomAction struct {
	ServerMessage_
	ConnID uint32
	gamemodel.RoomAction
}

// ServerPangBalanceData is sent after a pang purchase succeeds.
type ServerPangBalanceData struct {
	ServerMessage_
	PangsRemaining uint64
	PangsSpent     uint64
}

type ServerRoomShotEnd struct {
	ServerMessage_
	ConnID uint32
}

// ServerPlayerID is a message that contains the PlayerID and some
// other unknown data.
type ServerPlayerID struct {
	ServerMessage_
	Empty    byte
	PlayerID uint32
	Unknown  [239]byte
}

// UserCensusType enumerates the types of census messages.
type UserCensusType byte

const (
	UserAdd        UserCensusType = 1
	UserChange     UserCensusType = 2
	UserRemove     UserCensusType = 3
	UserListSet    UserCensusType = 4
	UserListAppend UserCensusType = 5
)

const CensusMaxUsers = 36

// ServerUserCensus contains information about users currently online in
// multiplayer
type ServerUserCensus struct {
	ServerMessage_
	Type       UserCensusType
	Count      uint8 `struct:"sizeof=PlayerList"`
	PlayerList []gamemodel.LobbyPlayer
}

// ListType enumerates the types of room list messages.
type ListType byte

const (
	ListSet    ListType = 0
	ListAdd    ListType = 1
	ListRemove ListType = 2
	ListChange ListType = 3

	// Only valid for lounge mode rooms
	ListLounge ListType = 7
)

type RoomListRoom struct {
	Name             string `struct:"[64]byte"`
	Public           bool   `struct:"byte"`
	Open             bool   `struct:"uint16"`
	UserMax          uint8
	UserCount        uint8
	Key              [16]byte // Known thanks to SuperSS; XOR pad for shot sync data
	Unknown3         uint8
	Unknown4         uint8
	NumHoles         uint8
	HoleProgression  uint8
	Number           int16
	Unknown5         uint8
	Course           uint8
	ShotTimerMS      uint32
	GameTimerMS      uint32
	Flags            uint32
	Unknown6         [68]byte
	Unknown7         uint32
	Unknown8         uint32
	OwnerID          uint32
	Class            byte
	ArtifactID       uint32
	Unknown9         uint32
	EventNum         uint32
	EventNumTop      uint32
	EventShotTimerMS uint32
	Unknown10        uint32
}

// ServerRoomList contains information about rooms currently open in
// multiplayer.
type ServerRoomList struct {
	ServerMessage_
	Count    uint8 `struct:"sizeof=RoomList"`
	Type     ListType
	Unknown  uint16
	RoomList []RoomListRoom
}

// ServerRoomCensus reports on the users in a game room.
type ServerRoomCensus struct {
	ServerMessage_
	Type       byte
	Unknown    int16
	ListSet    *RoomCensusListSet    `struct-if:"Type == 0"`
	ListAdd    *RoomCensusListAdd    `struct-if:"Type == 1"`
	ListRemove *RoomCensusListRemove `struct-if:"Type == 2"`
	ListChange *RoomCensusListChange `struct-if:"Type == 3"`
}

type RoomCensusListSet struct {
	PlayerCount uint8 `struct:"sizeof=PlayerList"`
	PlayerList  []gamemodel.RoomPlayerEntry
	Unknown     byte
}

type RoomCensusListAdd struct {
	User gamemodel.RoomPlayerEntry
}

type RoomCensusListRemove struct {
	ConnID uint32
}

type RoomCensusListChange struct {
	ConnID uint32
	User   gamemodel.RoomPlayerEntry
}

// ServerRoomStatus is sent when a room's settings or status changes.
type ServerRoomStatus struct {
	ServerMessage_
	Unknown         uint16
	RoomType        byte
	Course          byte
	NumHoles        byte
	HoleProgression byte
	NaturalWind     uint32
	MaxUsers        byte
	Unknown2        uint16
	ShotTimerMS     uint32
	GameTimerMS     uint32
	Flags           uint32
	Owner           bool `struct:"byte"`
	RoomName        common.PString
}

type ServerRoomEquipmentData struct {
	ServerMessage_
	Unknown []byte
}

type ServerRoomLeave struct {
	ServerMessage_
	RoomNumber int16
}

type Server004E struct {
	ServerMessage_
	Unknown []byte
}

type HoleInfo struct {
	HoleID uint32
	Pin    uint8
	Course uint8
	Num    uint8
}

type ServerRoomGameData struct {
	ServerMessage_
	Course          byte
	Unknown         byte
	HoleProgression byte
	NumHoles        uint8
	Unknown2        uint32
	ShotTimerMS     uint32
	GameTimerMS     uint32
	Holes           [18]HoleInfo // `struct:"sizefrom=NumHoles"`
	RandomSeed      uint32
	Unknown3        [18]byte
}

type ServerRoomStartHole struct {
	ServerMessage_
	ConnID uint32
}

type ServerRoomShotAnnounce struct {
	ServerMessage_
	ConnID           uint32
	ShotStrength     float32
	ShotAccuracy     float32
	ShotEnglishCurve float32
	ShotEnglishSpin  float32
	Unknown2         [30]byte
	Unknown3         [4]float32
}

type ServerRoomShotRotateAnnounce struct {
	ServerMessage_
	ConnID uint32
	Angle  float32
}

type ServerRoomShotPowerAnnounce struct {
	ServerMessage_
	ConnID uint32
	Level  uint8
}

type ServerRoomClubChangeAnnounce struct {
	ServerMessage_
	ConnID uint32
	Club   uint8
}

type ServerRoomItemUseAnnounce struct {
	ServerMessage_
	ItemTypeID uint32
	Unknown    uint32
	ConnID     uint32
}

type ServerRoomSetWind struct {
	ServerMessage_
	Wind    uint8
	Unknown uint8
	Heading uint16
	Reset   bool `struct:"bool"`
}

type ServerRoomUserTypingAnnounce struct {
	ServerMessage_
	ConnID uint32
	Status int16
}

type ServerRoomShotCometReliefAnnounce struct {
	ServerMessage_
	ConnID  uint32
	X, Y, Z float32
}

type ServerRoomActiveUserAnnounce struct {
	ServerMessage_
	ConnID uint32
}

type ServerRoomShotSync struct {
	ServerMessage_
	Data gamemodel.ShotSyncData
}

type ServerRoomFinishHole struct {
	ServerMessage_
}

type PlayerGameResult struct {
	ConnID    uint32
	Place     uint8
	Score     int8
	Unknown   uint8
	Unknown2  uint16
	Pang      uint64
	BonusPang uint64
	Unknown3  uint64
}

type ServerRoomFinishGame struct {
	ServerMessage_
	NumPlayers uint8
	Standings  []PlayerGameResult `struct:"sizefrom=NumPlayers"`
}

type ServerPurchaseItemResponse struct {
	ServerMessage_
	Status uint32
	Pang   uint64
	Points uint64
}

type Server016A struct {
	ServerMessage_
	Unknown  byte
	Unknown2 uint32
}

type ServerLockerCombinationResponse struct {
	ServerMessage_
	Status uint32
}

type ServerLockerInventoryResponse struct {
	ServerMessage_
	Unknown uint32
	Status  uint32
}

// ServerRoomJoin is sent when a room is joined.
type ServerRoomJoin struct {
	ServerMessage_
	Status      uint16
	RoomName    string `struct:"[64]byte"`
	Unknown2    [25]byte
	RoomNumber  int16
	Unknown3    [111]byte
	EventNumber uint32
	Unknown4    uint32
}

type ServerPlayerReady struct {
	ServerMessage_
	ConnID uint32
	State  byte
}

type ServerRoomInfoResponse struct {
	ServerMessage_
	RoomInfo gamemodel.RoomInfo
}

type ServerPlayerFirstShotReady struct {
	ServerMessage_
}

type ServerOpponentQuit struct {
	ServerMessage_
}

type MoneyUpdateType uint16

const (
	MoneyUpdateRewardUnknown MoneyUpdateType = 2
	MoneyUpdatePangBalance   MoneyUpdateType = 273
)

type UpdateRewardUnknownData struct {
	Unknown uint16
}

type UpdatePangBalanceData struct {
	Status     uint32
	PangAmount uint64
}

type ServerMoneyUpdate struct {
	ServerMessage_
	Type uint16

	RewardUnknown *UpdateRewardUnknownData `struct-if:"Type == 2"`
	PangBalance   *UpdatePangBalanceData   `struct-if:"Type == 273"`
}

type ServerPointsBalance struct {
	ServerMessage_
	Points uint64
}

type ServerRoomSetWeather struct {
	ServerMessage_
	Weather uint16
	Unknown uint8
}

// ServerMessageConnect seems to make the client connect to the message server.
// TODO: need to do more reverse engineering effort
type ServerMessageConnect struct {
	ServerMessage_
	Unknown byte
}

type ServerMultiplayerJoined struct {
	ServerMessage_
}

type ServerMultiplayerLeft struct {
	ServerMessage_
}

type RecentPlayer struct {
	Unknown  uint32
	Nickname string `struct:"[22]byte"`
	Username string `struct:"[22]byte"`
	PlayerID uint32
}

type ServerPlayerHistory struct {
	ServerMessage_
	RecentPlayers [5]RecentPlayer
}

type ServerTutorialStatus struct {
	ServerMessage_
	Unknown [6]byte
}

type ServerMyRoomEntered struct {
	ServerMessage_
	Unknown  uint32
	UserID   uint32
	Unknown2 uint32
	Unknown3 [99]byte
}

type FurnitureItem struct {
	Unknown  uint32
	ItemID   uint32
	Unknown2 [19]byte
}

type ServerMyRoomLayout struct {
	ServerMessage_
	Unknown        uint32
	FurnitureCount uint16
	Furniture      []FurnitureItem
}

type Server0151 struct {
	ServerMessage_
	Unknown []byte
}

type ServerPlayerStats struct {
	ServerMessage_

	SessionID uint32
	Unknown   byte
	Stats     pangya.PlayerStats
}

type ServerPlayerInfo struct {
	ServerMessage_
	Player gamemodel.RoomPlayerEntry
}

type Server01F6 struct {
	ServerMessage_
	Unknown []byte
}

type Server020E struct {
	ServerMessage_
	Unknown [8]byte
}

// ServerInboxNotify is unimplemented.
type ServerInboxNotify struct {
	ServerMessage_
	Unknown []byte
}

type MessageAttachment struct {
	ID           uint32
	ItemID       uint32
	Unknown      byte
	ItemQuantity uint32
	Unknown2     uint32
	Unknown3     uint64
	Unknown5     uint64
	Unknown6     uint32
	Unknown7     uint32
	Unknown8     [12]byte
	Unknown10    uint16
}

type InboxMessage struct {
	ID              uint32
	SenderNickname  string `struct:"[30]byte"`
	Message         string `struct:"[80]byte"`
	Unknown         [18]byte
	Unknown2        uint32
	Unknown3        byte
	AttachmentCount uint32 `struct:"sizeof=Attachments"`
	Attachments     []MessageAttachment
}

type ServerInboxList struct {
	ServerMessage_
	Status      uint32
	PageNum     uint32
	NumPages    uint32
	NumMessages uint32 `struct:"sizeof=Messages"`
	Messages    []InboxMessage
}

type MailMessage struct {
	ID              uint32
	SenderNickname  common.PString
	DateTime        common.PString
	Message         common.PString
	Unknown         byte
	AttachmentCount uint32 `struct:"sizeof=Attachments"`
	Attachments     []MessageAttachment
}

type ServerMailMessage struct {
	ServerMessage_
	Status  uint32
	Message MailMessage
}

type UserStatusChangeValue struct {
	StatusID          uint32
	StatusSlot        uint32
	Unknown           uint32
	StatusAmountOld   uint32
	StatusAmountNew   uint32
	StatusAmountDelta int32
	Unknown2          [25]byte
}

type UserStatusChangeMastery struct {
	CharacterID           uint32
	StatusSlot            uint32
	Unknown               [16]byte
	MasteryPowerUpCount   uint16
	MasteryControlUpCount uint16
	MasteryImpactUpCount  uint16
	MasterySpinUpCount    uint16
	MasteryCurveUpCount   uint16
	Unknown2              [16]byte
}

type UserStatusChange204 struct {
	Unknown [72]byte
}

type UserStatusChange struct {
	StatusChangeType byte
	Value            *UserStatusChangeValue   `struct-if:"StatusChangeType == 2"`
	Mastery          *UserStatusChangeMastery `struct-if:"StatusChangeType == 201"`
	Unknown204       *UserStatusChange204     `struct-if:"StatusChangeType == 204"`
}

type ServerUserStatusUpdate struct {
	ServerMessage_
	DateTimeUnix uint32
	Count        uint32
	Changes      []UserStatusChange
}

type AchievementProgress struct {
	Unknown    byte
	StatusID   uint32
	StatusSlot uint32
	Value      uint32
}

type ServerAchievementProgress struct {
	ServerMessage_
	Status       uint32
	Remaining    uint32
	Count        uint32 `struct:"sizeof=Achievements"`
	Achievements []AchievementProgress
}

type Server0230 struct {
	ServerMessage_
}

type Server0231 struct {
	ServerMessage_
}

type ServerLoginBonusStatus struct {
	ServerMessage_
	Unknown []byte
}

type ServerEventLobbyJoined struct {
	ServerMessage_
}

type ServerEventLobbyLeft struct {
	ServerMessage_
}

type ServerAssistModeToggled struct {
	ServerMessage_
	Unknown uint32
}
