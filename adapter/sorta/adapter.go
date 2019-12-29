// gogo. A Golang toolbox.
// Copyright (C) 2019 Yuan Gao
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

package sorta

import "github.com/donyori/gogo/function"

// An adapter for: []interface{} + github.com/donyori/gogo/function.LessFunc
// -> sort.Interface.
type Slice struct {
	Data     []interface{}
	LessFunc function.LessFunc
}

func (s *Slice) Len() int {
	return len(s.Data)
}

func (s *Slice) Less(i, j int) bool {
	return s.LessFunc(s.Data[i], s.Data[j])
}

func (s *Slice) Swap(i, j int) {
	s.Data[i], s.Data[j] = s.Data[j], s.Data[i]
}