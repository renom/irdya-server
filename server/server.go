// This file is part of Irdya Server.
//
// Irdya Server is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Irdya Server is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Irdya Server.  If not, see <https://www.gnu.org/licenses/>.

package server

import (
	"errors"
	"net"
	"strconv"

	"github.com/renom/irdya-server/client"
	"github.com/renom/irdya-server/client/state"
)

type Server struct {
	port    uint16
	clients []*client.Client
}

func NewServer(port uint16) Server {
	return Server{port: port}
}

func (s *Server) Start() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.port)))
	if err != nil {
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	c := client.NewClient(s, conn)
	s.clients = append(s.clients, &c)

	go c.Handle()
}

func (s *Server) client(username string) (*client.Client, error) {
	for _, v := range s.clients {
		if v.Username == username {
			return v, nil
		}
	}
	return &client.Client{}, errors.New("There isn't such client.")
}

func (s *Server) Whisper(client *client.Client, receiver string, text string) {
	if c, err := s.client(receiver); err == nil {
		c.Whisper(client.Username, text)
	}
}

func (s *Server) Message(client *client.Client, room string, text string) {
	for _, v := range s.clients {
		if client.Username != v.Username && v.State == state.Lobby {
			v.Message(client.Username, room, text)
		}
	}
}

func (s *Server) UpdateLobby(client *client.Client) {
	for _, v := range s.clients {
		if client.Username != v.Username && v.State == state.Lobby {
			v.UpdateLobby(client.Username)
		}
	}
}

func (s *Server) UserList() []string {
	users := []string{}
	for _, v := range s.clients {
		if v.State == state.Lobby {
			users = append(users, v.Username)
		}
	}
	return users
}
