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

package filesys

import (
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/inout"
)

// ErrNotTar is an error indicating that the file is not archived by tar,
// or is opened in raw mode.
//
// The client should use errors.Is to test whether an error is ErrNotTar.
var ErrNotTar = errors.AutoNewCustom(
	"file is not archived by tar, or is opened in raw mode",
	errors.PrependFullPkgName,
	0,
)

// ErrFileReaderClosed is an error indicating that
// the file reader is already closed.
//
// The client should use errors.Is to test whether
// an error is ErrFileReaderClosed.
var ErrFileReaderClosed = errors.AutoWrapCustom(
	inout.NewClosedError("file reader", inout.ErrReaderClosed),
	errors.PrependFullPkgName,
	0,
	nil,
)

// ErrFileWriterClosed is an error indicating that
// the file writer is already closed.
//
// The client should use errors.Is to test whether
// an error is ErrFileWriterClosed.
var ErrFileWriterClosed = errors.AutoWrapCustom(
	inout.NewClosedError("file writer", inout.ErrWriterClosed),
	errors.PrependFullPkgName,
	0,
	nil,
)
