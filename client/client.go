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

package client

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/renom/irdya-server/client/commands"
	"github.com/renom/irdya-server/client/parse"
	"github.com/renom/irdya-server/client/state"
	"github.com/renom/wml"
)

type Server interface {
	// Client-related methods
	Whisper(client *Client, receiver string, text string)
	Message(client *Client, room string, text string)
	UpdateLobby(client *Client)
	// Methods to obtain the server state
	UserList() []string
}

type Client struct {
	Username string
	State    state.State
	server   Server
	conn     net.Conn
}

func NewClient(server Server, conn net.Conn) Client {
	c := Client{State: state.Handshake, server: server, conn: conn}
	return c
}

func (c *Client) Handle() error {
	// Check for the handshake
	b, err := c.read(4)
	if err != nil && bytes.Equal(b, []byte{0, 0, 0, 0}) == false {
		return errors.New("Incorrect handshake.")
	}

	// Response to the handshake
	c.write(putUint32(42))

	// Request a game version
	c.sendData(wml.EmptyTag("version"))

	version, err := parse.Version(c.receiveString())

	if err != nil {
		return err
	}

	if parse.MajorVersion(version) != "1.14" {
		return err
	}

	c.sendData(wml.EmptyTag("mustlogin"))

	username, err := parse.Account(c.receiveString())
	if err != nil {
		return err
	}
	c.Username = username
	c.sendData(wml.EmptyTag("join_lobby"))

	userList := wml.Multiple{}
	for _, v := range append(c.server.UserList(), c.Username) {
		userList = append(userList, wml.Data{
			"available":  true,
			"game_id":    0,
			"location":   "",
			"name":       v,
			"registered": false,
			"status":     "lobby",
		})
	}
	c.sendData(wml.Data{
		"gamelist": wml.Data{},
		"user":     userList,
	})

	c.server.UpdateLobby(c)

	c.ServerMessage("Welcome to the Iradia Server!")

	c.State = state.Lobby

	return c.listen()
}

func (c *Client) listen() error {
	for {
		data := c.receiveWml()
		fmt.Printf("Received data: %q\n", data)
		tag, err := data.ToTag()
		if err != nil {
			continue
		}
		switch tag.Name {
		case "refresh_lobby":
		case "whisper":
			data := tag.Data
			var sender, receiver, message string
			var err error
			if sender, err = data.ReadString("sender"); err != nil {
				break
			}
			if receiver, err = data.ReadString("receiver"); err != nil {
				break
			}
			if message, err = data.ReadString("message"); err != nil {
				break
			}
			if sender == c.Username {
				c.server.Whisper(c, receiver, message)
			}
		case "query":
		case "nickserv":
		case "message":
			data := tag.Data
			var sender, room, message string
			var err error
			if sender, err = data.ReadString("sender"); err != nil {
				break
			}
			if room, err = data.ReadString("room"); err != nil {
				break
			}
			if message, err = data.ReadString("message"); err != nil {
				break
			}
			if sender == c.Username {
				c.server.Message(c, room, message)
			}
		case "create_game":
			c.sendData(wml.EmptyTag("leave_game"))
			c.ServerMessage("This server disallows any self-hosted games.")
		case "join":
		}
		time.Sleep(time.Second)
	}
	return nil
}

func (c *Client) UpdateLobby(user string) {
	userData := wml.Data{
		"available":  true,
		"game_id":    0,
		"location":   "",
		"name":       user,
		"registered": false,
		"status":     "lobby",
	}
	data := wml.Data{"gamelist_diff": wml.Data{"insert_child": wml.Data{"index": 1, "user": userData}}}
	c.sendData(data)
}

func (c *Client) Message(sender string, room string, text string) {
	c.sendData(commands.Message(sender, room, text))
}

func (c *Client) ServerMessage(text string) {
	c.sendData(commands.ServerMessage(text))
}

func (c *Client) Whisper(sender string, text string) {
	c.sendData(commands.Whisper(sender, c.Username, text))
}

func (c *Client) receiveWml() wml.Data {
	return wml.ParseData(c.receiveString())
}

func (c *Client) receiveString() string {
	return string(c.receiveBytes())
}

func (c *Client) receiveBytes() []byte {
	var err error
	buffer, err := c.read(4)
	if err != nil {
		return nil
	}
	if len(buffer) < 4 {
		return nil
	}
	size := int(binary.BigEndian.Uint32(buffer))
	data, err := c.read(size)
	if err != nil {
		return nil
	}
	reader, _ := gzip.NewReader(bytes.NewBuffer(data))
	var result []byte
	if result, err = ioutil.ReadAll(reader); err != nil {
		return nil
	}
	if err := reader.Close(); err != nil {
		return nil
	}
	return result
}

func (c *Client) sendData(data interface{}) {
	rawBytes := [][]byte{}

	switch data.(type) {
	case wml.Tag:
		rawBytes = append(rawBytes, data.(wml.Tag).Bytes())
	case []wml.Tag:
		for _, v := range data.([]wml.Tag) {
			rawBytes = append(rawBytes, v.Bytes())
		}
	case wml.Data:
		rawBytes = append(rawBytes, data.(wml.Data).Bytes())
	case []wml.Data:
		for _, v := range data.([]wml.Data) {
			rawBytes = append(rawBytes, v.Bytes())
		}
	case []byte:
		rawBytes = append(rawBytes, data.([]byte))
	case [][]byte:
		for _, v := range data.([][]byte) {
			rawBytes = append(rawBytes, v)
		}
	case string:
		rawBytes = append(rawBytes, []byte(data.(string)))
	case []string:
		for _, v := range data.([]string) {
			rawBytes = append(rawBytes, []byte(v))
		}
	default:
		return
	}

	for _, v := range rawBytes {
		var b bytes.Buffer

		gz := gzip.NewWriter(&b)
		gz.Write([]byte(v))
		gz.Close()

		var length int = len(b.Bytes())
		c.write(putUint32(uint32(length)))
		c.write(b.Bytes())
	}
}

func (c *Client) write(bytes []byte) (int, error) {
	return c.conn.Write(bytes)
}

func (c *Client) read(n int) ([]byte, error) {
	result := []byte{}
	count := 0
	for count < n {
		buffer := make([]byte, n-count)
		var num int
		num, err := c.conn.Read(buffer)
		if err != nil {
			return nil, err
		}
		count += num
		result = append(result, buffer[:num]...)
	}
	return result, nil
}
