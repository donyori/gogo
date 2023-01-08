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

package filesys

import (
	"crypto/sha256"
	"testing"

	"github.com/donyori/gogo/function/compare"
)

func TestNonNilDeduplicatedHashVerifiers(t *testing.T) {
	hvs := make([]HashVerifier, 3)
	for i := range hvs {
		hvs[i] = NewHashVerifier(sha256.New, "")
	}

	testCases := []struct {
		hvsName         string
		hvs             []HashVerifier
		want            []HashVerifier
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
			[]HashVerifier{},
			nil,
			false,
		},
		{
			"0",
			[]HashVerifier{hvs[0]},
			[]HashVerifier{hvs[0]},
			true,
		},
		{
			"0+1",
			[]HashVerifier{hvs[0], hvs[1]},
			[]HashVerifier{hvs[0], hvs[1]},
			true,
		},
		{
			"0+1+2",
			[]HashVerifier{hvs[0], hvs[1], hvs[2]},
			[]HashVerifier{hvs[0], hvs[1], hvs[2]},
			true,
		},
		{
			"nil",
			[]HashVerifier{nil},
			nil,
			false,
		},
		{
			"nil+nil",
			[]HashVerifier{nil, nil},
			nil,
			false,
		},
		{
			"nil+0",
			[]HashVerifier{nil, hvs[0]},
			[]HashVerifier{hvs[0]},
			false,
		},
		{
			"0+nil",
			[]HashVerifier{hvs[0], nil},
			[]HashVerifier{hvs[0]},
			false,
		},
		{
			"0+nil+1",
			[]HashVerifier{hvs[0], nil, hvs[1]},
			[]HashVerifier{hvs[0], hvs[1]},
			false,
		},
		{
			"0+nil+nil+1+2+nil",
			[]HashVerifier{hvs[0], nil, nil, hvs[1], hvs[2], nil},
			[]HashVerifier{hvs[0], hvs[1], hvs[2]},
			false,
		},
		{
			"0+0",
			[]HashVerifier{hvs[0], hvs[0]},
			[]HashVerifier{hvs[0]},
			false,
		},
		{
			"0+0+1",
			[]HashVerifier{hvs[0], hvs[0], hvs[1]},
			[]HashVerifier{hvs[0], hvs[1]},
			false,
		},
		{
			"0+1+1",
			[]HashVerifier{hvs[0], hvs[1], hvs[1]},
			[]HashVerifier{hvs[0], hvs[1]},
			false,
		},
		{
			"0+1+1+1+2+0",
			[]HashVerifier{hvs[0], hvs[1], hvs[1], hvs[1], hvs[2], hvs[0]},
			[]HashVerifier{hvs[0], hvs[1], hvs[2]},
			false,
		},
		{
			"nil+0+1+nil+1+1+nil+2+2+0",
			[]HashVerifier{nil, hvs[0], hvs[1], nil, hvs[1], hvs[1], nil, hvs[2], hvs[2], hvs[0]},
			[]HashVerifier{hvs[0], hvs[1], hvs[2]},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run("hvs="+tc.hvsName, func(t *testing.T) {
			var input []HashVerifier
			if tc.hvs != nil {
				input = make([]HashVerifier, len(tc.hvs))
				copy(input, tc.hvs)
			}
			got := nonNilDeduplicatedHashVerifiers(input)
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

// underlyingArrayEqual reports whether the underlying array of a
// is the same as that of b.
//
// In particular, if a or b is nil, it returns false.
func underlyingArrayEqual(a, b []HashVerifier) bool {
	return a != nil && b != nil && (*[0]HashVerifier)(a) == (*[0]HashVerifier)(b)
	// Before Go 1.17, can use:
	//  (*reflect.SliceHeader)(unsafe.Pointer(&a)).Data == (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data
}
