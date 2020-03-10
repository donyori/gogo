// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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

// Return the length of parsing with cfg of x source bytes.
func ParsedLen(x int, cfg *FormatConfig) int {
	if x == 0 {
		return 0
	}
	if formatCfgNotValid(cfg) {
		return DecodedLen(x)
	}
	blockSize := cfg.BlockLen * 2
	size := blockSize + len(cfg.Sep)
	lastBlockSize := x % size
	if lastBlockSize > blockSize {
		lastBlockSize = blockSize
	}
	return x/size*cfg.BlockLen + lastBlockSize/2
}

// Return the length of parsing with cfg of x source bytes.
func ParsedLen64(x int64, cfg *FormatConfig) int64 {
	if x == 0 {
		return 0
	}
	if formatCfgNotValid(cfg) {
		return DecodedLen64(x)
	}
	blockLen := int64(cfg.BlockLen)
	blockSize := blockLen * 2
	size := blockSize + int64(len(cfg.Sep))
	lastBlockSize := x % size
	if lastBlockSize > blockSize {
		lastBlockSize = blockSize
	}
	return x/size*blockLen + lastBlockSize/2
}
