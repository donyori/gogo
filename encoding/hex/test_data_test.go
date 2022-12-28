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

package hex_test

import (
	stdhex "encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type testEncodeCase struct {
	srcName  string
	dstName  string
	srcStr   string
	srcBytes []byte
	dstStr   string
	dstBytes []byte
	upper    bool
}

var testEncodeCases []*testEncodeCase
var testEncodeCasesDstMaxLen int

var testEncodeLongSrcCases [2]*testEncodeCase

func init() {
	srcs := []string{"", "Hello world! 你好，世界！", ""}
	longBytes := make([]byte, 8192)
	rand.New(rand.NewSource(10)).Read(longBytes)
	srcs[2] = string(longBytes)
	uppers := []bool{false, true}
	testEncodeCases = make([]*testEncodeCase, len(srcs)*len(uppers))
	var i int
	for _, src := range srcs {
		for _, upper := range uppers {
			dst := stdhex.EncodeToString([]byte(src))
			if upper {
				dst = strings.ToUpper(dst)
			}
			tc := &testEncodeCase{
				srcName: stringName(src),
				dstName: stringName(dst),
				srcStr:  src,
				dstStr:  dst,
				upper:   upper,
			}
			if len(src) == len(longBytes) {
				tc.srcBytes = longBytes
				var idx int
				if upper {
					idx = 1
				}
				testEncodeLongSrcCases[idx] = tc
			} else if len(src) > 0 {
				tc.srcBytes = []byte(src)
			}
			if len(dst) > 0 {
				tc.dstBytes = []byte(dst)
			}
			testEncodeCases[i] = tc
			i++
			if testEncodeCasesDstMaxLen < len(dst) {
				testEncodeCasesDstMaxLen = len(dst)
			}
		}
	}
}

// stringName returns the name of s for subtests and sub-benchmarks.
func stringName(s string) string {
	if len(s) <= 80 {
		return strconv.Quote(s)
	}
	return fmt.Sprintf(
		"<long string (%d) %s...%s>",
		len(s),
		strings.TrimSuffix(strconv.Quote(s[:4]), `"`),
		strings.TrimPrefix(strconv.Quote(s[len(s)-4:]), `"`),
	)
}
