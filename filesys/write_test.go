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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"path"
	"testing"
	"time"

	"github.com/donyori/gogo/filesys"
)

func TestWrite_Raw(t *testing.T) {
	for _, name := range testFSFilenames {
		t.Run(fmt.Sprintf("file=%q", name), func(t *testing.T) {
			file := &WritableFileImpl{Name: name}
			data := testFS[name].Data
			writeFile(t, file, data, &filesys.WriteOptions{Raw: true})
			if !t.Failed() && !bytes.Equal(file.Data, data) {
				t.Errorf(
					"file content - got (len: %d)\n%s\nwant (len: %d)\n%s",
					len(file.Data),
					file.Data,
					len(data),
					data,
				)
			}
		})
	}
}

func TestWrite_Basic(t *testing.T) {
	for _, name := range testFSBasicFilenames {
		t.Run(fmt.Sprintf("file=%q", name), func(t *testing.T) {
			file := &WritableFileImpl{Name: name}
			data := testFS[name].Data
			writeFile(t, file, data, nil)
			if !t.Failed() && !bytes.Equal(file.Data, data) {
				t.Errorf(
					"file content - got (len: %d)\n%s\nwant (len: %d)\n%s",
					len(file.Data),
					file.Data,
					len(data),
					data,
				)
			}
		})
	}
}

func TestWrite_Gz(t *testing.T) {
	for _, name := range testFSGzFilenames {
		t.Run(fmt.Sprintf("file=%q", name), func(t *testing.T) {
			file := &WritableFileImpl{Name: name}
			data := writeGzFile(t, file)
			if t.Failed() {
				return
			}
			gr, err := gzip.NewReader(bytes.NewReader(file.Data))
			if err != nil {
				t.Fatal("create gzip reader -", err)
			}
			defer func(gr *gzip.Reader) {
				if err := gr.Close(); err != nil {
					t.Error("close gzip reader -", err)
				}
			}(gr)
			got, err := io.ReadAll(gr)
			if err != nil {
				t.Fatal("decompress gzip -", err)
			}
			if !bytes.Equal(got, data) {
				t.Errorf("got (len: %d)\n%s\nwant (len: %d)\n%s",
					len(got), got, len(data), data)
			}
		})
	}
}

func TestWrite_TarTgz(t *testing.T) {
	for _, name := range append(testFSTarFilenames, testFSTgzFilenames...) {
		t.Run(fmt.Sprintf("file=%q", name), func(t *testing.T) {
			file := &WritableFileImpl{Name: name}
			writeTarFiles(t, file)
			if !t.Failed() {
				testTarTgzFile(t, file)
			}
		})
	}
}

func TestWrite_AfterClose(t *testing.T) {
	file := &WritableFileImpl{Name: "test-write-after-close.txt"}
	w, err := filesys.Write(file, nil, true)
	if err != nil {
		t.Fatal("create -", err)
	}
	err = w.Close()
	if err != nil {
		t.Fatal("close -", err)
	}
	_, err = w.WriteString("it should fail")
	if !errors.Is(err, filesys.ErrFileWriterClosed) {
		t.Errorf(
			"errors.Is(err, filesys.ErrFileWriterClosed) is false, err: %v; file content (len: %d): %s",
			err,
			len(file.Data),
			file.Data,
		)
	}
}

// writeFile writes data to file using Write.
//
// It closes file after writing.
func writeFile(t *testing.T, file *WritableFileImpl, data []byte, opts *filesys.WriteOptions) {
	w, err := filesys.Write(file, opts, true)
	if err != nil {
		t.Error("create -", err)
		return
	}
	defer func(w filesys.Writer) {
		if err := w.Close(); err != nil {
			t.Error("close -", err)
		}
	}(w)
	n, err := w.Write(data)
	if n != len(data) || err != nil {
		t.Errorf("write - got (%d, %v); want (%d, nil)", n, err, len(data))
	}
}

// writeGzFile loads data from testFS according to file.Name,
// and then writes the data to file using Write.
//
// It returns the data written to file before gzip compression.
//
// Caller should set file.Name before calling this function and
// guarantee that file.Name has extension ".gz".
func writeGzFile(t *testing.T, file *WritableFileImpl) (data []byte) {
	dataFilename := file.Name[:len(file.Name)-3]
	mapFile := testFS[dataFilename]
	if mapFile == nil {
		t.Errorf("file %q does not exist", dataFilename)
		return
	}
	data = mapFile.Data
	writeFile(t, file, data, nil)
	return
}

// writeTarFiles writes testFSTarFiles to file using Write.
//
// Caller should set file.Name before calling this function and
// guarantee that file.Name has extension ".tar", ".tar.gz", or ".tgz".
func writeTarFiles(t *testing.T, file *WritableFileImpl) {
	w, err := filesys.Write(file, nil, true)
	if err != nil {
		t.Error("create -", err)
		return
	}
	defer func(w filesys.Writer) {
		if err := w.Close(); err != nil {
			t.Error("close -", err)
		}
	}(w)
	for i := range testFSTarFiles {
		err = w.TarWriteHeader(&tar.Header{
			Name:    testFSTarFiles[i].name,
			Size:    int64(len(testFSTarFiles[i].body)),
			Mode:    0600,
			ModTime: time.Now(),
		})
		if err != nil {
			t.Errorf("write No.%d tar header - %v", i, err)
			return
		}
		var n int
		n, err = w.WriteString(testFSTarFiles[i].body)
		if n != len(testFSTarFiles[i].body) || err != nil {
			t.Errorf("write No.%d tar file body - got (%d, %v); want (%d, nil)",
				i, n, err, len(testFSTarFiles[i].body))
			return
		}
	}
}

// testTarTgzFile checks file written by function writeTarFiles.
//
// Caller should guarantee that file.Name has extension
// ".tar", ".tar.gz", or ".tgz".
func testTarTgzFile(t *testing.T, file *WritableFileImpl) {
	var r io.Reader = bytes.NewReader(file.Data)
	ext := path.Ext(file.Name)
	if ext == ".gz" || ext == ".tgz" {
		gr, err := gzip.NewReader(r)
		if err != nil {
			t.Error("create gzip reader -", err)
			return
		}
		defer func(gr *gzip.Reader) {
			if err := gr.Close(); err != nil {
				t.Error("close gzip reader -", err)
			}
		}(gr)
		r = gr
	}
	tr := tar.NewReader(r)
	for i := 0; ; i++ {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if i != len(testFSTarFiles) {
					t.Errorf("tar header number: %d != %d, but got EOF", i, len(testFSTarFiles))
				}
				return // end of archive
			}
			t.Errorf("read No.%d tar header - %v", i, err)
			return
		}
		if i >= len(testFSTarFiles) {
			t.Error("tar headers more than", len(testFSTarFiles))
			return
		}
		if hdr.Name != testFSTarFiles[i].name {
			t.Errorf("No.%d tar header name unequal - got %s; want %s", i, hdr.Name, testFSTarFiles[i].name)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Errorf("read No.%d tar file body - %v", i, err)
			return
		}
		if string(body) != testFSTarFiles[i].body {
			t.Errorf(
				"got No.%d tar file body (len: %d)\n%s\nwant (len: %d)\n%s",
				i,
				len(body),
				body,
				len(testFSTarFiles[i].body),
				testFSTarFiles[i].body,
			)
		}
	}
}
