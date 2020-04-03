// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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

package file

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

func TestTmp(t *testing.T) {
	tmpRoot, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpRoot)
	var locker sync.Mutex
	var wg sync.WaitGroup
	files := make([]string, 0, 10)
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(no int) {
			defer wg.Done()
			f, err := Tmp(tmpRoot, "f.*.tmp", 0740)
			if err != nil {
				t.Error(i, err)
				return
			}
			defer f.Close() // ignore error
			locker.Lock()
			defer locker.Unlock()
			files = append(files, f.Name())
		}(i)
	}
	wg.Wait()
	if n := len(files); n != 10 {
		t.Error("len(files):", n, "!= 10.")
	}
	set := make(map[string]bool)
	for _, filename := range files {
		if set[filename] {
			t.Error("Conflict::", filename)
			continue
		}
		set[filename] = true
	}
}

func TestTmpDir(t *testing.T) {
	tmpRoot, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpRoot)
	var locker sync.Mutex
	var wg sync.WaitGroup
	dirs := make([]string, 0, 10)
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(no int) {
			defer wg.Done()
			dir, err := TmpDir(tmpRoot, "tmp", 0700)
			if err != nil {
				t.Error(i, err)
				return
			}
			locker.Lock()
			defer locker.Unlock()
			dirs = append(dirs, dir)
		}(i)
	}
	wg.Wait()
	if n := len(dirs); n != 10 {
		t.Error("len(dirs):", n, "!= 10.")
	}
	set := make(map[string]bool)
	for _, dir := range dirs {
		if set[dir] {
			t.Error("Conflict::", dir)
			continue
		}
		set[dir] = true
	}
}
