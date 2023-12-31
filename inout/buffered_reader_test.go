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

package inout_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"unicode"

	"github.com/donyori/gogo/inout"
)

func TestBufferedReader_Basic(t *testing.T) {
	const Content = `die Ruinenstadt ist immer noch schön
ich warte lange Zeit auf deine Rückkehr
in der Hand ein Vergissmeinnicht
Regentropfen sind meine Tränen
Wind ist mein Atem und mein Erzählung
Zweige und Blätter sind meine Hände
denn mein Körper ist in Wurzeln gehüllt
wenn die Jahreszeit des Tauens kommt
werde ich wach und singe ein Lied
das Vergissmeinnicht,das du mir gegeben
hast ist hier
erinnerst du dich noch?
erinnerst du dich noch
an dein Wort Das du mir gegeben hast?
erinnerst du dich noch?
erinnerst du dich noch an den Tag Andem du mir
wenn die Jahreszeit des Vergissmeinnichts kommt,
singe ich ein Lied
wenn die Jahreszeit des Vergissmeinnichts kommt,
rufe ich dich
`
	br := inout.NewBufferedReader(strings.NewReader(Content))
	if err := iotest.TestReader(br, []byte(Content)); err != nil {
		t.Error(err)
	}
}

func TestResettableBufferedReader_ConsumeByte(t *testing.T) {
	const Target byte = 'a'
	testCases := []struct {
		content         string
		n, wantConsumed int64
		wantErr         error
	}{
		{"", 0, 0, nil},
		{"", -1, 0, io.EOF},
		{"", 1, 0, io.EOF},
		{"", 2, 0, io.EOF},

		{"aaab", 0, 0, nil},
		{"aaab", -1, 3, nil},
		{"aaab", 1, 1, nil},
		{"aaab", 2, 2, nil},
		{"aaab", 3, 3, nil},
		{"aaab", 4, 3, nil},
		{"aaab", 5, 3, nil},

		{"aaa", 0, 0, nil},
		{"aaa", -1, 3, io.EOF},
		{"aaa", 1, 1, nil},
		{"aaa", 2, 2, nil},
		{"aaa", 3, 3, nil},
		{"aaa", 4, 3, io.EOF},
		{"aaa", 5, 3, io.EOF},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("content=%+q&n=%d", tc.content, tc.n),
			func(t *testing.T) {
				br := inout.NewBufferedReader(strings.NewReader(tc.content))
				consumed, err := br.ConsumeByte(Target, tc.n)
				if consumed != tc.wantConsumed {
					t.Errorf("got consumed %d; want %d",
						consumed, tc.wantConsumed)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got err %v; want %v", err, tc.wantErr)
				}
			},
		)
	}
}

func TestResettableBufferedReader_ConsumeByteFunc(t *testing.T) {
	f := func(c byte) bool {
		return c >= 'a' && c <= 'z'
	}
	testCases := []struct {
		content         string
		n, wantConsumed int64
		wantErr         error
	}{
		{"", 0, 0, nil},
		{"", -1, 0, io.EOF},
		{"", 1, 0, io.EOF},
		{"", 2, 0, io.EOF},

		{"abc1", 0, 0, nil},
		{"abc1", -1, 3, nil},
		{"abc1", 1, 1, nil},
		{"abc1", 2, 2, nil},
		{"abc1", 3, 3, nil},
		{"abc1", 4, 3, nil},
		{"abc1", 5, 3, nil},

		{"abc", 0, 0, nil},
		{"abc", -1, 3, io.EOF},
		{"abc", 1, 1, nil},
		{"abc", 2, 2, nil},
		{"abc", 3, 3, nil},
		{"abc", 4, 3, io.EOF},
		{"abc", 5, 3, io.EOF},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("content=%+q&n=%d", tc.content, tc.n),
			func(t *testing.T) {
				br := inout.NewBufferedReader(strings.NewReader(tc.content))
				consumed, err := br.ConsumeByteFunc(f, tc.n)
				if consumed != tc.wantConsumed {
					t.Errorf("got consumed %d; want %d",
						consumed, tc.wantConsumed)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got err %v; want %v", err, tc.wantErr)
				}
			},
		)
	}
}

func TestResettableBufferedReader_ConsumeRune(t *testing.T) {
	const Target rune = '对'
	testCases := []struct {
		content         string
		n, wantConsumed int64
		wantErr         error
	}{
		{"", 0, 0, nil},
		{"", -1, 0, io.EOF},
		{"", 1, 0, io.EOF},
		{"", 2, 0, io.EOF},

		{"对对对啊", 0, 0, nil},
		{"对对对啊", -1, 3, nil},
		{"对对对啊", 1, 1, nil},
		{"对对对啊", 2, 2, nil},
		{"对对对啊", 3, 3, nil},
		{"对对对啊", 4, 3, nil},
		{"对对对啊", 5, 3, nil},

		{"对对对", 0, 0, nil},
		{"对对对", -1, 3, io.EOF},
		{"对对对", 1, 1, nil},
		{"对对对", 2, 2, nil},
		{"对对对", 3, 3, nil},
		{"对对对", 4, 3, io.EOF},
		{"对对对", 5, 3, io.EOF},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("content=%+q&n=%d", tc.content, tc.n),
			func(t *testing.T) {
				br := inout.NewBufferedReader(strings.NewReader(tc.content))
				consumed, err := br.ConsumeRune(Target, tc.n)
				if consumed != tc.wantConsumed {
					t.Errorf("got consumed %d; want %d",
						consumed, tc.wantConsumed)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got err %v; want %v", err, tc.wantErr)
				}
			},
		)
	}
}

func TestResettableBufferedReader_ConsumeRune_InvalidRune(t *testing.T) {
	bad1 := "啊"[1:2]
	bad2 := "对"[1:]
	s1 := bad1 + "啊"
	s2 := bad2 + "对"
	s3 := bad1 + bad2 + "对"
	s4 := bad1 + bad2
	testCases := []struct {
		content         string
		n, wantConsumed int64
		wantErr         error
	}{
		{s1, 0, 0, nil},
		{s1, -1, 1, nil},
		{s1, 1, 1, nil},
		{s1, 2, 1, nil},
		{s1, 3, 1, nil},

		{s2, 0, 0, nil},
		{s2, -1, 2, nil},
		{s2, 1, 1, nil},
		{s2, 2, 2, nil},
		{s2, 3, 2, nil},
		{s2, 4, 2, nil},

		{s3, 0, 0, nil},
		{s3, -1, 3, nil},
		{s3, 1, 1, nil},
		{s3, 2, 2, nil},
		{s3, 3, 3, nil},
		{s3, 4, 3, nil},
		{s3, 5, 3, nil},

		{s4, 0, 0, nil},
		{s4, -1, 3, io.EOF},
		{s4, 1, 1, nil},
		{s4, 2, 2, nil},
		{s4, 3, 3, nil},
		{s4, 4, 3, io.EOF},
		{s4, 5, 3, io.EOF},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("content=%+q&n=%d", tc.content, tc.n),
			func(t *testing.T) {
				br := inout.NewBufferedReader(strings.NewReader(tc.content))
				consumed, err := br.ConsumeRune(unicode.ReplacementChar, tc.n)
				if consumed != tc.wantConsumed {
					t.Errorf("got consumed %d; want %d",
						consumed, tc.wantConsumed)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got err %v; want %v", err, tc.wantErr)
				}
			},
		)
	}
}

func TestResettableBufferedReader_ConsumeRuneFunc(t *testing.T) {
	f := func(r rune, size int) bool {
		return r >= 'a' && r <= 'z' || size > 1
	}
	testCases := []struct {
		content         string
		n, wantConsumed int64
		wantErr         error
	}{
		{"", 0, 0, nil},
		{"", -1, 0, io.EOF},
		{"", 1, 0, io.EOF},
		{"", 2, 0, io.EOF},

		{"o夏天!", 0, 0, nil},
		{"o夏天!", -1, 3, nil},
		{"o夏天!", 1, 1, nil},
		{"o夏天!", 2, 2, nil},
		{"o夏天!", 3, 3, nil},
		{"o夏天!", 4, 3, nil},
		{"o夏天!", 5, 3, nil},

		{"o夏天", 0, 0, nil},
		{"o夏天", -1, 3, io.EOF},
		{"o夏天", 1, 1, nil},
		{"o夏天", 2, 2, nil},
		{"o夏天", 3, 3, nil},
		{"o夏天", 4, 3, io.EOF},
		{"o夏天", 5, 3, io.EOF},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("content=%+q&n=%d", tc.content, tc.n),
			func(t *testing.T) {
				br := inout.NewBufferedReader(strings.NewReader(tc.content))
				consumed, err := br.ConsumeRuneFunc(f, tc.n)
				if consumed != tc.wantConsumed {
					t.Errorf("got consumed %d; want %d",
						consumed, tc.wantConsumed)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got err %v; want %v", err, tc.wantErr)
				}
			},
		)
	}
}

func TestResettableBufferedReader_ReadEntireLine(t *testing.T) {
	longLine, data := buildLongLineAndInputData()
	br := inout.NewBufferedReader(bytes.NewReader(data))
	var err error
	for err == nil {
		var line []byte
		line, err = br.ReadEntireLine()
		if err == nil {
			if !bytes.Equal(line, longLine) {
				t.Errorf("read line wrong; line length: %d\nline: %q\nwant: %q",
					len(line), line, longLine)
			}
		} else if !errors.Is(err, io.EOF) {
			t.Error(err)
		}
	}
}

func TestBufferedReader_WriteLineTo(t *testing.T) {
	longLine, data := buildLongLineAndInputData()
	br := inout.NewBufferedReader(bytes.NewReader(data))
	var output bytes.Buffer
	output.Grow(len(longLine) + 100) // reserve enough space
	var err error
	for err == nil {
		output.Reset()
		_, err = br.WriteLineTo(&output)
		if err == nil {
			if !bytes.Equal(output.Bytes(), longLine) {
				t.Errorf("output line wrong; line length: %d\nline: %q\nwant: %q",
					output.Len(), output.Bytes(), longLine)
			}
		} else if !errors.Is(err, io.EOF) {
			t.Error(err)
		}
	}
}

// buildLongLineAndInputData generates a long string, longLine,
// with tens of thousands of bytes, without end-of-line characters.
//
// It also generates data with several repetitions of longLine
// separated by the end-of-line character '\n'.
// The data does not end with an end-of-line character.
func buildLongLineAndInputData() (longLine, data []byte) {
	const RepeatingUnit = "123456789ⅠⅡⅢⅣⅤⅥⅦⅧⅨ"
	const NumLine int = 4
	longLine = bytes.Repeat([]byte(RepeatingUnit), 2048)
	data = make([]byte, (len(longLine)+1)*NumLine-1)
	var n int
	for i := 0; i < NumLine; i++ {
		if i > 0 {
			data[n], n = '\n', n+1
		}
		n += copy(data[n:], longLine)
	}
	return
}
