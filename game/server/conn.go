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
	"fmt"

	"github.com/pangbox/server/common"
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
	player     dbmodels.Player
	playerData pangya.PlayerData
	characters []pangya.PlayerCharacterData

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
		CharTypeID:       c.playerData.EquippedCharacter.CharTypeID,
		StatusFlags:      0,
		GuildEmblemImage: "guildmark", // TODO
		PlayerID:         uint32(c.player.PlayerID),
		CharacterData:    c.playerData.EquippedCharacter,
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
				ShotTimerMS: t.ShotTimerMS,
				GameTimerMS: t.GameTimerMS,
				MaxUsers:    t.MaxUsers,
				RoomType:    t.RoomType,
				NumHoles:    t.NumHoles,
				Course:      t.Course,
				RoomName:    t.RoomName.Value,
				Password:    t.Password.Value,
				// TODO: natural wind, hole progression, more?
			})
			if err != nil {
				// TODO: handle error
				return err
			}
			c.currentRoom = newRoom
			c.currentRoom.Send(ctx, room.RoomPlayerJoin{
				Entry:      c.getRoomPlayer(),
				Conn:       c.ServerConn,
				PlayerData: c.playerData,
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
			c.currentRoom.Send(ctx, room.RoomGameShotItemUse{
				ConnID:     c.connID,
				ItemTypeID: t.ItemTypeID,
			})
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
		case *gamepacket.Client001A:
			// Do nothing.
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
					PlayerData: c.playerData,
				})
				c.currentRoom = joinRoom
			}
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
		case *gamepacket.ClientUserMacrosSet:
			// TODO: server-side macro storage
			log.Debugf("Set macros: %+v", t.MacroList)
		case *gamepacket.ClientEquipmentUpdate:
			// TODO
			log.Debug("TODO: 0020")
		case *gamepacket.Client00FE:
			// TODO
			log.Debug("TODO: 00FE")
		case *gamepacket.ClientShopJoin:
			// Enter shop, not sure what responses need to go here?
			log.Debug("TODO: 0140")
		default:
			return fmt.Errorf("unexpected message: %T", t)
		}
	}
}
