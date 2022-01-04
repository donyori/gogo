// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

package agpl3

import (
	"strings"
	"testing"
)

func TestRespShowWC(t *testing.T) {
	cases := []struct {
		input  []string
		doResp bool
		output string
	}{
		{[]string{}, false, ""},
		{[]string{"show a"}, false, ""},
		{[]string{"show", "a"}, false, ""},
		{[]string{"show w"}, true, "  " + DisclaimerOfWarranty + "\n"},
		{[]string{"show", "w"}, true, "  " + DisclaimerOfWarranty + "\n"},
		{[]string{"show c"}, true, "  " + TermsAndConditions + "\n"},
		{[]string{"show", "c"}, true, "  " + TermsAndConditions + "\n"},
		{[]string{"show", "w", "a"}, false, ""},
	}
	w := new(strings.Builder)
	for _, c := range cases {
		doResp, err := RespShowWC(w, c.input)
		if err != nil {
			t.Errorf("Error: %v, input: %#v.", err, c.input)
		}
		if doResp != c.doResp {
			t.Errorf("doResp: %t != %t, input: %#v.", doResp, c.doResp, c.input)
		}
		if s := w.String(); s != c.output {
			t.Errorf("Output: %q != %q, input: %#v.", s, c.output, c.input)
		}
		w.Reset()
	}
}
