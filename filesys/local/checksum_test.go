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

package local_test

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/filesys/local"
)

type ChecksumInfo struct {
	Filename string `json:"filename"`
	Sha256   string `json:"sha256"`
	Md5      string `json:"md5"`
}

func TestVerifyChecksum(t *testing.T) {
	const dir = "testdata"
	wrongChecksum := filesys.HashChecksum{
		NewHash: sha256.New,
		ExpHex:  strings.Repeat("0", hex.EncodedLen(sha256.Size)),
	}

	t.Run(`file="nonexist"&cs=<nil>`, func(t *testing.T) {
		if got := local.VerifyChecksum(filepath.Join(dir, "nonexist")); got {
			t.Error("got true; want false")
		}
	})

	checksumJsonData, err := os.ReadFile(filepath.Join(dir, "checksum.json"))
	if err != nil {
		t.Fatal(err)
	}
	var checksums []*ChecksumInfo
	err = json.Unmarshal(checksumJsonData, &checksums)
	if err != nil {
		t.Fatal(err)
	}
	for _, checksum := range checksums {
		filename := filepath.Join(dir, checksum.Filename)
		correctChecksums := []filesys.HashChecksum{
			{
				NewHash: sha256.New,
				ExpHex:  checksum.Sha256,
			},
			{
				NewHash: md5.New,
				ExpHex:  checksum.Md5,
			},
		}
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			testCases := []struct {
				csName string
				cs     []filesys.HashChecksum
				want   bool
			}{
				{"<nil>", nil, true},
				{"correct", correctChecksums, true},
				{"wrong", []filesys.HashChecksum{wrongChecksum}, false},
				{"correct+wrong", []filesys.HashChecksum{correctChecksums[0], wrongChecksum}, false},
				{"zero-value", []filesys.HashChecksum{{}}, false},
				{"no-ExpHex", []filesys.HashChecksum{{NewHash: correctChecksums[0].NewHash}}, false},
				{"no-NewHash", []filesys.HashChecksum{{ExpHex: correctChecksums[0].ExpHex}}, false},
				{"correct+zero-value", []filesys.HashChecksum{correctChecksums[0], {}}, false},
			}
			for _, tc := range testCases {
				t.Run("cs="+tc.csName, func(t *testing.T) {
					if got := local.VerifyChecksum(filename, tc.cs...); got != tc.want {
						t.Errorf("got %t; want %t", got, tc.want)
					}
				})
			}
		})
	}
}
