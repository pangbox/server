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

var ServerMessageTable = common.NewMessageTable(map[uint16]ServerMessage{
	0x0040: &ServerGlobalEvent{},
	0x0044: &ServerUserData{},
	0x0046: &ServerUserCensus{},
	0x0047: &ServerRoomList{},
	0x0048: &ServerRoomCensus{},
	0x0049: &ServerRoomJoin{},
	0x004A: &ServerRoomStatus{},
	0x004B: &ServerRoomEquipmentData{},
	0x004C: &ServerRoomLeave{},
	0x004E: &Server004E{},
	0x0052: &ServerRoomGameData{},
	0x0070: &ServerCharData{},
	0x0076: &ServerGameInit{},
	0x0078: &ServerPlayerReady{},
	0x0095: &ServerMoneyUpdate{},
	0x009F: &ServerChannelList{},
	0x00A1: &ServerUserInfo{},
	0x00C4: &ServerRoomLoungeAction{},
	0x00C8: &ServerPangPurchaseData{},
	0x00F1: &ServerMessageConnect{},
	0x00F5: &ServerMultiplayerJoined{},
	0x00F6: &ServerMultiplayerLeft{},
	0x010E: &Server010E{},
	0x011F: &ServerTutorialStatus{},
	0x01F6: &Server01F6{},
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

// MessageDataType enumerates message data event types.
type MessageDataType byte

const (
	ChatMessageData = 0
)

// GlobalChatMessage contains a global chat message
type GlobalChatMessage struct {
	Nickname common.PString
	Message  common.PString
}

// ServerGlobalEvent is a message that contains global chat events.
type ServerGlobalEvent struct {
	ServerMessage_
	Type MessageDataType
	Data GlobalChatMessage
}

// ServerChannelList is a message that contains a list of all of the
// channels for a given server. Channels are isolated game zones within a region.
type ServerChannelList struct {
	ServerMessage_
	Count   byte `struct:"sizeof=Servers"`
	Servers []pangya.ServerEntry
}

// ServerUserData contains important state information.
type ServerUserData struct {
	ServerMessage_
	Empty             byte
	ClientVersion     common.PString
	ServerVersion     common.PString
	Game              uint16
	UserInfo          pangya.UserInfo
	PlayerStats       pangya.PlayerStats
	Unknown           [78]byte
	Items             pangya.PlayerEquipment
	JunkData          [252 * 43]byte
	EquippedCharacter pangya.PlayerCharacterData
	EquippedCaddie    pangya.PlayerCaddieData
	EquippedClub      pangya.PlayerClubData
	EquippedMascot    pangya.PlayerMascotData
	Unknown2          [321]byte
}

// ServerCharData contains the user's characters.
type ServerCharData struct {
	ServerMessage_
	Count1     uint16 `struct:"sizeof=Characters"`
	Count2     uint16 `struct:"sizeof=Characters"`
	Characters []pangya.PlayerCharacterData
}

type PlayerInfo struct {
	Username         string `struct:"[22]byte"`
	Nickname         string `struct:"[22]byte"`
	GuildName        string `struct:"[17]byte"`
	GuildEmblemImage string `struct:"[12]byte"`
	Unknown          [71]byte
	Flag             byte
	Unknown2         uint16
	Unknown3         uint16
	Unknown4         uint16
	Unknown5         uint16
	Unknown6         [16]byte
	GlobalID         string `struct:"[128]byte"`
}

type PlayerGameInfo struct {
	Stroke     uint32
	Putt       uint32
	Time       uint32
	StrokeTime uint32
	Unknown    float32
	Unknown2   uint32
	Unknown3   uint32
	Unknown4   uint32
	Unknown5   uint32
	Unknown6   uint32
	Unknown7   uint32
	Unknown8   uint16
	Unknown9   uint32
	Unknown10  uint32
	Unknown11  uint32
	Unknown12  uint32
	Unknown13  float32
	Unknown14  float32
	Unknown15  uint32
	Level      byte
	Pang       uint64
	Unknown16  uint32
	Unknown17  [6]byte
	Unknown18  [5]uint64
	Unknown19  uint64
	Unknown20  uint32
	Unknown21  uint32
	Unknown22  uint32
	Unknown23  uint32
	Unknown24  uint32
	Unknown25  uint32
	Unknown26  uint32
	Unknown27  uint32
	Unknown28  uint32
	Unknown29  uint32
	Unknown30  uint32
	Unknown31  uint32
	Unknown32  uint64
	Unknown33  uint32
	Unknown34  uint32
	Unknown35  uint32
	Unknown36  uint32
	Unknown37  uint32
	Unknown38  uint32
	Unknown39  uint16
	Unknown40  uint32
	Unknown41  uint32
	Unknown42  uint32
	Unknown43  uint32
	Unknown44  uint32
	Unknown45  uint32
	Unknown46  uint32
	Unknown47  uint32
	Unknown48  uint32
	Unknown49  uint16
}

type GamePlayer struct {
	Number    uint16
	Info      PlayerInfo
	Game      PlayerGameInfo
	Unknown   [11430]byte
	Character pangya.PlayerCharacterData
	Caddie    pangya.PlayerCaddieData
	ClubSet   pangya.PlayerClubData
	Mascot    pangya.PlayerMascotData
	StartTime pangya.SystemTime
	NumCards  uint8
}

type ServerGameInit struct {
	ServerMessage_
	Unknown    byte
	NumPlayers byte `struct:"sizeof=Players"`
	Players    []GamePlayer
}

// ServerUserInfo contains requested user information.
type ServerUserInfo struct {
	ServerMessage_
	ResponseCode uint8
	PlayerID     uint32
	UserInfo     pangya.UserInfo
}

type ServerRoomLoungeAction struct {
	ServerMessage_
	ConnID uint32
	LoungeAction
}

// ServerPangPurchaseData is sent after a pang purchase succeeds.
type ServerPangPurchaseData struct {
	ServerMessage_
	PangsRemaining uint64
	PangsSpent     uint64
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
	UserRemove     UserCensusType = 3
	UserListSet    UserCensusType = 4
	UserListAppend UserCensusType = 5
)

const CensusMaxUsers = 36

type CensusUser struct {
	UserID        uint32
	ConnID        uint32
	RoomNumber    int16
	Nickname      string `struct:"[22]byte"`
	Rank          byte
	Unknown       uint32
	Badge         uint32
	Unknown2      uint32
	Unknown3      uint32
	Unknown4      byte
	GuildEmblemID string `struct:"[19]byte"`
	GlobalID      string `struct:"[128]byte"`
}

// ServerUserCensus contains information about users currently online in
// multiplayer
type ServerUserCensus struct {
	ServerMessage_
	Type     UserCensusType
	Count    uint8 `struct:"sizeof=UserList"`
	UserList []CensusUser
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
	Unknown          uint16
	UserMax          uint8
	UserCount        uint8
	Unknown2         [18]byte
	NumHoles         uint8
	Number           uint16
	HoleProgression  uint8
	Course           uint8
	ShotTimerMS      uint32
	GameTimerMS      uint32
	Flags            uint32
	Unknown3         [76]byte
	Unknown4         uint32
	Unknown5         uint32
	OwnerID          uint32
	Class            byte
	ArtifactID       uint32
	Unknown6         uint32
	EventNum         uint32
	EventNumTop      uint32
	EventShotTimerMS uint32
	Unknown7         uint32
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

type RoomListUser struct {
	ConnID            uint32
	Nickname          string `struct:"[22]byte"`
	GuildName         string `struct:"[20]byte"`
	Slot              uint8
	Flag              uint32
	TitleID           uint32
	CharTypeID        uint32
	PortraitBGID      uint32
	PortraitFrameID   uint32
	PortraitStickerID uint32
	PortraitSlotID    uint32
	SkinUnknown1      uint32
	SkinUnknown2      uint32
	Flag2             uint16
	Rank              uint8
	Unknown           uint8
	Unknown2          uint16
	GuildID           uint32
	GuildEmblemImage  string `struct:"[12]byte"`
	GuildEmblemID     uint8
	UserID            uint32
	LoungeState       uint32
	Unknown3          uint16
	Unknown4          uint32
	X                 float32
	Y                 float32
	Z                 float32
	Angle             float32
	ShopUnknown       uint32
	ShopName          string `struct:"[64]byte"`
	MascotTypeID      uint32
	GlobalID          string `struct:"[22]byte"`
	Unknown5          [106]byte
	Guest             bool `struct:"byte"`
	AverageScore      float32
	Unknown6          [3]byte
	UnknownMisalign   byte // TODO: something either before or after here is misaligned
	CharacterData     pangya.PlayerCharacterData
}

type RoomCensusListSet struct {
	UserCount uint8 `struct:"sizeof=UserList"`
	UserList  []RoomListUser
}

type RoomCensusListAdd struct {
	User RoomListUser
}

type RoomCensusListRemove struct {
	ConnID uint32
}

type RoomCensusListChange struct {
	ConnID uint32
	User   RoomListUser
}

// ServerRoomCensus reports on the users in a game room.
type ServerRoomCensus struct {
	ServerMessage_
	Type       byte
	Unknown    uint16
	ListSet    *RoomCensusListSet    `struct-if:"Type == 0"`
	ListAdd    *RoomCensusListAdd    `struct-if:"Type == 1"`
	ListRemove *RoomCensusListRemove `struct-if:"Type == 2"`
	ListChange *RoomCensusListChange `struct-if:"Type == 3"`
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
	RoomNumber uint16
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
	Holes           []HoleInfo `struct:"sizefrom=NumHoles"`
	RandomSeed      uint32
}

// ServerRoomJoin is sent when a room is joined.
type ServerRoomJoin struct {
	ServerMessage_
	Status      byte
	Unknown     byte
	RoomName    string `struct:"[64]byte"`
	Unknown2    [25]byte
	RoomNumber  uint16
	Unknown3    [111]byte
	EventNumber uint32
	Unknown4    [12]byte
}

type ServerPlayerReady struct {
	ServerMessage_
	ConnID uint32
	State  byte
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
	Unknown    uint32
	PangAmount uint32
	Unknown2   uint32
}

type ServerMoneyUpdate struct {
	ServerMessage_
	Type uint16

	RewardUnknown *UpdateRewardUnknownData `struct-if:"Type == 2"`
	PangBalance   *UpdatePangBalanceData   `struct-if:"Type == 273"`
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

type Server010E struct {
	ServerMessage_
	Unknown []byte
}

type ServerTutorialStatus struct {
	ServerMessage_
	Unknown [6]byte
}

type Server01F6 struct {
	ServerMessage_
	Unknown []byte
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
