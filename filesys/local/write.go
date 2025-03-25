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

package local

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/filesys"
)

// writeOpenFile opens a file using os.OpenFile and creates a writer on it.
//
// The first three arguments are passed to function os.OpenFile.
// The last is passed to function github.com/donyori/gogo/filesys.Write.
// mkDirs indicates whether to make necessary directories
// before opening the file.
//
// The file is closed when the returned writer is closed.
func writeOpenFile(
	name string,
	flag int,
	perm fs.FileMode,
	mkDirs bool,
	opts *filesys.WriteOptions,
) (w filesys.Writer, err error) {
	if name == "" {
		return nil, errors.AutoNew("name is empty")
	}
	name = filepath.Clean(name)
	if mkDirs {
		err = os.MkdirAll(filepath.Dir(name), perm)
		if err != nil {
			return nil, errors.AutoWrap(err)
		}
	}
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	w, err = filesys.Write(f, opts, true)
	return w, errors.AutoWrap(err)
}

// WriteTrunc creates (if necessary) and opens a file
// with specified name and options opts for writing.
//
// If the file exists, it is truncated.
// If the file does not exist, it is created
// with specified permission perm (before umask).
//
// mkDirs indicates whether to make necessary directories
// before opening the file.
//
// opts are handled the same as in
// function github.com/donyori/gogo/filesys.Write.
//
// The file is closed when the returned writer is closed.
func WriteTrunc(
	name string,
	perm fs.FileMode,
	mkDirs bool,
	opts *filesys.WriteOptions,
) (w filesys.Writer, err error) {
	w, err = writeOpenFile(
		name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm, mkDirs, opts)
	return w, errors.AutoWrap(err)
}

// WriteAppend creates (if necessary) and opens a file
// with specified name and options opts for writing.
//
// If the file exists, new data is appended to the file.
// If the file does not exist, it is created
// with specified permission perm (before umask).
//
// mkDirs indicates whether to make necessary directories
// before opening the file.
//
// opts are handled the same as in
// function github.com/donyori/gogo/filesys.Write.
//
// The file is closed when the returned writer is closed.
func WriteAppend(
	name string,
	perm fs.FileMode,
	mkDirs bool,
	opts *filesys.WriteOptions,
) (w filesys.Writer, err error) {
	w, err = writeOpenFile(
		name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm, mkDirs, opts)
	return w, errors.AutoWrap(err)
}

// WriteExcl creates and opens a file with specified name
// and options opts for writing.
//
// The file is created with specified permission perm (before umask).
// If the file exists, it reports an error that satisfies
// errors.Is(err, fs.ErrExist) is true.
//
// mkDirs indicates whether to make necessary directories
// before opening the file.
//
// opts are handled the same as in
// function github.com/donyori/gogo/filesys.Write.
//
// The file is closed when the returned writer is closed.
func WriteExcl(
	name string,
	perm fs.FileMode,
	mkDirs bool,
	opts *filesys.WriteOptions,
) (w filesys.Writer, err error) {
	w, err = writeOpenFile(
		name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm, mkDirs, opts)
	return w, errors.AutoWrap(err)
}
