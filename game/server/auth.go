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
)

func (c *Conn) handleAuth(ctx context.Context) error {
	if err := c.sendHello(ctx); err != nil {
		return fmt.Errorf("sending hello message: %w", err)
	}

	if err := c.waitForSessionAuth(ctx); err != nil {
		return fmt.Errorf("waiting for session auth: %w", err)
	}

	if err := c.fetchPlayer(ctx); err != nil {
		return fmt.Errorf("fetching player from db: %w", err)
	}

	if err := c.fetchCharacters(ctx); err != nil {
		return fmt.Errorf("fetching characters from db: %w", err)
	}

	if err := c.sendPlayerData(ctx); err != nil {
		return fmt.Errorf("sending player data to client: %w", err)
	}

	if err := c.sendCharacterData(ctx); err != nil {
		return fmt.Errorf("sending character data to client: %w", err)
	}

	if err := c.sendAchievementProgress(ctx); err != nil {
		return fmt.Errorf("sending achievemtn progress to client: %w", err)
	}

	if err := c.sendInventory(ctx); err != nil {
		return fmt.Errorf("sending inventory to client: %w", err)
	}

	if err := c.SendMessage(ctx, &gamepacket.ServerMessageConnect{}); err != nil {
		return fmt.Errorf("sending message server connect message: %w", err)
	}

	if err := c.sendServerList(ctx); err != nil {
		return fmt.Errorf("sending server list: %w", err)
	}

	return nil
}

func (c *Conn) sendHello(ctx context.Context) error {
	// TODO: remove hardcoded bytes
	return c.SendHello(&gamepacket.ConnectMessage{
		Unknown: [8]byte{0x00, 0x06, 0x00, 0x00, 0x3f, 0x00, 0x01, 0x01},
	})
}

func (c *Conn) waitForSessionAuth(ctx context.Context) error {
	msg, err := c.ReadMessage()
	if err != nil {
		return fmt.Errorf("reading message: %w", err)
	}

	switch t := msg.(type) {
	case *gamepacket.ClientAuth:
		c.session, err = c.s.accountsService.GetSessionByKey(ctx, t.LoginKey.Value)
		if err != nil {
			// TODO: error handling
			return err
		}
		c.connID = uint32(c.session.SessionID)

	default:
		return fmt.Errorf("expected client auth, got %T", t)
	}
	return nil
}

func (c *Conn) fetchPlayer(ctx context.Context) error {
	var err error
	c.player, err = c.s.accountsService.GetPlayer(ctx, c.session.PlayerID)
	if err != nil {
		// TODO: error handling
		return err
	}
	return nil
}

func (c *Conn) refreshCurrentCharacter() {
	for _, character := range c.characters {
		if character.ID == uint32(c.player.CharacterID.Int64) {
			c.currentCharacter = &character
			return
		}
	}
	if len(c.characters) > 0 {
		c.currentCharacter = &c.characters[0]
		return
	} else {
		c.currentCharacter = nil
	}
}

func (c *Conn) fetchCharacters(ctx context.Context) error {
	var err error
	c.characters, err = c.s.accountsService.GetCharacters(ctx, c.session.PlayerID)
	if err != nil {
		// TODO: handle error for client
		return fmt.Errorf("database error: %w", err)
	}
	c.refreshCurrentCharacter()

	return nil
}

func (c *Conn) sendPlayerData(ctx context.Context) error {
	return c.SendMessage(ctx, &gamepacket.ServerPlayerData{
		SubType: 0,
		MainData: &gamepacket.PlayerMainData{
			ClientVersion: common.ToPString("824.00"),
			ServerVersion: common.ToPString("Pangbox"),
			Game:          0xFFFF,
			PlayerData:    c.getPlayerData(),
		},
	})
}

func (c *Conn) sendCharacterData(ctx context.Context) error {
	return c.SendMessage(ctx, &gamepacket.ServerCharData{
		Count1:     uint16(len(c.characters)),
		Count2:     uint16(len(c.characters)),
		Characters: c.characters,
	})
}

func (c *Conn) sendAchievementProgress(ctx context.Context) error {
	return c.SendMessage(ctx, &gamepacket.ServerAchievementProgress{
		Remaining: 0,
		Count:     0,
	})
}

func (c *Conn) sendInventory(ctx context.Context) error {
	inventory, err := c.s.accountsService.GetPlayerInventory(ctx, c.session.PlayerID)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	for i, l := 0, 50; i < len(inventory); {
		if l > len(inventory) {
			l = len(inventory)
		}
		inventoryPkt := &gamepacket.ServerPlayerInventory{
			Remaining: uint16((len(inventory) - i) - l),
			Count:     uint16(l),
			Inventory: make([]gamepacket.InventoryItem, l),
		}
		for n := i + l; i < n; i++ {
			inventoryPkt.Inventory[i] = gamepacket.InventoryItem{
				ItemID:     uint32(inventory[i].ItemID),
				ItemTypeID: uint32(inventory[i].ItemTypeID),
				Unknown:    -1,
				Quantity:   uint32(inventory[i].Quantity.Int64),
				Unknown5: [156]byte{
					0x02, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
			}
		}
		if err := c.SendMessage(ctx, inventoryPkt); err != nil {
			return err
		}
	}

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
