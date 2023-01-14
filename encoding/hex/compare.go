// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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

import "github.com/donyori/gogo/constraints"

// CanEncodeTo reports whether src can be encoded to
// the hexadecimal representation x.
//
// It performs better than the comparison after encoding.
func CanEncodeTo[Bytes1, Bytes2 constraints.ByteString](src Bytes1, x Bytes2) bool {
	n := EncodedLen(len(src))
	if n != len(x) {
		return false
	}
	for i := 0; i < n; i += 2 {
		if x[i] < '0' || x[i+1] < '0' ||
			lowercaseHexTable[src[i>>1]>>4] != x[i]|letterCaseDiff ||
			lowercaseHexTable[src[i>>1]&0x0f] != x[i+1]|letterCaseDiff {
			return false
		}
	}
	return true
}

// CanEncodeToPrefix reports whether src can be encoded to
// the hexadecimal representation that has the specified prefix.
//
// It performs better than the comparison after encoding.
func CanEncodeToPrefix[Bytes1, Bytes2 constraints.ByteString](src Bytes1, prefix Bytes2) bool {
	n := len(prefix)
	if n > EncodedLen(len(src)) {
		return false
	}
	if n&1 > 0 { // i.e., n%2 == 1
		n--
	}
	for i := 0; i < n; i += 2 {
		if prefix[i] < '0' || prefix[i+1] < '0' ||
			lowercaseHexTable[src[i>>1]>>4] != prefix[i]|letterCaseDiff ||
			lowercaseHexTable[src[i>>1]&0x0f] != prefix[i+1]|letterCaseDiff {
			return false
		}
	}
	if len(prefix)&1 > 0 { // i.e., len(prefix)%2 == 1
		// Here, n == len(prefix) - 1
		return prefix[n] >= '0' &&
			lowercaseHexTable[src[n>>1]>>4] == prefix[n]|letterCaseDiff
	}
	return true
}
