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

package filesys_test

import (
	"io/fs"
	"path"
	"sync"
	"time"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/filesys"
)

// WritableFileImpl is an implementation of interface WritableFile, for testing.
type WritableFileImpl struct {
	Name    string      // Filename.
	Data    []byte      // File contents.
	Mode    fs.FileMode // FileInfo.Mode.
	ModTime time.Time   // FileInfo.ModTime.
	Sys     any         // FileInfo.Sys.

	closed bool
	lock   sync.Mutex // Lock for methods Write and Close.
}

var _ filesys.WritableFile = (*WritableFileImpl)(nil)

// Write appends p to wf.Data.
//
// It returns (len(p), nil) if wf is not closed.
// Otherwise, it writes nothing and reports io/fs.ErrClosed.
func (wf *WritableFileImpl) Write(p []byte) (n int, err error) {
	if wf == nil {
		return 0, errors.AutoNew("*WritableFileImpl is nil")
	}
	wf.lock.Lock()
	defer wf.lock.Unlock()
	if wf.closed {
		return 0, errors.AutoWrap(fs.ErrClosed)
	}
	wf.Data = append(wf.Data, p...)
	return len(p), nil
}

func (wf *WritableFileImpl) Close() error {
	if wf == nil {
		return errors.AutoNew("*WritableFileImpl is nil")
	}
	wf.lock.Lock()
	defer wf.lock.Unlock()
	wf.closed = true
	return nil
}

func (wf *WritableFileImpl) Stat() (info fs.FileInfo, err error) {
	if wf == nil {
		return nil, errors.AutoNew("*WritableFileImpl is nil")
	}
	return &writableFileInfo{f: wf}, nil
}

// writableFileInfo is an implementation of interface io/fs.FileInfo
// for WritableFileImpl.
type writableFileInfo struct {
	f *WritableFileImpl
}

func (wfi *writableFileInfo) Name() string {
	return path.Base(wfi.f.Name)
}

func (wfi *writableFileInfo) Size() int64 {
	return int64(len(wfi.f.Data))
}

func (wfi *writableFileInfo) Mode() fs.FileMode {
	return wfi.f.Mode
}

func (wfi *writableFileInfo) ModTime() time.Time {
	return wfi.f.ModTime
}

func (wfi *writableFileInfo) IsDir() bool {
	return wfi.f.Mode.IsDir()
}

func (wfi *writableFileInfo) Sys() any {
	return wfi.f.Sys
}
