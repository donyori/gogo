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

package time

import (
	"bytes"
	"encoding/json"
	"testing"
	stdtime "time"
)

var testTimestampCases = []struct {
	Timestamp Timestamp
	ForParse  []byte
	ForFormat []byte
}{
	// in seconds:
	{Timestamp(stdtime.Unix(158051640, 202429266)), []byte("158051640.202429266"), []byte("158051640.202429266")},
	{Timestamp(stdtime.Unix(1580516402, 24292660)), []byte("1580516402.02429266"), []byte("1580516402.02429266")},
	{Timestamp(stdtime.Unix(15805164020, 242926600)), []byte("15805164020.2429266"), []byte("15805164020.2429266")},
	// in milliseconds:
	{Timestamp(stdtime.Unix(158051640, 202429266)), []byte("158051640202.429266"), []byte("158051640.202429266")},
	{Timestamp(stdtime.Unix(1580516402, 24292660)), []byte("1580516402024.29266"), []byte("1580516402.02429266")},
	{Timestamp(stdtime.Unix(15805164020, 242926600)), []byte("15805164020242.9266"), []byte("15805164020.2429266")},
	// in microseconds:
	{Timestamp(stdtime.Unix(158051640, 202429266)), []byte("158051640202429.266"), []byte("158051640.202429266")},
	{Timestamp(stdtime.Unix(1580516402, 24292660)), []byte("1580516402024292.66"), []byte("1580516402.02429266")},
	// in nanoseconds:
	{Timestamp(stdtime.Unix(158051640, 202429266)), []byte("158051640202429266"), []byte("158051640.202429266")},
	{Timestamp(stdtime.Unix(1580516402, 24292660)), []byte("1580516402024292660"), []byte("1580516402.02429266")},

	// in seconds:
	{Timestamp(stdtime.Unix(-31547861, -602429266)), []byte("-31547861.602429266"), []byte("-31547861.602429266")},
	{Timestamp(stdtime.Unix(-315478616, -24292660)), []byte("-315478616.02429266"), []byte("-315478616.02429266")},
	{Timestamp(stdtime.Unix(-3154786160, -242926600)), []byte("-3154786160.2429266"), []byte("-3154786160.2429266")},
	{Timestamp(stdtime.Unix(-31547861602, -429266000)), []byte("-31547861602.429266"), []byte("-31547861602.429266")},
	// in milliseconds:
	{Timestamp(stdtime.Unix(-315478616, -24292660)), []byte("-315478616024.29266"), []byte("-315478616.02429266")},
	{Timestamp(stdtime.Unix(-3154786160, -242926600)), []byte("-3154786160242.9266"), []byte("-3154786160.2429266")},
	{Timestamp(stdtime.Unix(-31547861602, -429266000)), []byte("-31547861602429.266"), []byte("-31547861602.429266")},
	// in microseconds:
	{Timestamp(stdtime.Unix(-315478616, -24292660)), []byte("-315478616024292.66"), []byte("-315478616.02429266")},
	{Timestamp(stdtime.Unix(-3154786160, -242926600)), []byte("-3154786160242926.6"), []byte("-3154786160.2429266")},
	// in nanoseconds:
	{Timestamp(stdtime.Unix(-31547861, -602429266)), []byte("-31547861602429266"), []byte("-31547861.602429266")},
	{Timestamp(stdtime.Unix(-315478616, -24292660)), []byte("-315478616024292660"), []byte("-315478616.02429266")},
}

var testUnixTimestampCases = []struct {
	Timestamp UnixTimestamp
	Bytes     []byte
}{
	{UnixTimestamp(stdtime.Unix(0, 0)), []byte("0")},
	{UnixTimestamp(stdtime.Unix(1580516402, 0)), []byte("1580516402")},
	{UnixTimestamp(stdtime.Unix(1580516402, 242926600)), []byte("1580516402.2429266")},
	{UnixTimestamp(stdtime.Unix(1580516402, 2429266)), []byte("1580516402.002429266")},
	{UnixTimestamp(stdtime.Unix(1580516402, 242926678)), []byte("1580516402.242926678")},
	{UnixTimestamp(stdtime.Unix(0, 1580516402242926600)), []byte("1580516402.2429266")},
	{UnixTimestamp(stdtime.Unix(0, 1580516402002429266)), []byte("1580516402.002429266")},
	{UnixTimestamp(stdtime.Unix(0, 1580516402242926678)), []byte("1580516402.242926678")},
	{UnixTimestamp(stdtime.Date(1960, 1, 2, 15, 3, 4, 0, stdtime.UTC)), []byte("-315478616")},
	{UnixTimestamp(stdtime.Unix(-315478616, -242926600)), []byte("-315478616.2429266")},
	{UnixTimestamp(stdtime.Unix(-315478616, -2429266)), []byte("-315478616.002429266")},
	{UnixTimestamp(stdtime.Unix(-315478616, -242926678)), []byte("-315478616.242926678")},
	{UnixTimestamp(stdtime.Unix(0, -315478616242926600)), []byte("-315478616.2429266")},
	{UnixTimestamp(stdtime.Unix(0, -315478616002429266)), []byte("-315478616.002429266")},
	{UnixTimestamp(stdtime.Unix(0, -315478616242926678)), []byte("-315478616.242926678")},
}

var testMilliTimestampCases = []struct {
	Timestamp MilliTimestamp
	Bytes     []byte
}{
	{MilliTimestamp(stdtime.Unix(0, 0)), []byte("0")},
	{MilliTimestamp(stdtime.Unix(1580516402, 0)), []byte("1580516402000")},
	{MilliTimestamp(stdtime.Unix(1580516402, 242926600)), []byte("1580516402242.9266")},
	{MilliTimestamp(stdtime.Unix(1580516402, 2429266)), []byte("1580516402002.429266")},
	{MilliTimestamp(stdtime.Unix(1580516402, 242926678)), []byte("1580516402242.926678")},
	{MilliTimestamp(stdtime.Unix(0, 1580516402242926600)), []byte("1580516402242.9266")},
	{MilliTimestamp(stdtime.Unix(0, 1580516402002429266)), []byte("1580516402002.429266")},
	{MilliTimestamp(stdtime.Unix(0, 1580516402242926678)), []byte("1580516402242.926678")},
	{MilliTimestamp(stdtime.Date(1960, 1, 2, 15, 3, 4, 0, stdtime.UTC)), []byte("-315478616000")},
	{MilliTimestamp(stdtime.Unix(-315478616, -242926600)), []byte("-315478616242.9266")},
	{MilliTimestamp(stdtime.Unix(-315478616, -2429266)), []byte("-315478616002.429266")},
	{MilliTimestamp(stdtime.Unix(-315478616, -242926678)), []byte("-315478616242.926678")},
	{MilliTimestamp(stdtime.Unix(0, -315478616242926600)), []byte("-315478616242.9266")},
	{MilliTimestamp(stdtime.Unix(0, -315478616002429266)), []byte("-315478616002.429266")},
	{MilliTimestamp(stdtime.Unix(0, -315478616242926678)), []byte("-315478616242.926678")},
}

var testMicroTimestampCases = []struct {
	Timestamp MicroTimestamp
	Bytes     []byte
}{
	{MicroTimestamp(stdtime.Unix(0, 0)), []byte("0")},
	{MicroTimestamp(stdtime.Unix(1580516402, 0)), []byte("1580516402000000")},
	{MicroTimestamp(stdtime.Unix(1580516402, 242926600)), []byte("1580516402242926.6")},
	{MicroTimestamp(stdtime.Unix(1580516402, 2429266)), []byte("1580516402002429.266")},
	{MicroTimestamp(stdtime.Unix(1580516402, 242926678)), []byte("1580516402242926.678")},
	{MicroTimestamp(stdtime.Unix(0, 1580516402242926600)), []byte("1580516402242926.6")},
	{MicroTimestamp(stdtime.Unix(0, 1580516402002429266)), []byte("1580516402002429.266")},
	{MicroTimestamp(stdtime.Unix(0, 1580516402242926678)), []byte("1580516402242926.678")},
	{MicroTimestamp(stdtime.Date(1960, 1, 2, 15, 3, 4, 0, stdtime.UTC)), []byte("-315478616000000")},
	{MicroTimestamp(stdtime.Unix(-315478616, -242926600)), []byte("-315478616242926.6")},
	{MicroTimestamp(stdtime.Unix(-315478616, -2429266)), []byte("-315478616002429.266")},
	{MicroTimestamp(stdtime.Unix(-315478616, -242926678)), []byte("-315478616242926.678")},
	{MicroTimestamp(stdtime.Unix(0, -315478616242926600)), []byte("-315478616242926.6")},
	{MicroTimestamp(stdtime.Unix(0, -315478616002429266)), []byte("-315478616002429.266")},
	{MicroTimestamp(stdtime.Unix(0, -315478616242926678)), []byte("-315478616242926.678")},
}

var testNanoTimestampCases = []struct {
	Timestamp NanoTimestamp
	Bytes     []byte
}{
	{NanoTimestamp(stdtime.Unix(0, 0)), []byte("0")},
	{NanoTimestamp(stdtime.Unix(1580516402, 0)), []byte("1580516402000000000")},
	{NanoTimestamp(stdtime.Unix(1580516402, 242926600)), []byte("1580516402242926600")},
	{NanoTimestamp(stdtime.Unix(1580516402, 2429266)), []byte("1580516402002429266")},
	{NanoTimestamp(stdtime.Unix(1580516402, 242926678)), []byte("1580516402242926678")},
	{NanoTimestamp(stdtime.Unix(0, 1580516402242926600)), []byte("1580516402242926600")},
	{NanoTimestamp(stdtime.Unix(0, 1580516402002429266)), []byte("1580516402002429266")},
	{NanoTimestamp(stdtime.Unix(0, 1580516402242926678)), []byte("1580516402242926678")},
	{NanoTimestamp(stdtime.Date(1960, 1, 2, 15, 3, 4, 0, stdtime.UTC)), []byte("-315478616000000000")},
	{NanoTimestamp(stdtime.Unix(-315478616, -242926600)), []byte("-315478616242926600")},
	{NanoTimestamp(stdtime.Unix(-315478616, -2429266)), []byte("-315478616002429266")},
	{NanoTimestamp(stdtime.Unix(-315478616, -242926678)), []byte("-315478616242926678")},
	{NanoTimestamp(stdtime.Unix(0, -315478616242926600)), []byte("-315478616242926600")},
	{NanoTimestamp(stdtime.Unix(0, -315478616002429266)), []byte("-315478616002429266")},
	{NanoTimestamp(stdtime.Unix(0, -315478616242926678)), []byte("-315478616242926678")},
}

func TestTimestamp_String(t *testing.T) {
	for i, c := range testTimestampCases {
		s := c.Timestamp.String()
		if s != string(c.ForFormat) {
			t.Errorf("Case %d: s: %s != %s.", i+1, s, c.ForFormat)
		}
	}
}

func TestTimestamp_MarshalJSON(t *testing.T) {
	for i, c := range testTimestampCases {
		b, err := json.Marshal(c.Timestamp)
		if err != nil {
			t.Errorf("Case %d: %v.", i+1, err)
		} else if !bytes.Equal(b, c.ForFormat) {
			t.Errorf("Case %d: b: %s != %s.", i+1, b, c.ForFormat)
		}
	}
}

func TestTimestamp_UnmarshalJSON(t *testing.T) {
	for i, c := range testTimestampCases {
		var ts Timestamp
		err := json.Unmarshal(c.ForParse, &ts)
		if err != nil {
			t.Errorf("Case %d: %v.", i+1, err)
		} else if !(Time)(ts).Equal(Time(c.Timestamp)) {
			t.Errorf("Case %d: ts: %v != %v.", i+1, ts, c.Timestamp)
		}
	}
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
		} else if !(Time)(ut).Equal(Time(c.Timestamp)) {
			t.Errorf("Case %d: ut: %v != %v.", i+1, ut, c.Timestamp)
		}
	}
}

func TestMilliTimestamp_String(t *testing.T) {
	for i, c := range testMilliTimestampCases {
		s := c.Timestamp.String()
		if s != string(c.Bytes) {
			t.Errorf("Case %d: s: %s != %s.", i+1, s, c.Bytes)
		}
	}
}

func TestMilliTimestamp_MarshalJSON(t *testing.T) {
	for i, c := range testMilliTimestampCases {
		b, err := json.Marshal(c.Timestamp)
		if err != nil {
			t.Errorf("Case %d: %v.", i+1, err)
		} else if !bytes.Equal(b, c.Bytes) {
			t.Errorf("Case %d: b: %s != %s.", i+1, b, c.Bytes)
		}
	}
}

func TestMilliTimestamp_UnmarshalJSON(t *testing.T) {
	for i, c := range testMilliTimestampCases {
		var mt MilliTimestamp
		err := json.Unmarshal(c.Bytes, &mt)
		if err != nil {
			t.Errorf("Case %d: %v.", i+1, err)
		} else if !(Time)(mt).Equal(Time(c.Timestamp)) {
			t.Errorf("Case %d: mt: %v != %v.", i+1, mt, c.Timestamp)
		}
	}
}

func TestMicroTimestamp_String(t *testing.T) {
	for i, c := range testMicroTimestampCases {
		s := c.Timestamp.String()
		if s != string(c.Bytes) {
			t.Errorf("Case %d: s: %s != %s.", i+1, s, c.Bytes)
		}
	}
}

func TestMicroTimestamp_MarshalJSON(t *testing.T) {
	for i, c := range testMicroTimestampCases {
		b, err := json.Marshal(c.Timestamp)
		if err != nil {
			t.Errorf("Case %d: %v.", i+1, err)
		} else if !bytes.Equal(b, c.Bytes) {
			t.Errorf("Case %d: b: %s != %s.", i+1, b, c.Bytes)
		}
	}
}

func TestMicroTimestamp_UnmarshalJSON(t *testing.T) {
	for i, c := range testMicroTimestampCases {
		var ut MicroTimestamp
		err := json.Unmarshal(c.Bytes, &ut)
		if err != nil {
			t.Errorf("Case %d: %v.", i+1, err)
		} else if !(Time)(ut).Equal(Time(c.Timestamp)) {
			t.Errorf("Case %d: ut: %v != %v.", i+1, ut, c.Timestamp)
		}
	}
}

func TestNanoTimestamp_String(t *testing.T) {
	for i, c := range testNanoTimestampCases {
		s := c.Timestamp.String()
		if s != string(c.Bytes) {
			t.Errorf("Case %d: s: %s != %s.", i+1, s, c.Bytes)
		}
	}
}

func TestNanoTimestamp_MarshalJSON(t *testing.T) {
	for i, c := range testNanoTimestampCases {
		b, err := json.Marshal(c.Timestamp)
		if err != nil {
			t.Errorf("Case %d: %v.", i+1, err)
		} else if !bytes.Equal(b, c.Bytes) {
			t.Errorf("Case %d: b: %s != %s.", i+1, b, c.Bytes)
		}
	}
}

func TestNanoTimestamp_UnmarshalJSON(t *testing.T) {
	for i, c := range testNanoTimestampCases {
		var nt NanoTimestamp
		err := json.Unmarshal(c.Bytes, &nt)
		if err != nil {
			t.Errorf("Case %d: %v.", i+1, err)
		} else if !(Time)(nt).Equal(Time(c.Timestamp)) {
			t.Errorf("Case %d: nt: %v != %v.", i+1, nt, c.Timestamp)
		}
	}
}

func TestUnmarshalJSON_Null(t *testing.T) {
	now := stdtime.Now()
	b := []byte("null")

	ts := Timestamp(now)
	err := ts.UnmarshalJSON(b)
	if err != nil {
		t.Errorf("type: %T, err: %v.", ts, err)
	} else if !now.Equal(Time(ts)) {
		t.Errorf(`type: %T, not no-op for "null".`, ts)
	}

	unixTs := UnixTimestamp(now)
	err = unixTs.UnmarshalJSON(b)
	if err != nil {
		t.Errorf("type: %T, err: %v.", unixTs, err)
	} else if !now.Equal(Time(unixTs)) {
		t.Errorf(`type: %T, not no-op for "null".`, unixTs)
	}

	mts := MilliTimestamp(now)
	err = mts.UnmarshalJSON(b)
	if err != nil {
		t.Errorf("type: %T, err: %v.", mts, err)
	} else if !now.Equal(Time(mts)) {
		t.Errorf(`type: %T, not no-op for "null".`, mts)
	}

	uts := MicroTimestamp(now)
	err = uts.UnmarshalJSON(b)
	if err != nil {
		t.Errorf("type: %T, err: %v.", uts, err)
	} else if !now.Equal(Time(uts)) {
		t.Errorf(`type: %T, not no-op for "null".`, uts)
	}

	nts := NanoTimestamp(now)
	err = nts.UnmarshalJSON(b)
	if err != nil {
		t.Errorf("type: %T, err: %v.", nts, err)
	} else if !now.Equal(Time(nts)) {
		t.Errorf(`type: %T, not no-op for "null".`, nts)
	}
}
