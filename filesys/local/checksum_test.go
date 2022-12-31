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

package local_test

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/filesys/local"
)

func TestVerifyChecksum(t *testing.T) {
	const (
		wantReturnTrue int8 = iota
		wantReturnFalse
		wantPanic
	)

	wrongChecksum := filesys.HashChecksum{
		NewHash: sha256.New,
		WantHex: strings.Repeat("0", hex.EncodedLen(sha256.Size)),
	}
	nonExistFilename := filepath.Join(testDataDir, "nonexist")

	t.Run(`file="nonexist"&cs=<nil>`, func(t *testing.T) {
		if got := local.VerifyChecksum(nonExistFilename); got {
			t.Error("got true; want false")
		}
	})

	t.Run(`file="nonexist"&cs=zero-value`, func(t *testing.T) {
		defer func() {
			if e := recover(); e != nil {
				if s, ok := e.(string); ok {
					if strings.HasPrefix(s, "github.com/donyori/gogo/filesys/local.VerifyChecksum: ") {
						return
					}
				}
				t.Error("panic:", e)
			}
		}()
		got := local.VerifyChecksum(nonExistFilename, filesys.HashChecksum{})
		t.Error("should panic but got", got)
	})

	for _, entry := range testFileEntries {
		filename := entry.Name()
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(testDataDir, filename)
			checksums, err := lazyCalculateChecksums(name, sha256.New, md5.New)
			if err != nil {
				t.Fatal("calculate checksums -", err)
			}
			correctChecksums := []filesys.HashChecksum{
				{NewHash: sha256.New, WantHex: checksums[0]},
				{NewHash: md5.New, WantHex: checksums[1]},
			}

			testCases := []struct {
				csName string
				cs     []filesys.HashChecksum
				want   int8
			}{
				{"<nil>", nil, wantReturnTrue},
				{"correct", correctChecksums, wantReturnTrue},
				{"wrong", []filesys.HashChecksum{wrongChecksum}, wantReturnFalse},
				{"correct+wrong", []filesys.HashChecksum{correctChecksums[0], wrongChecksum}, wantReturnFalse},
				{"zero-value", []filesys.HashChecksum{{}}, wantPanic},
				{"no-WantHex", []filesys.HashChecksum{{NewHash: correctChecksums[0].NewHash}}, wantPanic},
				{"no-NewHash", []filesys.HashChecksum{{WantHex: correctChecksums[0].WantHex}}, wantPanic},
				{"correct+zero-value", []filesys.HashChecksum{correctChecksums[0], {}}, wantPanic},
			}

			for _, tc := range testCases {
				t.Run("cs="+tc.csName, func(t *testing.T) {
					var want, shouldPanic bool
					switch tc.want {
					case wantReturnTrue:
						want = true
					case wantReturnFalse:
						// Do nothing here.
					case wantPanic:
						shouldPanic = true
					default:
						// This should never happen, but will act as a safeguard for later,
						// as a default value doesn't make sense here.
						t.Fatal("unacceptable tc.want", tc.want)
					}
					defer func() {
						if e := recover(); e != nil {
							if shouldPanic {
								if s, ok := e.(string); ok {
									if strings.HasPrefix(s, "github.com/donyori/gogo/filesys/local.VerifyChecksum: ") {
										return
									}
								}
							}
							t.Error("panic:", e)
						}
					}()
					got := local.VerifyChecksum(name, tc.cs...)
					if shouldPanic {
						t.Fatal("should panic but got", got)
					}
					if got != want {
						t.Errorf("got %t; want %t", got, want)
					}
				})
			}
		})
	}
}
