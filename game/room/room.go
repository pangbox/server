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
	"math"
	"math/rand"
	"reflect"
	"time"

	"github.com/pangbox/server/common"
	"github.com/pangbox/server/common/actor"
	"github.com/pangbox/server/database/accounts"
	gamemodel "github.com/pangbox/server/game/model"
	gamepacket "github.com/pangbox/server/game/packet"
	"github.com/pangbox/server/pangya"
	log "github.com/sirupsen/logrus"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

type Room struct {
	actor.Base[RoomEvent]
	state    gamemodel.RoomState
	players  *orderedmap.OrderedMap[uint32, RoomPlayer]
	lobby    *Lobby
	accounts *accounts.Service
}

type RoomPlayer struct {
	Entry      *gamemodel.RoomPlayerEntry
	Conn       *gamepacket.ServerConn
	PlayerData pangya.PlayerData
	GameReady  bool
	ShotSync   *gamemodel.ShotSyncData
	TurnEnd    bool
	HoleEnd    bool
	Pang       uint64
	BonusPang  uint64
	LastTotal  int8
	Stroke     int8
	Score      int32
	TurnOrder  int
	Distance   float64
}

func (r *Room) Start(ctx context.Context, state gamemodel.RoomState, lobby *Lobby, accounts *accounts.Service) bool {
	return r.TryStart(ctx, func(ctx context.Context, t *actor.Task[RoomEvent]) error {
		r.state.Active = true
		r.state.Open = true
		r.state.ShotTimerMS = state.ShotTimerMS
		r.state.GameTimerMS = state.GameTimerMS
		r.state.NumUsers = state.NumUsers
		r.state.MaxUsers = state.MaxUsers
		r.state.RoomType = state.RoomType
		r.state.NumHoles = state.NumHoles
		r.state.Course = state.Course
		r.state.RoomName = state.RoomName
		r.state.Password = state.Password
		r.state.HoleProgression = state.HoleProgression
		r.state.NaturalWind = state.NaturalWind
		r.state.GamePhase = gamemodel.LobbyPhase
		r.players = orderedmap.New[uint32, RoomPlayer]()
		r.lobby = lobby
		r.accounts = accounts
		return r.task(ctx, t)
	})
}

func (r *Room) Number() int16 {
	if r == nil {
		return -1
	}
	return r.state.RoomNumber
}

func (r *Room) GetRoomInfo(ctx context.Context) (gamemodel.RoomInfo, error) {
	promise, err := r.Send(ctx, RoomGetInfo{})
	if err != nil {
		return gamemodel.RoomInfo{}, err
	}
	result, err := promise.Wait(ctx)
	if err != nil {
		return gamemodel.RoomInfo{}, err
	}
	return result.(gamemodel.RoomInfo), nil
}

func (r *Room) broadcast(ctx context.Context, msg gamepacket.ServerMessage) error {
	group, ctx := errgroup.WithContext(ctx)
	for pair := r.players.Oldest(); pair != nil; pair = pair.Next() {
		player := pair.Value
		group.Go(func() error {
			return player.Conn.SendMessage(ctx, msg)
		})
	}
	return group.Wait()
}

func (r *Room) stateUpdated(ctx context.Context) error {
	r.lobby.Send(ctx, LobbyRoomUpdate{
		Room: r.state,
	})
	r.broadcast(ctx, r.roomStatus())
	return nil
}

func (r *Room) getRoomPlayerList() []gamemodel.RoomPlayerEntry {
	playerList := make([]gamemodel.RoomPlayerEntry, r.players.Len())
	for i, pair := 0, r.players.Oldest(); pair != nil; pair = pair.Next() {
		playerList[i] = *pair.Value.Entry
		playerList[i].Slot = uint8(i + 1)
		i++
	}
	return playerList
}

func (r *Room) task(ctx context.Context, t *actor.Task[RoomEvent]) error {
	defer func() {
		r.state.Active = false
		r.lobby.Send(ctx, LobbyRoomRemove{
			Room: r.state,
		})
	}()

	for {
		msg, err := t.Receive()
		if err != nil {
			return err
		}
		if err := r.handleEvent(ctx, t, msg); err != nil {
			return err
		}
		if r.players.Len() == 0 {
			break
		}
	}

	return nil
}

func (r *Room) handleEvent(ctx context.Context, t *actor.Task[RoomEvent], msg actor.Message[RoomEvent]) error {
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
	case RoomGetInfo:
		info, err := r.handleRoomInfo(ctx, event)
		if err != nil {
			msg.Promise.Reject(err)
		} else {
			msg.Promise.Resolve(info)
		}
		return nil

	case RoomPlayerJoin:
		return rejectOnError(r.handlePlayerJoin(ctx, event))

	case RoomPlayerLeave:
		return rejectOnError(r.handlePlayerLeave(ctx, event))

	case RoomAction:
		return rejectOnError(r.handleRoomAction(ctx, event))

	case RoomPlayerIdle:
		return rejectOnError(r.handleRoomPlayerIdle(ctx, event))

	case RoomPlayerReady:
		return rejectOnError(r.handleRoomPlayerReady(ctx, event))

	case RoomPlayerKick:
		return rejectOnError(r.handleRoomPlayerKick(ctx, event))

	case RoomSettingsChange:
		return rejectOnError(r.handleRoomSettingsChange(ctx, event))

	case RoomStartGame:
		return rejectOnError(r.handleRoomStartGame(ctx, event))

	case RoomLoadingProgress:
		return rejectOnError(r.handleRoomLoadingProgress(ctx, event))

	case RoomGameReady:
		return rejectOnError(r.handleRoomGameReady(ctx, event))

	case RoomGameShotCommit:
		return rejectOnError(r.handleRoomGameShotCommit(ctx, event))

	case RoomGameShotRotate:
		return rejectOnError(r.handleRoomGameShotRotate(ctx, event))

	case RoomGameShotPower:
		return rejectOnError(r.handleRoomGameShotPower(ctx, event))

	case RoomGameShotClubChange:
		return rejectOnError(r.handleRoomGameShotClubChange(ctx, event))

	case RoomGameShotItemUse:
		return rejectOnError(r.handleRoomGameShotItemUse(ctx, event))

	case RoomGameTypingIndicator:
		return rejectOnError(r.handleRoomGameTypingIndicator(ctx, event))

	case RoomGameShotCometRelief:
		return rejectOnError(r.handleRoomGameShotCometRelief(ctx, event))

	case RoomGameTurn:
		return rejectOnError(r.handleRoomGameTurn(ctx, event))

	case RoomGameTurnEnd:
		return rejectOnError(r.handleRoomGameTurnEnd(ctx, event))

	case RoomGameHoleEnd:
		return rejectOnError(r.handleRoomGameHoleEnd(ctx, event))

	case RoomGameShotSync:
		return rejectOnError(r.handleRoomGameShotSync(ctx, event))

	case RoomGameHoleInfo:
		return rejectOnError(r.handleRoomGameHoleInfo(ctx, event))

	case ChatMessage:
		return rejectOnError(r.handleChatMessage(ctx, event))

	default:
		return fmt.Errorf("unknown event: %T", event)
	}
}

func (r *Room) handleRoomInfo(ctx context.Context, event RoomGetInfo) (gamemodel.RoomInfo, error) {
	info := gamemodel.RoomInfo{
		PlayerCount: uint32(r.players.Len()),
		NumHoles:    r.state.NumHoles,
		Unknown:     0,
		Course:      r.state.Course,
		RoomType:    r.state.RoomType,
		Users:       make([]gamemodel.RoomInfoPlayer, r.players.Len()),
	}

	for i, pair := 0, r.players.Oldest(); pair != nil; pair = pair.Next() {
		player := pair.Value
		info.Users[i] = gamemodel.RoomInfoPlayer{
			ConnID:      player.Entry.ConnID,
			Rank:        player.Entry.Rank,
			PlayerFlags: player.Entry.PlayerFlags,
			TitleID:     player.Entry.TitleID,
		}
		i++
	}

	return info, nil
}

func (r *Room) handlePlayerJoin(ctx context.Context, event RoomPlayerJoin) error {
	if r.players.Len() >= int(r.state.MaxUsers) {
		return errors.New("room full")
	}
	if r.players.Len() == 0 {
		// New room
		event.Entry.StatusFlags |= gamemodel.RoomStateMaster
		r.state.OwnerConnID = event.Entry.ConnID
	}
	_, present := r.players.Set(event.Entry.ConnID, RoomPlayer{
		Entry:      event.Entry,
		Conn:       event.Conn,
		PlayerData: event.PlayerData,
	})
	if present {
		return errors.New("already in room")
	}

	event.Conn.SendMessage(ctx, &gamepacket.ServerRoomJoin{
		RoomName:    r.state.RoomName,
		RoomNumber:  r.state.RoomNumber,
		EventNumber: 0,
	})

	event.Conn.SendMessage(ctx, r.roomStatus())

	r.lobby.Send(ctx, LobbyPlayerUpdateRoom{
		ConnID:     event.Entry.ConnID,
		RoomNumber: -1,
	})

	err := r.broadcastPlayerList(ctx)
	if err != nil {
		log.WithError(err).Error("error broadcasting room status")
	}

	r.state.NumUsers = uint8(r.players.Len())
	r.stateUpdated(ctx)

	return nil
}

func (r *Room) handlePlayerLeave(ctx context.Context, event RoomPlayerLeave) error {
	return r.removePlayer(ctx, event.ConnID)
}

func (r *Room) handleRoomAction(ctx context.Context, event RoomAction) error {
	if event.Action.Rotation != nil {
		if pair := r.players.GetPair(event.ConnID); pair != nil {
			pair.Value.Entry.Angle = event.Action.Rotation.Z
			return r.broadcast(ctx, &gamepacket.ServerRoomAction{
				ConnID:     pair.Value.Entry.ConnID,
				RoomAction: event.Action,
			})
		}
	}
	return nil
}

func (r *Room) handleRoomPlayerIdle(ctx context.Context, event RoomPlayerIdle) error {
	if pair := r.players.GetPair(event.ConnID); pair != nil {
		if event.Idle {
			pair.Value.Entry.StatusFlags |= gamemodel.RoomStateAway
		} else {
			pair.Value.Entry.StatusFlags &^= gamemodel.RoomStateAway
		}
		// TODO: not sure what to broadcast
		return r.broadcastPlayerList(ctx)
	}
	return nil
}

func (r *Room) handleRoomPlayerReady(ctx context.Context, event RoomPlayerReady) error {
	if pair := r.players.GetPair(event.ConnID); pair != nil {
		state := byte(0)
		if event.Ready {
			pair.Value.Entry.StatusFlags |= gamemodel.RoomStateReady
		} else {
			pair.Value.Entry.StatusFlags &^= gamemodel.RoomStateReady
			state = 1
		}
		return r.broadcast(ctx, &gamepacket.ServerPlayerReady{
			ConnID: pair.Value.Entry.ConnID,
			State:  state,
		})
	}
	return nil
}

func (r *Room) handleRoomPlayerKick(ctx context.Context, event RoomPlayerKick) error {
	pair := r.players.GetPair(event.ConnID)
	if pair == nil {
		return nil
	}
	subject := r.players.GetPair(event.KickConnID)
	if subject == nil {
		return nil
	}
	if pair.Value.Entry.ConnID != r.state.OwnerConnID {
		return nil
	}
	return r.removePlayer(ctx, subject.Value.Entry.ConnID)
}

func (r *Room) handleRoomSettingsChange(ctx context.Context, event RoomSettingsChange) error {
	if event.ConnID != r.state.OwnerConnID {
		return nil
	}

	for _, change := range event.Changes {
		if change.RoomName != nil {
			r.state.RoomName = change.RoomName.Value
		}
		if change.RoomType != nil {
			r.state.RoomType = *change.RoomType
		}
		if change.Course != nil {
			r.state.Course = *change.Course
		}
		if change.NumHoles != nil {
			r.state.NumHoles = *change.NumHoles
		}
		if change.HoleProgression != nil {
			r.state.HoleProgression = *change.HoleProgression
		}
		if change.ShotTimerSeconds != nil {
			r.state.ShotTimerMS = uint32(*change.ShotTimerSeconds) * 1000
		}
		if change.MaxUsers != nil {
			r.state.MaxUsers = *change.MaxUsers
		}
		if change.GameTimerMinutes != nil {
			r.state.GameTimerMS = uint32(*change.GameTimerMinutes) * 60 * 1000
		}
		if change.NaturalWind != nil {
			r.state.NaturalWind = *change.NaturalWind
		}
	}

	r.stateUpdated(ctx)
	return nil
}

func (r *Room) handleRoomStartGame(ctx context.Context, event RoomStartGame) error {
	r.state.Open = false
	r.state.GamePhase = gamemodel.WaitingLoad
	r.stateUpdated(ctx)

	// Pick hole numbers to play.
	h := make([]uint8, 18)
	for i := 0; i < len(h); i++ {
		h[i] = uint8(i + 1)
	}
	switch r.state.HoleProgression {
	case 0:
		break
	case 1:
		h = h[len(h)-int(r.state.NumHoles):]
	case 2:
		h = h[rand.Intn(len(h)-int(r.state.NumHoles)+1):]
	case 3:
		rand.Shuffle(len(h), func(i, j int) { h[i], h[j] = h[j], h[i] })
	}
	h = h[:r.state.NumHoles]

	// Generate holes in state.
	r.state.Holes = make([]gamemodel.RoomHole, r.state.NumHoles)
	for i := 0; i < int(r.state.NumHoles); i++ {
		r.state.Holes[i] = gamemodel.RoomHole{
			Course:  r.state.Course,
			HoleNum: h[i],
			HoleID:  rand.Uint32(),
			Pin:     0, // TODO
		}
	}
	r.state.CurrentHole = 1

	// TODO
	r.broadcast(ctx, &gamepacket.Server0230{})
	r.broadcast(ctx, &gamepacket.Server0231{})
	r.broadcast(ctx, &gamepacket.Server0077{Unknown: 0x64})

	// Send game init packet.
	now := time.Now()
	gameInit := &gamepacket.ServerGameInit{
		SubType: gamepacket.GameInitTypeFull,
		Full: &gamepacket.GameInitFull{
			NumPlayers: byte(r.players.Len()),
			Players:    make([]gamepacket.GamePlayer, r.players.Len()),
		},
	}
	for i, pair := 0, r.players.Oldest(); pair != nil; pair = pair.Next() {
		// Clear ready status.
		pair.Value.Entry.StatusFlags &^= gamemodel.RoomStateReady

		// Set initial turn order.
		pair.Value.TurnOrder = i
		pair.Value.Distance = math.Inf(1)
		pair.Value.Stroke = 0
		pair.Value.LastTotal = 0

		player := pair.Value
		gameInit.Full.Players[i] = gamepacket.GamePlayer{
			Number:     uint16(i + 1),
			PlayerData: player.PlayerData,
			StartTime:  pangya.NewSystemTime(now),
			NumCards:   0,
		}

		i++
	}
	r.broadcast(ctx, gameInit)

	// Send room game data packet.
	gameData := &gamepacket.ServerRoomGameData{
		Course:          r.state.Course,
		Unknown:         0x0,
		HoleProgression: r.state.HoleProgression,
		NumHoles:        r.state.NumHoles,
		Unknown2:        0x0,
		ShotTimerMS:     r.state.ShotTimerMS,
		GameTimerMS:     r.state.GameTimerMS,
		RandomSeed:      rand.Uint32(),
	}
	// Copy hole data from state.
	for i := uint8(0); i < r.state.NumHoles; i++ {
		stateHole := r.state.Holes[i]
		gameData.Holes[i] = gamepacket.HoleInfo{
			HoleID:  stateHole.HoleID,
			HoleNum: stateHole.HoleNum,
			Pin:     stateHole.Pin,
			Course:  stateHole.Course,
		}
	}
	r.broadcast(ctx, gameData)

	// TODO
	r.broadcast(ctx, &gamepacket.Server016A{Unknown: 1, Unknown2: 0x24bd})

	return nil
}

func (r *Room) handleRoomLoadingProgress(ctx context.Context, event RoomLoadingProgress) error {
	return r.broadcast(ctx, &gamepacket.ServerPlayerLoadProgress{
		ConnID:   event.ConnID,
		Progress: event.Progress,
	})
}

func (r *Room) handleRoomGameReady(ctx context.Context, event RoomGameReady) error {
	if pair := r.players.GetPair(event.ConnID); pair != nil {
		pair.Value.GameReady = true
	}
	if r.checkGameReady() {
		r.startHole(ctx)
	}
	return nil
}

func (r *Room) handleRoomGameShotCommit(ctx context.Context, event RoomGameShotCommit) error {
	return r.broadcast(ctx, &gamepacket.ServerRoomShotAnnounce{
		ConnID:           event.ConnID,
		ShotStrength:     event.ShotStrength,
		ShotAccuracy:     event.ShotAccuracy,
		ShotEnglishCurve: event.ShotEnglishCurve,
		ShotEnglishSpin:  event.ShotEnglishSpin,
		Unknown2:         event.Unknown2,
		Unknown3:         event.Unknown3,
	})
}

func (r *Room) handleRoomGameShotRotate(ctx context.Context, event RoomGameShotRotate) error {
	return r.broadcast(ctx, &gamepacket.ServerRoomShotRotateAnnounce{
		ConnID: event.ConnID,
		Angle:  event.Angle,
	})
}

func (r *Room) handleRoomGameShotPower(ctx context.Context, event RoomGameShotPower) error {
	return r.broadcast(ctx, &gamepacket.ServerRoomShotPowerAnnounce{
		ConnID: event.ConnID,
		Level:  event.Level,
	})
}

func (r *Room) handleRoomGameShotClubChange(ctx context.Context, event RoomGameShotClubChange) error {
	return r.broadcast(ctx, &gamepacket.ServerRoomClubChangeAnnounce{
		ConnID: event.ConnID,
		Club:   event.Club,
	})
}

func (r *Room) handleRoomGameShotItemUse(ctx context.Context, event RoomGameShotItemUse) error {
	return r.broadcast(ctx, &gamepacket.ServerRoomItemUseAnnounce{
		ConnID:     event.ConnID,
		ItemTypeID: event.ItemTypeID,
	})
}

func (r *Room) handleRoomGameTypingIndicator(ctx context.Context, event RoomGameTypingIndicator) error {
	return r.broadcast(ctx, &gamepacket.ServerRoomUserTypingAnnounce{
		ConnID: event.ConnID,
		Status: event.Status,
	})
}

func (r *Room) handleRoomGameShotCometRelief(ctx context.Context, event RoomGameShotCometRelief) error {
	return r.broadcast(ctx, &gamepacket.ServerRoomShotCometReliefAnnounce{
		ConnID: event.ConnID,
		X:      event.X,
		Y:      event.Y,
		Z:      event.Z,
	})
}

func (r *Room) handleRoomGameTurn(ctx context.Context, event RoomGameTurn) error {
	return nil
}

func (r *Room) handleRoomGameTurnEnd(ctx context.Context, event RoomGameTurnEnd) error {
	if pair := r.players.GetPair(event.ConnID); pair != nil {
		pair.Value.TurnEnd = true
	}
	if r.checkShouldEndTurn() {
		return r.endTurn(ctx)
	}
	return nil
}

func (r *Room) handleRoomGameHoleEnd(ctx context.Context, event RoomGameHoleEnd) error {
	if pair := r.players.GetPair(event.ConnID); pair != nil {
		pair.Value.HoleEnd = true
		pair.Value.Score += int32(pair.Value.Stroke) - int32(r.currentHole().Par)
		pair.Value.LastTotal = pair.Value.Stroke
		pair.Value.Stroke = 0
	}
	return nil
}

func (r *Room) handleRoomGameShotSync(ctx context.Context, event RoomGameShotSync) error {
	syncData := event.Data
	if r.state.ShotSync == nil {
		r.state.ShotSync = &syncData
	} else {
		// TODO: this is inefficient, also does nothing terribly useful.
		if !reflect.DeepEqual(*r.state.ShotSync, syncData) {
			log.Warningf("Shot sync mismatch: %#v vs %#v", r.state.ShotSync, syncData)
		}
	}
	if pair := r.players.GetPair(event.ConnID); pair != nil {
		pair.Value.ShotSync = r.state.ShotSync
	}
	if r.checkShotSync() {
		r.broadcast(ctx, &gamepacket.ServerRoomShotSync{
			Data: *r.state.ShotSync,
		})
		if pair := r.players.GetPair(r.state.ShotSync.ActiveConnID); pair != nil {
			pair.Value.Pang = uint64(r.state.ShotSync.Pang)
			pair.Value.BonusPang = uint64(r.state.ShotSync.BonusPang)

			// TODO: Sometimes we need to increment twice, need to compare packets
			pair.Value.Stroke++

			dx := float64(r.currentHole().PinX) - float64(r.state.ShotSync.X)
			dy := float64(r.currentHole().PinZ) - float64(r.state.ShotSync.Z)
			pair.Value.Distance = math.Sqrt(dx*dx + dy*dy)
		} else {
			log.WithField("ConnID", r.state.ShotSync.ActiveConnID).Warn("couldn't find conn")
		}
		for pair := r.players.Oldest(); pair != nil; pair = pair.Next() {
			pair.Value.ShotSync = nil
		}
		r.state.ShotSync = nil
	}
	return nil
}

func (r *Room) handleRoomGameHoleInfo(ctx context.Context, event RoomGameHoleInfo) error {
	hole := r.currentHole()

	// TODO: It'd probably be better to not rely on the client for this if possible.
	hole.Par = event.Par
	hole.TeeX = event.TeeX
	hole.TeeZ = event.TeeZ
	hole.PinX = event.PinX
	hole.PinZ = event.PinZ

	return nil
}

func (r *Room) handleChatMessage(ctx context.Context, event ChatMessage) error {
	msg := &gamepacket.ServerEvent{Type: gamepacket.ChatMessageEvent}
	msg.Data.Message = common.ToPString(event.Message)
	msg.Data.Nickname = common.ToPString(event.Nickname)
	return r.broadcast(ctx, msg)
}

func (r *Room) endTurn(ctx context.Context) error {
	r.broadcast(ctx, &gamepacket.ServerRoomShotEnd{
		ConnID: r.state.ActiveConnID,
	})
	for pair := r.players.Oldest(); pair != nil; pair = pair.Next() {
		pair.Value.TurnEnd = false
	}
	nextPlayer := r.getNextPlayer()
	if nextPlayer == nil {
		return r.endHole(ctx)
	}
	r.state.ActiveConnID = nextPlayer.Entry.ConnID
	r.broadcast(ctx, &gamepacket.ServerRoomActiveUserAnnounce{
		ConnID: r.state.ActiveConnID,
	})
	return nil
}

func (r *Room) getNextPlayer() *RoomPlayer {
	var nextPlayer *RoomPlayer
	for pair := r.players.Oldest(); pair != nil; pair = pair.Next() {
		// Don't consider players who are finished with this hole.
		if pair.Value.HoleEnd {
			continue
		}
		// If we don't have a candidate yet, then use the first player we see.
		if nextPlayer == nil {
			nextPlayer = &pair.Value
			continue
		}
		// Prefer players further away.
		if pair.Value.Distance > nextPlayer.Distance {
			nextPlayer = &pair.Value
			continue
		}
		// Finally, prefer players with a lower turn order.
		if pair.Value.Distance == nextPlayer.Distance && pair.Value.TurnOrder < nextPlayer.TurnOrder {
			nextPlayer = &pair.Value
			continue
		}
	}
	if nextPlayer != nil {
		log.Printf("next player: %s, with distance: %f", nextPlayer.Entry.Nickname, nextPlayer.Distance)
	}
	// Note: can return nil if everyone has holed out.
	return nextPlayer
}

func (r *Room) endHole(ctx context.Context) error {
	r.state.CurrentHole++
	if r.state.CurrentHole > r.state.NumHoles {
		return r.endGame(ctx)
	} else {
		r.broadcast(ctx, &gamepacket.ServerRoomFinishHole{})
		for pair := r.players.Oldest(); pair != nil; pair = pair.Next() {
			pair.Value.HoleEnd = false
		}
		r.setupNextTurnOrder()
	}
	r.stateUpdated(ctx)
	return nil
}

func (r *Room) setupNextTurnOrder() {
	players := []*RoomPlayer{}
	for pair := r.players.Oldest(); pair != nil; pair = pair.Next() {
		players = append(players, &pair.Value)
		log.Printf("before: %s: last=%d, order=%d", pair.Value.Entry.Nickname, pair.Value.LastTotal, pair.Value.TurnOrder)
	}
	slices.SortFunc(players, func(a, b *RoomPlayer) bool {
		// If two players tied on last hole, they should maintain their previous
		// turn order; so fall back to sorting by current turn order.
		if a.LastTotal == b.LastTotal {
			return a.TurnOrder < b.TurnOrder
		}
		return a.LastTotal < b.LastTotal
	})
	for i, player := range players {
		player.TurnOrder = i
		player.Distance = math.Inf(1)
		log.Printf("after: %s: last=%d, order=%d", player.Entry.Nickname, player.LastTotal, player.TurnOrder)
	}
}

func (r *Room) endGame(ctx context.Context) error {
	results := &gamepacket.ServerRoomFinishGame{
		NumPlayers: uint8(r.players.Len()),
		Standings:  make([]gamepacket.PlayerGameResult, r.players.Len()),
	}
	for i, pair := 0, r.players.Oldest(); pair != nil; pair = pair.Next() {
		bonusPang := pair.Value.BonusPang
		bonusPang += r.lobby.configProvider.GetCourseBonus(r.state.Course, r.players.Len(), int(r.state.NumHoles))
		results.Standings[i].ConnID = pair.Value.Entry.ConnID
		results.Standings[i].Pang = pair.Value.Pang
		results.Standings[i].Score = int8(pair.Value.Score)
		results.Standings[i].BonusPang = bonusPang

		totalPang := bonusPang + pair.Value.Pang

		newPang, err := r.accounts.AddPang(ctx, int64(pair.Value.Entry.PlayerID), int64(totalPang))
		if err != nil {
			log.WithError(err).Error("failed giving game-ending pang")
		}

		if err := pair.Value.Conn.SendMessage(ctx, &gamepacket.ServerPangBalanceData{PangsRemaining: uint64(newPang)}); err != nil {
			log.WithError(err).Error("failed informing player of game-ending pang")
		}

		pair.Value.Score = 0
		pair.Value.Pang = 0
		pair.Value.BonusPang = 0
		pair.Value.HoleEnd = false
		pair.Value.ShotSync = nil

		i++
	}
	slices.SortFunc(results.Standings, func(a, b gamepacket.PlayerGameResult) bool {
		return a.Score < b.Score
	})
	results.Standings[0].Place = 1
	for i := 1; i < len(results.Standings); i++ {
		if results.Standings[i-1].Score == results.Standings[i].Score {
			// If tie: use placement of tied player(s)
			results.Standings[i].Place = results.Standings[i-1].Place
		} else {
			// If not tie: use position in standing as placement
			results.Standings[i].Place = uint8(i + 1)
		}
	}
	r.broadcast(ctx, results)
	r.state.Open = true
	r.state.CurrentHole = 0
	r.state.GamePhase = gamemodel.LobbyPhase
	return nil
}

func (r *Room) checkGameReady() bool {
	for pair := r.players.Oldest(); pair != nil; pair = pair.Next() {
		if !pair.Value.GameReady {
			return false
		}
	}
	return true
}

func (r *Room) checkShotSync() bool {
	for pair := r.players.Oldest(); pair != nil; pair = pair.Next() {
		if pair.Value.ShotSync == nil {
			return false
		}
	}
	return true
}

func (r *Room) checkShouldEndTurn() bool {
	for pair := r.players.Oldest(); pair != nil; pair = pair.Next() {
		if !pair.Value.TurnEnd {
			return false
		}
	}
	return true
}

func (r *Room) startHole(ctx context.Context) error {
	r.state.GamePhase = gamemodel.InGame
	r.broadcast(ctx, &gamepacket.ServerRoomSetWeather{
		Weather: 0,
	})
	wind := rand.Intn(8) + 1
	r.broadcast(ctx, &gamepacket.ServerRoomSetWind{
		Wind:    uint8(wind),
		Unknown: 0,
		Heading: uint16(rand.Intn(256)),
		Reset:   true,
	})
	nextPlayer := r.getNextPlayer()
	if nextPlayer == nil {
		log.Error("nextPlayer == nil in startHole?")
		return nil
	}
	r.state.ActiveConnID = nextPlayer.Entry.ConnID
	r.broadcast(ctx, &gamepacket.ServerRoomStartHole{
		ConnID: r.state.ActiveConnID,
	})
	// TODO: These blobs are taken from an old packet dump. Not exactly sure what they are for.
	r.broadcast(ctx, &gamepacket.Server0151{Unknown: []byte{
		0x0d, 0x00, 0x57, 0x5f, 0x42, 0x49, 0x47, 0x42, 0x4f, 0x4e, 0x47, 0x44, 0x41, 0x52, 0x49, 0x00,
		0x03, 0x01, 0x03, 0x02, 0x03, 0x03, 0x02, 0x00, 0x02, 0x02, 0x02, 0x03, 0x01, 0x01, 0x00, 0x01,
		0x00, 0x03, 0x02, 0x00, 0x00, 0x00, 0x02, 0x03, 0x03, 0x00, 0x01, 0x01, 0x03, 0x00, 0x02, 0x03,
		0x01, 0x03, 0x03, 0x01, 0x02, 0x00, 0x03, 0x00, 0x02, 0x00, 0x00, 0x02, 0x00, 0x03, 0x03, 0x03,
		0x02, 0x02, 0x02, 0x03, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x02, 0x01, 0x01, 0x00, 0x03,
		0x00, 0x01, 0x00, 0x03, 0x03, 0x03, 0x02, 0x00, 0x01, 0x01, 0x02, 0x03, 0x03, 0x01, 0x02, 0x00,
		0x00, 0x02, 0x03, 0x02, 0x00, 0x00, 0x03, 0x02, 0x03, 0x00, 0x03, 0x00, 0x03, 0x02, 0x03, 0x02,
		0x03, 0x00, 0x03,
	}})
	r.broadcast(ctx, &gamepacket.Server0151{Unknown: []byte{
		0x0d, 0x00, 0x52, 0x5f, 0x42, 0x49, 0x47, 0x42, 0x4f, 0x4e, 0x47, 0x44, 0x41, 0x52, 0x49, 0x01,
		0x02, 0x00, 0x00, 0x01, 0x03, 0x01, 0x00, 0x01, 0x02, 0x02, 0x02, 0x03, 0x03, 0x02, 0x02, 0x01,
		0x01, 0x03, 0x03, 0x00, 0x02, 0x02, 0x02, 0x03, 0x01, 0x02, 0x02, 0x03, 0x00, 0x00, 0x00, 0x02,
		0x03, 0x03, 0x01, 0x02, 0x01, 0x03, 0x01, 0x00, 0x01, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0x03,
		0x01, 0x00, 0x02, 0x02, 0x00, 0x02, 0x03, 0x00, 0x03, 0x03, 0x01, 0x03, 0x02, 0x01, 0x02, 0x03,
		0x03, 0x03, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x00, 0x01, 0x03, 0x00, 0x01, 0x02, 0x00, 0x00,
		0x02, 0x02, 0x03, 0x00, 0x01, 0x02, 0x01, 0x02, 0x01, 0x03, 0x02, 0x01, 0x01, 0x03, 0x01, 0x02,
		0x00, 0x02, 0x01,
	}})
	r.broadcast(ctx, &gamepacket.Server0151{Unknown: []byte{
		0x0f, 0x00, 0x43, 0x4c, 0x55, 0x42, 0x53, 0x45, 0x54, 0x5f, 0x4d, 0x49, 0x52, 0x41, 0x43, 0x4c,
		0x45, 0x01, 0x01, 0x01, 0x02, 0x02, 0x00, 0x02, 0x01, 0x02, 0x03, 0x01, 0x03, 0x00, 0x02, 0x02,
		0x03, 0x03, 0x01, 0x01, 0x02, 0x02, 0x00, 0x03, 0x02, 0x01, 0x01, 0x01, 0x03, 0x01, 0x00, 0x02,
		0x01, 0x03, 0x03, 0x03, 0x02, 0x01, 0x03, 0x03, 0x03, 0x02, 0x03, 0x01, 0x00, 0x00, 0x03, 0x00,
		0x01, 0x02, 0x00, 0x02, 0x03, 0x02, 0x02, 0x02, 0x00, 0x03, 0x02, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x00, 0x01, 0x01, 0x01, 0x01, 0x02, 0x02, 0x03, 0x02, 0x01, 0x00, 0x01, 0x03, 0x03, 0x03, 0x00,
		0x03, 0x02, 0x02, 0x02, 0x03, 0x00, 0x00, 0x02, 0x02, 0x00, 0x00, 0x00, 0x02, 0x01, 0x01, 0x03,
		0x01, 0x00, 0x01, 0x02, 0x00,
	}})

	return nil
}

func (r *Room) removePlayer(ctx context.Context, connID uint32) error {
	if pair := r.players.GetPair(connID); pair != nil {
		err := pair.Value.Conn.SendMessage(ctx, &gamepacket.ServerRoomLeave{
			RoomNumber: -1,
		})
		r.broadcast(ctx, &gamepacket.ServerRoomCensus{
			Type:    byte(gamepacket.ListRemove),
			Unknown: -1,
			ListRemove: &gamepacket.RoomCensusListRemove{
				ConnID: pair.Value.Entry.ConnID,
			},
		})
		r.players.Delete(connID)
		r.lobby.Send(ctx, LobbyPlayerUpdateRoom{
			ConnID:     connID,
			RoomNumber: -1,
		})
		if r.state.OwnerConnID == pair.Value.Entry.ConnID && r.players.Len() > 0 {
			newOwner := r.players.Oldest()
			newOwner.Value.Entry.StatusFlags |= gamemodel.RoomStateMaster
			r.state.OwnerConnID = newOwner.Value.Entry.ConnID
			r.broadcastPlayerList(ctx)
		}
		r.state.NumUsers = uint8(r.players.Len())
		r.stateUpdated(ctx)
		return err
	} else {
		return errors.New("user not in room")
	}
}

func (r *Room) broadcastPlayerList(ctx context.Context) error {
	playerList := r.getRoomPlayerList()

	return r.broadcast(ctx, &gamepacket.ServerRoomCensus{
		Type:    byte(gamepacket.ListSet),
		Unknown: -1,
		ListSet: &gamepacket.RoomCensusListSet{
			PlayerCount: uint8(len(playerList)),
			PlayerList:  playerList,
		},
	})
}

func (r *Room) roomStatus() *gamepacket.ServerRoomStatus {
	return &gamepacket.ServerRoomStatus{
		RoomType:        r.state.RoomType,
		Course:          r.state.Course,
		NumHoles:        r.state.NumHoles,
		HoleProgression: 1,
		NaturalWind:     0,
		MaxUsers:        r.state.MaxUsers,
		ShotTimerMS:     r.state.ShotTimerMS,
		GameTimerMS:     r.state.GameTimerMS,
		Flags:           0,
		Owner:           true,
		RoomName:        common.ToPString(r.state.RoomName),
	}
}

func (r *Room) currentHole() *gamemodel.RoomHole {
	return &r.state.Holes[r.state.CurrentHole-1]
}
