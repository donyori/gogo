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

package local_test

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/donyori/gogo/filesys/local"
)

func TestCaptureStdoutToString(t *testing.T) {
	testCaptureOutputFileToString(t, &os.Stdout, local.CaptureStdoutToString)
}

func TestCaptureStdoutToString_AlreadyCaptured(t *testing.T) {
	testCaptureOutputFileToStringAlreadyCaptured(
		t, &os.Stdout, local.CaptureStdoutToString)
}

func TestCaptureStdoutToString_Concurrent(t *testing.T) {
	testCaptureOutputFileToStringConcurrent(
		t, &os.Stdout, local.CaptureStdoutToString)
}

func TestCaptureStderrToString(t *testing.T) {
	testCaptureOutputFileToString(t, &os.Stderr, local.CaptureStderrToString)
}

func TestCaptureStderrToString_AlreadyCaptured(t *testing.T) {
	testCaptureOutputFileToStringAlreadyCaptured(
		t, &os.Stderr, local.CaptureStderrToString)
}

func TestCaptureStderrToString_Concurrent(t *testing.T) {
	testCaptureOutputFileToStringConcurrent(
		t, &os.Stderr, local.CaptureStderrToString)
}

// testCaptureOutputFileToString is the common process of
// TestCaptureStdoutToString and TestCaptureStderrToString.
func testCaptureOutputFileToString(
	t *testing.T,
	filePtr **os.File,
	fn func() (f local.CaptureToStringFunc, err error),
) {
	backup := *filePtr
	const (
		OneKB int = 1024
		OneMB     = OneKB * 1024
		OneGB     = OneMB * 1024
	)
	// 1GB is much larger than the OS pipe buffer size
	// (e.g., on Linux, the pipe capacity is 16 pages (i.e., 65,536 bytes
	// in a system with a page size of 4096 bytes) since Linux 2.6.11;
	// see <https://man7.org/linux/man-pages/man7/pipe.7.html>).
	// Use 1GB to test whether the OS pipe buffer will overflow.
	buf := make([]byte, OneGB)
	rand.New(rand.NewSource(20)).Read(buf)
	// Since buf is large, use package unsafe to convert buf to a string
	// to avoid allocation and copying.
	bufStr := unsafe.String(unsafe.SliceData(buf), len(buf))
	for _, length := range []int{0, 5, 300, OneKB, OneMB, OneGB} {
		t.Run(fmt.Sprintf("len=%d", length), func(t *testing.T) {
			t.Cleanup(func() {
				*filePtr = backup
			})
			f, err := fn()
			if err != nil {
				t.Fatal(err)
			} else if f == nil {
				t.Fatal("got f <nil>")
			}
			var want string
			if length > 0 {
				want = bufStr[:length]
				_, err = (*filePtr).Write(buf[:length])
				if err != nil {
					t.Fatal(err)
				}
			}
			for i := 0; i < 3; i++ {
				s, err, first := f()
				if *filePtr != backup {
					t.Errorf("call %d - file is not restored after calling f",
						i+1)
				}
				if err != nil {
					t.Fatalf("call %d - %v", i+1, err)
				} else if s != want {
					t.Errorf("call %d - got s %q\nwant %q", i+1, s, want)
				}
				if first != (i == 0) {
					t.Errorf("call %d - got first %t; want %t",
						i+1, first, i == 0)
				}
			}
		})
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
	backup := *filePtr
	t.Cleanup(func() {
		*filePtr = backup
	})
	f1, err := fn()
	if err != nil {
		t.Fatal(err)
	} else if f1 == nil {
		t.Fatal("got f1 <nil>")
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
		t.Error("got f2 non-nil")
		_, _, _ = f2()
	}
}

// testCaptureOutputFileToStringConcurrent is the common process of
// TestCaptureStdoutToString_Concurrent and
// TestCaptureStderrToString_Concurrent.
func testCaptureOutputFileToStringConcurrent(
	t *testing.T,
	filePtr **os.File,
	fn func() (f local.CaptureToStringFunc, err error),
) {
	backup := *filePtr
	const (
		OneKB int = 1024
		OneMB     = OneKB * 1024
		OneGB     = OneMB * 1024
	)
	// 1GB is much larger than the OS pipe buffer size
	// (e.g., on Linux, the pipe capacity is 16 pages (i.e., 65,536 bytes
	// in a system with a page size of 4096 bytes) since Linux 2.6.11;
	// see <https://man7.org/linux/man-pages/man7/pipe.7.html>).
	// Use 1GB to test whether the OS pipe buffer will overflow.
	buf := make([]byte, OneGB)
	rand.New(rand.NewSource(20)).Read(buf)
	// Since buf is large, use package unsafe to convert buf to a string
	// to avoid allocation and copying.
	bufStr := unsafe.String(unsafe.SliceData(buf), len(buf))
	for _, length := range []int{0, 5, 300, OneKB, OneMB, OneGB} {
		t.Run(fmt.Sprintf("len=%d", length), func(t *testing.T) {
			t.Cleanup(func() {
				*filePtr = backup
			})
			f, err := fn()
			if err != nil {
				t.Fatal(err)
			} else if f == nil {
				t.Fatal("got f <nil>")
			}
			var want string
			if length > 0 {
				want = bufStr[:length]
				_, err = (*filePtr).Write(buf[:length])
				if err != nil {
					t.Fatal(err)
				}
			}
			n := runtime.NumCPU()
			if n < 4 {
				n = 4
			}
			var wg, barrier sync.WaitGroup
			wg.Add(n)
			barrier.Add(n)
			for i := 0; i < n; i++ {
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
						t.Errorf("goroutine %d - got s %q\nwant %q",
							rank, s, want)
					}
				}(i)
			}
			wg.Wait()
		})
	}
}
