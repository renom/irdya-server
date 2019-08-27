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

package parse

import (
	"github.com/renom/wml"
)

func Version(text string) (string, error) {
	data := wml.ParseData(text)

	return data.ReadString("version.version")
}

func Account(text string) (string, error) {
	data := wml.ParseData(text)

	var err error
	login, err := data.ReadData("login")
	if err != nil {
		return "", err
	}

	return login.ReadString("username")
}
