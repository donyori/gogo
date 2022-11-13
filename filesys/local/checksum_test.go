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

package local_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/filesys/local"
)

func TestVerifyChecksum(t *testing.T) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer func(dir string) {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	}(dir)
	f1, f2, h, err := makeTestFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	ck := filesys.HashChecksum{
		NewHash: sha256.New,
		ExpHex:  h,
	}
	wrongCk := filesys.HashChecksum{
		NewHash: sha256.New,
		ExpHex:  strings.Repeat("0", len(h)),
	}
	namePrefix := fmt.Sprintf("file=%q&cs=", f1)

	testCases := []struct {
		name     string
		filename string
		cs       []filesys.HashChecksum
		want     bool
	}{
		{`file="nonexist"&cs=<nil>`, "nonexist", nil, false},
		{namePrefix + "<nil>", f1, nil, true},
		{namePrefix + "correct", f1, []filesys.HashChecksum{ck}, true},
		{namePrefix + "wrong", f1, []filesys.HashChecksum{wrongCk}, false},
		{namePrefix + "correct+wrong", f1, []filesys.HashChecksum{ck, wrongCk}, false},
		{namePrefix + "zero-value", f1, []filesys.HashChecksum{{}}, false},
		{namePrefix + "noExpHex", f1, []filesys.HashChecksum{{NewHash: sha256.New}}, false},
		{namePrefix + "noHash", f1, []filesys.HashChecksum{{ExpHex: h}}, false},
		{namePrefix + "correct+empty", f1, []filesys.HashChecksum{ck, {}}, false},
		{fmt.Sprintf("file=%q&cs=wrong", f2), f2, []filesys.HashChecksum{ck}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := local.VerifyChecksum(tc.filename, tc.cs...); got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

// makeTestFiles makes two files with different content for TestVerifyChecksum.
//
// It returns the filenames and the SHA256 hash (in hexadecimal representation)
// of the first file (f1), and any error encountered.
func makeTestFiles(dir string) (f1, f2, h string, err error) {
	f1 = filepath.Join(dir, "test1.txt")
	f2 = filepath.Join(dir, "test2.txt")
	file1, err := os.Create(f1)
	if err != nil {
		return "", "", "", errors.AutoWrap(err)
	}
	closeFile := func(f *os.File) {
		err1 := f.Close()
		if err1 != nil {
			err = errors.AutoWrapSkip(errors.Combine(err, err1), 1) // skip = 1 to skip the inner function
		}
	}
	defer closeFile(file1)
	file2, err := os.Create(f2)
	if err != nil {
		return "", "", "", errors.AutoWrap(err)
	}
	defer closeFile(file2)
	hFn := sha256.New()
	mw1 := io.MultiWriter(file1, hFn)
	mw2 := io.MultiWriter(mw1, file2)
	for i := 0; i < 100; i++ {
		_, err = fmt.Fprintln(mw2, i)
		if err != nil {
			return "", "", "", errors.AutoWrap(err)
		}
	}
	_, err = fmt.Fprintln(mw1, 100)
	if err != nil {
		return "", "", "", errors.AutoWrap(err)
	}
	h = hex.EncodeToString(hFn.Sum(nil))
	return
}
