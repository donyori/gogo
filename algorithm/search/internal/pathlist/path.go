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

package pathlist

// Path recursively represents a path.
//
// The path represented by this Path is constructed by
// extending path P with element E.
//
// For example, the path "0 -> 1 -> 2" is represented as
//
//	&Path[int]{
//	    E: 2,
//	    P: &Path[int]{
//	        E: 1,
//	        P: &Path[int]{
//	            E: 0,
//	            P: nil,
//	        }
//	    },
//	}
type Path[Elem any] struct {
	E Elem
	P *Path[Elem]
}

// ToList describes the path represented by this Path as a list,
// whose items are the elements along the path, from the start to the end.
//
// The client should guarantee that there is no change along the path
// during the call to this method.
func (p *Path[Elem]) ToList() []Elem {
	var length int
	for t := p; t != nil; t = t.P {
		length++
	}
	list := make([]Elem, length)
	for i, t := length-1, p; i >= 0; i, t = i-1, t.P {
		list[i] = t.E
	}
	return list
}
