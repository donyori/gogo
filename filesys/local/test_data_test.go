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

package local_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/donyori/gogo/errors"
)

const TestDataDir = "testdata"

var testFileEntries []fs.DirEntry

func init() {
	entries, err := os.ReadDir(TestDataDir)
	if err != nil {
		panic(err)
	}
	testFileEntries = make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry != nil && !entry.IsDir() {
			testFileEntries = append(testFileEntries, entry)
		}
	}
}

var (
	testDataMap      map[string][]byte
	loadTestDataLock sync.Mutex
)

// lazyLoadTestData loads a file with specified name.
//
// It stores the file contents in the memory the first time reading that file.
// Subsequent reads get the file contents from the memory instead of
// reading the file again.
// Therefore, all modifications to the file after the first read cannot
// take effect on this function.
func lazyLoadTestData(name string) (data []byte, err error) {
	loadTestDataLock.Lock()
	defer loadTestDataLock.Unlock()
	if testDataMap != nil {
		data = testDataMap[name]
		if data != nil {
			return
		}
	}
	data, err = os.ReadFile(name)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	if testDataMap == nil {
		testDataMap = make(map[string][]byte, len(testFileEntries))
	}
	testDataMap[name] = data
	return
}

// lazyLoadTarFile loads a ".tar", ".tgz", ".tar.gz", ".tbz", or ".tar.bz2"
// file with specified name through lazyLoadTestData.
//
// It returns a list of (filename, file body) pairs.
// It also returns any error encountered.
//
// Caller should guarantee that the file name has
// the suffix ".tar", ".tgz", ".tar.gz", ".tbz", or ".tar.bz2".
func lazyLoadTarFile(name string) (files []struct {
	name string
	body []byte
}, err error) {
	data, err := lazyLoadTestData(name)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	var r io.Reader = bytes.NewReader(data)
	ext := filepath.Ext(name)
	switch ext {
	case ".gz", ".tgz":
		gr, err := gzip.NewReader(r)
		if err != nil {
			return nil, errors.AutoWrap(err)
		}
		defer func(gr *gzip.Reader) {
			_ = gr.Close() // ignore error
		}(gr)
		r = gr
	case ".bz2", ".tbz":
		r = bzip2.NewReader(r)
	}
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return files, nil // end of archive
		}
		if err != nil {
			return files, errors.AutoWrap(err)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return files, errors.AutoWrap(err)
		}
		files = append(files, struct {
			name string
			body []byte
		}{hdr.Name, data})
	}
}

// lazyLoadZipFile loads a ".zip" file with specified name
// through lazyLoadTestData.
//
// It returns a map from filenames to (zip header, file body) pairs.
// It also returns any error encountered.
//
// Caller should guarantee that the file name has the suffix ".zip".
func lazyLoadZipFile(name string) (fileMap map[string]*struct {
	header *zip.FileHeader
	body   []byte
}, err error) {
	data, err := lazyLoadTestData(name)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	fileMap = make(map[string]*struct {
		header *zip.FileHeader
		body   []byte
	}, len(zr.File))
	for _, file := range zr.File {
		rc, err := file.Open()
		if err != nil {
			return nil, errors.AutoWrap(err)
		}
		body, err := io.ReadAll(rc)
		if err != nil {
			return nil, errors.AutoWrap(err)
		}
		_ = rc.Close() // ignore error
		fileMap[file.Name] = &struct {
			header *zip.FileHeader
			body   []byte
		}{header: &file.FileHeader, body: body}
	}
	return
}

// lazyCalculateChecksums loads a file with specified name
// through lazyLoadTestData and then calculates its checksums.
//
// newHashes are hash function makers
// (e.g., crypto/sha256.New, crypto.SHA256.New).
//
// The returned checksums correspond to newHashes.
// They are in hexadecimal representation, lowercase.
//
// lazyCalculateChecksums panics if anyone in newHashes is nil or returns nil.
func lazyCalculateChecksums(name string, newHashes ...func() hash.Hash) (checksums []string, err error) {
	data, err := lazyLoadTestData(name)
	if err != nil || len(newHashes) == 0 {
		return nil, errors.AutoWrap(err)
	}
	hs := make([]hash.Hash, len(newHashes))
	ws := make([]io.Writer, len(newHashes))
	for i := range newHashes {
		if newHashes[i] == nil {
			panic(errors.AutoMsg(fmt.Sprintf("newHashes[%d] is nil", i)))
		}
		hs[i] = newHashes[i]()
		if hs[i] == nil {
			panic(errors.AutoMsg(fmt.Sprintf("newHashes[%d] returns nil", i)))
		}
		ws[i] = hs[i]
	}
	w := ws[0]
	if len(ws) > 1 {
		w = io.MultiWriter(ws...)
	}
	_, err = w.Write(data)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	checksums = make([]string, len(hs))
	for i := range hs {
		checksums[i] = hex.EncodeToString(hs[i].Sum(nil))
	}
	return
}
