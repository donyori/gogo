// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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

package local

import (
	"os"
	"path/filepath"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/filesys"
)

// Read opens a file with specified name for reading.
//
// The file will be closed when closing the returned reader.
//
// If the file is a symlink, it will be evaluated by filepath.EvalSymlinks.
//
// The file is opened by os.Open;
// the associated file descriptor has mode syscall.O_RDONLY.
func Read(name string, opts *filesys.ReadOptions) (r filesys.Reader, err error) {
	name, err = filepath.EvalSymlinks(name)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	return filesys.Read(f, opts, true)
}
