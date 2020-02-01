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
	"errors"
	"regexp"
	"strconv"
	stdtime "time"
)

// Unix timestamp.
type UnixTimestamp stdtime.Time

func (ut UnixTimestamp) String() string {
	return string(timeToUnixTimestamp(stdtime.Time(ut)))
}

func (ut UnixTimestamp) MarshalText() (text []byte, err error) {
	return timeToUnixTimestamp(stdtime.Time(ut)), nil
}

func (ut *UnixTimestamp) UnmarshalText(text []byte) error {
	t, err := unixTimestampToTime(text)
	if err != nil {
		return err
	}
	*(*stdtime.Time)(ut) = t
	return nil
}

func (ut UnixTimestamp) MarshalJSON() ([]byte, error) {
	return timeToUnixTimestamp(stdtime.Time(ut)), nil
}

func (ut *UnixTimestamp) UnmarshalJSON(b []byte) error {
	t, err := unixTimestampToTime(b)
	if err != nil {
		return err
	}
	*(*stdtime.Time)(ut) = t
	return nil
}

// Unix timestamp pattern. (without escape to fit more languages)
const UnixTimestampPatternStr = "^[+-]?[0-9]+([.][0-9]{1,9})?$"

var unixTimestampPattern = regexp.MustCompile(UnixTimestampPatternStr)

func unixTimestampToTime(ts []byte) (t stdtime.Time, err error) {
	if len(ts) == 0 {
		return stdtime.Time{}, errors.New("time: empty timestamp")
	}
	if ok := unixTimestampPattern.Match(ts); !ok {
		return stdtime.Time{}, errors.New("time: invalid timestamp")
	}
	i := bytes.IndexRune(ts, '.')
	var s, ns []byte
	if i < 0 {
		s = ts
	} else {
		s = ts[:i]
		ns = ts[i+1:]
	}
	var sec, nsec int64
	sec, err = strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		return stdtime.Time{}, err
	}
	if ns != nil {
		nsec, err = strconv.ParseInt(string(ns), 10, 64)
		if err != nil {
			return stdtime.Time{}, err
		}
		for i := len(ns); i < 9; i++ {
			nsec *= 10
		}
		if sec < 0 {
			nsec = -nsec
		}
	}
	return stdtime.Unix(sec, nsec), nil
}

func timeToUnixTimestamp(t stdtime.Time) []byte {
	sec := t.Unix()
	nsec := t.Nanosecond()
	if sec < 0 && nsec != 0 {
		nsec = 1e9 - nsec
		sec++
	}
	s := strconv.FormatInt(sec, 10)
	if nsec == 0 {
		return []byte(s)
	}
	var b bytes.Buffer
	b.WriteString(s)
	b.WriteRune('.')
	for radix := int(1e8); radix > 1 && nsec != 0; radix /= 10 {
		b.WriteRune('0' + rune(nsec/radix))
		nsec %= radix
	}
	if nsec != 0 {
		b.WriteRune('0' + rune(nsec))
	}
	return b.Bytes()
}
