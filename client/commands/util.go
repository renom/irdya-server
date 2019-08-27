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

import "strings"

func SplitMessage(text string) []string {
	upperLimit := 256

	result := []string{}

	for pos := 0; pos < len(text); {
		from := pos
		var to int
		if pos+upperLimit < len(text) {
			splitIndex := strings.LastIndex(text[pos:pos+upperLimit+1], "\n")
			if splitIndex == -1 {
				if text[pos+upperLimit-1] == '"' && strings.Count(text[pos:pos+upperLimit], "\"\"")%2 != 0 {
					splitIndex = pos + upperLimit
				}
			}
			if splitIndex != -1 {
				to = pos + splitIndex
			} else {
				to = pos + upperLimit
			}
		} else {
			to = len(text)
		}
		result = append(result, text[from:to])
		pos = to
	}
	return result
}
