// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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

package framework

import "fmt"

// PanicRecord is a panic record, including the name of the goroutine
// and the panic content (i.e., the argument passed to function panic).
type PanicRecord struct {
	Name    string // Name of the goroutine.
	Content any    // The argument passed to function panic.
}

// Error formats the panic record into a string
// and reports it as an error message.
func (pr PanicRecord) Error() string {
	if pr.Content == nil {
		return "no panic"
	}
	return fmt.Sprintf("panic on goroutine %s: %v", pr.Name, pr.Content)
}
