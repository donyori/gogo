// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"math/rand"
	"path"
	"sort"
	"testing/fstest"
	"time"

	"github.com/donyori/gogo/filesys"
)

var (
	testFs fstest.MapFS

	testFsTarFiles []struct{ name, body string }

	testFsFilenames      []string
	testFsBasicFilenames []string
	testFsGzFilenames    []string
	testFsTarFilenames   []string
	testFsTgzFilenames   []string

	testFsChecksumMap map[string][]filesys.HashChecksum
)

func init() {
	testFs = fstest.MapFS{
		"file1.txt": {
			Data:    []byte("This is File 1."),
			ModTime: time.Now(),
		},
		"file2.txt": {
			Data:    []byte("Here is File 2!"),
			ModTime: time.Now(),
		},
		"roses are red.txt": {
			Data:    []byte("Roses are red.\n  Violets are blue.\nSugar is sweet.\n  And so are you.\n"),
			ModTime: time.Now(),
		},
	}

	big := make([]byte, 1_048_576)
	rand.New(rand.NewSource(10)).Read(big)
	testFs["1MB.dat"] = &fstest.MapFile{
		Data:    big,
		ModTime: time.Now(),
	}
	testFsTarFiles = []struct{ name, body string }{
		{"tar file1.txt", "This is tar file 1."},
		{"tar file2.txt", "Here is tar file 2!"},
		{"roses are red.txt", "Roses are red.\n  Violets are blue.\nSugar is sweet.\n  And so are you.\n"},
		{"1MB.dat", string(big)},
	}

	buf := new(bytes.Buffer)
	gzw, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	_, err = gzw.Write(big)
	if err != nil {
		panic(err)
	}
	err = gzw.Close()
	if err != nil {
		panic(err)
	}
	testFs["gzip.gz"] = &fstest.MapFile{
		Data:    copyBuffer(buf),
		ModTime: time.Now(),
	}

	buf.Reset()
	tw := tar.NewWriter(buf)
	for i := range testFsTarFiles {
		err = tw.WriteHeader(&tar.Header{
			Name:    testFsTarFiles[i].name,
			Size:    int64(len(testFsTarFiles[i].body)),
			Mode:    0600,
			ModTime: time.Now(),
		})
		if err != nil {
			panic(err)
		}
		_, err = tw.Write([]byte(testFsTarFiles[i].body))
		if err != nil {
			panic(err)
		}
	}
	err = tw.Close()
	if err != nil {
		panic(err)
	}
	testFs["tar file.tar"] = &fstest.MapFile{
		Data:    copyBuffer(buf),
		ModTime: time.Now(),
	}

	buf.Reset()
	gzw, err = gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	tw = tar.NewWriter(gzw)
	for i := range testFsTarFiles {
		err = tw.WriteHeader(&tar.Header{
			Name:    testFsTarFiles[i].name,
			Size:    int64(len(testFsTarFiles[i].body)),
			Mode:    0600,
			ModTime: time.Now(),
		})
		if err != nil {
			panic(err)
		}
		_, err = tw.Write([]byte(testFsTarFiles[i].body))
		if err != nil {
			panic(err)
		}
	}
	err = tw.Close()
	if err != nil {
		panic(err)
	}
	err = gzw.Close()
	if err != nil {
		panic(err)
	}
	testFs["tar gzip.tgz"] = &fstest.MapFile{
		Data:    copyBuffer(buf),
		ModTime: time.Now(),
	}
	testFs["tar gzip.tar.gz"] = &fstest.MapFile{
		Data:    copyBuffer(buf),
		ModTime: time.Now(),
	}

	testFsFilenames = make([]string, len(testFs))
	var idx, gzIdx, tarIdx, tgzIdx int
	for name := range testFs {
		testFsFilenames[idx] = name
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
		}
	}
	sort.Strings(testFsFilenames)
	testFsBasicFilenames = make([]string, idx-gzIdx-tarIdx-tgzIdx)
	testFsGzFilenames = make([]string, gzIdx)
	testFsTarFilenames = make([]string, tarIdx)
	testFsTgzFilenames = make([]string, tgzIdx)
	idx, gzIdx, tarIdx, tgzIdx = 0, 0, 0, 0
	for _, name := range testFsFilenames {
		cname := path.Clean(name)
		switch path.Ext(cname) {
		case ".gz":
			if path.Ext(cname[:len(cname)-3]) == ".tar" {
				testFsTgzFilenames[tgzIdx] = name
				tgzIdx++
			} else {
				testFsGzFilenames[gzIdx] = name
				gzIdx++
			}
		case ".tar":
			testFsTarFilenames[tarIdx] = name
			tarIdx++
		case ".tgz":
			testFsTgzFilenames[tgzIdx] = name
			tgzIdx++
		default:
			testFsBasicFilenames[idx] = name
			idx++
		}
	}

	testFsChecksumMap = make(map[string][]filesys.HashChecksum, len(testFs))
	for name, file := range testFs {
		r := bytes.NewReader(file.Data)
		hSha256 := sha256.New()
		hMd5 := md5.New()
		w := io.MultiWriter(hSha256, hMd5)
		_, err = io.Copy(w, r)
		if err != nil {
			panic(err)
		}
		testFsChecksumMap[name] = []filesys.HashChecksum{
			{
				NewHash: sha256.New,
				ExpHex:  hex.EncodeToString(hSha256.Sum(nil)),
			},
			{
				NewHash: md5.New,
				ExpHex:  hex.EncodeToString(hMd5.Sum(nil)),
			},
		}
	}
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
