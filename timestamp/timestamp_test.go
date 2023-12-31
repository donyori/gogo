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

package timestamp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/donyori/gogo/timestamp"
)

var testTimestampCases = []struct {
	timestamp timestamp.Timestamp
	forParse  []byte
	forFormat string
}{
	// in seconds:
	{timestamp.Timestamp(time.Unix(158051640, 202429266)), []byte("158051640.202429266"), "158051640.202429266"},
	{timestamp.Timestamp(time.Unix(1580516402, 24292660)), []byte("1580516402.02429266"), "1580516402.02429266"},
	{timestamp.Timestamp(time.Unix(15805164020, 242926600)), []byte("15805164020.2429266"), "15805164020.2429266"},
	// in milliseconds:
	{timestamp.Timestamp(time.Unix(158051640, 202429266)), []byte("158051640202.429266"), "158051640.202429266"},
	{timestamp.Timestamp(time.Unix(1580516402, 24292660)), []byte("1580516402024.29266"), "1580516402.02429266"},
	{timestamp.Timestamp(time.Unix(15805164020, 242926600)), []byte("15805164020242.9266"), "15805164020.2429266"},
	// in microseconds:
	{timestamp.Timestamp(time.Unix(158051640, 202429266)), []byte("158051640202429.266"), "158051640.202429266"},
	{timestamp.Timestamp(time.Unix(1580516402, 24292660)), []byte("1580516402024292.66"), "1580516402.02429266"},
	// in nanoseconds:
	{timestamp.Timestamp(time.Unix(158051640, 202429266)), []byte("158051640202429266"), "158051640.202429266"},
	{timestamp.Timestamp(time.Unix(1580516402, 24292660)), []byte("1580516402024292660"), "1580516402.02429266"},

	// in seconds:
	{timestamp.Timestamp(time.Unix(-31547861, -602429266)), []byte("-31547861.602429266"), "-31547861.602429266"},
	{timestamp.Timestamp(time.Unix(-315478616, -24292660)), []byte("-315478616.02429266"), "-315478616.02429266"},
	{timestamp.Timestamp(time.Unix(-3154786160, -242926600)), []byte("-3154786160.2429266"), "-3154786160.2429266"},
	{timestamp.Timestamp(time.Unix(-31547861602, -429266000)), []byte("-31547861602.429266"), "-31547861602.429266"},
	// in milliseconds:
	{timestamp.Timestamp(time.Unix(-315478616, -24292660)), []byte("-315478616024.29266"), "-315478616.02429266"},
	{timestamp.Timestamp(time.Unix(-3154786160, -242926600)), []byte("-3154786160242.9266"), "-3154786160.2429266"},
	{timestamp.Timestamp(time.Unix(-31547861602, -429266000)), []byte("-31547861602429.266"), "-31547861602.429266"},
	// in microseconds:
	{timestamp.Timestamp(time.Unix(-315478616, -24292660)), []byte("-315478616024292.66"), "-315478616.02429266"},
	{timestamp.Timestamp(time.Unix(-3154786160, -242926600)), []byte("-3154786160242926.6"), "-3154786160.2429266"},
	// in nanoseconds:
	{timestamp.Timestamp(time.Unix(-31547861, -602429266)), []byte("-31547861602429266"), "-31547861.602429266"},
	{timestamp.Timestamp(time.Unix(-315478616, -24292660)), []byte("-315478616024292660"), "-315478616.02429266"},
}

var testUnixTimestampCases = []struct {
	timestamp timestamp.UnixTimestamp
	bytes     []byte
}{
	{timestamp.UnixTimestamp(time.Unix(0, 0)), []byte("0")},
	{timestamp.UnixTimestamp(time.Unix(1580516402, 0)), []byte("1580516402")},
	{timestamp.UnixTimestamp(time.Unix(1580516402, 242926600)), []byte("1580516402.2429266")},
	{timestamp.UnixTimestamp(time.Unix(1580516402, 2429266)), []byte("1580516402.002429266")},
	{timestamp.UnixTimestamp(time.Unix(1580516402, 242926678)), []byte("1580516402.242926678")},
	{timestamp.UnixTimestamp(time.Unix(0, 1580516402242926600)), []byte("1580516402.2429266")},
	{timestamp.UnixTimestamp(time.Unix(0, 1580516402002429266)), []byte("1580516402.002429266")},
	{timestamp.UnixTimestamp(time.Unix(0, 1580516402242926678)), []byte("1580516402.242926678")},
	{timestamp.UnixTimestamp(time.Date(1960, 1, 2, 15, 3, 4, 0, time.UTC)), []byte("-315478616")},
	{timestamp.UnixTimestamp(time.Unix(-315478616, -242926600)), []byte("-315478616.2429266")},
	{timestamp.UnixTimestamp(time.Unix(-315478616, -2429266)), []byte("-315478616.002429266")},
	{timestamp.UnixTimestamp(time.Unix(-315478616, -242926678)), []byte("-315478616.242926678")},
	{timestamp.UnixTimestamp(time.Unix(0, -315478616242926600)), []byte("-315478616.2429266")},
	{timestamp.UnixTimestamp(time.Unix(0, -315478616002429266)), []byte("-315478616.002429266")},
	{timestamp.UnixTimestamp(time.Unix(0, -315478616242926678)), []byte("-315478616.242926678")},
}

var testMilliTimestampCases = []struct {
	timestamp timestamp.MilliTimestamp
	bytes     []byte
}{
	{timestamp.MilliTimestamp(time.Unix(0, 0)), []byte("0")},
	{timestamp.MilliTimestamp(time.Unix(1580516402, 0)), []byte("1580516402000")},
	{timestamp.MilliTimestamp(time.Unix(1580516402, 242926600)), []byte("1580516402242.9266")},
	{timestamp.MilliTimestamp(time.Unix(1580516402, 2429266)), []byte("1580516402002.429266")},
	{timestamp.MilliTimestamp(time.Unix(1580516402, 242926678)), []byte("1580516402242.926678")},
	{timestamp.MilliTimestamp(time.Unix(0, 1580516402242926600)), []byte("1580516402242.9266")},
	{timestamp.MilliTimestamp(time.Unix(0, 1580516402002429266)), []byte("1580516402002.429266")},
	{timestamp.MilliTimestamp(time.Unix(0, 1580516402242926678)), []byte("1580516402242.926678")},
	{timestamp.MilliTimestamp(time.Date(1960, 1, 2, 15, 3, 4, 0, time.UTC)), []byte("-315478616000")},
	{timestamp.MilliTimestamp(time.Unix(-315478616, -242926600)), []byte("-315478616242.9266")},
	{timestamp.MilliTimestamp(time.Unix(-315478616, -2429266)), []byte("-315478616002.429266")},
	{timestamp.MilliTimestamp(time.Unix(-315478616, -242926678)), []byte("-315478616242.926678")},
	{timestamp.MilliTimestamp(time.Unix(0, -315478616242926600)), []byte("-315478616242.9266")},
	{timestamp.MilliTimestamp(time.Unix(0, -315478616002429266)), []byte("-315478616002.429266")},
	{timestamp.MilliTimestamp(time.Unix(0, -315478616242926678)), []byte("-315478616242.926678")},
}

var testMicroTimestampCases = []struct {
	timestamp timestamp.MicroTimestamp
	bytes     []byte
}{
	{timestamp.MicroTimestamp(time.Unix(0, 0)), []byte("0")},
	{timestamp.MicroTimestamp(time.Unix(1580516402, 0)), []byte("1580516402000000")},
	{timestamp.MicroTimestamp(time.Unix(1580516402, 242926600)), []byte("1580516402242926.6")},
	{timestamp.MicroTimestamp(time.Unix(1580516402, 2429266)), []byte("1580516402002429.266")},
	{timestamp.MicroTimestamp(time.Unix(1580516402, 242926678)), []byte("1580516402242926.678")},
	{timestamp.MicroTimestamp(time.Unix(0, 1580516402242926600)), []byte("1580516402242926.6")},
	{timestamp.MicroTimestamp(time.Unix(0, 1580516402002429266)), []byte("1580516402002429.266")},
	{timestamp.MicroTimestamp(time.Unix(0, 1580516402242926678)), []byte("1580516402242926.678")},
	{timestamp.MicroTimestamp(time.Date(1960, 1, 2, 15, 3, 4, 0, time.UTC)), []byte("-315478616000000")},
	{timestamp.MicroTimestamp(time.Unix(-315478616, -242926600)), []byte("-315478616242926.6")},
	{timestamp.MicroTimestamp(time.Unix(-315478616, -2429266)), []byte("-315478616002429.266")},
	{timestamp.MicroTimestamp(time.Unix(-315478616, -242926678)), []byte("-315478616242926.678")},
	{timestamp.MicroTimestamp(time.Unix(0, -315478616242926600)), []byte("-315478616242926.6")},
	{timestamp.MicroTimestamp(time.Unix(0, -315478616002429266)), []byte("-315478616002429.266")},
	{timestamp.MicroTimestamp(time.Unix(0, -315478616242926678)), []byte("-315478616242926.678")},
}

var testNanoTimestampCases = []struct {
	timestamp timestamp.NanoTimestamp
	bytes     []byte
}{
	{timestamp.NanoTimestamp(time.Unix(0, 0)), []byte("0")},
	{timestamp.NanoTimestamp(time.Unix(1580516402, 0)), []byte("1580516402000000000")},
	{timestamp.NanoTimestamp(time.Unix(1580516402, 242926600)), []byte("1580516402242926600")},
	{timestamp.NanoTimestamp(time.Unix(1580516402, 2429266)), []byte("1580516402002429266")},
	{timestamp.NanoTimestamp(time.Unix(1580516402, 242926678)), []byte("1580516402242926678")},
	{timestamp.NanoTimestamp(time.Unix(0, 1580516402242926600)), []byte("1580516402242926600")},
	{timestamp.NanoTimestamp(time.Unix(0, 1580516402002429266)), []byte("1580516402002429266")},
	{timestamp.NanoTimestamp(time.Unix(0, 1580516402242926678)), []byte("1580516402242926678")},
	{timestamp.NanoTimestamp(time.Date(1960, 1, 2, 15, 3, 4, 0, time.UTC)), []byte("-315478616000000000")},
	{timestamp.NanoTimestamp(time.Unix(-315478616, -242926600)), []byte("-315478616242926600")},
	{timestamp.NanoTimestamp(time.Unix(-315478616, -2429266)), []byte("-315478616002429266")},
	{timestamp.NanoTimestamp(time.Unix(-315478616, -242926678)), []byte("-315478616242926678")},
	{timestamp.NanoTimestamp(time.Unix(0, -315478616242926600)), []byte("-315478616242926600")},
	{timestamp.NanoTimestamp(time.Unix(0, -315478616002429266)), []byte("-315478616002429266")},
	{timestamp.NanoTimestamp(time.Unix(0, -315478616242926678)), []byte("-315478616242926678")},
}

func TestTimestamp_String(t *testing.T) {
	for i, tc := range testTimestampCases {
		t.Run(
			fmt.Sprintf("case %d?forFormat=%+q", i, tc.forFormat),
			func(t *testing.T) {
				if s := tc.timestamp.String(); s != tc.forFormat {
					t.Errorf("got %s; want %s", s, tc.forFormat)
				}
			},
		)
	}
}

func TestTimestamp_MarshalJSON(t *testing.T) {
	for i, tc := range testTimestampCases {
		t.Run(
			fmt.Sprintf("case %d?forFormat=%+q", i, tc.forFormat),
			func(t *testing.T) {
				b, err := json.Marshal(tc.timestamp)
				if err != nil {
					t.Error(err)
				} else if string(b) != tc.forFormat {
					t.Errorf("got %s; want %s", b, tc.forFormat)
				}
			},
		)
	}
}

func TestTimestamp_UnmarshalJSON(t *testing.T) {
	for i, tc := range testTimestampCases {
		t.Run(
			fmt.Sprintf("case %d?forParse=%+q", i, tc.forParse),
			func(t *testing.T) {
				var ts timestamp.Timestamp
				err := json.Unmarshal(tc.forParse, &ts)
				if err != nil {
					t.Error(err)
				} else if !(time.Time)(ts).Equal(time.Time(tc.timestamp)) {
					t.Errorf("got %v; want %v", ts, tc.timestamp)
				}
			},
		)
	}
	t.Run("null", func(t *testing.T) {
		testUnmarshalJsonNull[timestamp.Timestamp](t)
	})
}

func TestUnixTimestamp_String(t *testing.T) {
	testString(t, testUnixTimestampCases)
}

func TestUnixTimestamp_MarshalJSON(t *testing.T) {
	testMarshalJson(t, testUnixTimestampCases)
}

func TestUnixTimestamp_UnmarshalJSON(t *testing.T) {
	testUnmarshalJson(t, testUnixTimestampCases)
}

func TestMilliTimestamp_String(t *testing.T) {
	testString(t, testMilliTimestampCases)
}

func TestMilliTimestamp_MarshalJSON(t *testing.T) {
	testMarshalJson(t, testMilliTimestampCases)
}

func TestMilliTimestamp_UnmarshalJSON(t *testing.T) {
	testUnmarshalJson(t, testMilliTimestampCases)
}

func TestMicroTimestamp_String(t *testing.T) {
	testString(t, testMicroTimestampCases)
}

func TestMicroTimestamp_MarshalJSON(t *testing.T) {
	testMarshalJson(t, testMicroTimestampCases)
}

func TestMicroTimestamp_UnmarshalJSON(t *testing.T) {
	testUnmarshalJson(t, testMicroTimestampCases)
}

func TestNanoTimestamp_String(t *testing.T) {
	testString(t, testNanoTimestampCases)
}

func TestNanoTimestamp_MarshalJSON(t *testing.T) {
	testMarshalJson(t, testNanoTimestampCases)
}

func TestNanoTimestamp_UnmarshalJSON(t *testing.T) {
	testUnmarshalJson(t, testNanoTimestampCases)
}

type timestampConstraint interface {
	timestamp.Timestamp | timestamp.UnixTimestamp | timestamp.MilliTimestamp |
		timestamp.MicroTimestamp | timestamp.NanoTimestamp
}

func testString[Ts timestampConstraint](t *testing.T, cases []struct {
	timestamp Ts
	bytes     []byte
}) {
	for i, tc := range cases {
		t.Run(
			fmt.Sprintf("case %d?bytes=%+q", i, tc.bytes),
			func(t *testing.T) {
				s := any(tc.timestamp).(fmt.Stringer).String()
				if s != string(tc.bytes) {
					t.Errorf("got %s; want %s", s, tc.bytes)
				}
			},
		)
	}
}

func testMarshalJson[Ts timestampConstraint](t *testing.T, cases []struct {
	timestamp Ts
	bytes     []byte
}) {
	for i, tc := range cases {
		t.Run(
			fmt.Sprintf("case %d?bytes=%+q", i, tc.bytes),
			func(t *testing.T) {
				b, err := json.Marshal(tc.timestamp)
				if err != nil {
					t.Error(err)
				} else if !bytes.Equal(b, tc.bytes) {
					t.Errorf("got %s; want %s", b, tc.bytes)
				}
			},
		)
	}
}

func testUnmarshalJson[Ts timestampConstraint](t *testing.T, cases []struct {
	timestamp Ts
	bytes     []byte
}) {
	for i, tc := range cases {
		t.Run(
			fmt.Sprintf("case %d?bytes=%+q", i, tc.bytes),
			func(t *testing.T) {
				var ts Ts
				err := json.Unmarshal(tc.bytes, &ts)
				if err != nil {
					t.Error(err)
				} else if !(time.Time)(ts).Equal(time.Time(tc.timestamp)) {
					t.Errorf("got %v; want %v", ts, tc.timestamp)
				}
			},
		)
	}
	t.Run("null", func(t *testing.T) {
		testUnmarshalJsonNull[Ts](t)
	})
}

var (
	now       = time.Now()
	nullBytes = []byte("null")
)

func testUnmarshalJsonNull[Ts timestampConstraint](t *testing.T) {
	ts := Ts(now)
	err := any(&ts).(json.Unmarshaler).UnmarshalJSON(nullBytes)
	if err != nil {
		t.Error(err)
	} else if !now.Equal(time.Time(ts)) {
		t.Error(`not no-op for "null".`)
	}
}
