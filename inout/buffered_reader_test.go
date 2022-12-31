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
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/donyori/gogo/inout"
)

func TestBufferedReader_Basic(t *testing.T) {
	content := `die Ruinenstadt ist immer noch schön
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
	r := inout.NewBufferedReader(strings.NewReader(content))
	if err := iotest.TestReader(r, []byte(content)); err != nil {
		t.Error(err)
	}
}

func TestResettableBufferedReader_ReadEntireLine(t *testing.T) {
	longLine, data := buildLongLineAndInputData()
	br := inout.NewBufferedReader(bytes.NewReader(data))
	longLineWitoutEndOfLine := strings.TrimRight(longLine, "\n")
	var err error
	for err == nil {
		var line []byte
		line, err = br.ReadEntireLine()
		if err == nil {
			if s := string(line); s != longLineWitoutEndOfLine {
				t.Errorf("read line wrong; line length: %d\nline: %q\nwant: %q", len(s), s, longLineWitoutEndOfLine)
			}
		} else if !errors.Is(err, io.EOF) {
			t.Error(err)
		}
	}
}

func TestBufferedReader_WriteLineTo(t *testing.T) {
	longLine, data := buildLongLineAndInputData()
	br := inout.NewBufferedReader(bytes.NewReader(data))
	longLineWitoutEndOfLine := strings.TrimRight(longLine, "\n")
	var output strings.Builder
	output.Grow(len(longLine) + 100)
	var err error
	for err == nil {
		output.Reset()
		_, err = br.WriteLineTo(&output)
		if err == nil {
			if output.String() != longLineWitoutEndOfLine {
				t.Errorf("output line wrong; line length: %d\nline: %q\nwant: %q", output.Len(), output.String(), longLineWitoutEndOfLine)
			}
		} else if !errors.Is(err, io.EOF) {
			t.Error(err)
		}
	}
}

func buildLongLineAndInputData() (longLine string, data []byte) {
	var longLineBuilder strings.Builder
	longLineBuilder.Grow(16390)
	for longLineBuilder.Len() < 16384 {
		longLineBuilder.WriteString("12345678")
	}
	longLineBuilder.WriteByte('\n')
	longLine = longLineBuilder.String()
	data = make([]byte, 0, 65560)
	for i := 0; i < 4; i++ {
		data = append(data, longLine...)
	}
	data = data[:len(data)-1] // Remove the last '\n'.
	return longLine, data
}
