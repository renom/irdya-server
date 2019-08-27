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

package commands

import "github.com/renom/wml"

func Message(sender string, room string, text string) []wml.Data {
	data := []wml.Data{}
	for _, v := range SplitMessage(wml.EscapeString(text)) {
		data = append(data, wml.Data{"message": wml.Data{"message": v, "room": room, "sender": sender}})
	}
	return data
}

func ServerMessage(text string) []wml.Data {
	data := []wml.Data{}
	for _, v := range SplitMessage(wml.EscapeString(text)) {
		data = append(data, wml.Data{"message": wml.Data{"message": v, "sender": "server"}})
	}
	return data
}

func Whisper(sender string, receiver string, text string) []wml.Data {
	data := []wml.Data{}
	for _, v := range SplitMessage(wml.EscapeString(text)) {
		data = append(data, wml.Data{"whisper": wml.Data{"sender": sender, "receiver": receiver, "message": v}})
	}
	return data
}
