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

package filesys_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto"
	_ "crypto/md5"    // link crypto.MD5 to the binary
	_ "crypto/sha256" // link crypto.SHA256 to the binary
	"encoding/hex"
	"io"
	"math/rand"
	"path"
	"sort"
	"testing/fstest"
	"time"

	"github.com/donyori/gogo/errors"
)

var (
	testFS fstest.MapFS

	testFSTarFiles           []struct{ name, body string }
	testFSZipFileNameBodyMap map[string]string

	testFSZipOffset int64

	testFSFilenames      []string
	testFSBasicFilenames []string
	testFSGzFilenames    []string
	testFSTarFilenames   []string
	testFSTgzFilenames   []string
	testFSZipFilenames   []string

	testFSChecksumMap map[string][]struct {
		hash     crypto.Hash
		checksum string
	}
)

const testFSZipOffsetName = "zip offset.zip"

const testFSZipComment = "The end-of-central-directory comment 你好"

func init() {
	testFS = fstest.MapFS{
		"file1.txt": {
			Data:    []byte("This is File 1."),
			Mode:    0755,
			ModTime: time.Now(),
		},
		"file2.txt": {
			Data:    []byte("Here is File 2!"),
			Mode:    0755,
			ModTime: time.Now(),
		},
		"roses are red.txt": {
			Data: []byte(`Roses are red.
  Violets are blue.
Sugar is sweet.
  And so are you.
`),
			Mode:    0755,
			ModTime: time.Now(),
		},
	}

	big := make([]byte, 13<<10)
	random := rand.New(rand.NewSource(100))
	random.Read(big)
	testFS["13KB.dat"] = &fstest.MapFile{
		Data:    big,
		Mode:    0755,
		ModTime: time.Now(),
	}

	bigStr := string(big)
	testFSTarFiles = []struct{ name, body string }{
		{"tardir/", ""},
		{"tardir/tar file1.txt", "This is tar file 1."},
		{"tardir/tar file2.txt", "Here is tar file 2!"},
		{"emptydir/", ""},
		{"roses are red.txt", `Roses are red.
  Violets are blue.
Sugar is sweet.
  And so are you.
`},
		{"13KB.dat", bigStr},
	}
	testFSZipFileNameBodyMap = map[string]string{
		"zipdir/":              "",
		"zipdir/zip file1.txt": "This is ZIP file 1.",
		"zipdir/zip file2.txt": "Here is ZIP file 2!",
		"emptydir/":            "",
		"roses are red.txt": `Roses are red.
  Violets are blue.
Sugar is sweet.
  And so are you.
`,
		"13KB.dat": bigStr,
	}

	buf := new(bytes.Buffer)
	err := initAddGzFile(buf, big)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	err = initAddTarFile(buf)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	err = initAddTgzFile(buf)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	err = initAddZipFile(buf)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	err = initAddZipFileWithOffset(buf, random)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	initRecordFilenames()
	err = initMakeFSChecksumMap()
	if err != nil {
		panic(errors.AutoWrap(err))
	}
}

// initAddGzFile makes a gzip file and adds it to the global variable testFS.
func initAddGzFile(buf *bytes.Buffer, data []byte) error {
	gzw, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		return errors.AutoWrap(err)
	}
	_, err = gzw.Write(data)
	if err != nil {
		return errors.AutoWrap(err)
	}
	err = gzw.Close()
	if err != nil {
		return errors.AutoWrap(err)
	}
	testFS["13KB.dat.gz"] = &fstest.MapFile{
		Data:    copyBuffer(buf),
		Mode:    0755,
		ModTime: time.Now(),
	}
	return nil
}

// initAddTarFile makes a tar archive file and
// adds it to the global variable testFS.
func initAddTarFile(buf *bytes.Buffer) error {
	buf.Reset()
	tw := tar.NewWriter(buf)
	for i := range testFSTarFiles {
		hdr := &tar.Header{
			Name:    testFSTarFiles[i].name,
			Size:    int64(len(testFSTarFiles[i].body)),
			Mode:    0600,
			ModTime: time.Now(),
		}
		if len(hdr.Name) == 0 || hdr.Name[len(hdr.Name)-1] != '/' {
			hdr.Typeflag = tar.TypeReg
		} else {
			hdr.Typeflag = tar.TypeDir
		}
		err := tw.WriteHeader(hdr)
		if err != nil {
			return errors.AutoWrap(err)
		} else if len(testFSTarFiles[i].body) > 0 {
			_, err = tw.Write([]byte(testFSTarFiles[i].body))
			if err != nil {
				return errors.AutoWrap(err)
			}
		}
	}
	err := tw.Close()
	if err != nil {
		return errors.AutoWrap(err)
	}
	testFS["tar file.tar"] = &fstest.MapFile{
		Data:    copyBuffer(buf),
		Mode:    0755,
		ModTime: time.Now(),
	}
	return nil
}

// initAddTgzFile makes a tar archive file, compresses it with gzip,
// and finally adds it to the global variable testFS
// with extensions ".tgz" and ".tar.gz".
func initAddTgzFile(buf *bytes.Buffer) error {
	buf.Reset()
	gzw, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		return errors.AutoWrap(err)
	}
	tw := tar.NewWriter(gzw)
	for i := range testFSTarFiles {
		hdr := &tar.Header{
			Name:    testFSTarFiles[i].name,
			Size:    int64(len(testFSTarFiles[i].body)),
			Mode:    0600,
			ModTime: time.Now(),
		}
		if len(hdr.Name) == 0 || hdr.Name[len(hdr.Name)-1] != '/' {
			hdr.Typeflag = tar.TypeReg
		} else {
			hdr.Typeflag = tar.TypeDir
		}
		err = tw.WriteHeader(hdr)
		if err != nil {
			return errors.AutoWrap(err)
		} else if len(testFSTarFiles[i].body) > 0 {
			_, err = tw.Write([]byte(testFSTarFiles[i].body))
			if err != nil {
				return errors.AutoWrap(err)
			}
		}
	}
	err = tw.Close()
	if err != nil {
		return errors.AutoWrap(err)
	}
	err = gzw.Close()
	if err != nil {
		return errors.AutoWrap(err)
	}
	testFS["tar gzip.tgz"] = &fstest.MapFile{
		Data:    copyBuffer(buf),
		Mode:    0755,
		ModTime: time.Now(),
	}
	testFS["tar gzip.tar.gz"] = &fstest.MapFile{
		Data:    copyBuffer(buf),
		Mode:    0755,
		ModTime: time.Now(),
	}
	return nil
}

// initAddZipFile makes a ZIP archive file and
// adds that file to the global variable testFS.
func initAddZipFile(buf *bytes.Buffer) error {
	buf.Reset()
	zw := zip.NewWriter(buf)
	err := zw.SetComment(testFSZipComment)
	if err != nil {
		return errors.AutoWrap(err)
	}
	zw.RegisterCompressor(
		zip.Deflate,
		func(w io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(w, flate.BestCompression)
		},
	)
	for name, body := range testFSZipFileNameBodyMap {
		var w io.Writer
		w, err = zw.Create(name)
		if err != nil {
			return errors.AutoWrap(err)
		} else if len(name) > 0 && name[len(name)-1] == '/' {
			continue
		}
		_, err = w.Write([]byte(body))
		if err != nil {
			return errors.AutoWrap(err)
		}
	}
	err = zw.Close()
	if err != nil {
		panic(err)
	}
	testFS["zip basic.zip"] = &fstest.MapFile{
		Data:    copyBuffer(buf),
		Mode:    0755,
		ModTime: time.Now(),
	}
	return nil
}

// initAddZipFileWithOffset makes a ZIP archive file prepended with
// random content, adds that file to the global variable testFS,
// and sets the global variable testFSZipOffset.
func initAddZipFileWithOffset(buf *bytes.Buffer, random *rand.Rand) error {
	buf.Reset()
	_, err := buf.ReadFrom(io.LimitReader(random, 5<<10))
	if err != nil {
		return errors.AutoWrap(err)
	}
	testFSZipOffset = int64(buf.Len())
	zw := zip.NewWriter(buf)
	zw.SetOffset(int64(buf.Len()))
	err = zw.SetComment(testFSZipComment)
	if err != nil {
		panic(err)
	}
	zw.RegisterCompressor(
		zip.Deflate,
		func(w io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(w, flate.BestCompression)
		},
	)
	for name, body := range testFSZipFileNameBodyMap {
		var w io.Writer
		w, err = zw.Create(name)
		if err != nil {
			panic(err)
		} else if len(name) > 0 && name[len(name)-1] == '/' {
			continue
		}
		_, err = w.Write([]byte(body))
		if err != nil {
			panic(err)
		}
	}
	err = zw.Close()
	if err != nil {
		panic(err)
	}
	testFS[testFSZipOffsetName] = &fstest.MapFile{
		Data:    copyBuffer(buf),
		Mode:    0755,
		ModTime: time.Now(),
	}
	return nil
}

// initRecordFilenames sets global variables testFSFilenames,
// testFSBasicFilenames, testFSGzFilenames, testFSTarFilenames,
// testFSTgzFilenames, and testFSZipFilenames.
func initRecordFilenames() {
	testFSFilenames = make([]string, len(testFS))
	var idx, gzIdx, tarIdx, tgzIdx, zipIdx int
	for name := range testFS {
		testFSFilenames[idx] = name
		idx++
		cname := path.Clean(name)
		switch path.Ext(cname) {
		case ".gz":
			if path.Ext(cname[:len(cname)-3]) == ".tar" {
				tgzIdx++
			} else {
				gzIdx++
			}
		case ".tar":
			tarIdx++
		case ".tgz":
			tgzIdx++
		case ".zip":
			zipIdx++
		}
	}
	sort.Strings(testFSFilenames)
	testFSBasicFilenames = make([]string, idx-gzIdx-tarIdx-tgzIdx-zipIdx)
	testFSGzFilenames = make([]string, gzIdx)
	testFSTarFilenames = make([]string, tarIdx)
	testFSTgzFilenames = make([]string, tgzIdx)
	testFSZipFilenames = make([]string, zipIdx)
	idx, gzIdx, tarIdx, tgzIdx, zipIdx = 0, 0, 0, 0, 0
	for _, name := range testFSFilenames {
		cname := path.Clean(name)
		switch path.Ext(cname) {
		case ".gz":
			if path.Ext(cname[:len(cname)-3]) == ".tar" {
				testFSTgzFilenames[tgzIdx] = name
				tgzIdx++
			} else {
				testFSGzFilenames[gzIdx] = name
				gzIdx++
			}
		case ".tar":
			testFSTarFilenames[tarIdx] = name
			tarIdx++
		case ".tgz":
			testFSTgzFilenames[tgzIdx] = name
			tgzIdx++
		case ".zip":
			testFSZipFilenames[zipIdx] = name
			zipIdx++
		default:
			testFSBasicFilenames[idx] = name
			idx++
		}
	}
}

// initMakeFSChecksumMap sets the global variable testFSChecksumMap.
func initMakeFSChecksumMap() error {
	testFSChecksumMap = make(map[string][]struct {
		hash     crypto.Hash
		checksum string
	}, len(testFS))
	sha256Hash := crypto.SHA256.New()
	md5Hash := crypto.MD5.New()
	w := io.MultiWriter(sha256Hash, md5Hash)
	for name, file := range testFS {
		r := bytes.NewReader(file.Data)
		sha256Hash.Reset()
		md5Hash.Reset()
		_, err := io.Copy(w, r)
		if err != nil {
			return errors.AutoWrap(err)
		}
		testFSChecksumMap[name] = []struct {
			hash     crypto.Hash
			checksum string
		}{
			{
				hash:     crypto.SHA256,
				checksum: hex.EncodeToString(sha256Hash.Sum(nil)),
			},
			{
				hash:     crypto.MD5,
				checksum: hex.EncodeToString(md5Hash.Sum(nil)),
			},
		}
	}
	return nil
}

// copyBuffer returns a copy of buf.Bytes().
func copyBuffer(buf *bytes.Buffer) []byte {
	if buf == nil {
		return nil
	}
	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())
	return data
}
