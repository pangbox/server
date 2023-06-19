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

	"github.com/bufbuild/connect-go"
	"github.com/pangbox/server/common"
	gamepacket "github.com/pangbox/server/game/packet"
	"github.com/pangbox/server/gen/proto/go/topologypb"
	"github.com/pangbox/server/pangya"
	log "github.com/sirupsen/logrus"
)

func (c *Conn) handleAuth(ctx context.Context) error {
	err := c.SendHello(&gamepacket.ConnectMessage{
		Unknown: [8]byte{0x00, 0x06, 0x00, 0x00, 0x3f, 0x00, 0x01, 0x01},
	})
	if err != nil {
		return fmt.Errorf("sending hello message: %w", err)
	}

	msg, err := c.ReadMessage()
	if err != nil {
		return fmt.Errorf("reading handshake: %w", err)
	}

	switch t := msg.(type) {
	case *gamepacket.ClientAuth:
		c.session, err = c.s.accountsService.GetSessionByKey(ctx, t.LoginKey.Value)
		if err != nil {
			// TODO: error handling
			return nil
		}
		c.player, err = c.s.accountsService.GetPlayer(ctx, c.session.PlayerID)
		if err != nil {
			// TODO: error handling
			return nil
		}
		log.Debugf("Client auth: %#v", msg)

	default:
		return fmt.Errorf("expected client auth, got %T", t)
	}

	c.characters, err = c.s.accountsService.GetCharacters(ctx, c.session.PlayerID)
	if err != nil {
		// TODO: handle error for client
		return fmt.Errorf("database error: %w", err)
	}

	c.connID = uint32(c.session.SessionID)

	// TODO: need data modelling
	c.playerData = pangya.PlayerData{
		UserInfo: pangya.PlayerInfo{
			Username: c.player.Username,
			Nickname: c.player.Nickname.String,
			PlayerID: uint32(c.player.PlayerID),
			ConnID:   c.connID,
		},
		PlayerStats: pangya.PlayerStats{
			Pang: uint64(c.player.Pang),
		},
		Items: pangya.PlayerEquipment{
			CaddieID:    0,
			CharacterID: c.characters[0].ID,
			ClubSetID:   0x1754,
			CometTypeID: 0x14000000,
			Items: pangya.PlayerEquippedItems{
				ItemIDs: [10]uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
		},
		EquippedCharacter: c.characters[0],
		EquippedClub: pangya.PlayerClubData{
			Item: pangya.PlayerItem{
				ID:     0x1754,
				TypeID: 0x10000000,
			},
			Stats: pangya.ClubStats{
				UpgradeStats: [5]uint16{8, 9, 8, 3, 3},
			},
		},
	}

	c.SendMessage(ctx, &gamepacket.ServerUserData{
		SubType: 0,
		MainData: &gamepacket.PlayerMainData{
			ClientVersion: common.ToPString("824.00"),
			ServerVersion: common.ToPString("Pangbox"),
			Game:          0xFFFF,
			PlayerData:    c.playerData,
		},
	})

	c.SendMessage(ctx, &gamepacket.ServerCharData{
		Count1:     uint16(len(c.characters)),
		Count2:     uint16(len(c.characters)),
		Characters: c.characters,
	})

	c.SendMessage(ctx, &gamepacket.ServerAchievementProgress{
		Remaining: 0,
		Count:     0,
	})

	c.SendMessage(ctx, &gamepacket.ServerMessageConnect{})

	c.sendServerList(ctx)

	return nil
}

func (c *Conn) sendServerList(ctx context.Context) error {
	message := &gamepacket.ServerChannelList{}
	response, err := c.s.topologyClient.ListServers(ctx, connect.NewRequest(&topologypb.ListServersRequest{
		Type: topologypb.Server_TYPE_GAME_SERVER,
	}))
	if err != nil {
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
	return c.SendMessage(ctx, message)
}
