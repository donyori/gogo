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

package errors_test

import (
	"strings"
	"testing"

	"github.com/donyori/gogo/errors"
)

func TestAs_PanicForErrorPointer(t *testing.T) {
	target := new(error)
	err := errors.New("test error")
	defer func() {
		e := recover()
		if e == nil {
			t.Error("want panic but not")
			return
		}
		s, ok := e.(string)
		if !ok || !strings.HasSuffix(s,
			"target is of type *error; As always returns true for that") {
			t.Error("panic -", e)
		}
	}()
	errors.As(err, target)
}
