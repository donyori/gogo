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
	"os"
	"sync"
	"testing"

	"github.com/donyori/gogo/filesys/local"
)

func TestTmp_Sync(t *testing.T) {
	tmpRoot := t.TempDir()
	const N int = 10
	var mutex sync.Mutex
	var wg sync.WaitGroup
	files := make([]string, 0, N)
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(no int) {
			defer wg.Done()
			f, err := local.Tmp(tmpRoot, "f.", ".tmp", 0740)
			mutex.Lock()
			defer mutex.Unlock()
			if err != nil {
				t.Error("Goroutine", no, "local.Tmp -", err)
				return
			}
			defer func(f *os.File) {
				if err := f.Close(); err != nil {
					t.Error("Goroutine", no, "close file -", err)
				}
			}(f)
			files = append(files, f.Name())
		}(i)
	}
	wg.Wait()
	if t.Failed() {
		return
	} else if n := len(files); n != N {
		t.Errorf("got %d files; want %d", n, N)
	}
	set := make(map[string]struct{})
	for _, filename := range files {
		if _, ok := set[filename]; ok {
			t.Error("collided filename", filename)
			continue
		}
		set[filename] = struct{}{}
	}
}

func TestTmpDir_Sync(t *testing.T) {
	tmpRoot := t.TempDir()
	const N int = 10
	var mutex sync.Mutex
	var wg sync.WaitGroup
	dirs := make([]string, 0, N)
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(no int) {
			defer wg.Done()
			dir, err := local.TmpDir(tmpRoot, "tmp-", "", 0700)
			mutex.Lock()
			defer mutex.Unlock()
			if err != nil {
				t.Error("Goroutine", no, "local.TmpDir -", err)
				return
			}
			dirs = append(dirs, dir)
		}(i)
	}
	wg.Wait()
	if t.Failed() {
		return
	} else if n := len(dirs); n != N {
		t.Errorf("got %d dirs; want %d", n, N)
	}
	set := make(map[string]struct{})
	for _, dir := range dirs {
		if _, ok := set[dir]; ok {
			t.Error("collided directory", dir)
			continue
		}
		set[dir] = struct{}{}
	}
}
