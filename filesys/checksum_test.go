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

package filesys_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/donyori/gogo/filesys"
)

func TestVerifyChecksumFromFs(t *testing.T) {
	wrongChecksum := filesys.HashChecksum{
		NewHash: sha256.New,
		ExpHex:  strings.Repeat("0", hex.EncodedLen(sha256.Size)),
	}

	testCases := make([]struct {
		name     string
		filename string
		cs       []filesys.HashChecksum
		want     bool
	}, 1+len(testFsFilenames)*8)
	testCases[0].name = `file="nonexist"&cs=<nil>`
	testCases[0].filename = "nonexist"
	idx := 1
	for _, filename := range testFsFilenames {
		for i := 0; i < 8; i++ {
			testCases[idx+i].filename = filename
		}
		namePrefix := fmt.Sprintf("file=%q&cs=", filename)

		testCases[idx].name = namePrefix + "<nil>"
		testCases[idx].want = true

		testCases[idx+1].name = namePrefix + "correct"
		testCases[idx+1].cs = testFsChecksumMap[filename]
		testCases[idx+1].want = true

		testCases[idx+2].name = namePrefix + "wrong"
		testCases[idx+2].cs = []filesys.HashChecksum{wrongChecksum}

		testCases[idx+3].name = namePrefix + "correct+wrong"
		testCases[idx+3].cs = []filesys.HashChecksum{testFsChecksumMap[filename][0], wrongChecksum}

		testCases[idx+4].name = namePrefix + "zero-value"
		testCases[idx+4].cs = []filesys.HashChecksum{{}}

		testCases[idx+5].name = namePrefix + "noExpHex"
		testCases[idx+5].cs = []filesys.HashChecksum{{NewHash: sha256.New}}

		testCases[idx+6].name = namePrefix + "noHash"
		testCases[idx+6].cs = []filesys.HashChecksum{{ExpHex: testFsChecksumMap[filename][0].ExpHex}}

		testCases[idx+7].name = namePrefix + "correct+empty"
		testCases[idx+7].cs = []filesys.HashChecksum{testFsChecksumMap[filename][0], {}}

		idx += 8
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := filesys.VerifyChecksumFromFs(testFs, tc.filename, tc.cs...); got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}
