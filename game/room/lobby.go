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

package room

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pangbox/server/common"
	"github.com/pangbox/server/common/actor"
	"github.com/pangbox/server/database/accounts"
	gamemodel "github.com/pangbox/server/game/model"
	gamepacket "github.com/pangbox/server/game/packet"
	"github.com/pangbox/server/gameconfig"
	log "github.com/sirupsen/logrus"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/sync/errgroup"
)

type Lobby struct {
	actor.Base[LobbyEvent]
	logger         *log.Entry
	storage        *Storage
	players        *orderedmap.OrderedMap[uint32, *LobbyPlayer]
	accounts       *accounts.Service
	configProvider gameconfig.Provider
}

type LobbyPlayer struct {
	Entry  gamemodel.LobbyPlayer
	Conn   *gamepacket.ServerConn
	Joined time.Time
}

func NewLobby(ctx context.Context, logger *log.Entry, accounts *accounts.Service, configProvider gameconfig.Provider) *Lobby {
	lobby := &Lobby{
		logger:         logger,
		storage:        new(Storage),
		players:        orderedmap.New[uint32, *LobbyPlayer](),
		accounts:       accounts,
		configProvider: configProvider,
	}
	lobby.TryStart(ctx, lobby.task)
	return lobby
}

func (l *Lobby) NewRoom(ctx context.Context, room gamemodel.RoomState) (*Room, error) {
	promise, err := l.Send(ctx, LobbyRoomCreate{Room: room})
	if err != nil {
		return nil, err
	}
	result, err := promise.Wait(ctx)
	if err != nil {
		return nil, err
	}
	return result.(*Room), nil
}

func (l *Lobby) GetRoom(ctx context.Context, roomNumber int16) *Room {
	return l.storage.GetRoom(ctx, roomNumber)
}

func (l *Lobby) broadcast(ctx context.Context, message gamepacket.ServerMessage) error {
	group, ctx := errgroup.WithContext(ctx)
	for pair := l.players.Oldest(); pair != nil; pair = pair.Next() {
		player := pair.Value

		// Only broadcast to users in the main lobby area.
		if player.Entry.RoomNumber != -1 {
			continue
		}

		group.Go(func() error {
			return player.Conn.SendMessage(ctx, message)
		})
	}
	return group.Wait()
}

func (l *Lobby) task(ctx context.Context, t *actor.Task[LobbyEvent]) error {
	for {
		msg, err := t.Receive()
		if err != nil {
			return err
		}
		if err := l.handleEvent(ctx, t, msg); err != nil {
			return err
		}
	}
}

func (l *Lobby) handleEvent(ctx context.Context, t *actor.Task[LobbyEvent], msg actor.Message[LobbyEvent]) error {
	defer msg.Promise.Close()

	rejectOnError := func(err error) error {
		if err != nil {
			msg.Promise.Reject(err)
		} else {
			msg.Promise.Resolve(nil)
		}
		return nil
	}

	switch event := msg.Value.(type) {
	case LobbyPlayerJoin:
		return rejectOnError(l.lobbyPlayerJoin(ctx, &event))

	case LobbyPlayerUpdate:
		return rejectOnError(l.lobbyPlayerUpdate(ctx, &event))

	case LobbyPlayerUpdateRoom:
		return rejectOnError(l.lobbyPlayerUpdateRoom(ctx, &event))

	case LobbyPlayerLeave:
		return rejectOnError(l.lobbyPlayerLeave(ctx, &event))

	case LobbyRoomCreate:
		room, err := l.lobbyRoomCreate(ctx, &event)
		if err != nil {
			msg.Promise.Reject(err)
		} else {
			msg.Promise.Resolve(room)
		}
		return nil

	case LobbyRoomUpdate:
		return rejectOnError(l.lobbyRoomUpdate(ctx, &event))

	case LobbyRoomRemove:
		return rejectOnError(l.lobbyRoomRemove(ctx, &event))

	case ChatMessage:
		return rejectOnError(l.lobbyChat(ctx, &event))

	default:
		return fmt.Errorf("unknown event: %T", event)
	}
}

func (l *Lobby) lobbyRoomCreate(ctx context.Context, e *LobbyRoomCreate) (*Room, error) {
	room := l.storage.NewRoom(ctx)
	room.Start(ctx, e.Room, l, l.accounts)
	e.Room.RoomNumber = room.Number()
	l.broadcast(ctx, &gamepacket.ServerRoomList{
		Count:   1,
		Type:    gamepacket.ListAdd,
		Unknown: 0xFFFF,
		RoomList: []gamepacket.RoomListRoom{
			roomToList(&e.Room),
		},
	})
	return room, nil
}

func (l *Lobby) lobbyRoomUpdate(ctx context.Context, e *LobbyRoomUpdate) error {
	err := l.storage.UpdateRoom(ctx, e.Room)
	if err != nil {
		return err
	}
	return l.broadcast(ctx, &gamepacket.ServerRoomList{
		Count:   1,
		Type:    gamepacket.ListChange,
		Unknown: 0xFFFF,
		RoomList: []gamepacket.RoomListRoom{
			roomToList(&e.Room),
		},
	})
}

func (l *Lobby) lobbyRoomRemove(ctx context.Context, e *LobbyRoomRemove) error {
	err := l.storage.UpdateRoom(ctx, e.Room)
	if err != nil {
		return err
	}
	return l.broadcast(ctx, &gamepacket.ServerRoomList{
		Count:   1,
		Type:    gamepacket.ListRemove,
		Unknown: 0xFFFF,
		RoomList: []gamepacket.RoomListRoom{
			roomToList(&e.Room),
		},
	})
}

func (l *Lobby) lobbyPlayerUpdate(ctx context.Context, e *LobbyPlayerUpdate) error {
	if player, ok := l.players.Get(e.Entry.ConnID); ok {
		player.Entry = e.Entry

		l.broadcast(ctx, &gamepacket.ServerUserCensus{
			Type:  gamepacket.UserChange,
			Count: 1,
			PlayerList: []gamemodel.LobbyPlayer{
				player.Entry,
			},
		})
	}
	return nil
}

func (l *Lobby) lobbyPlayerUpdateRoom(ctx context.Context, e *LobbyPlayerUpdateRoom) error {
	if player, ok := l.players.Get(e.ConnID); ok {
		// Player has left a room
		if player.Entry.RoomNumber != -1 && e.RoomNumber == -1 {
			l.playerSyncLobbyState(ctx, player.Conn)
		}

		player.Entry.RoomNumber = e.RoomNumber

		l.broadcast(ctx, &gamepacket.ServerUserCensus{
			Type:  gamepacket.UserChange,
			Count: 1,
			PlayerList: []gamemodel.LobbyPlayer{
				player.Entry,
			},
		})
	}
	return nil
}

func (l *Lobby) lobbyPlayerJoin(ctx context.Context, e *LobbyPlayerJoin) error {
	l.broadcast(ctx, &gamepacket.ServerUserCensus{
		Type:  gamepacket.UserAdd,
		Count: 1,
		PlayerList: []gamemodel.LobbyPlayer{
			e.Entry,
		},
	})

	l.players.Set(e.Entry.ConnID, &LobbyPlayer{
		Entry: e.Entry,
		Conn:  e.Conn,
	})

	l.playerSyncLobbyState(ctx, e.Conn)

	if err := e.Conn.SendMessage(ctx, &gamepacket.ServerMultiplayerJoined{}); err != nil {
		log.WithError(err).Error("error sending multiplayer joined")
	}

	return nil
}

func (l *Lobby) playerSyncLobbyState(ctx context.Context, conn *gamepacket.ServerConn) error {
	msg := &gamepacket.ServerUserCensus{
		Type: gamepacket.UserListSet,
	}

	playerList := make([]gamemodel.LobbyPlayer, 0, gamepacket.CensusMaxUsers)
	for pair := l.players.Oldest(); pair != nil; pair = pair.Next() {
		player := pair.Value
		playerList = append(playerList, player.Entry)
		if len(playerList) == gamepacket.CensusMaxUsers {
			msg.Count = uint8(len(playerList))
			msg.PlayerList = playerList
			if err := conn.SendMessage(ctx, msg); err != nil {
				log.WithError(err).Error("error sending player list")
			}
			playerList = playerList[0:0]
			msg.Type = gamepacket.UserListAppend
		}
	}
	if len(playerList) > 0 {
		msg.Count = uint8(len(playerList))
		msg.PlayerList = playerList
		if err := conn.SendMessage(ctx, msg); err != nil {
			log.WithError(err).Error("error sending player list")
		}
	}

	roomList := l.storage.GetRoomList()

	roomListMsg := &gamepacket.ServerRoomList{
		Count:    uint8(len(roomList)),
		Type:     gamepacket.ListSet,
		Unknown:  0xFFFF,
		RoomList: []gamepacket.RoomListRoom{},
	}

	for _, room := range roomList {
		roomListMsg.RoomList = append(roomListMsg.RoomList, roomToList(&room.state))
	}

	if err := conn.SendMessage(ctx, roomListMsg); err != nil {
		log.WithError(err).Error("error sending room list")
	}

	return nil
}

func (l *Lobby) lobbyPlayerLeave(ctx context.Context, e *LobbyPlayerLeave) error {
	player, ok := l.players.Delete(e.ConnID)
	if !ok {
		return errors.New("no such player")
	}
	player.Conn.SendMessage(ctx, &gamepacket.ServerMultiplayerLeft{})
	return l.broadcast(ctx, &gamepacket.ServerUserCensus{
		Type:  gamepacket.UserRemove,
		Count: 1,
		PlayerList: []gamemodel.LobbyPlayer{
			player.Entry,
		},
	})
}

func (l *Lobby) lobbyChat(ctx context.Context, e *ChatMessage) error {
	event := &gamepacket.ServerEvent{Type: gamepacket.ChatMessageEvent}
	event.Data.Message = common.ToPString(e.Message)
	event.Data.Nickname = common.ToPString(e.Nickname)
	err := l.broadcast(ctx, event)
	if err != nil {
		log.WithError(err).Error("error broadcasting lobby chat message")
	}

	return nil
}

func roomToList(state *gamemodel.RoomState) gamepacket.RoomListRoom {
	return gamepacket.RoomListRoom{
		Name:            state.RoomName,
		Public:          true, // TODO
		Open:            state.Open,
		UserMax:         state.MaxUsers,
		UserCount:       state.NumUsers,
		Unknown3:        0x1E,
		NumHoles:        state.NumHoles,
		Number:          state.RoomNumber,
		HoleProgression: 0, // TODO
		Course:          state.Course,
		ShotTimerMS:     state.ShotTimerMS,
		GameTimerMS:     state.GameTimerMS,
		OwnerID:         state.OwnerConnID,
		Class:           255, // TODO
		ArtifactID:      0,   // TODO
	}
}
