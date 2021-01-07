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

package agpl3

import (
	"errors"
	"strings"
	"testing"
)

func _TestShowPrintNotice(t *testing.T) {
	w := new(strings.Builder)
	err := PrintNotice(w, "", "", "", "")
	if !errors.Is(err, ErrAuthorMissing) {
		t.Error("err is not ErrAuthorMissing, err:", err)
	}
	w.Reset()
	err = PrintNotice(w, "", "", "donyori", "")
	if err != nil {
		t.Error(err)
	}
	t.Log("\n", w.String())
	w.Reset()
	err = PrintNotice(w, "", "", "donyori", "https://github.com/donyori/gogo")
	if err != nil {
		t.Error(err)
	}
	t.Log("\n", w.String())
	w.Reset()
	err = PrintNotice(w, "ShowPrintNotice", "2019-2020", "donyori", "")
	if err != nil {
		t.Error(err)
	}
	t.Log("\n", w.String())
	w.Reset()
	err = PrintNotice(w, "ShowPrintNotice", "2019-2020", "donyori", "https://github.com/donyori/gogo")
	if err != nil {
		t.Error(err)
	}
	t.Log("\n", w.String())
}
