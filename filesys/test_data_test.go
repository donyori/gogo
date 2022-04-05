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

package filesys

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path"
	"sort"
	"strings"
	"testing/fstest"
	"time"
)

var (
	testFs fstest.MapFS

	testFsTarFiles []struct{ name, body string }

	testFsFilenames      []string
	testFsBasicFilenames []string
	testFsGzFilenames    []string
	testFsTarFilenames   []string
	testFsTgzFilenames   []string

	testFsChecksumMap map[string][]Checksum
)

func init() {
	testFs = fstest.MapFS{
		"testFile1.txt": {
			Data:    []byte("This is test file 1."),
			ModTime: time.Now(),
		},
		"testFile2.txt": {
			Data:    []byte("Here is test file 2!"),
			ModTime: time.Now(),
		},
	}

	var sb strings.Builder
	sb.Grow(100_000)
	for i := 0; i < 100_000; i++ {
		sb.WriteByte(byte(i % (1 << 8)))
	}
	testFs["big.dat"] = &fstest.MapFile{
		Data:    []byte(sb.String()),
		ModTime: time.Now(),
	}
	testFsTarFiles = []struct{ name, body string }{
		{"tarfile1.txt", "This is tar file 1."},
		{"tarfile2.txt", "Here is tar file 2!"},
		{"tarfile3.dat", sb.String()},
	}

	var buf strings.Builder
	gzw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	_, err = gzw.Write([]byte(sb.String()))
	if err != nil {
		panic(err)
	}
	err = gzw.Close()
	if err != nil {
		panic(err)
	}
	testFs["gzip.gz"] = &fstest.MapFile{
		Data:    []byte(buf.String()),
		ModTime: time.Now(),
	}
	buf.Reset()

	tw := tar.NewWriter(&buf)
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
	testFs["tarfile.tar"] = &fstest.MapFile{
		Data:    []byte(buf.String()),
		ModTime: time.Now(),
	}
	buf.Reset()

	gzw, err = gzip.NewWriterLevel(&buf, gzip.BestCompression)
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
	testFs["targzip.tgz"] = &fstest.MapFile{
		Data:    []byte(buf.String()),
		ModTime: time.Now(),
	}
	testFs["targzip.tar.gz"] = &fstest.MapFile{
		Data:    []byte(buf.String()),
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

	testFsChecksumMap = make(map[string][]Checksum, len(testFs))
	for name, file := range testFs {
		r := bytes.NewReader(file.Data)
		hSha256 := sha256.New()
		hMd5 := md5.New()
		w := io.MultiWriter(hSha256, hMd5)
		_, err = io.Copy(w, r)
		if err != nil {
			panic(err)
		}
		testFsChecksumMap[name] = []Checksum{
			{
				HashGen:   sha256.New,
				HexExpSum: hex.EncodeToString(hSha256.Sum(nil)),
			},
			{
				HashGen:   md5.New,
				HexExpSum: hex.EncodeToString(hMd5.Sum(nil)),
			},
		}
	}
}
