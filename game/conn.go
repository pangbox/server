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
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/bufbuild/connect-go"
	"github.com/go-restruct/restruct"
	"github.com/pangbox/server/common"
	"github.com/pangbox/server/gen/dbmodels"
	"github.com/pangbox/server/gen/proto/go/topologypb"
	"github.com/pangbox/server/pangya"
)

// Conn holds the state for a connection to the server.
type Conn struct {
	common.ServerConn[ClientMessage, ServerMessage]
	s *Server
}

// SendHello sends the initial handshake bytes to the client.
func (c *Conn) SendHello() error {
	data, err := restruct.Pack(binary.LittleEndian, &ConnectMessage{
		Unknown: [8]byte{0x00, 0x06, 0x00, 0x00, 0x3f, 0x00, 0x01, 0x01},
		Key:     c.Key,
	})
	if err != nil {
		return err
	}

	_, err = c.Socket.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// Handle runs the main connection loop.
func (c *Conn) Handle(ctx context.Context) error {
	log := c.Log
	c.Key = uint8(rand.Intn(16))

	err := c.SendHello()
	if err != nil {
		return fmt.Errorf("sending hello message: %w", err)
	}

	msg, err := c.ReadMessage()
	if err != nil {
		return fmt.Errorf("reading handshake: %w", err)
	}

	var session dbmodels.Session
	var player dbmodels.Player
	switch t := msg.(type) {
	case *ClientAuth:
		session, err = c.s.accountsService.GetSessionByKey(ctx, t.LoginKey.Value)
		if err != nil {
			// TODO: error handling
			return nil
		}
		player, err = c.s.accountsService.GetPlayer(ctx, session.PlayerID)
		if err != nil {
			// TODO: error handling
			return nil
		}
		log.Debugf("Client auth: %#v", msg)

	default:
		return fmt.Errorf("expected client auth, got %T", t)
	}

	playerCharacters, err := c.s.accountsService.GetCharacters(ctx, session.PlayerID)
	if err != nil {
		// TODO: handle error for client
		return fmt.Errorf("database error: %w", err)
	}

	playerGameData := ServerUserData{
		ClientVersion: common.ToPString("824.00"),
		ServerVersion: common.ToPString("Pangbox"),
		Game:          0xFFFF,
		UserInfo: pangya.UserInfo{
			Username:      player.Username,
			Nickname:      player.Nickname.String,
			PlayerID:      uint32(player.PlayerID),
			ConnnectionID: uint32(session.SessionID),
		},
		PlayerStats: pangya.PlayerStats{
			Pangs: uint64(player.Pang),
		},
		Items: pangya.PlayerEquipment{
			CaddieID:    0,
			CharacterID: playerCharacters[0].ID,
			ClubSetID:   0x1754,
			AztecIffID:  0x14000000,
			Items: pangya.PlayerEquippedItems{
				ItemIDs: [10]uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
		},
		EquippedCharacter: playerCharacters[0],
		EquippedClub: pangya.PlayerClubData{
			Item: pangya.PlayerItem{
				ID:    0x1754,
				IFFID: 0x10000000,
			},
			Stats: pangya.ClubStats{
				UpgradeStats: [5]uint16{8, 9, 8, 3, 3},
			},
		},
	}

	c.SendMessage(&playerGameData)

	c.SendMessage(&ServerCharData{
		Count1:     uint16(len(playerCharacters)),
		Count2:     uint16(len(playerCharacters)),
		Characters: playerCharacters,
	})

	c.SendMessage(&ServerAchievementProgress{
		Remaining: 0,
		Count:     0,
	})

	c.SendMessage(&ServerMessageConnect{})

	message := &ServerChannelList{}
	response, err := c.s.topologyClient.ListServers(ctx, connect.NewRequest(&topologypb.ListServersRequest{
		Type: topologypb.Server_TYPE_GAME_SERVER,
	}))
	if err != nil {
		// TODO: error handling to client
		return err
	}
	for _, server := range response.Msg.Server {
		entry := pangya.ServerEntry{
			ServerName: server.Name,
			ServerID:   server.Id,
			NumUsers:   server.NumUsers,
			MaxUsers:   server.MaxUsers,
			IPAddress:  server.Address,
			Port:       uint16(server.Port),
			Flags:      uint16(server.Flags),
		}
		if server.Id == c.s.serverID {
			// TODO: support multiple channels?
			entry.Channels = append(entry.Channels, pangya.ChannelEntry{
				ChannelName: c.s.channelName,
				MaxUsers:    200,    // TODO
				NumUsers:    0,      // TODO
				Unknown2:    0x0008, // TODO
			})
		}
		message.Servers = append(message.Servers, entry)
	}
	message.Count = uint8(len(response.Msg.Server))
	c.SendMessage(message)

	status := &ServerRoomStatus{}

	for {
		msg, err = c.ReadMessage()
		if err != nil {
			return fmt.Errorf("reading next message: %w", err)
		}

		switch t := msg.(type) {
		case *ClientException:
			log.WithField("exception", t.Message).Debug("Client exception")
			return fmt.Errorf("client reported exception: %v", t.Message)
		case *ClientMessageSend:
			event := &ServerGlobalEvent{Type: ChatMessageData}
			event.Data.Message = t.Message
			event.Data.Nickname = t.Nickname
			c.SendMessage(event)
			log.Debug(t.Message.Value)
		case *ClientRequestMessengerList:
			// TODO
			log.Debug("TODO: messenger list")
		case *ClientGetUserOnlineStatus:
			// TODO
			log.Debug("TODO: online status")
		case *ClientGetUserData:
			// TODO
			log.Debug("TODO: user data")
		case *ClientRoomLoungeAction:
			c.SendMessage(&ServerRoomLoungeAction{
				ConnID:       uint32(session.SessionID),
				LoungeAction: t.LoungeAction,
			})
		case *ClientRoomCreate:
			c.SendMessage(&ServerRoomJoin{
				RoomName:    t.RoomName.Value,
				RoomNumber:  1,
				EventNumber: 0,
			})
			status = &ServerRoomStatus{
				RoomType:        t.RoomType,
				Course:          t.Course,
				NumHoles:        t.NumHoles,
				HoleProgression: 1,
				NaturalWind:     0,
				MaxUsers:        t.MaxUsers,
				ShotTimerMS:     t.ShotTimerMS,
				GameTimerMS:     t.GameTimerMS,
				Flags:           0,
				Owner:           true,
				RoomName:        t.RoomName,
			}
			c.SendMessage(status)
			self := RoomListUser{
				ConnID:           uint32(session.SessionID),
				Nickname:         player.Nickname.String,
				Rank:             uint8(player.Rank),
				GuildName:        "",
				Slot:             1,
				CharTypeID:       playerGameData.EquippedCharacter.CharTypeID,
				Flag2:            520,
				GuildEmblemImage: "guildmark",
				UserID:           uint32(player.PlayerID),
				CharacterData:    playerGameData.EquippedCharacter,
			}
			other := RoomListUser{
				ConnID:           0xFEEE,
				Nickname:         "other",
				GuildName:        "",
				PortraitSlotID:   0x38C00083,
				Rank:             uint8(pangya.JuniorA),
				Slot:             2,
				CharTypeID:       0x04000007,
				Flag2:            0,
				GuildEmblemImage: "guildmark",
				UserID:           0x2000,
				CharacterData: pangya.PlayerCharacterData{
					CharTypeID: 0x04000007,
					ID:         0x50000,
					HairColor:  0,
				},
			}
			c.SendMessage(&ServerRoomCensus{
				Type:    byte(ListSet),
				Unknown: 0xFFFF,
				ListSet: &RoomCensusListSet{
					UserCount: 2,
					UserList:  []RoomListUser{self, other},
				},
			})
			c.SendMessage(&ServerPlayerReady{
				ConnID: 0xFEEE,
				State:  0,
			})
		case *ClientAssistModeToggle:
			c.SendMessage(&ServerAssistModeToggled{})
			// TODO: Should send user status update; need to look at packet dumps.
		case *ClientPlayerReady, *ClientPlayerStartGame:
			c.SendMessage(&Server0230{})
			c.SendMessage(&Server0231{})
			c.SendRaw([]byte{0x77, 0x00, 0x64, 0x00, 0x00, 0x00})
			c.SendMessage(&ServerGameInit{
				SubType: GameInitTypeFull,
				Full: &GameInitFull{
					NumPlayers: 2,
					Players: []GamePlayer{
						{
							Number: 1,
							Info: PlayerInfo{
								Username: player.Username,
								Nickname: player.Nickname.String,
								ConnID:   uint32(session.SessionID),
								UserID:   uint32(player.PlayerID),
							},
							Stats:     PlayerStats{},
							Character: playerGameData.EquippedCharacter,
							Caddie:    playerGameData.EquippedCaddie,
							ClubSet:   playerGameData.EquippedClub,
							Mascot:    playerGameData.EquippedMascot,
							StartTime: pangya.SystemTime{},
							NumCards:  0,
						},
						{
							Number: 2,
							Info: PlayerInfo{
								Username: "otheru",
								Nickname: "other",
								ConnID:   uint32(0xFEEE),
								UserID:   uint32(0x2000),
							},
							Stats:     PlayerStats{},
							Character: playerGameData.EquippedCharacter,
							Caddie:    playerGameData.EquippedCaddie,
							ClubSet:   playerGameData.EquippedClub,
							Mascot:    playerGameData.EquippedMascot,
							StartTime: pangya.SystemTime{},
							NumCards:  0,
						},
					},
				},
			})
			gameData := &ServerRoomGameData{
				Course:          status.Course,
				Unknown:         0x0,
				HoleProgression: status.HoleProgression,
				NumHoles:        status.NumHoles,
				Unknown2:        0x0,
				ShotTimerMS:     status.ShotTimerMS,
				GameTimerMS:     status.GameTimerMS,
				RandomSeed:      rand.Uint32(),
			}
			for i := byte(0); i < gameData.NumHoles; i++ {
				gameData.Holes = append(gameData.Holes, HoleInfo{
					HoleID: rand.Uint32(),
					Pin:    0x0,
					Course: status.Course,
					Num:    i,
				})
			}
			c.SendMessage(gameData)
			/*c.SendMessage(&ServerRoomGameData{
				Course:          11,
				Unknown:         0,
				HoleProgression: 3,
				NumHoles:        3,
				Unknown2:        0,
				ShotTimerMS:     300000,
				GameTimerMS:     0,
				Holes: []HoleInfo{
					{
						HoleID: 2159514729,
						Pin:    0,
						Course: 11,
						Num:    14,
					},
					{
						HoleID: 358258534,
						Pin:    1,
						Course: 11,
						Num:    6,
					},
					{
						HoleID: 3739427069,
						Pin:    2,
						Course: 11,
						Num:    3,
					},
				},
			})*/
			// (currently crashes...)
		case *ClientRequestInboxList:
			// TODO: need new sql message table
			msg := &ServerInboxList{
				PageNum:     t.PageNum,
				NumPages:    1,
				NumMessages: 1,
				Messages: []InboxMessage{
					{ID: 0x1, SenderNickname: "@Pangbox"},
				},
			}
			c.DebugMsg(msg)
			c.SendMessage(msg)
		case *ClientRequestInboxMessage:
			c.SendMessage(&ServerMailMessage{
				Message: MailMessage{
					ID:             0x1,
					SenderNickname: common.ToPString("@Pangbox"),
					DateTime:       common.ToPString("2023-06-03 01:21:00"),
					Message:        common.ToPString("Welcome to the first Pangbox server release! Not much works yet..."),
				},
			})
		case *ClientUnknownCounter:
			// Do nothing.
		case *Client001A:
			// Do nothing.
		case *ClientJoinChannel:
			c.SendMessage(&Server004E{Unknown: []byte{0x01}})
			c.SendMessage(&Server01F6{Unknown: []byte{0x00, 0x00, 0x00, 0x00}})
			c.SendMessage(&ServerLoginBonusStatus{Unknown: []byte{0x0, 0x0, 0x0, 0x0, 0x1, 0x4, 0x0, 0x0, 0x18, 0x3, 0x0, 0x0, 0x0, 0x27, 0x0, 0x0, 0x18, 0x3, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0}})
		case *ClientRequestDailyReward:
			c.SendMessage(&ServerMoneyUpdate{
				Type: uint16(MoneyUpdateRewardUnknown),
				RewardUnknown: &UpdateRewardUnknownData{
					Unknown: 1,
				},
			})
		case *Client009C:
			c.SendMessage(&Server010E{Unknown: make([]byte, 0x104)})
		case *ClientMultiplayerJoin:
			c.SendMessage(&ServerRoomList{
				Count:    0,
				Type:     ListSet,
				RoomList: []RoomListRoom{},
			})
			// TODO: lobby room new sql impl
			c.SendMessage(&ServerUserCensus{
				Count: 1,
				Type:  UserListSet,
				UserList: []CensusUser{
					{
						UserID:        uint32(player.PlayerID),
						ConnID:        uint32(session.SessionID),
						RoomNumber:    -1,
						Nickname:      player.Nickname.String,
						Rank:          byte(player.Rank),
						GuildEmblemID: "guildmark",     // TODO
						GlobalID:      player.Username, // TODO
					},
				},
			})
			c.SendMessage(&ServerMultiplayerJoined{})
		case *ClientMultiplayerLeave:
			c.SendMessage(&ServerMultiplayerLeft{})
		case *ClientEventLobbyJoin:
			c.SendMessage(&ServerEventLobbyJoined{})
		case *ClientEventLobbyLeave:
			c.SendMessage(&ServerEventLobbyLeft{})
		case *ClientRoomLeave:
			log.Println("Client leave room")
			c.SendMessage(&ServerRoomLeave{RoomNumber: t.RoomNumber})
		case *ClientRoomEdit:
			log.Printf("%#v\n", t)
			for _, change := range t.Changes {
				if change.RoomName != nil {
					status.RoomName = *change.RoomName
				}
				if change.RoomType != nil {
					status.RoomType = *change.RoomType
				}
				if change.Course != nil {
					status.Course = *change.Course
				}
				if change.NumHoles != nil {
					status.NumHoles = *change.NumHoles
				}
				if change.HoleProgression != nil {
					status.HoleProgression = *change.HoleProgression
				}
				if change.ShotTimerSeconds != nil {
					status.ShotTimerMS = uint32(*change.ShotTimerSeconds) * 1000
				}
				if change.MaxUsers != nil {
					status.MaxUsers = *change.MaxUsers
				}
				if change.GameTimerMinutes != nil {
					status.GameTimerMS = uint32(*change.GameTimerMinutes) * 60 * 1000
				}
				if change.NaturalWind != nil {
					status.NaturalWind = *change.NaturalWind
				}
			}
			c.SendMessage(status)
		case *Client0088:
			// Unknown tutorial-related message.
		case *ClientRoomUserEquipmentChange:
			// TODO
		case *ClientTutorialStart:
			// TODO
			c.SendMessage(&ServerRoomEquipmentData{
				Unknown: []byte{
					0x00, 0x00, 0x00, 0x01, 0x04, 0x04, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x04, 0xdd,
					0x77, 0x94, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x04, 0x14, 0x08, 0x00, 0x24, 0x14, 0x08, 0x00,
					0x44, 0x14, 0x08, 0x00, 0x64, 0x14, 0x08, 0x00, 0x84, 0x14, 0x08, 0x00, 0xa4, 0x14, 0x08, 0x00,
					0xc4, 0x14, 0x08, 0x00, 0xe4, 0x14, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
			})
		case *ClientTutorialClear:
			// After clearing first tutorial
			// TODO
			c.SendMessage(&ServerTutorialStatus{
				Unknown: [6]byte{0x00, 0x01, 0x03, 0x00, 0x00, 0x00},
			})
		case *ClientUserMacrosSet:
			// TODO: server-side macro storage
			log.Debugf("Set macros: %+v", t.MacroList)
		case *ClientEquipmentUpdate:
			// TODO
			log.Debug("TODO: 0020")
		case *Client00FE:
			// TODO
			log.Debug("TODO: 00FE")
		case *ClientShopJoin:
			// Enter shop, not sure what responses need to go here?
			log.Debug("TODO: 0140")
		default:
			return fmt.Errorf("unexpected message: %T", t)
		}
	}
}
