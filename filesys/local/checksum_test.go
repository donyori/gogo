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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/filesys/local"
)

func TestVerifyChecksum(t *testing.T) {
	nonExistFilename := filepath.Join(testDataDir, "nonexist")

	t.Run(`file="nonexist"`, func(t *testing.T) {
		testCases := verifyChecksumTestCases(t, nonExistFilename)
		if t.Failed() {
			return
		}
		for _, tc := range testCases {
			t.Run("hvs="+tc.hvsName, func(t *testing.T) {
				if got := local.VerifyChecksum(nonExistFilename, tc.hvs...); got {
					t.Error("got true; want false")
				}
			})
		}
	})

	for _, entry := range testFileEntries {
		filename := entry.Name()
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(testDataDir, filename)
			testCases := verifyChecksumTestCases(t, name)
			if t.Failed() {
				return
			}
			for _, tc := range testCases {
				t.Run("hvs="+tc.hvsName, func(t *testing.T) {
					got := local.VerifyChecksum(name, tc.hvs...)
					if got != tc.want {
						t.Errorf("got %t; want %t", got, tc.want)
					}
				})
			}
		})
	}
}

// verifyChecksumTestCases returns test cases for TestVerifyChecksum.
func verifyChecksumTestCases(t *testing.T, name string) []struct {
	hvsName string
	hvs     []filesys.HashVerifier
	want    bool
} {
	newHash := sha256.New
	checksums, err := lazyCalculateChecksums(name, newHash)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []struct {
				hvsName string
				hvs     []filesys.HashVerifier
				want    bool
			}{
				{
					"<nil>",
					nil,
					false,
				},
				{
					"nil-HashVerifier",
					[]filesys.HashVerifier{nil},
					false,
				},
				{
					"all-0-prefix",
					[]filesys.HashVerifier{filesys.NewHashVerifier(
						newHash,
						strings.Repeat("0", hex.EncodedLen(sha256.Size)),
					)},
					false,
				},
				{
					"empty-prefix",
					[]filesys.HashVerifier{filesys.NewHashVerifier(
						newHash,
						"",
					)},
					false,
				},
			}
		}
		t.Error("calculate checksums -", err)
		return nil
	}

	checksum := checksums[0]
	wrongChecksum := strings.Repeat("0", len(checksum))
	if wrongChecksum == checksum {
		wrongChecksum = wrongChecksum[:len(wrongChecksum)-1] + "1"
	}
	prefixVerifiers := make([]filesys.HashVerifier, len(checksum)+1)
	for i := range prefixVerifiers {
		prefixVerifiers[i] = filesys.NewHashVerifier(newHash, checksum[:i])
	}
	duplicateVerifier1 := filesys.NewHashVerifier(newHash, checksum)
	duplicateVerifier2 := filesys.NewHashVerifier(newHash, checksum)

	return []struct {
		hvsName string
		hvs     []filesys.HashVerifier
		want    bool
	}{
		{
			"<nil>",
			nil,
			true,
		},
		{
			"correct",
			[]filesys.HashVerifier{
				filesys.NewHashVerifier(newHash, checksum),
			},
			true,
		},
		{
			"wrong",
			[]filesys.HashVerifier{
				filesys.NewHashVerifier(newHash, wrongChecksum),
			},
			false,
		},
		{
			"correct+wrong",
			[]filesys.HashVerifier{
				filesys.NewHashVerifier(newHash, checksum),
				filesys.NewHashVerifier(newHash, wrongChecksum),
			},
			false,
		},
		{
			"nil-HashVerifier",
			[]filesys.HashVerifier{nil},
			true,
		},
		{
			"nil+correct",
			[]filesys.HashVerifier{
				nil,
				filesys.NewHashVerifier(newHash, checksum),
			},
			true,
		},
		{
			"nil+wrong",
			[]filesys.HashVerifier{
				nil,
				filesys.NewHashVerifier(newHash, wrongChecksum),
			},
			false,
		},
		{
			"correct+nil+nil+wrong+nil",
			[]filesys.HashVerifier{
				filesys.NewHashVerifier(newHash, checksum),
				nil,
				nil,
				filesys.NewHashVerifier(newHash, wrongChecksum),
				nil,
			},
			false,
		},
		{
			"all-prefixes",
			prefixVerifiers,
			true,
		},
		{
			"too-long-prefix",
			[]filesys.HashVerifier{
				filesys.NewHashVerifier(newHash, checksum+"00"),
			},
			false,
		},
		{
			"duplicate-correct",
			[]filesys.HashVerifier{
				duplicateVerifier1,
				duplicateVerifier1,
			},
			true,
		},
		{
			"duplicate-correct+nil",
			[]filesys.HashVerifier{
				duplicateVerifier2,
				duplicateVerifier2,
				duplicateVerifier2,
				nil,
			},
			true,
		},
	}
}
