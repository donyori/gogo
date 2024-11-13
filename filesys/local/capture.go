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

package local

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/donyori/gogo/concurrency"
	"github.com/donyori/gogo/errors"
)

// Backups of the standard output and standard error files.
var (
	stdoutBackup = os.Stdout
	stderrBackup = os.Stderr
)

// Locks for capturing the standard output and standard error files.
var (
	captureStdoutLock sync.Mutex
	captureStderrLock sync.Mutex
)

// CaptureToStringFunc is a function that stops capturing
// and returns the captured content as a string.
//
// CaptureToStringFunc is safe for concurrent use.
// The first call to the CaptureToStringFunc returns
// the captured content, any error encountered, and true.
// The subsequent calls to the CaptureToStringFunc return
// the same content and error, but false.
type CaptureToStringFunc func() (s string, err error, first bool)

// CaptureStdoutToString captures the standard output stream to a string.
//
// It returns a CaptureToStringFunc and any error encountered.
// The CaptureToStringFunc stops capturing and retrieves the captured content.
//
// If the returned CaptureToStringFunc is not nil,
// the client is responsible for calling it to stop capturing and
// restore the standard output stream when capturing is no longer needed.
//
// CaptureStdoutToString reports an error if the standard output stream
// has already been captured elsewhere.
func CaptureStdoutToString() (f CaptureToStringFunc, err error) {
	f, err = captureOutputFileToString(
		"stdout", &os.Stdout, stdoutBackup, &captureStdoutLock)
	return f, errors.AutoWrap(err)
}

// CaptureStderrToString captures the standard error stream to a string.
//
// It returns a CaptureToStringFunc and any error encountered.
// The CaptureToStringFunc stops capturing and retrieves the captured content.
//
// If the returned CaptureToStringFunc is not nil,
// the client is responsible for calling it to stop capturing and
// restore the standard error stream when capturing is no longer needed.
//
// CaptureStderrToString reports an error if the standard error stream
// has already been captured elsewhere.
func CaptureStderrToString() (f CaptureToStringFunc, err error) {
	f, err = captureOutputFileToString(
		"stderr", &os.Stderr, stderrBackup, &captureStderrLock)
	return f, errors.AutoWrap(err)
}

// captureOutputFileToString is the common process of
// CaptureStdoutToString and CaptureStderrToString.
//
// Caller should guarantee that filePtr, backup, and locker are not nil.
func captureOutputFileToString(
	name string,
	filePtr **os.File,
	backup *os.File,
	locker sync.Locker,
) (f CaptureToStringFunc, err error) {
	if name == "" {
		name = "the file"
	}
	locker.Lock()
	defer locker.Unlock()
	if *filePtr != backup {
		return nil, errors.AutoWrap(fmt.Errorf(
			"%s has been captured elsewhere", name))
	}
	r, w, err := os.Pipe()
	if err != nil {
		return nil, errors.AutoWrap(err)
	}

	// Read from the pipe immediately in a dedicated goroutine
	// to avoid pipe buffer overflow.

	type stringError struct {
		s   string
		err error
	}
	c := make(chan stringError, 1)
	go func(r *os.File, c chan<- stringError) {
		defer close(c)
		var b strings.Builder
		_, err := io.Copy(&b, r)
		c <- stringError{s: b.String(), err: errors.AutoWrap(err)}
	}(r, c)
	*filePtr = w // replace the file with the writer side of the pipe
	var result stringError
	once := concurrency.NewOnce(func() {
		locker.Lock()
		defer locker.Unlock()
		*filePtr = backup // restore the file before closing the pipe
		closeErr := w.Close()
		strErr, ok := <-c
		if ok {
			result = strErr
		} else {
			result.err = errors.AutoNew("capturer goroutine interrupted")
		}
		if closeErr != nil {
			result.err = errors.Combine(errors.AutoWrap(closeErr), result.err)
		}
	})
	return func() (s string, err error, first bool) {
		first, p := once.DoRecover()
		s, err = result.s, result.err
		if p != nil {
			pe, ok := p.(error)
			if !ok {
				pe = fmt.Errorf("%v", p)
			}
			err = errors.Combine(err, errors.AutoWrap(pe))
		}
		return
	}, nil
}
