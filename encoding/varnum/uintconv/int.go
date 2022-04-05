// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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

package uintconv

// FromInt64Zigzag maps a 64-bit signed integer to
// a 64-bit unsigned integer with zigzag encoding.
//
// FromInt64Zigzag(ToInt64Zigzag(x)) == x.
func FromInt64Zigzag(i int64) uint64 {
	if i >= 0 {
		return uint64(i) << 1
	}
	return (^uint64(i) << 1) | 1
}

// ToInt64Zigzag maps a 64-bit unsigned integer back to
// a 64-bit signed integer with zigzag encoding.
//
// ToInt64Zigzag(FromInt64Zigzag(x)) == x.
func ToInt64Zigzag(u uint64) int64 {
	if u&1 == 0 {
		return int64(u >> 1)
	}
	return ^int64(u >> 1)
}
