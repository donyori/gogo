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
	"hash"
	"io/fs"
	"strings"
	"testing"

	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/function/compare"
)

func TestChecksum(t *testing.T) {
	for _, filename := range testFSFilenames {
		checksums := testFSChecksumMap[filename]
		newHashes := make([]func() hash.Hash, len(checksums)+2)
		wantLower := make([]string, len(newHashes))
		wantUpper := make([]string, len(newHashes))
		for i := range checksums {
			newHashes[i] = checksums[i].hash.New
			wantLower[i] = strings.ToLower(checksums[i].checksum)
			wantUpper[i] = strings.ToUpper(checksums[i].checksum)
		}
		newHashes[len(newHashes)-1] = newNilHash

		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			for _, upper := range []bool{false, true} {
				want := wantLower
				if upper {
					want = wantUpper
				}
				t.Run(fmt.Sprintf("upper=%t", upper), func(t *testing.T) {
					file, err := testFS.Open(filename)
					if err != nil {
						t.Fatal("open file -", err)
					}
					defer func(f fs.File) {
						if err := f.Close(); err != nil {
							t.Error("close file -", err)
						}
					}(file)
					got, err := filesys.Checksum(file, false, upper, newHashes...)
					if err != nil {
						t.Error("checksum -", err)
					} else if !compare.ComparableSliceEqual(got, want) {
						t.Errorf("got %v\nwant %v", got, want)
					}
				})
			}
		})
	}
}

func TestChecksumFromFS(t *testing.T) {
	for _, filename := range testFSFilenames {
		checksums := testFSChecksumMap[filename]
		newHashes := make([]func() hash.Hash, len(checksums)+2)
		wantLower := make([]string, len(newHashes))
		wantUpper := make([]string, len(newHashes))
		for i := range checksums {
			newHashes[i] = checksums[i].hash.New
			wantLower[i] = strings.ToLower(checksums[i].checksum)
			wantUpper[i] = strings.ToUpper(checksums[i].checksum)
		}
		newHashes[len(newHashes)-1] = newNilHash

		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			for _, upper := range []bool{false, true} {
				want := wantLower
				if upper {
					want = wantUpper
				}
				t.Run(fmt.Sprintf("upper=%t", upper), func(t *testing.T) {
					got, err := filesys.ChecksumFromFS(testFS, filename, upper, newHashes...)
					if err != nil {
						t.Error("checksum -", err)
					} else if !compare.ComparableSliceEqual(got, want) {
						t.Errorf("got %v\nwant %v", got, want)
					}
				})
			}
		})
	}
}

func TestNewHashVerifier(t *testing.T) {
	const PanicPrefix = "github.com/donyori/gogo/filesys.NewHashVerifier: "
	const ToBeFilledIn = "-"
	testCases := []struct {
		newHashName string
		newHash     func() hash.Hash
		prefixHex   string
		panicMsg    string
	}{
		{
			"<nil>",
			nil,
			"",
			PanicPrefix + "newHash is nil",
		},
		{
			"return-nil",
			func() hash.Hash { return nil },
			"",
			PanicPrefix + "newHash returns nil",
		},
		{
			"sha256.New",
			sha256.New,
			"",
			"",
		},
		{
			"sha256.New",
			sha256.New,
			"0123456789ABCDEFabcdef",
			"",
		},
		{
			"sha256.New",
			sha256.New,
			strings.Repeat("0", hex.EncodedLen(sha256.Size)),
			"",
		},
		{
			"sha256.New",
			sha256.New,
			strings.Repeat("0", hex.EncodedLen(sha256.Size)+2),
			"",
		},
		{
			"sha256.New",
			sha256.New,
			"12ab" + string('0'-1),
			ToBeFilledIn,
		},
		{
			"sha256.New",
			sha256.New,
			string('A'-1) + "ABC",
			ToBeFilledIn,
		},
		{
			"sha256.New",
			sha256.New,
			"0123456789ABCDEFGabcdef",
			ToBeFilledIn,
		},
		{
			"sha256.New",
			sha256.New,
			"g",
			ToBeFilledIn,
		},
	}
	for i := range testCases {
		if testCases[i].panicMsg == ToBeFilledIn {
			testCases[i].panicMsg = PanicPrefix + fmt.Sprintf(
				"prefixHex (%q) is not hexadecimal",
				testCases[i].prefixHex,
			)
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf(
			"newHash=%s&prefixHex=%q(len=%d)",
			tc.newHashName,
			tc.prefixHex,
			len(tc.prefixHex),
		), func(t *testing.T) {
			defer func() {
				e := recover()
				if tc.panicMsg != "" {
					if s, ok := e.(string); !ok || s != tc.panicMsg {
						t.Errorf("got panic %v (type: %[1]T); want %s", e, tc.panicMsg)
					}
				} else if e != nil {
					t.Error("panic -", e)
				}
			}()
			got := filesys.NewHashVerifier(tc.newHash, tc.prefixHex)
			if tc.panicMsg != "" {
				t.Error("function returned but want panic")
			} else if got == nil {
				t.Error("got nil")
			}
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	t.Run("file=<nil>", func(t *testing.T) {
		for _, tc := range verifyChecksumTestCases("") {
			t.Run("hvs="+tc.hvsName, func(t *testing.T) {
				if got := filesys.VerifyChecksum(nil, true, tc.hvs...); got {
					t.Error("got true; want false")
				}
			})
		}
	})

	for _, filename := range testFSFilenames {
		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			for _, tc := range verifyChecksumTestCases(filename) {
				t.Run("hvs="+tc.hvsName, func(t *testing.T) {
					file, err := testFS.Open(filename)
					if err != nil {
						t.Fatal("open file -", err)
					}
					defer func(f fs.File) {
						if err := f.Close(); err != nil {
							t.Error("close file -", err)
						}
					}(file)
					got := filesys.VerifyChecksum(file, false, tc.hvs...)
					if got != tc.want {
						t.Errorf("got %t; want %t", got, tc.want)
					}
				})
			}
		})
	}
}

func TestVerifyChecksumFromFS(t *testing.T) {
	nilAndNonExistTestCases := verifyChecksumTestCases("")

	t.Run(`fsys=<nil>&file=""`, func(t *testing.T) {
		for _, tc := range nilAndNonExistTestCases {
			t.Run("hvs="+tc.hvsName, func(t *testing.T) {
				if got := filesys.VerifyChecksumFromFS(nil, "", tc.hvs...); got {
					t.Error("got true; want false")
				}
			})
		}
	})

	t.Run(`file="nonexist"`, func(t *testing.T) {
		for _, tc := range nilAndNonExistTestCases {
			t.Run("hvs="+tc.hvsName, func(t *testing.T) {
				if got := filesys.VerifyChecksumFromFS(testFS, "nonexist", tc.hvs...); got {
					t.Error("got true; want false")
				}
			})
		}
	})

	for _, filename := range testFSFilenames {
		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			for _, tc := range verifyChecksumTestCases(filename) {
				t.Run("hvs="+tc.hvsName, func(t *testing.T) {
					got := filesys.VerifyChecksumFromFS(testFS, filename, tc.hvs...)
					if got != tc.want {
						t.Errorf("got %t; want %t", got, tc.want)
					}
				})
			}
		})
	}
}

func TestNonNilDeduplicatedHashVerifiers(t *testing.T) {
	hvs := make([]filesys.HashVerifier, 3)
	for i := range hvs {
		hvs[i] = filesys.NewHashVerifier(sha256.New, "")
	}

	testCases := []struct {
		hvsName         string
		hvs             []filesys.HashVerifier
		want            []filesys.HashVerifier
		equalUnderlying bool
	}{
		{
			"<nil>",
			nil,
			nil,
			false,
		},
		{
			"empty",
			[]filesys.HashVerifier{},
			nil,
			false,
		},
		{
			"0",
			[]filesys.HashVerifier{hvs[0]},
			[]filesys.HashVerifier{hvs[0]},
			true,
		},
		{
			"0+1",
			[]filesys.HashVerifier{hvs[0], hvs[1]},
			[]filesys.HashVerifier{hvs[0], hvs[1]},
			true,
		},
		{
			"0+1+2",
			[]filesys.HashVerifier{hvs[0], hvs[1], hvs[2]},
			[]filesys.HashVerifier{hvs[0], hvs[1], hvs[2]},
			true,
		},
		{
			"nil",
			[]filesys.HashVerifier{nil},
			nil,
			false,
		},
		{
			"nil+nil",
			[]filesys.HashVerifier{nil, nil},
			nil,
			false,
		},
		{
			"nil+0",
			[]filesys.HashVerifier{nil, hvs[0]},
			[]filesys.HashVerifier{hvs[0]},
			false,
		},
		{
			"0+nil",
			[]filesys.HashVerifier{hvs[0], nil},
			[]filesys.HashVerifier{hvs[0]},
			false,
		},
		{
			"0+nil+1",
			[]filesys.HashVerifier{hvs[0], nil, hvs[1]},
			[]filesys.HashVerifier{hvs[0], hvs[1]},
			false,
		},
		{
			"0+nil+nil+1+2+nil",
			[]filesys.HashVerifier{hvs[0], nil, nil, hvs[1], hvs[2], nil},
			[]filesys.HashVerifier{hvs[0], hvs[1], hvs[2]},
			false,
		},
		{
			"0+0",
			[]filesys.HashVerifier{hvs[0], hvs[0]},
			[]filesys.HashVerifier{hvs[0]},
			false,
		},
		{
			"0+0+1",
			[]filesys.HashVerifier{hvs[0], hvs[0], hvs[1]},
			[]filesys.HashVerifier{hvs[0], hvs[1]},
			false,
		},
		{
			"0+1+1",
			[]filesys.HashVerifier{hvs[0], hvs[1], hvs[1]},
			[]filesys.HashVerifier{hvs[0], hvs[1]},
			false,
		},
		{
			"0+1+1+1+2+0",
			[]filesys.HashVerifier{hvs[0], hvs[1], hvs[1], hvs[1], hvs[2], hvs[0]},
			[]filesys.HashVerifier{hvs[0], hvs[1], hvs[2]},
			false,
		},
		{
			"nil+0+1+nil+1+1+nil+2+2+0",
			[]filesys.HashVerifier{nil, hvs[0], hvs[1], nil, hvs[1], hvs[1], nil, hvs[2], hvs[2], hvs[0]},
			[]filesys.HashVerifier{hvs[0], hvs[1], hvs[2]},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run("hvs="+tc.hvsName, func(t *testing.T) {
			var input []filesys.HashVerifier
			if tc.hvs != nil {
				input = make([]filesys.HashVerifier, len(tc.hvs))
				copy(input, tc.hvs)
			}
			got := filesys.NonNilDeduplicatedHashVerifiers(input)
			if tc.want != nil {
				if !compare.AnySliceEqual(got, tc.want) {
					t.Errorf("got (len: %d) %v; want (len: %d) %v", len(got), got, len(tc.want), tc.want)
				}
			} else if got != nil {
				t.Errorf("got (len: %d) %v; want <nil>", len(got), got)
			}
			if underlyingArrayEqual(input, got) != tc.equalUnderlying {
				if tc.equalUnderlying {
					t.Error("return value and input have different underlying arrays, but want the same one")
				} else {
					t.Error("return value and input have the same underlying array, but want different")
				}
			}
			if !compare.AnySliceEqual(input, tc.hvs) || cap(input) != cap(tc.hvs) {
				t.Error("input has been modified")
			}
		})
	}
}

// newNilHash always returns a nil hash.Hash.
func newNilHash() hash.Hash {
	return nil
}

// verifyChecksumTestCases returns test cases for
// TestVerifyChecksum and TestVerifyChecksumFromFS.
func verifyChecksumTestCases(filename string) []struct {
	hvsName string
	hvs     []filesys.HashVerifier
	want    bool
} {
	checksums, ok := testFSChecksumMap[filename]
	if !ok {
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
					sha256.New,
					strings.Repeat("0", hex.EncodedLen(sha256.Size)),
				)},
				false,
			},
			{
				"empty-prefix",
				[]filesys.HashVerifier{filesys.NewHashVerifier(
					sha256.New,
					"",
				)},
				false,
			},
		}
	}

	newHash := checksums[0].hash.New
	checksum := checksums[0].checksum
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

// underlyingArrayEqual reports whether the underlying array of a
// is the same as that of b.
//
// In particular, if a or b is nil, it returns false.
func underlyingArrayEqual(a, b []filesys.HashVerifier) bool {
	return a != nil && b != nil &&
		(*[0]filesys.HashVerifier)(a) == (*[0]filesys.HashVerifier)(b)
	// Before Go 1.17, can use:
	//  (*reflect.SliceHeader)(unsafe.Pointer(&a)).Data == (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data
}
