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
	"errors"
	"testing"

	"github.com/donyori/gogo/inout"
)

// errErrorWriter is used by errorWriter,
// whose Write method always returns this error.
var errErrorWriter = errors.New("error writer")

// errorWriter implements io.Writer.
// Its Write method does nothing and always returns (0, errErrorWriter).
type errorWriter struct{}

func (w errorWriter) Write([]byte) (n int, err error) {
	return 0, errErrorWriter
}

func TestMustFunctionsWritePanic(t *testing.T) {
	var ew errorWriter
	newBufferedWriter := func() inout.BufferedWriter {
		return inout.NewBufferedWriterSize(ew, 1)
	}
	testCases := []struct {
		name string
		f    func()
	}{
		{"method-MustWrite", func() {
			newBufferedWriter().MustWrite([]byte("ABCD"))
		}},
		{"method-MustWriteByte", func() {
			w := newBufferedWriter()
			w.MustWriteByte('A')
			w.MustWriteByte('B')
			w.MustWriteByte('C')
			w.MustWriteByte('D')
		}},
		{"method-MustWriteRune", func() {
			newBufferedWriter().MustWriteRune('汉')
		}},
		{"method-MustWriteString", func() {
			newBufferedWriter().MustWriteString("ABCD")
		}},
		{"method-MustPrintf", func() {
			newBufferedWriter().MustPrintf("A%sD", "BC")
		}},
		{"method-MustPrint", func() {
			newBufferedWriter().MustPrint("ABCD")
		}},
		{"method-MustPrintln", func() {
			newBufferedWriter().MustPrintln("ABCD")
		}},
		{"function-MustFprintf", func() {
			inout.MustFprintf(ew, "")
		}},
		{"function-MustFprint", func() {
			inout.MustFprint(ew)
		}},
		{"function-MustFprintln", func() {
			inout.MustFprintln(ew)
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				err := recover()
				wp, ok := err.(*inout.WritePanic)
				if !ok {
					t.Errorf("recover type %T; want *inout.WritePanic", err)
				} else if !errors.Is(wp, errErrorWriter) {
					t.Error("errors.Is(wp, errErrorWriter) is false")
				}
			}()
			tc.f()
		})
	}
}
