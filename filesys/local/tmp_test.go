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

package local_test

import (
	"os"
	"testing"
	"testing/synctest"

	"github.com/donyori/gogo/filesys/local"
)

// rankString consists of a rank (of type int) and a string.
type rankString struct {
	r int
	s string
}

func TestTmp_Sync(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		tmpRoot := t.TempDir()

		const N int = 10

		// Create a dedicated channel to make the following goroutines call
		// github.com/donyori/gogo/filesys/local.Tmp logically simultaneously.
		startC := make(chan struct{})
		outC := make(chan rankString, N)

		for i := range N {
			go func(
				t *testing.T,
				rank int,
				startC <-chan struct{},
				outC chan<- rankString,
			) {
				t.Helper()

				rs := rankString{r: rank}

				defer func() {
					outC <- rs
				}()

				<-startC

				f, err := local.Tmp(tmpRoot, "f.", ".tmp", 0740)
				if err != nil {
					t.Errorf("goroutine %d, local.Tmp: %v", rank, err)
					return
				}
				defer func(f *os.File) {
					err := f.Close()
					if err != nil {
						t.Errorf("goroutine %d, close file: %v", rank, err)
					}
				}(f)

				rs.s = f.Name()
			}(t, i, startC, outC)
		}

		synctest.Wait()
		close(startC)
		checkTmpSyncTmpDirSyncOutputs(t, N, outC)
	})
}

func TestTmpDir_Sync(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		tmpRoot := t.TempDir()

		const N int = 10

		// Create a dedicated channel to make the following goroutines call
		// github.com/donyori/gogo/filesys/local.TmpDir
		// logically simultaneously.
		startC := make(chan struct{})
		outC := make(chan rankString, N)

		for i := range N {
			go func(
				t *testing.T,
				rank int,
				startC <-chan struct{},
				outC chan<- rankString,
			) {
				t.Helper()

				rs := rankString{r: rank}

				defer func() {
					outC <- rs
				}()

				<-startC

				dir, err := local.TmpDir(tmpRoot, "tmp-", "", 0700)
				if err != nil {
					t.Errorf("goroutine %d, local.TmpDir: %v", rank, err)
					return
				}

				rs.s = dir
			}(t, i, startC, outC)
		}

		synctest.Wait()
		close(startC)
		checkTmpSyncTmpDirSyncOutputs(t, N, outC)
	})
}

// checkTmpSyncTmpDirSyncOutputs is a common subprocess of
// TestTmp_Sync and TestTmpDir_Sync that checks the outputs.
func checkTmpSyncTmpDirSyncOutputs(t *testing.T, n int, c <-chan rankString) {
	t.Helper()

	stringRankMap := make(map[string]int, n)
	for range n {
		rs := <-c

		if rs.s == "" {
			if rank := stringRankMap[rs.s]; rank > rs.r {
				stringRankMap[rs.s] = rs.r // keep the smallest rank in the map
			}

			t.Errorf("goroutine %d output an empty string", rs.r)
		} else if rank, ok := stringRankMap[rs.s]; ok {
			r1, r2 := rank, rs.r
			if r1 > r2 {
				r1, r2 = r2, r1
				stringRankMap[rs.s] = r1 // keep the smallest rank in the map
			}

			t.Errorf("goroutine %d and %d output the same result %q",
				r1, r2, rs.s)
		} else {
			stringRankMap[rs.s] = rs.r
		}
	}

	if len(stringRankMap) != n {
		t.Errorf("got %d nonempty outputs; want %d", len(stringRankMap), n)
	}
}
