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

package gameserver

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/pangbox/server/common"
	"github.com/pangbox/server/database/accounts"
	gamemodel "github.com/pangbox/server/game/model"
	gamepacket "github.com/pangbox/server/game/packet"
	"github.com/pangbox/server/game/room"
	"github.com/pangbox/server/gen/dbmodels"
	"github.com/pangbox/server/pangya"
)

// Conn holds the state for a connection to the server.
type Conn struct {
	*gamepacket.ServerConn
	s *Server

	connID     uint32
	session    dbmodels.Session
	player     dbmodels.GetPlayerRow
	characters []pangya.PlayerCharacterData

	currentCharacter *pangya.PlayerCharacterData

	currentLobby *room.Lobby
	currentRoom  *room.Room
}

func (c *Conn) getLobbyPlayer() gamemodel.LobbyPlayer {
	return gamemodel.LobbyPlayer{
		PlayerID:         uint32(c.player.PlayerID),
		ConnID:           c.connID,
		RoomNumber:       c.currentRoom.Number(),
		Nickname:         c.player.Nickname.String,
		Rank:             byte(c.player.Rank),
		GuildEmblemImage: "guildmark",       // TODO
		GlobalID:         c.player.Username, // TODO
	}
}

func (c *Conn) getRoomPlayer() *gamemodel.RoomPlayerEntry {
	return &gamemodel.RoomPlayerEntry{
		ConnID:           c.connID,
		Nickname:         c.player.Nickname.String,
		Rank:             uint8(c.player.Rank),
		GuildName:        "",
		CharTypeID:       c.currentCharacter.CharTypeID,
		StatusFlags:      0,
		GuildEmblemImage: "guildmark", // TODO
		PlayerID:         uint32(c.player.PlayerID),
		BackgroundTypeID: uint32(c.player.BackgroundTypeID.Int64),
		FrameTypeID:      uint32(c.player.FrameTypeID.Int64),
		StickerTypeID:    uint32(c.player.StickerTypeID.Int64),
		SlotTypeID:       uint32(c.player.SlotTypeID.Int64),
		CutInTypeID:      uint32(c.player.CutInTypeID.Int64),
		TitleTypeID:      uint32(c.player.TitleTypeID.Int64),
		CharacterData:    c.getPlayerEquippedCharacter(),
	}
}

func (c *Conn) leaveRoom(ctx context.Context) error {
	if c.currentRoom != nil {
		promise, err := c.currentRoom.Send(ctx, room.RoomPlayerLeave{
			ConnID: c.connID,
		})
		if err != nil {
			return err
		}
		_, err = promise.Wait(ctx)
		if err != nil {
			return err
		}
		c.currentRoom = nil
	}
	return nil
}

func (c *Conn) leaveMultiplayerLobby(ctx context.Context) error {
	if c.currentLobby != nil {
		promise, err := c.currentLobby.Send(ctx, room.LobbyPlayerLeave{
			ConnID: c.connID,
		})
		if err != nil {
			return err
		}
		_, err = promise.Wait(ctx)
		if err != nil {
			return err
		}
		c.currentLobby = nil
	}
	return nil
}

// Handle runs the main connection loop.
func (c *Conn) Handle(ctx context.Context) error {
	log := c.s.logger

	// Handle the authentication phase.
	if err := c.handleAuth(ctx); err != nil {
		return err
	}

	defer func() {
		c.leaveRoom(ctx)
		c.leaveMultiplayerLobby(ctx)
	}()

	for {
		msg, err := c.ReadMessage()
		if err != nil {
			return fmt.Errorf("reading next message: %w", err)
		}

		switch t := msg.(type) {
		case *gamepacket.ClientException:
			log.WithField("exception", t.Message).Debug("Client exception")
		case *gamepacket.ClientMessageSend:
			chatMsg := room.ChatMessage{
				Nickname: t.Nickname.Value,
				Message:  t.Message.Value,
			}
			if c.currentRoom != nil {
				c.currentRoom.Send(ctx, chatMsg)
			} else if c.currentLobby != nil {
				c.currentLobby.Send(ctx, chatMsg)
			}
		case *gamepacket.ClientRequestMessengerList:
			// TODO
			log.Debug("TODO: messenger list")
		case *gamepacket.ClientGetUserOnlineStatus:
			// TODO
			log.Debug("TODO: online status")
		case *gamepacket.ClientGetUserData:
			// TODO
			log.Debug("TODO: user data")
		case *gamepacket.ClientRoomLoungeAction:
			if c.currentRoom == nil {
				break
			}
			c.currentRoom.Send(ctx, room.RoomAction{
				ConnID: c.connID,
				Action: t.RoomAction,
			})
		case *gamepacket.ClientRequestServerList:
			c.sendServerList(ctx)
		case *gamepacket.ClientRoomCreate:
			if c.currentLobby == nil {
				break
			}
			newRoom, err := c.currentLobby.NewRoom(context.Background(), gamemodel.RoomState{
				ShotTimerMS:     t.ShotTimerMS,
				GameTimerMS:     t.GameTimerMS,
				MaxUsers:        t.MaxUsers,
				RoomType:        t.RoomType,
				NumHoles:        t.NumHoles,
				Course:          t.Course,
				RoomName:        t.RoomName.Value,
				Password:        t.Password.Value,
				HoleProgression: t.HoleProgression,
				// TODO: natural wind, more?
			})
			if err != nil {
				// TODO: handle error
				return err
			}
			c.currentRoom = newRoom
			c.currentRoom.Send(ctx, room.RoomPlayerJoin{
				Entry:      c.getRoomPlayer(),
				Conn:       c.ServerConn,
				PlayerData: c.getPlayerData(),
			})
		case *gamepacket.ClientAssistModeToggle:
			c.SendMessage(ctx, &gamepacket.ServerAssistModeToggled{})
			// TODO: Should send user status update; need to look at packet dumps.
		case *gamepacket.ClientSetIdleStatus:
			c.currentRoom.Send(ctx, room.RoomPlayerIdle{
				ConnID: c.connID,
				Idle:   t.Idle,
			})
		case *gamepacket.ClientPlayerReady:
			ready := false
			if t.State == 0 {
				ready = true
			}
			c.currentRoom.Send(ctx, room.RoomPlayerReady{
				ConnID: c.connID,
				Ready:  ready,
			})
		case *gamepacket.ClientPlayerStartGame:
			c.currentRoom.Send(ctx, room.RoomStartGame{
				ConnID: c.connID,
			})
		case *gamepacket.ClientLoadProgress:
			// TODO: publish to game room
			if c.currentRoom == nil {
				break
			}
			c.currentRoom.Send(ctx, room.RoomLoadingProgress{
				ConnID:   c.connID,
				Progress: t.Progress,
			})
		case *gamepacket.ClientReadyStartHole:
			c.currentRoom.Send(ctx, room.RoomGameReady{
				ConnID: c.connID,
			})
		case *gamepacket.ClientShotCommit:
			c.currentRoom.Send(ctx, room.RoomGameShotCommit{
				ConnID:           c.connID,
				ShotStrength:     t.ShotStrength,
				ShotAccuracy:     t.ShotAccuracy,
				ShotEnglishCurve: t.ShotEnglishCurve,
				ShotEnglishSpin:  t.ShotEnglishSpin,
				Unknown2:         t.Unknown2,
				Unknown3:         t.Unknown3,
			})
		case *gamepacket.ClientShotSync:
			c.currentRoom.Send(ctx, room.RoomGameShotSync{
				ConnID: c.connID,
				Data:   t.Data,
			})
		case *gamepacket.ClientShotRotate:
			c.currentRoom.Send(ctx, room.RoomGameShotRotate{
				ConnID: c.connID,
				Angle:  t.Angle,
			})
		case *gamepacket.ClientShotMeterInput:
			// TODO?
		case *gamepacket.ClientShotArrow:
			// TODO?
		case *gamepacket.ClientShotPower:
			c.currentRoom.Send(ctx, room.RoomGameShotPower{
				ConnID: c.connID,
				Level:  t.Level,
			})
		case *gamepacket.ClientShotClubChange:
			c.currentRoom.Send(ctx, room.RoomGameShotClubChange{
				ConnID: c.connID,
				Club:   t.Club,
			})
		case *gamepacket.ClientShotItemUse:
			if c.currentRoom == nil {
				break
			}
			err := c.s.accountsService.UseItem(ctx, c.session.PlayerID, int64(t.ItemTypeID), &c.player)
			if err != nil {
				log.WithError(err).Error("using item failed")
			} else {
				c.currentRoom.Send(ctx, room.RoomGameShotItemUse{
					ConnID:     c.connID,
					ItemTypeID: t.ItemTypeID,
				})
			}
		case *gamepacket.ClientUserTypingIndicator:
			c.currentRoom.Send(ctx, room.RoomGameTypingIndicator{
				ConnID: c.connID,
				Status: t.Status,
			})
		case *gamepacket.ClientShotCometRelief:
			c.currentRoom.Send(ctx, room.RoomGameShotCometRelief{
				ConnID: c.connID,
				X:      t.X,
				Y:      t.Y,
				Z:      t.Z,
			})
		case *gamepacket.ClientRoomSync:
			c.currentRoom.Send(ctx, room.RoomGameTurnEnd{
				ConnID: c.connID,
			})
		case *gamepacket.ClientHoleEnd:
			c.currentRoom.Send(ctx, room.RoomGameHoleEnd{
				ConnID: c.connID,
				Stats:  t.Stats,
			})
		case *gamepacket.ClientGameEnd:
			// TODO
		case *gamepacket.ClientPauseGame:
			// TODO
		case *gamepacket.ClientShotActiveUserAcknowledge:
			c.currentRoom.Send(ctx, room.RoomGameTurn{
				ConnID: c.connID,
			})
		case *gamepacket.ClientFirstShotReady:
			c.SendMessage(ctx, &gamepacket.ServerPlayerFirstShotReady{})
		case *gamepacket.ClientRoomInfo:
			room := c.currentLobby.GetRoom(ctx, int16(t.RoomNumber))
			if room != nil {
				info, err := room.GetRoomInfo(ctx)
				if err != nil {
					log.Printf("Error getting room info: %v", err)
				} else {
					c.SendMessage(ctx, &gamepacket.ServerRoomInfoResponse{
						RoomInfo: info,
					})
				}
			}
		case *gamepacket.ClientRequestInboxList:
			// TODO: need new sql message table
			msg := &gamepacket.ServerInboxList{
				PageNum:     t.PageNum,
				NumPages:    1,
				NumMessages: 1,
				Messages: []gamepacket.InboxMessage{
					{ID: 0x1, SenderNickname: "@Pangbox"},
				},
			}
			c.DebugMsg(msg)
			c.SendMessage(ctx, msg)
		case *gamepacket.ClientRequestInboxMessage:
			c.SendMessage(ctx, &gamepacket.ServerMailMessage{
				Message: gamepacket.MailMessage{
					ID:             0x1,
					SenderNickname: common.ToPString("@Pangbox"),
					DateTime:       common.ToPString("2023-06-03 01:21:00"),
					Message:        common.ToPString("Welcome to the first Pangbox server release! Not much works yet..."),
				},
			})
		case *gamepacket.ClientJoinChannel:
			c.SendMessage(ctx, &gamepacket.Server004E{Unknown: []byte{0x01}})
			c.SendMessage(ctx, &gamepacket.Server01F6{Unknown: []byte{0x00, 0x00, 0x00, 0x00}})
			c.SendMessage(ctx, &gamepacket.ServerLoginBonusStatus{Unknown: []byte{0x0, 0x0, 0x0, 0x0, 0x1, 0x4, 0x0, 0x0, 0x18, 0x3, 0x0, 0x0, 0x0, 0x27, 0x0, 0x0, 0x18, 0x3, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0}})
		case *gamepacket.ClientRequestDailyReward:
			c.SendMessage(ctx, &gamepacket.ServerMoneyUpdate{
				Type: uint16(gamepacket.MoneyUpdateRewardUnknown),
				RewardUnknown: &gamepacket.UpdateRewardUnknownData{
					Unknown: 1,
				},
			})
		case *gamepacket.ClientRequestPlayerHistory:
			c.SendMessage(ctx, &gamepacket.ServerPlayerHistory{})
		case *gamepacket.ClientMultiplayerJoin:
			if c.currentLobby != nil {
				break
			}
			log.Println("Join Lobby")
			c.currentLobby = c.s.lobby
			c.currentLobby.Send(ctx, room.LobbyPlayerJoin{
				Entry: c.getLobbyPlayer(),
				Conn:  c.ServerConn,
			})
		case *gamepacket.ClientMultiplayerLeave:
			if err := c.leaveMultiplayerLobby(ctx); err != nil {
				// TODO: handle error
				return err
			}
		case *gamepacket.ClientEventLobbyJoin:
			// TODO
			c.SendMessage(ctx, &gamepacket.ServerEventLobbyJoined{})
		case *gamepacket.ClientEventLobbyLeave:
			// TODO
			c.SendMessage(ctx, &gamepacket.ServerEventLobbyLeft{})
		case *gamepacket.ClientRoomJoin:
			if c.currentLobby == nil || c.currentRoom != nil {
				break
			}
			joinRoom := c.currentLobby.GetRoom(context.Background(), t.RoomNumber)
			if joinRoom != nil {
				joinRoom.Send(ctx, room.RoomPlayerJoin{
					Entry:      c.getRoomPlayer(),
					Conn:       c.ServerConn,
					PlayerData: c.getPlayerData(),
				})
				c.currentRoom = joinRoom
			}
		case *gamepacket.ClientHoleInfo:
			if c.currentRoom == nil {
				break
			}
			c.currentRoom.Send(ctx, room.RoomGameHoleInfo{
				Par:  t.Par,
				TeeX: t.TeeX,
				TeeZ: t.TeeZ,
				PinX: t.PinX,
				PinZ: t.PinZ,
			})
		case *gamepacket.ClientRoomLeave:
			if err := c.leaveRoom(ctx); err != nil {
				// TODO: handle error
				return err
			}
		case *gamepacket.ClientRoomKick:
			if c.currentRoom == nil {
				break
			}
			c.currentRoom.Send(ctx, room.RoomPlayerKick{
				ConnID:     c.connID,
				KickConnID: t.ConnID,
			})
		case *gamepacket.ClientRoomEdit:
			if c.currentRoom == nil {
				break
			}
			c.currentRoom.Send(ctx, room.RoomSettingsChange{
				ConnID:  c.connID,
				Changes: t.Changes,
			})
		case *gamepacket.Client0088:
			// Unknown tutorial-related message.
		case *gamepacket.ClientRareShopOpen:
			c.SendMessage(ctx, &gamepacket.ServerRareShopOpen{
				UnknownA: 50, // Prevents "No Draws Left" bug with Big Papel Shop
			})

		case *gamepacket.ClientAchievementStatusRequest:
			c.SendMessage(ctx, &gamepacket.ServerAchievementStatusResponse{
				// TODO: Actually support achievements.  This is the bare minimum not to hang/crash the game.
				Remaining: 1,
				Count:     1,
				Groups: []gamepacket.AchievementGroup{
					{
						GroupID: 0x4c80002d,
						ID:      248388034,
						Count:   5,
						Achievements: []gamepacket.Achievement{
							{
								ID:        0x7480087c,
								Value:     1,
								Timestamp: uint32(time.Date(2018, 11, 3, 0, 25, 10, 0, time.UTC).Unix()),
							},
							{
								ID:        0x7480087d,
								Value:     3,
								Timestamp: uint32(time.Date(2018, 11, 7, 2, 5, 26, 0, time.UTC).Unix()),
							},
							{
								ID:        0x7480087e,
								Value:     5,
								Timestamp: uint32(time.Date(2018, 11, 7, 3, 3, 45, 0, time.UTC).Unix()),
							},
							{
								ID:        0x7480087f,
								Value:     7,
								Timestamp: uint32(time.Date(2018, 11, 8, 23, 27, 19, 0, time.UTC).Unix()),
							},
							{
								ID:        0x74800880,
								Value:     10,
								Timestamp: uint32(time.Date(2019, 8, 25, 19, 39, 45, 0, time.UTC).Unix()),
							},
						},
					},
				},
			})
			c.SendMessage(ctx, &gamepacket.ServerAchievementUnknownResponse{
				UnknownA: [4]byte{0x00, 0x00, 0x00, 0x00},
			})
		case *gamepacket.ClientBigPapelPlay:
			c.SendMessage(ctx, &gamepacket.ServerBlackPapelResponse{
				RemainingTurns: 50, // Displays as remaining turns in the box.
				UnknownA:       -1,
			})
			// TODO: make Pang interaction transactional
			// TODO: make items show up in inventory
			// TODO: make sure not to subtract pang if a ticket is used.
			// TODO: Fix pang able to go negative.
			c.player.Pang, err = c.s.accountsService.AddPang(ctx, c.player.PlayerID, -10000)
			if err != nil {
				return err
			}
			packet := &gamepacket.ServerBigPapelWinnings{}
			const BigPapelItems = 10 // Unknown if this is actually how the number of items are actually chosen.
			itemQuantities := make(map[uint32]int)
			for i := 0; i < BigPapelItems; i++ {
				typeID := c.s.papelShop.Choose()
				rarity := c.s.papelRarity[typeID]
				if rarity == 2 {
					itemQuantities[typeID] = 1 // Technically should be possible to get multiple of the same rare item here.
				} else {
					itemQuantities[typeID] += 1 + rand.Intn(2)
				}
			}
			for typeID, quantity := range itemQuantities {
				item := gamepacket.ServerBlackPapelItems{}
				item.ItemTypeID = typeID
				item.Quantity = uint32(quantity)
				packet.Items = append(packet.Items, item)
			}
			packet.PangsRemaining = uint64(c.player.Pang)     // This should be calculated by player data
			packet.CookiesRemaining = uint64(c.player.Points) // Cookies remaining even if the action doesn't cost cookies
			c.SendMessage(ctx, packet)
		case *gamepacket.ClientBlackPapelPlay:
			// TODO: make Pang interaction transactional.
			// TODO: make items show up in inventory.
			// TODO: make sure not to subtract pang if a ticket is used.
			c.player.Pang, err = c.s.accountsService.AddPang(ctx, c.player.PlayerID, -500)
			if err != nil {
				return err
			}
			packet := &gamepacket.ServerBlackPapelWinnings{}
			drawnItemsSet := make(map[uint32]struct{})
			for i := rand.Intn(4) + 1; i > 0; {
				item := gamepacket.ServerBlackPapelItems{}
				var typeID uint32
				for {
					typeID = c.s.papelShop.Choose()
					if _, ok := drawnItemsSet[typeID]; !ok {
						drawnItemsSet[typeID] = struct{}{}
						break
					}
				}
				item.ItemTypeID = typeID
				item.DolfiniBallColor = uint32(rand.Intn(3))
				item.Rarity = c.s.papelRarity[typeID]
				if item.Rarity == 2 {
					item.Quantity = 1
				} else {
					item.Quantity = uint32(rand.Intn(3) + 1)
				}
				packet.UniqueItemsWon += 1
				packet.Items = append(packet.Items, item)
				i--
			}
			packet.PangsRemaining = uint64(c.player.Pang)
			packet.CookiesRemaining = uint64(c.player.Points)
			c.SendMessage(ctx, packet)
		case *gamepacket.ClientRoomUserEquipmentChange:
			// TODO
		case *gamepacket.ClientTutorialStart:
			// TODO
			c.SendMessage(ctx, &gamepacket.ServerRoomEquipmentData{
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
		case *gamepacket.ClientTutorialClear:
			// After clearing first tutorial
			// TODO
			c.SendMessage(ctx, &gamepacket.ServerTutorialStatus{
				Unknown: [6]byte{0x00, 0x01, 0x03, 0x00, 0x00, 0x00},
			})
		case *gamepacket.ClientEnterMyRoom:
			if t.UserID != t.RoomUserID {
				return errors.New("entering another user's myroom is not implemented yet")
			}
			c.SendMessage(ctx, &gamepacket.ServerMyRoomEntered{
				Unknown:  1,
				UserID:   t.UserID,
				Unknown2: 1,
			})
		case *gamepacket.ClientRequestInventory:
			if t.UserID != uint32(c.session.PlayerID) {
				return errors.New("wrong user ID specified in packet")
			}
			c.SendMessage(ctx, &gamepacket.ServerMyRoomLayout{
				Unknown:        1,
				FurnitureCount: 0,
			})
			c.SendMessage(ctx, &gamepacket.ServerPlayerInfo{
				Player: *c.getRoomPlayer(),
			})
		case *gamepacket.ClientLockerCombinationAttempt:
			// TODO
			c.SendMessage(ctx, &gamepacket.ServerLockerCombinationResponse{
				Status: 0,
			})
		case *gamepacket.ClientLockerInventoryRequest:
			// This has something to do with the combination/password system.
			c.SendMessage(ctx, &gamepacket.ServerLockerInventoryResponse{
				Status: 76,
			})
		case *gamepacket.Client00C1:
			// Seems to happen when entering my room.
			log.Debug("TODO: 00C1 (my room?)")
		case *gamepacket.ClientUserMacrosSet:
			// TODO: server-side macro storage
			log.Debugf("Set macros: %+v", t.MacroList)
		case *gamepacket.ClientEquipmentUpdate:
			switch {
			case t.CharParts != nil:
				if err := c.s.accountsService.SetCharacterParts(
					ctx,
					c.player.PlayerID,
					*t.CharParts,
				); err != nil {
					return err
				}
				if err := c.fetchPlayer(ctx); err != nil {
					return err
				}
				if err := c.SendMessage(ctx, &gamepacket.ServerPlayerEquipmentUpdated{
					Status:    0x04,
					Type:      uint8(gamepacket.UpdatedCharParts),
					CharParts: c.currentCharacter,
				}); err != nil {
					return err
				}

			case t.Caddie != nil:
				if err := c.s.accountsService.SetCaddie(ctx, c.player.PlayerID, int64(t.Caddie.CaddieID)); err != nil {
					return err
				}
				if err := c.fetchPlayer(ctx); err != nil {
					return err
				}
				if err := c.SendMessage(ctx, &gamepacket.ServerPlayerEquipmentUpdated{
					Status: 0x04,
					Type:   uint8(gamepacket.UpdatedCaddie),
					Caddie: &gamepacket.CaddieUpdated{
						CaddieID: t.Caddie.CaddieID,
					},
				}); err != nil {
					return err
				}

			case t.Consumables != nil:
				if err := c.s.accountsService.SetConsumables(ctx, c.player.PlayerID, t.Consumables.ItemTypeID, &c.player); err != nil {
					return err
				}
				if err := c.SendMessage(ctx, &gamepacket.ServerPlayerEquipmentUpdated{
					Status: 0x04,
					Type:   uint8(gamepacket.UpdatedConsumables),
					Consumables: &gamepacket.ConsumablesUpdated{
						ItemTypeID: c.getPlayerEquippedConsumables(),
					},
				}); err != nil {
					return err
				}

			case t.Comet != nil:
				cometID, err := c.s.accountsService.SetComet(ctx, c.player.PlayerID, t.Comet.ItemTypeID, &c.player)
				if err != nil {
					log.WithError(err).Errorf("attempt to set comet failed")
				}
				if err := c.SendMessage(ctx, &gamepacket.ServerPlayerEquipmentUpdated{
					Status: 0x04,
					Type:   uint8(gamepacket.UpdatedComet),
					Comet: &gamepacket.CometUpdated{
						ItemID:     uint32(cometID),
						ItemTypeID: uint32(c.player.BallTypeID),
					},
				}); err != nil {
					return err
				}

			case t.Decoration != nil:
				if err := c.s.accountsService.SetDecoration(ctx, c.player.PlayerID, accounts.DecorationTypeIDs{
					BackgroundTypeID: t.Decoration.BackgroundTypeID,
					FrameTypeID:      t.Decoration.FrameTypeID,
					StickerTypeID:    t.Decoration.StickerTypeID,
					SlotTypeID:       t.Decoration.SlotTypeID,
					CutInTypeID:      t.Decoration.CutInTypeID,
					TitleTypeID:      t.Decoration.TitleTypeID,
				}); err != nil {
					return err
				}
				if err := c.fetchPlayer(ctx); err != nil {
					return err
				}
				if err := c.SendMessage(ctx, &gamepacket.ServerPlayerEquipmentUpdated{
					Status: 0x04,
					Type:   uint8(gamepacket.UpdatedDecoration),
					Decoration: &gamepacket.DecorationUpdated{
						BackgroundTypeID: uint32(c.player.BackgroundTypeID.Int64),
						FrameTypeID:      uint32(c.player.FrameTypeID.Int64),
						StickerTypeID:    uint32(c.player.StickerTypeID.Int64),
						SlotTypeID:       uint32(c.player.SlotTypeID.Int64),
						CutInTypeID:      uint32(c.player.CutInTypeID.Int64),
						TitleTypeID:      uint32(c.player.TitleTypeID.Int64),
					},
				}); err != nil {
					return err
				}

			case t.Character != nil:
				if err := c.s.accountsService.SetCharacter(ctx, c.player.PlayerID, int64(t.Character.CharacterID)); err != nil {
					return err
				}
				if err := c.fetchPlayer(ctx); err != nil {
					return err
				}
				c.refreshCurrentCharacter()
				if err := c.SendMessage(ctx, &gamepacket.ServerPlayerEquipmentUpdated{
					Status:    0x04,
					Type:      uint8(gamepacket.UpdatedCharParts),
					CharParts: c.currentCharacter,
				}); err != nil {
					return err
				}
			}
		case *gamepacket.Client00FE:
			// TODO
			log.Debug("TODO: 00FE")
		case *gamepacket.ClientShopJoin:
			c.SendMessage(ctx, &gamepacket.Server020E{})
		case *gamepacket.ClientBuyItem:
			pangTotal := int64(0)
			pointTotal := int64(0)
			for _, item := range t.Items {
				// TODO: validate with iff
				pangTotal += int64(item.ItemCostPang)
				pointTotal += int64(item.ItemCostPoint)
			}
			if pangTotal > c.player.Pang || pointTotal > c.player.Points {
				c.SendMessage(ctx, &gamepacket.ServerPurchaseItemResponse{
					Status: 1,
				})
				continue
			}
			newCurrency := dbmodels.SetPlayerCurrencyRow{}
			for _, item := range t.Items {
				newCurrency, err = c.s.accountsService.PurchaseItem(ctx, c.session.PlayerID, int64(item.ItemCostPang), int64(item.ItemCostPoint), int64(item.ItemTypeID), int64(item.Quantity))
				if err != nil {
					return fmt.Errorf("purchasing item %v: %w", item.ItemTypeID, err)
				}
			}
			if err := c.SendMessage(ctx, &gamepacket.ServerPangBalanceData{
				PangsRemaining: uint64(newCurrency.Pang),
				PangsSpent:     uint64(pangTotal),
			}); err != nil {
				return err
			}
			if err := c.SendMessage(ctx, &gamepacket.ServerPointsBalance{
				Points: uint64(newCurrency.Points),
			}); err != nil {
				return err
			}
			if err := c.SendMessage(ctx, &gamepacket.ServerPurchaseItemResponse{
				Status: 0,
				Pang:   uint64(newCurrency.Pang),
				Points: uint64(newCurrency.Points),
			}); err != nil {
				return err
			}
			c.sendInventory(ctx)
			c.fetchCharacters(ctx)
			c.sendCharacterData(ctx)
		case *gamepacket.Client004F:
			// ignore
		default:
			return fmt.Errorf("unexpected message: %T", t)
		}
	}
}
