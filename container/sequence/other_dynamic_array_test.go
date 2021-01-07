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

package sequence

import "testing"

func TestNewIntDynamicArray(t *testing.T) {
	ida := NewIntDynamicArray(3)
	if n, c := ida.Len(), ida.Cap(); n != 0 || c != 3 {
		t.Errorf("NewIntDynamicArray(3) - Len(): %d, Cap(): %d.", n, c)
	}
}

func TestNewFloat64DynamicArray(t *testing.T) {
	fda := NewFloat64DynamicArray(3)
	if n, c := fda.Len(), fda.Cap(); n != 0 || c != 3 {
		t.Errorf("NewFloat64DynamicArray(3) - Len(): %d, Cap(): %d.", n, c)
	}
}

func TestNewStringDynamicArray(t *testing.T) {
	sda := NewStringDynamicArray(3)
	if n, c := sda.Len(), sda.Cap(); n != 0 || c != 3 {
		t.Errorf("NewStringDynamicArray(3) - Len(): %d, Cap(): %d.", n, c)
	}
}
