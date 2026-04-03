// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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

//nolint:paralleltest // each test in this file captures global variables and therefore cannot be parallel
package local_test

import (
	"fmt"
	"math/rand/v2"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/donyori/gogo/filesys/local"
	"github.com/donyori/gogo/randbytes"
)

func TestCaptureStdoutToString(t *testing.T) {
	testCaptureOutputFileToString(t, &os.Stdout, local.CaptureStdoutToString)
}

func TestCaptureStdoutToString_AlreadyCaptured(t *testing.T) {
	testCaptureOutputFileToStringAlreadyCaptured(
		t,
		&os.Stdout,
		local.CaptureStdoutToString,
	)
}

func TestCaptureStdoutToString_Concurrent(t *testing.T) {
	testCaptureOutputFileToStringConcurrent(
		t,
		&os.Stdout,
		local.CaptureStdoutToString,
	)
}

func TestCaptureStderrToString(t *testing.T) {
	testCaptureOutputFileToString(t, &os.Stderr, local.CaptureStderrToString)
}

func TestCaptureStderrToString_AlreadyCaptured(t *testing.T) {
	testCaptureOutputFileToStringAlreadyCaptured(
		t,
		&os.Stderr,
		local.CaptureStderrToString,
	)
}

func TestCaptureStderrToString_Concurrent(t *testing.T) {
	testCaptureOutputFileToStringConcurrent(
		t,
		&os.Stderr,
		local.CaptureStderrToString,
	)
}

// testCaptureOutputFileToString is the common process of
// TestCaptureStdoutToString and TestCaptureStderrToString.
func testCaptureOutputFileToString( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	filePtr **os.File,
	fn func() (f local.CaptureToStringFunc, err error),
) {
	backup := *filePtr

	const (
		OneKB int = 1 << 10
		OneMB     = OneKB << 10
		OneGB     = OneMB << 10
	)

	// 1GB is much larger than the OS pipe buffer size
	// (e.g., on Linux, the pipe capacity is 16 pages (i.e., 65,536 bytes
	// in a system with a page size of 4096 bytes) since Linux 2.6.11;
	// see <https://man7.org/linux/man-pages/man7/pipe.7.html>).
	// Use 1GB to test whether the OS pipe buffer will overflow.
	buf := randbytes.Make(
		rand.NewChaCha8([32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"))),
		OneGB,
	)

	// Since buf is large, use package unsafe to convert buf to a string
	// to avoid allocation and copying.
	bufStr := unsafe.String(unsafe.SliceData(buf), len(buf))

	for _, length := range []int{0, 5, 300, OneKB, OneMB, OneGB} {
		t.Run(fmt.Sprintf("len=%d", length), func(t *testing.T) {
			testCaptureOutputFileToStringMain(
				t,
				filePtr,
				fn,
				length,
				backup,
				bufStr,
				buf,
			)
		})
	}
}

// testCaptureOutputFileToStringMain is the main process of
// testCaptureOutputFileToString.
func testCaptureOutputFileToStringMain(
	t *testing.T,
	filePtr **os.File,
	fn func() (f local.CaptureToStringFunc, err error),
	length int,
	backup *os.File,
	bufStr string,
	buf []byte,
) {
	t.Helper()

	t.Cleanup(func() {
		*filePtr = backup
	})

	f, err := fn()
	if err != nil {
		t.Error(err)
		return
	} else if f == nil {
		t.Error("got f <nil>")
		return
	}

	var want string
	if length > 0 {
		want = bufStr[:length]

		_, err = (*filePtr).Write(buf[:length])
		if err != nil {
			t.Error(err)
			return
		}
	}

	for i := range 3 {
		s, err, first := f()

		if *filePtr != backup {
			t.Errorf("call %d - file is not restored after calling f", i+1)
		}

		if err != nil {
			t.Errorf("call %d - %v", i+1, err)
			return
		} else if s != want {
			t.Errorf("call %d - got s %q\nwant %q", i+1, s, want)
		}

		if first != (i == 0) {
			t.Errorf("call %d - got first %t; want %t", i+1, first, i == 0)
		}
	}
}

// testCaptureOutputFileToStringAlreadyCaptured is the common process of
// TestCaptureStdoutToString_AlreadyCaptured and
// TestCaptureStderrToString_AlreadyCaptured.
func testCaptureOutputFileToStringAlreadyCaptured(
	t *testing.T,
	filePtr **os.File,
	fn func() (f local.CaptureToStringFunc, err error),
) {
	t.Helper()

	backup := *filePtr

	t.Cleanup(func() {
		*filePtr = backup
	})

	f1, err := fn()
	if err != nil {
		t.Error(err)
		return
	} else if f1 == nil {
		t.Error("got f1 <nil>")
		return
	}
	defer func(f local.CaptureToStringFunc) {
		_, err, _ := f()

		if *filePtr != backup {
			t.Error("file is not restored after calling f")
		}

		if err != nil {
			t.Error(err)
		}
	}(f1)

	const WantErrSuffix = " has been captured elsewhere"

	f2, err := fn()
	if err == nil || !strings.HasSuffix(err.Error(), WantErrSuffix) {
		t.Errorf("got err %v; want one with suffix %q", err, WantErrSuffix)
	}

	if f2 != nil {
		t.Error("got f2 not nil")

		_, _, _ = f2()
	}
}

// testCaptureOutputFileToStringConcurrent is the common process of
// TestCaptureStdoutToString_Concurrent and
// TestCaptureStderrToString_Concurrent.
func testCaptureOutputFileToStringConcurrent( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	filePtr **os.File,
	fn func() (f local.CaptureToStringFunc, err error),
) {
	backup := *filePtr

	const (
		OneKB int = 1 << 10
		OneMB     = OneKB << 10
		OneGB     = OneMB << 10
	)

	// 1GB is much larger than the OS pipe buffer size
	// (e.g., on Linux, the pipe capacity is 16 pages (i.e., 65,536 bytes
	// in a system with a page size of 4096 bytes) since Linux 2.6.11;
	// see <https://man7.org/linux/man-pages/man7/pipe.7.html>).
	// Use 1GB to test whether the OS pipe buffer will overflow.
	buf := randbytes.Make(
		rand.NewChaCha8([32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"))),
		OneGB,
	)

	// Since buf is large, use package unsafe to convert buf to a string
	// to avoid allocation and copying.
	bufStr := unsafe.String(unsafe.SliceData(buf), len(buf))

	for _, length := range []int{0, 5, 300, OneKB, OneMB, OneGB} {
		t.Run(fmt.Sprintf("len=%d", length), func(t *testing.T) {
			testCaptureOutputFileToStringConcurrentMain(
				t,
				filePtr,
				fn,
				length,
				backup,
				bufStr,
				buf,
			)
		})
	}
}

// testCaptureOutputFileToStringConcurrentMain is the main process of
// testCaptureOutputFileToStringConcurrent.
func testCaptureOutputFileToStringConcurrentMain(
	t *testing.T,
	filePtr **os.File,
	fn func() (f local.CaptureToStringFunc, err error),
	length int,
	backup *os.File,
	bufStr string,
	buf []byte,
) {
	t.Helper()

	t.Cleanup(func() {
		*filePtr = backup
	})

	f, err := fn()
	if err != nil {
		t.Error(err)
		return
	} else if f == nil {
		t.Error("got f <nil>")
		return
	}

	var want string
	if length > 0 {
		want = bufStr[:length]

		_, err = (*filePtr).Write(buf[:length])
		if err != nil {
			t.Error(err)
			return
		}
	}

	n := max(runtime.NumCPU(), 4)

	var wg, barrier sync.WaitGroup
	wg.Add(n)
	barrier.Add(n)

	for i := range n {
		go func(rank int) {
			defer wg.Done()

			barrier.Done()
			barrier.Wait()

			s, err, _ := f()

			if *filePtr != backup {
				t.Errorf("goroutine %d - file is not restored after calling f",
					rank)
			}

			if err != nil {
				t.Errorf("goroutine %d - %v", rank, err)
			} else if s != want {
				t.Errorf("goroutine %d - got s %q\nwant %q", rank, s, want)
			}
		}(i)
	}

	wg.Wait()
}
