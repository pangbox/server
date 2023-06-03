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

package login

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/bufbuild/connect-go"
	"github.com/go-restruct/restruct"
	"github.com/pangbox/server/common"
	"github.com/pangbox/server/database/accounts"
	"github.com/pangbox/server/gen/dbmodels"
	"github.com/pangbox/server/gen/proto/go/topologypb"
	"github.com/pangbox/server/gen/proto/go/topologypb/topologypbconnect"
	"github.com/pangbox/server/pangya"
)

// Conn holds the state for a connection to the server.
type Conn struct {
	common.ServerConn[ClientMessage, ServerMessage]
	topologyClient  topologypbconnect.TopologyServiceClient
	accountsService *accounts.Service
}

// SendHello sends the initial handshake bytes to the client.
func (c *Conn) SendHello() error {
	data, err := restruct.Pack(binary.LittleEndian, &ConnectMessage{
		Unknown1: 0x0b00,
		Unknown2: 0x0000,
		Unknown3: 0x0000,
		Key:      uint16(c.Key),
		Unknown4: 0x0000,
		ServerID: 0x2775,
		Unknown6: 0x0000,
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

// GetServerList returns a server list using the topology store.
func (c *Conn) GetServerList(ctx context.Context, typ topologypb.Server_Type) (*ServerList, error) {
	message := &ServerList{}
	response, err := c.topologyClient.ListServers(ctx, connect.NewRequest(&topologypb.ListServersRequest{Type: typ}))
	if err != nil {
		return nil, fmt.Errorf("getting server list: %w", err)
	}

	for _, server := range response.Msg.Server {
		message.Servers = append(message.Servers, ServerEntry{
			ServerName: server.Name,
			ServerID:   server.Id,
			NumUsers:   server.NumUsers,
			MaxUsers:   server.MaxUsers,
			IPAddress:  server.Address,
			Port:       uint16(server.Port),
			Flags:      uint16(server.Flags),
		})
	}

	message.Count = uint8(len(message.Servers))

	return message, nil
}

// Handle runs the main connection loop.
func (c *Conn) Handle(ctx context.Context) error {
	log := c.Log
	c.Key = uint8(rand.Intn(16))

	err := c.SendHello()
	if err != nil {
		return fmt.Errorf("sending hello: %w", err)
	}

	msg, err := c.ReadMessage()
	if err != nil {
		return fmt.Errorf("reading handshake: %w", err)
	}

	var player dbmodels.Player
	switch t := msg.(type) {
	case *ClientLogin:
		player, err = c.accountsService.Authenticate(ctx, t.Username.Value, t.Password.Value)
	default:
		return fmt.Errorf("expected ClientLogin, got %T", t)
	}

	if err == accounts.ErrUnknownUsername || err == accounts.ErrInvalidPassword {
		log.Infof("Bad credentials.")
		c.SendMessage(&ServerLogin{
			Status: LoginStatusError,
			Error: &LoginError{
				Error: LoginErrorInvalidCredentials,
			},
		})
		c.Socket.Close()
		return nil
	} else if err != nil {
		return fmt.Errorf("database error during authentication: %w", err)
	}

	if !player.Nickname.Valid {
		c.SendMessage(&ServerLogin{
			Status: LoginStatusSetNickname,
			SetNickname: &LoginSetNickname{
				Unknown: 0xFFFFFFFF,
			},
		})

	NickSetup:
		for {
			msg, err := c.ReadMessage()
			if err != nil {
				return fmt.Errorf("reading handshake: %w", err)
			}

			switch t := msg.(type) {
			case *ClientCheckNickname:
				// TODO
				log.Infof("TODO: check nickname %s", t.Nickname.Value)
				c.SendMessage(&ServerNicknameCheckResponse{
					Nickname: t.Nickname,
				})
			case *ClientSetNickname:
				player, err = c.accountsService.SetNickname(ctx, player.PlayerID, t.Nickname.Value)
				if err != nil {
					// TODO: need to handle error
					log.Errorf("Database error setting nickname: %v", err)
					return nil
				}
				break NickSetup
			default:
				return fmt.Errorf("expected ClientCheckNickname, ClientSetNickname, got %T", t)
			}
		}
	}

	haveCharacters, err := c.accountsService.HasCharacters(ctx, player.PlayerID)
	if err != nil {
		log.Errorf("Database error getting characters: %v", err)
		return nil
	}

	if !haveCharacters {
		c.SendMessage(&ServerLogin{
			Status:       LoginStatusSetCharacter,
			SetCharacter: &LoginSetCharacter{},
		})

	CharacterSetup:
		for {
			msg, err := c.ReadMessage()
			if err != nil {
				return fmt.Errorf("reading handshake: %w", err)
			}

			switch t := msg.(type) {
			case *ClientSelectCharacter:
				c.accountsService.AddCharacter(ctx, player.PlayerID, pangya.PlayerCharacterData{
					CharTypeID: t.CharacterID,
					HairColor:  t.HairColor,
				})
				c.SendMessage(&Server0011{})
				break CharacterSetup
			default:
				return fmt.Errorf("expected ClientSelectCharacter, got %T", t)
			}
		}
	}

	session, err := c.accountsService.AddSession(ctx, player.PlayerID, c.Socket.RemoteAddr().String())
	if err != nil {
		log.Errorf("Error creating session in DB: %v", err)
	}

	// TODO: make token
	c.SendMessage(&ServerLoginSessionKey{
		SessionKey: common.ToPString(session.SessionKey),
	})

	c.SendMessage(&ServerLogin{
		Success: &LoginSuccess{
			Username: common.ToPString(player.Username),
			Nickname: common.ToPString(player.Nickname.String),
			UserID:   uint32(player.PlayerID),
		},
	})

	log.Info("sending message server list")
	messageServers, err := c.GetServerList(ctx, topologypb.Server_TYPE_MESSAGE_SERVER)
	if err != nil {
		return fmt.Errorf("listing message servers: %w", err)
	}
	c.SendMessage(&ServerMessageServerList{ServerList: *messageServers})

	log.Info("sending game server list")
	gameServers, err := c.GetServerList(ctx, topologypb.Server_TYPE_GAME_SERVER)
	if err != nil {
		return fmt.Errorf("listing game servers: %w", err)
	}
	c.SendMessage(&ServerGameServerList{ServerList: *gameServers})

	log.Info("waiting for response.")
	msg, err = c.ReadMessage()
	if err != nil {
		return fmt.Errorf("reading next message: %w", err)
	}

	switch t := msg.(type) {
	case *ClientSelectServer:
		log.Debugf("Select server: %+v", t)
	default:
		return fmt.Errorf("expected ClientSelectServer, got %T", t)
	}

	c.SendMessage(&ServerGameSessionKey{
		SessionKey: common.ToPString(session.SessionKey),
	})

	return nil
}
