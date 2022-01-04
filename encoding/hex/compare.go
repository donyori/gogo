// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
//
// This file is part of gogo.
//
// gogo is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package hex

// CanEncode reports whether src can be encoded to
// the hexadecimal representation x (in type []byte).
//
// It performs better than the comparison after encoding.
func CanEncode(src, x []byte) bool {
	if EncodedLen(len(src)) != len(x) {
		return false
	}
	var k int
	for _, b := range src {
		if x[k] < '0' || x[k+1] < '0' ||
			lowercaseHexTable[b>>4] != x[k]|letterCaseDiff ||
			lowercaseHexTable[b&0x0f] != x[k+1]|letterCaseDiff {
			return false
		}
		k += 2
	}
	return true
}

// CanEncodeToString reports whether src can be encoded to
// the hexadecimal representation x (in type string).
//
// It performs better than the comparison after encoding.
func CanEncodeToString(src []byte, x string) bool {
	if EncodedLen(len(src)) != len(x) {
		return false
	}
	var k int
	for _, b := range src {
		if x[k] < '0' || x[k+1] < '0' ||
			lowercaseHexTable[b>>4] != x[k]|letterCaseDiff ||
			lowercaseHexTable[b&0x0f] != x[k+1]|letterCaseDiff {
			return false
		}
		k += 2
	}
	return true
}
