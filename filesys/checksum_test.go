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

package filesys_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/donyori/gogo/filesys"
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

	for _, filename := range testFSFilenames {
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			testCases := []struct {
				csName string
				cs     []filesys.HashChecksum
				want   int8
			}{
				{"<nil>", nil, wantReturnTrue},
				{"correct", testFSChecksumMap[filename], wantReturnTrue},
				{"wrong", []filesys.HashChecksum{wrongChecksum}, wantReturnFalse},
				{"correct+wrong", []filesys.HashChecksum{testFSChecksumMap[filename][0], wrongChecksum}, wantReturnFalse},
				{"zero-value", []filesys.HashChecksum{{}}, wantPanic},
				{"no-WantHex", []filesys.HashChecksum{{NewHash: testFSChecksumMap[filename][0].NewHash}}, wantPanic},
				{"no-NewHash", []filesys.HashChecksum{{WantHex: testFSChecksumMap[filename][0].WantHex}}, wantPanic},
				{"correct+zero-value", []filesys.HashChecksum{testFSChecksumMap[filename][0], {}}, wantPanic},
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
									if strings.HasPrefix(s, "github.com/donyori/gogo/filesys.VerifyChecksum: ") {
										return
									}
								}
							}
							t.Error("panic:", e)
						}
					}()
					file, err := testFS.Open(filename)
					if err != nil {
						t.Fatal("open file -", err)
					}
					defer func(f fs.File) {
						if err := f.Close(); err != nil {
							t.Error("close file -", err)
						}
					}(file)
					got := filesys.VerifyChecksum(file, false, tc.cs...)
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

func TestVerifyChecksumFromFS(t *testing.T) {
	const (
		wantReturnTrue int8 = iota
		wantReturnFalse
		wantPanic
	)

	wrongChecksum := filesys.HashChecksum{
		NewHash: sha256.New,
		WantHex: strings.Repeat("0", hex.EncodedLen(sha256.Size)),
	}

	t.Run("fsys=nil", func(t *testing.T) {
		defer func() {
			if e := recover(); e != nil {
				if s, ok := e.(string); ok {
					if s == "github.com/donyori/gogo/filesys.VerifyChecksumFromFS: fsys is nil" {
						return
					}
				}
				t.Error("panic:", e)
			}
		}()
		got := filesys.VerifyChecksumFromFS(nil, "")
		t.Error("should panic but got", got)
	})

	t.Run(`file="nonexist"&cs=<nil>`, func(t *testing.T) {
		if got := filesys.VerifyChecksumFromFS(testFS, "nonexist"); got {
			t.Error("got true; want false")
		}
	})

	t.Run(`file="nonexist"&cs=zero-value`, func(t *testing.T) {
		defer func() {
			if e := recover(); e != nil {
				if s, ok := e.(string); ok {
					if strings.HasPrefix(s, "github.com/donyori/gogo/filesys.VerifyChecksumFromFS: ") {
						return
					}
				}
				t.Error("panic:", e)
			}
		}()
		got := filesys.VerifyChecksumFromFS(testFS, "nonexist", filesys.HashChecksum{})
		t.Error("should panic but got", got)
	})

	for _, filename := range testFSFilenames {
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			testCases := []struct {
				csName string
				cs     []filesys.HashChecksum
				want   int8
			}{
				{"<nil>", nil, wantReturnTrue},
				{"correct", testFSChecksumMap[filename], wantReturnTrue},
				{"wrong", []filesys.HashChecksum{wrongChecksum}, wantReturnFalse},
				{"correct+wrong", []filesys.HashChecksum{testFSChecksumMap[filename][0], wrongChecksum}, wantReturnFalse},
				{"zero-value", []filesys.HashChecksum{{}}, wantPanic},
				{"no-WantHex", []filesys.HashChecksum{{NewHash: testFSChecksumMap[filename][0].NewHash}}, wantPanic},
				{"no-NewHash", []filesys.HashChecksum{{WantHex: testFSChecksumMap[filename][0].WantHex}}, wantPanic},
				{"correct+zero-value", []filesys.HashChecksum{testFSChecksumMap[filename][0], {}}, wantPanic},
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
									if strings.HasPrefix(s, "github.com/donyori/gogo/filesys.VerifyChecksumFromFS: ") {
										return
									}
								}
							}
							t.Error("panic:", e)
						}
					}()
					got := filesys.VerifyChecksumFromFS(testFS, filename, tc.cs...)
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
