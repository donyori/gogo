// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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

package timestamp

import (
	"bytes"
	"regexp"
	"strconv"
	"time"

	"github.com/donyori/gogo/errors"
)

// Regular expression patterns for various timestamps.
const (
	// UnixTimestampPattern is a regular expression pattern for UNIX timestamp
	// (without escape to fit more languages).
	UnixTimestampPattern = "^[+-]?[0-9]+([.][0-9]{1,9})?$"
	// MilliTimestampPattern is a regular expression pattern for
	// millisecond timestamp (without escape to fit more languages).
	MilliTimestampPattern = "^[+-]?[0-9]+([.][0-9]{1,6})?$"
	// MicroTimestampPattern is a regular expression pattern for
	// microsecond timestamp (without escape to fit more languages).
	MicroTimestampPattern = "^[+-]?[0-9]+([.][0-9]{1,3})?$"
	// NanoTimestampPattern is a regular expression pattern for
	// nanosecond timestamp (without escape to fit more languages).
	NanoTimestampPattern = "^[+-]?[0-9]+$"
)

// Timestamp is a timestamp wrapped on time.Time.
//
// It automatically detects the time unit (seconds, milliseconds, microseconds,
// or nanoseconds) when parsed from strings.
// It is treated as a UNIX timestamp (in seconds) when formatted to strings.
//
// It determines the time unit by the integer part digits, as follows:
//
//	less than 12-digits - second,
//	12-digits to 14-digits - millisecond,
//	15-digits or 16-digits - microsecond,
//	more than 16-digits - nanosecond.
type Timestamp time.Time

// String formats this timestamp in seconds to a decimal representation.
func (ts Timestamp) String() string {
	return string(timeToTimestamp(autoTimestamp, time.Time(ts)))
}

// MarshalText formats this timestamp in seconds to a decimal representation.
//
// It always returns a nil error.
//
// It conforms to interface encoding.TextMarshaler.
func (ts Timestamp) MarshalText() (text []byte, err error) {
	return timeToTimestamp(autoTimestamp, time.Time(ts)), nil
}

// UnmarshalText parses decimal timestamp.
//
// It automatically determines the time unit by the integer part digits,
// as follows:
//
//	less than 12-digits - second,
//	12-digits to 14-digits - millisecond,
//	15-digits or 16-digits - microsecond,
//	more than 16-digits - nanosecond.
//
// It reports an error if text is empty or text is not decimal timestamp.
//
// It conforms to interface encoding.TextUnmarshaler.
func (ts *Timestamp) UnmarshalText(text []byte) error {
	t, err := timestampToTime(autoTimestamp, text)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*ts = Timestamp(t)
	return nil
}

// MarshalJSON formats this timestamp in seconds to a decimal representation.
//
// It always returns a nil error.
//
// It conforms to interface encoding/json.Marshaler.
func (ts Timestamp) MarshalJSON() ([]byte, error) {
	return timeToTimestamp(autoTimestamp, time.Time(ts)), nil
}

// UnmarshalJSON parses decimal timestamp.
//
// It automatically determines the time unit by the integer part digits,
// as follows:
//
//	less than 12-digits - second,
//	12-digits to 14-digits - millisecond,
//	15-digits or 16-digits - microsecond,
//	more than 16-digits - nanosecond.
//
// It reports an error if b is empty or
// b is neither decimal timestamp nor []byte("null").
//
// By convention, to approximate the behavior of encoding/json.Unmarshal itself,
// It does nothing if b is []byte("null").
//
// It conforms to interface encoding/json.Unmarshaler.
func (ts *Timestamp) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	t, err := timestampToTime(autoTimestamp, b)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*ts = Timestamp(t)
	return nil
}

// UnixTimestamp is a UNIX timestamp (in seconds) wrapped on time.Time.
type UnixTimestamp time.Time

// String formats this timestamp in seconds to a decimal representation.
func (ut UnixTimestamp) String() string {
	return string(timeToTimestamp(unixTimestamp, time.Time(ut)))
}

// MarshalText formats this timestamp in seconds to a decimal representation.
//
// It always returns a nil error.
//
// It conforms to interface encoding.TextMarshaler.
func (ut UnixTimestamp) MarshalText() (text []byte, err error) {
	return timeToTimestamp(unixTimestamp, time.Time(ut)), nil
}

// UnmarshalText parses decimal timestamp in seconds.
//
// It reports an error if text is empty or text is not decimal timestamp.
//
// It conforms to interface encoding.TextUnmarshaler.
func (ut *UnixTimestamp) UnmarshalText(text []byte) error {
	t, err := timestampToTime(unixTimestamp, text)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*ut = UnixTimestamp(t)
	return nil
}

// MarshalJSON formats this timestamp in seconds to a decimal representation.
//
// It always returns a nil error.
//
// It conforms to interface encoding/json.Marshaler.
func (ut UnixTimestamp) MarshalJSON() ([]byte, error) {
	return timeToTimestamp(unixTimestamp, time.Time(ut)), nil
}

// UnmarshalJSON parses decimal timestamp in seconds.
//
// It reports an error if b is empty or
// b is neither decimal timestamp nor []byte("null").
//
// By convention, to approximate the behavior of encoding/json.Unmarshal itself,
// It does nothing if b is []byte("null").
//
// It conforms to interface encoding/json.Unmarshaler.
func (ut *UnixTimestamp) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	t, err := timestampToTime(unixTimestamp, b)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*ut = UnixTimestamp(t)
	return nil
}

// MilliTimestamp is a millisecond timestamp, often used in JavaScript,
// wrapped on time.Time.
type MilliTimestamp time.Time

// String formats this timestamp in milliseconds to a decimal representation.
func (mt MilliTimestamp) String() string {
	return string(timeToTimestamp(milliTimestamp, time.Time(mt)))
}

// MarshalText formats this timestamp in milliseconds to
// a decimal representation.
//
// It always returns a nil error.
//
// It conforms to interface encoding.TextMarshaler.
func (mt MilliTimestamp) MarshalText() (text []byte, err error) {
	return timeToTimestamp(milliTimestamp, time.Time(mt)), nil
}

// UnmarshalText parses decimal timestamp in milliseconds.
//
// It reports an error if text is empty or text is not decimal timestamp.
//
// It conforms to interface encoding.TextUnmarshaler.
func (mt *MilliTimestamp) UnmarshalText(text []byte) error {
	t, err := timestampToTime(milliTimestamp, text)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*mt = MilliTimestamp(t)
	return nil
}

// MarshalJSON formats this timestamp in milliseconds to
// a decimal representation.
//
// It always returns a nil error.
//
// It conforms to interface encoding/json.Marshaler.
func (mt MilliTimestamp) MarshalJSON() ([]byte, error) {
	return timeToTimestamp(milliTimestamp, time.Time(mt)), nil
}

// UnmarshalJSON parses decimal timestamp in milliseconds.
//
// It reports an error if b is empty or
// b is neither decimal timestamp nor []byte("null").
//
// By convention, to approximate the behavior of encoding/json.Unmarshal itself,
// It does nothing if b is []byte("null").
//
// It conforms to interface encoding/json.Unmarshaler.
func (mt *MilliTimestamp) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	t, err := timestampToTime(milliTimestamp, b)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*mt = MilliTimestamp(t)
	return nil
}

// MicroTimestamp is a microsecond timestamp wrapped on time.Time.
type MicroTimestamp time.Time

// String formats this timestamp in microseconds to a decimal representation.
func (ct MicroTimestamp) String() string {
	return string(timeToTimestamp(microTimestamp, time.Time(ct)))
}

// MarshalText formats this timestamp in microseconds to
// a decimal representation.
//
// It always returns a nil error.
//
// It conforms to interface encoding.TextMarshaler.
func (ct MicroTimestamp) MarshalText() (text []byte, err error) {
	return timeToTimestamp(microTimestamp, time.Time(ct)), nil
}

// UnmarshalText parses decimal timestamp in microseconds.
//
// It reports an error if text is empty or text is not decimal timestamp.
//
// It conforms to interface encoding.TextUnmarshaler.
func (ct *MicroTimestamp) UnmarshalText(text []byte) error {
	t, err := timestampToTime(microTimestamp, text)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*ct = MicroTimestamp(t)
	return nil
}

// MarshalJSON formats this timestamp in microseconds to
// a decimal representation.
//
// It always returns a nil error.
//
// It conforms to interface encoding/json.Marshaler.
func (ct MicroTimestamp) MarshalJSON() ([]byte, error) {
	return timeToTimestamp(microTimestamp, time.Time(ct)), nil
}

// UnmarshalJSON parses decimal timestamp in microseconds.
//
// It reports an error if b is empty or
// b is neither decimal timestamp nor []byte("null").
//
// By convention, to approximate the behavior of encoding/json.Unmarshal itself,
// It does nothing if b is []byte("null").
//
// It conforms to interface encoding/json.Unmarshaler.
func (ct *MicroTimestamp) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	t, err := timestampToTime(microTimestamp, b)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*ct = MicroTimestamp(t)
	return nil
}

// NanoTimestamp is a nanosecond timestamp wrapped on time.Time.
type NanoTimestamp time.Time

// String formats this timestamp in nanoseconds to a decimal representation.
func (nt NanoTimestamp) String() string {
	return string(timeToTimestamp(nanoTimestamp, time.Time(nt)))
}

// MarshalText formats this timestamp in nanoseconds to
// a decimal representation.
//
// It always returns a nil error.
//
// It conforms to interface encoding.TextMarshaler.
func (nt NanoTimestamp) MarshalText() (text []byte, err error) {
	return timeToTimestamp(nanoTimestamp, time.Time(nt)), nil
}

// UnmarshalText parses decimal timestamp in nanoseconds.
//
// It reports an error if text is empty or text is not decimal timestamp.
//
// It conforms to interface encoding.TextUnmarshaler.
func (nt *NanoTimestamp) UnmarshalText(text []byte) error {
	t, err := timestampToTime(nanoTimestamp, text)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*nt = NanoTimestamp(t)
	return nil
}

// MarshalJSON formats this timestamp in nanoseconds to
// a decimal representation.
//
// It always returns a nil error.
//
// It conforms to interface encoding/json.Marshaler.
func (nt NanoTimestamp) MarshalJSON() ([]byte, error) {
	return timeToTimestamp(nanoTimestamp, time.Time(nt)), nil
}

// UnmarshalJSON parses decimal timestamp in nanoseconds.
//
// It reports an error if b is empty or
// b is neither decimal timestamp nor []byte("null").
//
// By convention, to approximate the behavior of encoding/json.Unmarshal itself,
// It does nothing if b is []byte("null").
//
// It conforms to interface encoding/json.Unmarshaler.
func (nt *NanoTimestamp) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	t, err := timestampToTime(nanoTimestamp, b)
	if err != nil {
		return errors.AutoWrap(err)
	}
	*nt = NanoTimestamp(t)
	return nil
}

// timestampType indicates the type of timestamp.
//
// Possible types are autoTimestamp, unixTimestamp, milliTimestamp,
// microTimestamp, and nanoTimestamp.
type timestampType int8

// Types of timestamp.
const (
	// autoTimestamp is a timestamp that automatically detects the time unit
	// when parsed from strings.
	// It is treated as a UNIX timestamp (in seconds) when formatted to strings.
	//
	// It determines the time unit by the integer part digits, as follows:
	//   - less than 12-digits - second,
	//   - 12-digits to 14-digits - millisecond,
	//   - 15-digits or 16-digits - microsecond,
	//   - more than 16-digits - nanosecond.
	autoTimestamp timestampType = iota
	// unixTimestamp is a UNIX timestamp (in seconds).
	unixTimestamp
	// milliTimestamp is a millisecond timestamp.
	milliTimestamp
	// microTimestamp is a microsecond timestamp.
	microTimestamp
	// nanoTimestamp is a nanosecond timestamp.
	nanoTimestamp
)

// Regular expressions compiled from the patterns.
var (
	// unixTimestampRegExpr is a regular expression for UNIX timestamp.
	unixTimestampRegExpr = regexp.MustCompile(UnixTimestampPattern)
	// milliTimestampRegExpr is a regular expression for millisecond timestamp.
	milliTimestampRegExpr = regexp.MustCompile(MilliTimestampPattern)
	// microTimestampRegExpr is a regular expression for microsecond timestamp.
	microTimestampRegExpr = regexp.MustCompile(MicroTimestampPattern)
	// nanoTimestampRegExpr is a regular expression for nanosecond timestamp.
	nanoTimestampRegExpr = regexp.MustCompile(NanoTimestampPattern)
)

// Variables used in functions timestampToTime and timeToTimestamp.
var (
	// timestampRegExprMapping is a mapping from timestampType to
	// regular expressions.
	timestampRegExprMapping = [...]*regexp.Regexp{
		unixTimestampRegExpr,
		unixTimestampRegExpr,
		milliTimestampRegExpr,
		microTimestampRegExpr,
		nanoTimestampRegExpr,
	}

	// timestampFloatShiftMapping is a mapping from timestampType to
	// the float point position relative to nanoseconds.
	timestampFloatShiftMapping = [...]int{1e9, 1e9, 1e6, 1e3, 1}

	// timestampFractionalLenMapping is a mapping from timestampType to
	// the maximum length of fractional part of decimal timestamp.
	//
	// The first item -1 means it is invalid for autoTimestamp.
	timestampFractionalLenMapping = [...]int{-1, 9, 6, 3, 0}
)

// timestampToTime parses decimal timestamp ts into time.Time
// according to the timestamp type tsType.
//
// It reports an error if ts is empty, or ts is not valid for tsType.
//
// Caller should guarantee that tsType is valid.
func timestampToTime(tsType timestampType, ts []byte) (t time.Time, err error) {
	if len(ts) == 0 {
		err = errors.AutoNew("empty timestamp")
		return
	} else if !timestampRegExprMapping[tsType].Match(ts) {
		err = errors.AutoNew("invalid timestamp")
		return
	}
	pointIdx := bytes.IndexByte(ts, '.')
	tst := tsType
	if tsType == autoTimestamp {
		tst, err = detectTimestampType(ts, pointIdx)
		if err != nil {
			err = errors.AutoWrap(err)
			return
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
		err = errors.AutoWrap(err)
		return
	} else if ns != nil {
		nsec, err = strconv.ParseInt(string(ns), 10, 64)
		if err != nil {
			err = errors.AutoWrap(err)
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
	return time.Unix(sec, nsec), nil
}

// timeToTimestamp formats time t into decimal timestamp
// according to the timestamp type tsType.
//
// Caller should guarantee that tsType is valid.
func timeToTimestamp(tsType timestampType, t time.Time) []byte {
	intPart := t.Unix()
	sign := 1                  // sign = 1 for nonnegative values, -1 for negative values.
	fracPart := t.Nanosecond() // Nanosecond() always returns nonnegative value, so adjust as follows:
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
	b.WriteByte('.')
	for radix := shift / 10; radix > 0 && fracPart != 0; radix /= 10 {
		b.WriteByte('0' + byte(fracPart/radix))
		fracPart %= radix
	}
	return b.Bytes()
}

// detectTimestampType detects the type of timestamp
// according to the specifications of autoTimestamp.
//
// It is called by function timestampToTime.
//
// ts is the decimal timestamp.
//
// pointIdx is the index of the decimal point.
// pointIdx is -1 if the decimal point is not present in ts.
func detectTimestampType(ts []byte, pointIdx int) (
	tst timestampType, err error) {
	var k int // length of integer part
	if pointIdx < 0 {
		k = len(ts)
	} else {
		k = pointIdx
	}
	if ts[0] == '-' || ts[0] == '+' {
		k--
	}
	tst = nanoTimestamp
	switch {
	case k < 12:
		tst = unixTimestamp
	case k < 15:
		tst = milliTimestamp
	case k < 17:
		tst = microTimestamp
	}
	if pointIdx >= 0 &&
		len(ts)-pointIdx-1 > timestampFractionalLenMapping[tst] {
		err = errors.AutoNew("invalid timestamp")
	}
	return
}
