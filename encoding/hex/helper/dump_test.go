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

package helper

import (
	"bytes"
	"testing"

	"github.com/donyori/gogo/encoding/hex"
)

func _TestShowExampleDumpConfig(t *testing.T) {
	cfg := ExampleDumpConfig(true, 0)
	var b bytes.Buffer
	for b.Len() < 500 {
		b.WriteString("Hello world! 你好，世界！")
	}
	s := hex.DumpToString(b.Bytes(), cfg)
	t.Log("\n" + s)
	cfg = ExampleDumpConfig(false, 0)
	s = hex.DumpToString(b.Bytes(), cfg)
	t.Log("\n" + s)
}
