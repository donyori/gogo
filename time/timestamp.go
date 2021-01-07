// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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
	"regexp"
	"strconv"
	stdtime "time"

	"github.com/donyori/gogo/errors"
)

const (
	// Unix timestamp pattern. (without escape to fit more languages)
	UnixTimestampPattern = "^[+-]?[0-9]+([.][0-9]{1,9})?$"
	// Millisecond timestamp pattern. (without escape to fit more languages)
	MilliTimestampPattern = "^[+-]?[0-9]+([.][0-9]{1,6})?$"
	// Microsecond timestamp pattern. (without escape to fit more languages)
	MicroTimestampPattern = "^[+-]?[0-9]+([.][0-9]{1,3})?$"
	// Nanosecond timestamp pattern. (without escape to fit more languages)
	NanoTimestampPattern = "^[+-]?[0-9]+$"
)

// Timestamp.
// Auto detect the time unit (seconds, milliseconds, microseconds, or nanoseconds)
// when parse from string.
// Treated as UnixTimestamp when format string.
//
// It determines the time unit by the integer part digits, as follows:
// less than 12-digits: in seconds,
// 12-digits to 14-digits: in milliseconds,
// 15-digits or 16-digits: in microseconds,
// more than 16-digits: in nanoseconds.
type Timestamp stdtime.Time

func (ts Timestamp) String() string {
	return string(timeToTimestamp(autoTimestamp, stdtime.Time(ts)))
}

func (ts Timestamp) MarshalText() (text []byte, err error) {
	return timeToTimestamp(autoTimestamp, stdtime.Time(ts)), nil
}

func (ts *Timestamp) UnmarshalText(text []byte) error {
	t, err := timestampToTime(autoTimestamp, text)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*(*stdtime.Time)(ts) = t
	return nil
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	return timeToTimestamp(autoTimestamp, stdtime.Time(ts)), nil
}

func (ts *Timestamp) UnmarshalJSON(b []byte) error {
	t, err := timestampToTime(autoTimestamp, b)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*(*stdtime.Time)(ts) = t
	return nil
}

// Unix timestamp, in seconds.
type UnixTimestamp stdtime.Time

func (ut UnixTimestamp) String() string {
	return string(timeToTimestamp(unixTimestamp, stdtime.Time(ut)))
}

func (ut UnixTimestamp) MarshalText() (text []byte, err error) {
	return timeToTimestamp(unixTimestamp, stdtime.Time(ut)), nil
}

func (ut *UnixTimestamp) UnmarshalText(text []byte) error {
	t, err := timestampToTime(unixTimestamp, text)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*(*stdtime.Time)(ut) = t
	return nil
}

func (ut UnixTimestamp) MarshalJSON() ([]byte, error) {
	return timeToTimestamp(unixTimestamp, stdtime.Time(ut)), nil
}

func (ut *UnixTimestamp) UnmarshalJSON(b []byte) error {
	t, err := timestampToTime(unixTimestamp, b)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*(*stdtime.Time)(ut) = t
	return nil
}

// Timestamp in milliseconds.
// Often used in JavaScript.
type MilliTimestamp stdtime.Time

func (mt MilliTimestamp) String() string {
	return string(timeToTimestamp(milliTimestamp, stdtime.Time(mt)))
}

func (mt MilliTimestamp) MarshalText() (text []byte, err error) {
	return timeToTimestamp(milliTimestamp, stdtime.Time(mt)), nil
}

func (mt *MilliTimestamp) UnmarshalText(text []byte) error {
	t, err := timestampToTime(milliTimestamp, text)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*(*stdtime.Time)(mt) = t
	return nil
}

func (mt MilliTimestamp) MarshalJSON() ([]byte, error) {
	return timeToTimestamp(milliTimestamp, stdtime.Time(mt)), nil
}

func (mt *MilliTimestamp) UnmarshalJSON(b []byte) error {
	t, err := timestampToTime(milliTimestamp, b)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*(*stdtime.Time)(mt) = t
	return nil
}

// Timestamp in microseconds.
type MicroTimestamp stdtime.Time

func (ct MicroTimestamp) String() string {
	return string(timeToTimestamp(microTimestamp, stdtime.Time(ct)))
}

func (ct MicroTimestamp) MarshalText() (text []byte, err error) {
	return timeToTimestamp(microTimestamp, stdtime.Time(ct)), nil
}

func (ct *MicroTimestamp) UnmarshalText(text []byte) error {
	t, err := timestampToTime(microTimestamp, text)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*(*stdtime.Time)(ct) = t
	return nil
}

func (ct MicroTimestamp) MarshalJSON() ([]byte, error) {
	return timeToTimestamp(microTimestamp, stdtime.Time(ct)), nil
}

func (ct *MicroTimestamp) UnmarshalJSON(b []byte) error {
	t, err := timestampToTime(microTimestamp, b)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*(*stdtime.Time)(ct) = t
	return nil
}

// Timestamp in nanoseconds.
type NanoTimestamp stdtime.Time

func (nt NanoTimestamp) String() string {
	return string(timeToTimestamp(nanoTimestamp, stdtime.Time(nt)))
}

func (nt NanoTimestamp) MarshalText() (text []byte, err error) {
	return timeToTimestamp(nanoTimestamp, stdtime.Time(nt)), nil
}

func (nt *NanoTimestamp) UnmarshalText(text []byte) error {
	t, err := timestampToTime(nanoTimestamp, text)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*(*stdtime.Time)(nt) = t
	return nil
}

func (nt NanoTimestamp) MarshalJSON() ([]byte, error) {
	return timeToTimestamp(nanoTimestamp, stdtime.Time(nt)), nil
}

func (nt *NanoTimestamp) UnmarshalJSON(b []byte) error {
	t, err := timestampToTime(nanoTimestamp, b)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*(*stdtime.Time)(nt) = t
	return nil
}

type timestampType int8

const (
	autoTimestamp timestampType = iota
	unixTimestamp
	milliTimestamp
	microTimestamp
	nanoTimestamp
)

var (
	unixTimestampRegExpr  = regexp.MustCompile(UnixTimestampPattern)
	milliTimestampRegExpr = regexp.MustCompile(MilliTimestampPattern)
	microTimestampRegExpr = regexp.MustCompile(MicroTimestampPattern)
	nanoTimestampRegExpr  = regexp.MustCompile(NanoTimestampPattern)
)

var (
	timestampRegExprMapping = []*regexp.Regexp{
		unixTimestampRegExpr,
		unixTimestampRegExpr,
		milliTimestampRegExpr,
		microTimestampRegExpr,
		nanoTimestampRegExpr,
	}
	// Reflect the float point position relative to nanoseconds.
	timestampFloatShiftMapping    = []int{1e9, 1e9, 1e6, 1e3, 1}
	timestampFractionalLenMapping = []int{-1, 9, 6, 3, 0}
)

func timestampToTime(tsType timestampType, ts []byte) (t stdtime.Time, err error) {
	if len(ts) == 0 {
		return stdtime.Time{}, errors.New("empty timestamp")
	}
	// Check tsType if this function is exported.
	if ok := timestampRegExprMapping[tsType].Match(ts); !ok {
		return stdtime.Time{}, errors.New("invalid timestamp")
	}
	pointIdx := bytes.IndexRune(ts, '.')
	tst := tsType
	if tsType == autoTimestamp {
		var k int // length of integer part
		if pointIdx < 0 {
			k = len(ts)
		} else {
			k = pointIdx
		}
		if ts[0] == '-' || ts[0] == '+' {
			k--
		}
		if k < 12 {
			tst = unixTimestamp
		} else if k < 15 {
			tst = milliTimestamp
		} else if k < 17 {
			tst = microTimestamp
		} else {
			tst = nanoTimestamp
		}
		if pointIdx >= 0 && len(ts)-pointIdx-1 > timestampFractionalLenMapping[tst] {
			return stdtime.Time{}, errors.New("invalid timestamp")
		}
	}
	var s, ns []byte
	if pointIdx < 0 {
		s = ts
	} else {
		s = ts[:pointIdx]
		ns = ts[pointIdx+1:]
	}
	var sec, nsec int64
	sec, err = strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		return
	}
	if ns != nil {
		nsec, err = strconv.ParseInt(string(ns), 10, 64)
		if err != nil {
			return
		}
		for i, end := len(ns), timestampFractionalLenMapping[tst]; i < end; i++ {
			nsec *= 10
		}
		if sec < 0 {
			nsec = -nsec
		}
	}
	shift := int64(timestampFloatShiftMapping[tst])
	radix := 1e9 / shift
	nsec += (sec % radix) * shift // valid for negative values
	sec /= radix
	return stdtime.Unix(sec, nsec), nil
}

func timeToTimestamp(tsType timestampType, t stdtime.Time) []byte {
	// Check tsType if this function is exported.
	intPart := t.Unix()
	sign := 1                  // specially, sign = 1 for zero
	fracPart := t.Nanosecond() // Nanosecond() always returns non-negative value, so adjust as follows:
	if intPart < 0 {
		sign = -sign
		if fracPart != 0 {
			fracPart = 1e9 - fracPart
			intPart++
		}
	}
	shift := timestampFloatShiftMapping[tsType]
	intPart = intPart*int64(1e9/shift) + int64(sign*fracPart/shift)
	fracPart %= shift // valid for negative values
	s := strconv.FormatInt(intPart, 10)
	if fracPart == 0 {
		return []byte(s)
	}
	var b bytes.Buffer
	b.WriteString(s)
	b.WriteRune('.')
	for radix := shift / 10; radix > 0 && fracPart != 0; radix /= 10 {
		b.WriteRune('0' + rune(fracPart/radix))
		fracPart %= radix
	}
	return b.Bytes()
}
