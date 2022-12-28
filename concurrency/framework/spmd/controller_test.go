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

package spmd_test

import (
	"testing"

	"github.com/donyori/gogo/concurrency/framework/spmd"
)

func TestNew_lnchCommMaps(t *testing.T) {
	const N int = 8
	groupMap := map[string][]int{
		"g1": {0, 1, 2, 3},
		"g2": {4, 5, 6, 7},
		"g3": {0, 2, 4, 6},
		"g4": {0, 1, 2, 3, 7, 6, 5, 4},
		"g5": {4, 1, 1, 1},
		"g6": {4, 1, 1, 2, 1, 1, 4},
	}
	ctrl := spmd.New(N, func(world spmd.Communicator[int], commMap map[string]spmd.Communicator[int]) {
		// Empty function body.
	}, groupMap)

	deduplicatedGroupMap := make(map[string][]int, len(groupMap))
	for k, v := range groupMap {
		newV := make([]int, 0, len(v))
		set := make(map[int]bool, N)
		for _, x := range v {
			if set[x] {
				continue
			}
			set[x] = true
			newV = append(newV, x)
		}
		deduplicatedGroupMap[k] = newV
	}
	lnchCommMaps := spmd.WrapController[int](ctrl).GetLnchCommMaps()
	// Verify that all commMaps are consistent with groupMap.
	for wr, commMap := range lnchCommMaps {
		for id, comm := range commMap {
			if comm == nil {
				t.Errorf("goroutine %d, group %q, comm is nil", wr, id)
			} else if r := comm.Rank(); wr != deduplicatedGroupMap[id][r] {
				t.Errorf("goroutine %d, group %q, comm.Rank %d is inconsistent with groupMap[%[2]q][%[1]d] %[4]d",
					wr, id, r, deduplicatedGroupMap[id][r])
			}
		}
	}
	// Verify that all items in groupMap have corresponding communicators.
	for id, group := range deduplicatedGroupMap {
		for r, wr := range group {
			comm := lnchCommMaps[wr][id]
			if comm == nil {
				t.Errorf("goroutine %d, group %q, comm is nil", wr, id)
			} else if rank := comm.Rank(); rank != r {
				t.Errorf("goroutine %d, group %q, comm.Rank %d is inconsistent with groupMap[%[2]q][%[1]d] %[4]d",
					wr, id, rank, r)
			}
		}
	}
}

func TestController_Wait_BeforeLaunch(t *testing.T) {
	ctrl := spmd.New(0, func(world spmd.Communicator[int], commMap map[string]spmd.Communicator[int]) {
		// Do nothing.
	}, nil)
	if r := ctrl.Wait(); r != -1 {
		t.Errorf("got %d; want -1", r)
	}
}
