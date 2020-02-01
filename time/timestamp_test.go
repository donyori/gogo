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

package time

import (
	"bytes"
	"encoding/json"
	"testing"
	stdtime "time"
)

var testUnixTimestampCases = []struct {
	Timestamp UnixTimestamp
	Bytes     []byte
}{
	{UnixTimestamp(stdtime.Unix(0, 0)), []byte("0")},
	{UnixTimestamp(stdtime.Unix(1580516402, 0)), []byte("1580516402")},
	{UnixTimestamp(stdtime.Unix(1580516402, 242926600)), []byte("1580516402.2429266")},
	{UnixTimestamp(stdtime.Unix(1580516402, 242926678)), []byte("1580516402.242926678")},
	{UnixTimestamp(stdtime.Unix(0, 1580516402242926600)), []byte("1580516402.2429266")},
	{UnixTimestamp(stdtime.Unix(0, 1580516402242926678)), []byte("1580516402.242926678")},
	{UnixTimestamp(stdtime.Date(1960, 1, 2, 15, 3, 4, 0, stdtime.UTC)), []byte("-315478616")},
	{UnixTimestamp(stdtime.Unix(-315478616, -242926600)), []byte("-315478616.2429266")},
	{UnixTimestamp(stdtime.Unix(-315478616, -242926678)), []byte("-315478616.242926678")},
	{UnixTimestamp(stdtime.Unix(0, -315478616242926600)), []byte("-315478616.2429266")},
	{UnixTimestamp(stdtime.Unix(0, -315478616242926678)), []byte("-315478616.242926678")},
}

func TestUnixTimestamp_String(t *testing.T) {
	for i, c := range testUnixTimestampCases {
		s := c.Timestamp.String()
		if s != string(c.Bytes) {
			t.Errorf("Case %d: s: %s != %s.", i+1, s, c.Bytes)
		}
	}
}

func TestUnixTimestamp_MarshalJSON(t *testing.T) {
	for i, c := range testUnixTimestampCases {
		b, err := json.Marshal(c.Timestamp)
		if err != nil {
			t.Errorf("Case %d: %v.", i+1, err)
		} else if !bytes.Equal(b, c.Bytes) {
			t.Errorf("Case %d: b: %s != %s.", i+1, b, c.Bytes)
		}
	}
}

func TestUnixTimestamp_UnmarshalJSON(t *testing.T) {
	for i, c := range testUnixTimestampCases {
		var ut UnixTimestamp
		err := json.Unmarshal(c.Bytes, &ut)
		if err != nil {
			t.Errorf("Case %d: %v.", i+1, err)
		} else if !(stdtime.Time)(ut).Equal(stdtime.Time(c.Timestamp)) {
			t.Errorf("Case %d: ut: %v != %v.", i+1, ut, c.Timestamp)
		}
	}
}
