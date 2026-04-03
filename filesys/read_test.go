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

package filesys_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"math"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/internal/unequal"
)

func TestRead_NotCloseFile(t *testing.T) {
	t.Parallel()

	const Name = "file1.txt"

	file, err := testFS.Open(Name)
	if err != nil {
		t.Fatal("open file -", err)
	}
	defer func(f fs.File) {
		err := f.Close()
		if err != nil {
			t.Error("close file -", err)
		}
	}(file)

	r, err := filesys.Read(file, &filesys.ReadOptions{Raw: true}, false)
	if err != nil {
		t.Fatal("create reader -", err)
	}

	halfSize := int64(len(testFS[Name].Data) >> 1)
	hr := io.LimitReader(r, halfSize)

	err = iotest.TestReader(hr, testFS[Name].Data[:halfSize])
	if err != nil {
		t.Error("test read a half -", err)
	}

	var buffered []byte

	if n := r.Buffered(); n > 0 {
		peek, err := r.Peek(n)
		if err != nil {
			t.Fatal("peek buffered data -", err)
		}

		buffered = make([]byte, len(peek), len(peek)+int(halfSize))
		copy(buffered, peek)
	}

	err = r.Close()
	if err != nil {
		t.Fatal("close reader -", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal("read all from rest part -", err)
	}

	data = append(buffered, data...)
	want := testFS[Name].Data[halfSize:]

	if !bytes.Equal(data, want) {
		t.Errorf("read all from rest part - got %s; want %s", data, want)
	}
}

func TestRead_NotCloseFile_ErrorOnCreate(t *testing.T) {
	t.Parallel()

	const Name = "file1.txt"

	file, err := testFS.Open(Name)
	if err != nil {
		t.Fatal("open file -", err)
	}
	defer func(f fs.File) {
		err := f.Close()
		if err != nil {
			t.Error("close file -", err)
		}
	}(file)

	r, err := filesys.Read(file, &filesys.ReadOptions{
		Offset: math.MaxInt64,
		Raw:    true,
	}, false)
	if err == nil {
		_ = r.Close()

		t.Fatal("create reader - no error but offset is out of range")
	} else if !strings.HasSuffix(err.Error(), fmt.Sprintf(
		"option Offset (%d) is out of range; file size: %d",
		math.MaxInt64,
		len(testFS[Name].Data),
	)) {
		t.Fatal("create reader -", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal("read all from file -", err)
	}

	want := testFS[Name].Data
	if !bytes.Equal(data, want) {
		t.Errorf("read all from file - got %s; want %s", data, want)
	}
}

func TestReadFromFS_Raw(t *testing.T) {
	t.Parallel()

	for _, name := range testFSFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			t.Parallel()

			testReadFromTestFS(
				t,
				name,
				testFS[name].Data,
				&filesys.ReadOptions{Raw: true},
			)
		})
	}
}

func TestReadFromFS_Basic(t *testing.T) {
	t.Parallel()

	for _, name := range testFSBasicFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			t.Parallel()

			testReadFromTestFS(t, name, testFS[name].Data, nil)
		})
	}
}

func TestReadFromFS_Gz(t *testing.T) {
	t.Parallel()

	for _, name := range testFSGzFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			t.Parallel()

			dataFilename := name[:len(name)-3]
			mapFile := testFS[dataFilename]

			if mapFile == nil {
				t.Fatalf("file %q does not exist", dataFilename)
			}

			testReadFromTestFS(t, name, mapFile.Data, nil)
		})
	}
}

func TestReadFromFS_TarTgz(t *testing.T) {
	t.Parallel()

	for _, name := range append(testFSTarFilenames, testFSTgzFilenames...) {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			t.Parallel()

			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				err := r.Close()
				if err != nil {
					t.Error("close -", err)
				}
			}(r)

			testReadFromFSTarTgzCheckTarEntries(t, r)
		})
	}
}

// testReadFromFSTarTgzCheckTarEntries is a subprocess of TestReadFromFS_TarTgz
// that checks entries in the tar archive.
func testReadFromFSTarTgzCheckTarEntries(t *testing.T, r filesys.Reader) {
	t.Helper()

	for i := 0; ; i++ {
		hdr, err := r.TarNext()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if i != len(testFSTarFiles) {
					t.Errorf("tar header number: %d != %d, but got EOF",
						i, len(testFSTarFiles))
				}

				return // end of archive
			} else if !errors.Is(err, tar.ErrInsecurePath) ||
				hdr == nil || hdr.Name != testFSTarNonLocalFilename {
				t.Errorf("read No.%d tar header - %v", i, err)
				return
			}
		}

		if testTarFile(t, r, i, hdr) {
			return
		}
	}
}

func TestReadFromFS_TarTgz_Seq(t *testing.T) {
	t.Parallel()

	for _, name := range append(testFSTarFilenames, testFSTgzFilenames...) {
		for _, acceptNonLocalNames := range []bool{false, true} {
			t.Run(
				fmt.Sprintf(
					"file=%+q&acceptNonLocalNames=%t",
					name,
					acceptNonLocalNames,
				),
				func(t *testing.T) {
					t.Parallel()

					r, err := filesys.ReadFromFS(testFS, name, nil)
					if err != nil {
						t.Fatal("create -", err)
					}
					defer func(r filesys.Reader) {
						err := r.Close()
						if err != nil {
							t.Error("close -", err)
						}
					}(r)

					outErr := errors.New("init error") // initialize as a non-nil error

					seq, err := r.IterTarFiles(&outErr, acceptNonLocalNames)
					if err != nil {
						t.Fatal("create iterator -", err)
					}

					testTarSeq(t, r, &outErr, acceptNonLocalNames, seq)
				},
			)
		}
	}
}

// testTarSeq is the main process of TestReadFromFS_TarTgz_Seq.
func testTarSeq(
	t *testing.T,
	r filesys.Reader,
	pErr *error,
	acceptNonLocalNames bool,
	seq iter.Seq[*tar.Header],
) {
	t.Helper()

	if seq == nil {
		t.Error("got nil iterator")
		return
	}

	var i int
	for hdr := range seq {
		if testTarFile(t, r, i, hdr) {
			return
		} else if !acceptNonLocalNames &&
			hdr != nil && hdr.Name == testFSTarNonLocalFilename {
			t.Errorf("No.%d encountered non-local name %q", i, hdr.Name)
		}

		i++
	}

	if acceptNonLocalNames {
		if *pErr != nil {
			t.Errorf("iteration ended with %v; want <nil>", *pErr)
		}
	} else if !errors.Is(*pErr, tar.ErrInsecurePath) {
		t.Errorf("iteration ended with %v; want %v", *pErr, tar.ErrInsecurePath)
	}

	if *pErr == nil && i != len(testFSTarFiles) {
		t.Errorf("tar header number: %d != %d, but iteration has ended",
			i, len(testFSTarFiles))
	}

	// Test whether the iterator is single-use.
	testTarSeqSingleUse(t, pErr, seq)
}

// testTarSeqSingleUse is a subprocess of testTarSeq
// that tests whether the iterator is single-use.
func testTarSeqSingleUse(t *testing.T, pErr *error, seq iter.Seq[*tar.Header]) {
	t.Helper()

	prevErr := *pErr

	for hdr := range seq {
		if hdr != nil {
			t.Errorf("not single-use iterator; got %q", hdr.Name)
		} else {
			t.Error("not single-use iterator; got <nil>")
		}

		break
	}

	if unequal.ErrorUnwrapAuto(*pErr, prevErr) {
		t.Errorf("output error changed from %v to %v", prevErr, *pErr)
	}
}

func TestReadFromFS_TarTgz_Seq2(t *testing.T) {
	t.Parallel()

	for _, name := range append(testFSTarFilenames, testFSTgzFilenames...) {
		for _, acceptNonLocalNames := range []bool{false, true} {
			t.Run(
				fmt.Sprintf(
					"file=%+q&acceptNonLocalNames=%t",
					name,
					acceptNonLocalNames,
				),
				func(t *testing.T) {
					t.Parallel()

					r, err := filesys.ReadFromFS(testFS, name, nil)
					if err != nil {
						t.Fatal("create -", err)
					}
					defer func(r filesys.Reader) {
						err := r.Close()
						if err != nil {
							t.Error("close -", err)
						}
					}(r)

					outErr := errors.New("init error") // initialize as a non-nil error

					seq2, err := r.IterIndexTarFiles(
						&outErr,
						acceptNonLocalNames,
					)
					if err != nil {
						t.Fatal("create iterator -", err)
					}

					testTarSeq2(t, r, &outErr, acceptNonLocalNames, seq2)
				},
			)
		}
	}
}

// testTarSeq2 is the main process of TestReadFromFS_TarTgz_Seq2.
func testTarSeq2(
	t *testing.T,
	r filesys.Reader,
	pErr *error,
	acceptNonLocalNames bool,
	seq2 iter.Seq2[int, *tar.Header],
) {
	t.Helper()

	if seq2 == nil {
		t.Error("got nil iterator")
		return
	}

	var ctr int
	for i, hdr := range seq2 {
		if i != ctr {
			t.Errorf("got index %d; want %d", i, ctr)
		}

		if testTarFile(t, r, ctr, hdr) {
			return
		} else if !acceptNonLocalNames &&
			hdr != nil && hdr.Name == testFSTarNonLocalFilename {
			t.Errorf("No.%d encountered non-local name %q", ctr, hdr.Name)
		}

		ctr++
	}

	if acceptNonLocalNames {
		if *pErr != nil {
			t.Errorf("iteration ended with %v; want <nil>", *pErr)
		}
	} else if !errors.Is(*pErr, tar.ErrInsecurePath) {
		t.Errorf("iteration ended with %v; want %v", *pErr, tar.ErrInsecurePath)
	}

	if *pErr == nil && ctr != len(testFSTarFiles) {
		t.Errorf("tar header number: %d != %d, but iteration has ended",
			ctr, len(testFSTarFiles))
	}

	// Test whether the iterator is single-use.
	testTarSeq2SingleUse(t, pErr, seq2)
}

// testTarSeq2SingleUse is a subprocess of testTarSeq2
// that tests whether the iterator is single-use.
func testTarSeq2SingleUse(
	t *testing.T,
	pErr *error,
	seq2 iter.Seq2[int, *tar.Header],
) {
	t.Helper()

	prevErr := *pErr

	for i, hdr := range seq2 {
		if hdr != nil {
			t.Errorf("not single-use iterator; got %d, %q", i, hdr.Name)
		} else {
			t.Errorf("not single-use iterator; got %d, <nil>", i)
		}

		break
	}

	if unequal.ErrorUnwrapAuto(*pErr, prevErr) {
		t.Errorf("output error changed from %v to %v", prevErr, *pErr)
	}
}

func TestReadFromFS_Zip_Open(t *testing.T) {
	t.Parallel()

	for _, name := range testFSZipFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			t.Parallel()

			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}

			t.Cleanup(func() {
				err := r.Close()
				if err != nil {
					t.Error("close -", err)
				}
			})

			testZipOpen(t, r)
		})
	}
}

func TestReadFromFS_Zip_Files(t *testing.T) {
	t.Parallel()

	for _, name := range testFSZipFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			t.Parallel()

			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}

			t.Cleanup(func() {
				err := r.Close()
				if err != nil {
					t.Error("close -", err)
				}
			})

			testZipFiles(t, r)
		})
	}
}

func TestReadFromFS_Zip_Seq(t *testing.T) {
	t.Parallel()

	for _, name := range testFSZipFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			t.Parallel()

			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				err := r.Close()
				if err != nil {
					t.Error("close -", err)
				}
			}(r)

			testZipSeqCommon(t, r)
		})
	}
}

func TestReadFromFS_Zip_Seq2(t *testing.T) {
	t.Parallel()

	for _, name := range testFSZipFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			t.Parallel()

			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				err := r.Close()
				if err != nil {
					t.Error("close -", err)
				}
			}(r)

			testZipSeq2Common(t, r)
		})
	}
}

func TestReadFromFS_Zip_Comment(t *testing.T) {
	t.Parallel()

	for _, name := range testFSZipFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			t.Parallel()

			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				err := r.Close()
				if err != nil {
					t.Error("close -", err)
				}
			}(r)

			comment, err := r.ZipComment()
			if err != nil {
				t.Error("zip comment -", err)
			} else if comment != testFSZipComment {
				t.Errorf("got %q; want %q", comment, testFSZipComment)
			}
		})
	}
}

func TestReadFromFS_Zip_Offset_Open(t *testing.T) {
	t.Parallel()

	r, err := filesys.ReadFromFS(
		testFS,
		testFSZipOffsetName,
		&filesys.ReadOptions{Offset: testFSZipOffset},
	)
	if err != nil {
		t.Fatal("create -", err)
	}

	t.Cleanup(func() {
		err := r.Close()
		if err != nil {
			t.Error("close -", err)
		}
	})

	testZipOpen(t, r)
}

func TestReadFromFS_Zip_Offset_Files(t *testing.T) {
	t.Parallel()

	r, err := filesys.ReadFromFS(
		testFS,
		testFSZipOffsetName,
		&filesys.ReadOptions{Offset: testFSZipOffset},
	)
	if err != nil {
		t.Fatal("create -", err)
	}

	t.Cleanup(func() {
		err := r.Close()
		if err != nil {
			t.Error("close -", err)
		}
	})

	testZipFiles(t, r)
}

func TestReadFromFS_Zip_Offset_Seq(t *testing.T) {
	t.Parallel()

	r, err := filesys.ReadFromFS(
		testFS,
		testFSZipOffsetName,
		&filesys.ReadOptions{Offset: testFSZipOffset},
	)
	if err != nil {
		t.Fatal("create -", err)
	}
	defer func(r filesys.Reader) {
		err := r.Close()
		if err != nil {
			t.Error("close -", err)
		}
	}(r)

	testZipSeqCommon(t, r)
}

func TestReadFromFS_Zip_Offset_Seq2(t *testing.T) {
	t.Parallel()

	r, err := filesys.ReadFromFS(
		testFS,
		testFSZipOffsetName,
		&filesys.ReadOptions{Offset: testFSZipOffset},
	)
	if err != nil {
		t.Fatal("create -", err)
	}
	defer func(r filesys.Reader) {
		err := r.Close()
		if err != nil {
			t.Error("close -", err)
		}
	}(r)

	testZipSeq2Common(t, r)
}

func TestReadFromFS_Zip_Offset_Comment(t *testing.T) {
	t.Parallel()

	r, err := filesys.ReadFromFS(
		testFS,
		testFSZipOffsetName,
		&filesys.ReadOptions{Offset: testFSZipOffset},
	)
	if err != nil {
		t.Fatal("create -", err)
	}
	defer func(r filesys.Reader) {
		err := r.Close()
		if err != nil {
			t.Error("close -", err)
		}
	}(r)

	comment, err := r.ZipComment()
	if err != nil {
		t.Error("zip comment -", err)
	} else if comment != testFSZipComment {
		t.Errorf("got %q; want %q", comment, testFSZipComment)
	}
}

func TestReadFromFS_Offset(t *testing.T) {
	t.Parallel()

	const Name = "file1.txt"

	fileData := testFS[Name].Data
	size := int64(len(fileData))

	for offset := -size; offset <= size; offset++ {
		var pos int64
		if offset > 0 {
			pos = offset
		} else if offset < 0 {
			pos = size + offset
		} else {
			continue
		}

		t.Run(fmt.Sprintf("offset=%d", offset), func(t *testing.T) {
			t.Parallel()

			testReadFromTestFS(t, Name, fileData[pos:], &filesys.ReadOptions{
				Offset: offset,
				Raw:    true,
			})
		})
	}

	for _, offset := range []int64{
		math.MinInt64,
		-size - 1,
		size + 1,
		math.MaxInt64,
	} {
		t.Run(fmt.Sprintf("offset=%d", offset), func(t *testing.T) {
			t.Parallel()

			r, err := filesys.ReadFromFS(testFS, Name, &filesys.ReadOptions{
				Offset: offset,
				Raw:    true,
			})
			if err == nil {
				_ = r.Close() // ignore error

				t.Fatal("create - no error but offset is out of range")
			}

			if !strings.HasSuffix(err.Error(), fmt.Sprintf(
				"option Offset (%d) is out of range; file size: %d",
				offset,
				size,
			)) {
				t.Error("create -", err)
			}
		})
	}
}

const (
	readFromFSAfterCloseRegFile = "file1.txt"
	readFromFSAfterCloseTarFile = "tar file.tar"
	readFromFSAfterCloseZipFile = "zip basic.zip"
)

var readFromFSAfterCloseTestCases = []struct {
	methodName string
	filename   string
	f          func(t *testing.T, r filesys.Reader) error
	wantErr    error
}{
	{
		"Close",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			return r.Close()
		},
		nil,
	},
	{
		"Closed",
		readFromFSAfterCloseRegFile,
		func(t *testing.T, r filesys.Reader) error {
			t.Helper()

			if !r.Closed() {
				t.Error("r.Closed - got false; want true")
			}

			return nil
		},
		nil,
	},
	{
		"Read",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.Read(nil)
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"ReadByte",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.ReadByte()
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"UnreadByte",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			return r.UnreadByte()
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"ReadRune",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, _, err := r.ReadRune()
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"UnreadRune",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			return r.UnreadRune()
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"WriteTo",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.WriteTo(nil)
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"ReadLine",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, _, err := r.ReadLine()
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"ReadEntireLine",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.ReadEntireLine()
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"WriteLineTo",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.WriteLineTo(nil)
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"Peek",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.Peek(0)
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"Discard",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.Discard(0)
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"TarNext-notTar",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.TarNext()
			return err
		},
		filesys.ErrNotTar,
	},
	{
		"TarNext-isTar",
		readFromFSAfterCloseTarFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.TarNext()
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"IterTarFiles-notTar",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.IterTarFiles(nil, false)
			return err
		},
		filesys.ErrNotTar,
	},
	{
		"IterTarFiles-isTar",
		readFromFSAfterCloseTarFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.IterTarFiles(nil, false)
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"IterIndexTarFiles-notTar",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.IterIndexTarFiles(nil, false)
			return err
		},
		filesys.ErrNotTar,
	},
	{
		"IterIndexTarFiles-isTar",
		readFromFSAfterCloseTarFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.IterIndexTarFiles(nil, false)
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"ZipOpen-notZip",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.ZipOpen("")
			return err
		},
		filesys.ErrNotZip,
	},
	{
		"ZipOpen-isZip",
		readFromFSAfterCloseZipFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.ZipOpen("")
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"ZipFiles-notZip",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.ZipFiles()
			return err
		},
		filesys.ErrNotZip,
	},
	{
		"ZipFiles-isZip",
		readFromFSAfterCloseZipFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.ZipFiles()
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"IterZipFiles-notZip",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.IterZipFiles()
			return err
		},
		filesys.ErrNotZip,
	},
	{
		"IterZipFiles-isZip",
		readFromFSAfterCloseZipFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.IterZipFiles()
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"IterIndexZipFiles-notZip",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.IterIndexZipFiles()
			return err
		},
		filesys.ErrNotZip,
	},
	{
		"IterIndexZipFiles-isZip",
		readFromFSAfterCloseZipFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.IterIndexZipFiles()
			return err
		},
		filesys.ErrFileReaderClosed,
	},
	{
		"ZipComment-notZip",
		readFromFSAfterCloseRegFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.ZipComment()
			return err
		},
		filesys.ErrNotZip,
	},
	{
		"ZipComment-isZip",
		readFromFSAfterCloseZipFile,
		func(_ *testing.T, r filesys.Reader) error {
			_, err := r.ZipComment()
			return err
		},
		filesys.ErrFileReaderClosed,
	},
}

func TestReadFromFS_AfterClose(t *testing.T) {
	t.Parallel()

	methodNameSet := make(
		map[string]struct{},
		len(readFromFSAfterCloseTestCases),
	)
	for _, tc := range readFromFSAfterCloseTestCases {
		if _, ok := methodNameSet[tc.methodName]; ok {
			t.Errorf("duplicate test cases for %q", tc.methodName)
		}

		methodNameSet[tc.methodName] = struct{}{}
	}

	if t.Failed() {
		return
	}

	for _, tc := range readFromFSAfterCloseTestCases {
		t.Run("method="+tc.methodName, func(t *testing.T) {
			t.Parallel()

			r, err := filesys.ReadFromFS(testFS, tc.filename, nil)
			if err != nil {
				t.Error("create -", err)
				return
			}

			err = r.Close()
			if err != nil {
				t.Error("close -", err)
				return
			}

			err = tc.f(t, r)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got error %v; want %v", err, tc.wantErr)
			}
		})
	}
}

// testReadFromTestFS reads the file with specified name from testFS
// using ReadFromFS and tests the reader using iotest.TestReader.
//
// want is the expected data read from the file.
func testReadFromTestFS(
	t *testing.T,
	name string,
	want []byte,
	opts *filesys.ReadOptions,
) {
	t.Helper()

	r, err := filesys.ReadFromFS(testFS, name, opts)
	if err != nil {
		t.Error("create -", err)
		return
	}
	defer func(r filesys.Reader) {
		err := r.Close()
		if err != nil {
			t.Error("close -", err)
		}
	}(r)

	err = iotest.TestReader(r, want)
	if err != nil {
		t.Error("test read -", err)
	}
}

// testTarFile tests the i-th tar file.
//
// It reports whether to stop testing.
func testTarFile(
	t *testing.T,
	r filesys.Reader,
	i int,
	hdr *tar.Header,
) (stop bool) {
	t.Helper()

	switch {
	case i >= len(testFSTarFiles):
		t.Error("tar headers more than", len(testFSTarFiles))
		return true
	case hdr == nil:
		t.Errorf("No.%d got nil tar header", i)
		return false
	case hdr.Name != testFSTarFiles[i].name:
		t.Errorf("No.%d tar header name unequal - got %q; want %q",
			i, hdr.Name, testFSTarFiles[i].name)
	}

	if filesys.TarHeaderIsDir(hdr) {
		_, err := r.Read([]byte{})
		if !errors.Is(err, filesys.ErrIsDir) {
			t.Errorf("No.%d tar read file body - got %v; want %v",
				i, err, filesys.ErrIsDir)
		}
	} else {
		err := iotest.TestReader(r, []byte(testFSTarFiles[i].body))
		if err != nil {
			t.Errorf("No.%d tar test read - %v", i, err)
		}
	}

	return false
}

// testZipOpen tests filesys.Reader.ZipOpen.
func testZipOpen(t *testing.T, r filesys.Reader) { //nolint:thelper // this is the main body, not a helper
	dirSet := make(map[string]struct{}, len(testFSZipFileNameBodyMap))
	for zipFilename, body := range testFSZipFileNameBodyMap {
		for i := 1; i < len(zipFilename)-1; i++ {
			if zipFilename[i-1] != '/' && zipFilename[i] == '/' {
				dirSet[zipFilename[:i]] = struct{}{}
			}
		}

		if len(zipFilename) > 1 &&
			zipFilename[len(zipFilename)-2] != '/' &&
			zipFilename[len(zipFilename)-1] == '/' {
			dirSet[zipFilename[:len(zipFilename)-1]] = struct{}{}
			continue
		}

		testZipOpenRegularFile(t, r, zipFilename, body)
	}

	for dir := range dirSet {
		testZipOpenDirectory(t, r, dir)
	}
}

// testZipOpenRegularFile is the subtest of testZipOpen
// to test opening a regular file in the ZIP archive.
func testZipOpenRegularFile( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	r filesys.Reader,
	zipFilename string,
	body string,
) {
	t.Run(fmt.Sprintf("zipFile=%+q", zipFilename), func(t *testing.T) {
		t.Parallel()

		file, err := r.ZipOpen(zipFilename)
		if err != nil {
			t.Fatal("open -", err)
		}
		defer func(f fs.File) {
			err := f.Close()
			if err != nil {
				t.Error("close -", err)
			}
		}(file)

		info, err := file.Stat()
		if err != nil {
			t.Fatal("stat -", err)
		} else if d := info.IsDir(); d {
			t.Errorf("got IsDir %t; want false", d)
		}

		data, err := io.ReadAll(file)
		if err != nil {
			t.Error("read -", err)
		} else if string(data) != body {
			t.Errorf("file contents - got (len: %d)\n%s\nwant (len: %d)\n%s",
				len(data), data, len(body), body)
		}
	})
}

// testZipOpenDirectory is the subtest of testZipOpen
// to test opening a directory in the ZIP archive.
func testZipOpenDirectory( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	r filesys.Reader,
	dir string,
) {
	t.Run(fmt.Sprintf("zipFile=%+q", dir), func(t *testing.T) {
		t.Parallel()

		file, err := r.ZipOpen(dir)
		if err != nil {
			t.Fatal("open -", err)
		}
		defer func(f fs.File) {
			err := f.Close()
			if err != nil {
				t.Error("close -", err)
			}
		}(file)

		info, err := file.Stat()
		if err != nil {
			t.Error("stat -", err)
		} else if d := info.IsDir(); !d {
			t.Errorf("got IsDir %t; want true", d)
		}
	})
}

// testZipFiles tests filesys.Reader.ZipFiles.
func testZipFiles(t *testing.T, r filesys.Reader) { //nolint:thelper // this is the main body, not a helper
	files, err := r.ZipFiles()
	if err != nil {
		t.Error("ZipFiles -", err)
		return
	} else if len(files) != len(testFSZipFileNameBodyMap) {
		t.Errorf("got %d zip files; want %d",
			len(files), len(testFSZipFileNameBodyMap))
	}

	for i, file := range files {
		if file == nil {
			t.Errorf("No.%d zip file is nil", i)
			continue
		}

		t.Run(fmt.Sprintf("zipFile=%+q", file.Name), func(t *testing.T) {
			t.Parallel()

			testZipFilesMain(t, file)
		})
	}
}

// testZipFilesMain is the main process of testZipFiles.
func testZipFilesMain(t *testing.T, file *zip.File) {
	t.Helper()

	body, ok := testFSZipFileNameBodyMap[file.Name]
	if !ok {
		t.Errorf("unknown zip file %q", file.Name)
		return
	}

	isDir := len(file.Name) > 0 && file.Name[len(file.Name)-1] == '/'
	if d := file.FileInfo().IsDir(); d != isDir {
		t.Errorf("got IsDir %t; want %t", d, isDir)
	}

	if isDir {
		return
	}

	rc, err := file.Open()
	if err != nil {
		t.Error("open -", err)
		return
	}
	defer func(rc io.ReadCloser) {
		err := rc.Close()
		if err != nil {
			t.Error("close -", err)
		}
	}(rc)

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Error("read -", err)
	} else if string(data) != body {
		t.Errorf("file contents - got (len: %d)\n%s\nwant (len: %d)\n%s",
			len(data), data, len(body), body)
	}
}

// testZipSeqCommon is the common process of
// TestReadFromFS_Zip_Seq and TestReadFromFS_Zip_Offset_Seq.
func testZipSeqCommon(t *testing.T, r filesys.Reader) {
	t.Helper()

	files, err := r.ZipFiles()
	if err != nil {
		t.Error("ZipFiles -", err)
		return
	}

	seq, err := r.IterZipFiles()
	if err != nil {
		t.Error("create iterator -", err)
		return
	} else if seq == nil {
		t.Error("got nil iterator")
		return
	}

	testZipSeqCommonSub(t, files, seq, false)

	// Rewind the iterator and test it again.
	testZipSeqCommonSub(t, files, seq, true)
}

// testZipSeqCommonSub is a subprocess of testZipSeqCommon
// that tests the iterator with a single call.
func testZipSeqCommonSub(
	t *testing.T,
	files []*zip.File,
	seq iter.Seq[*zip.File],
	isRewound bool,
) {
	t.Helper()

	var prefix string
	if isRewound {
		prefix = "rewind - "
	}

	var i int
	for file := range seq {
		if i >= len(files) {
			t.Errorf("%szip files more than %d", prefix, len(files))
			return
		} else if file != files[i] {
			const NilFileStr = "<nil>"

			file1Str := NilFileStr
			if file != nil {
				file1Str = fmt.Sprintf("%q(%p)", file.Name, file)
			}

			file2Str := NilFileStr
			if files[i] != nil {
				file2Str = fmt.Sprintf("%q(%p)", files[i].Name, files[i])
			}

			t.Errorf("%sNo.%d got %s; want %s", prefix, i, file1Str, file2Str)
		}

		i++
	}

	if i != len(files) {
		t.Errorf("%szip file number: %d != %d, but iteration has ended",
			prefix, i, len(files))
	}
}

// testZipSeq2Common is the common process of
// TestReadFromFS_Zip_Seq2 and TestReadFromFS_Zip_Offset_Seq2.
func testZipSeq2Common(t *testing.T, r filesys.Reader) {
	t.Helper()

	files, err := r.ZipFiles()
	if err != nil {
		t.Error("ZipFiles -", err)
		return
	}

	seq2, err := r.IterIndexZipFiles()
	if err != nil {
		t.Error("create iterator -", err)
		return
	} else if seq2 == nil {
		t.Error("got nil iterator")
		return
	}

	testZipSeq2CommonSub(t, files, seq2, false)

	// Rewind the iterator and test it again.
	testZipSeq2CommonSub(t, files, seq2, true)
}

// testZipSeq2CommonSub is a subprocess of testZipSeq2Common
// that tests the iterator with a single call.
func testZipSeq2CommonSub(
	t *testing.T,
	files []*zip.File,
	seq2 iter.Seq2[int, *zip.File],
	isRewound bool,
) {
	t.Helper()

	var prefix string
	if isRewound {
		prefix = "rewind - "
	}

	var ctr int
	for i, file := range seq2 {
		if i != ctr {
			t.Errorf("%sgot index %d; want %d", prefix, i, ctr)
		}

		if ctr >= len(files) {
			t.Errorf("%szip files more than %d", prefix, len(files))
			return
		} else if file != files[ctr] {
			const NilFileStr = "<nil>"

			file1Str := NilFileStr
			if file != nil {
				file1Str = fmt.Sprintf("%q(%p)", file.Name, file)
			}

			file2Str := NilFileStr
			if files[ctr] != nil {
				file2Str = fmt.Sprintf("%q(%p)", files[ctr].Name, files[ctr])
			}

			t.Errorf("%sNo.%d got %s; want %s", prefix, ctr, file1Str, file2Str)
		}

		ctr++
	}

	if ctr != len(files) {
		t.Errorf("%szip file number: %d != %d, but iteration has ended",
			prefix, ctr, len(files))
	}
}
