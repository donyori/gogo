// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

package spmd

import "testing"

func TestNew_lnchCommMaps(t *testing.T) {
	n := 8
	groupMap := map[string][]int{
		"g1": {0, 1, 2, 3},
		"g2": {4, 5, 6, 7},
		"g3": {0, 2, 4, 6},
		"g4": {0, 1, 2, 3, 7, 6, 5, 4},
		"g5": {4, 1, 1, 1},
	}
	ctrl := New(n, func(world Communicator, commMap map[string]Communicator) {
		// Empty function body.
	}, groupMap).(*controller)
	// Verify that all commMaps are consistent with groupMap.
	for wr, commMap := range ctrl.lnchCommMaps {
		for id, comm := range commMap {
			if comm == nil {
				t.Errorf("Goroutine %d, group %q: comm is nil.", wr, id)
			} else if r := comm.Rank(); wr != groupMap[id][r] {
				t.Errorf("Goroutine %d, group %q: comm.Rank(): %d, groupMap[%[2]q][%[1]d]: %[4]d.", wr, id, r, groupMap[id][wr])
			}
		}
	}
	// Verify that all items in groupMap have corresponding communicators.
	for id, group := range groupMap {
		set, r := make(map[int]bool), 0
		for _, wr := range group {
			if set[wr] {
				continue
			}
			set[wr] = true
			comm := ctrl.lnchCommMaps[wr][id]
			if comm == nil {
				t.Errorf("Goroutine %d, group %q: comm is nil.", wr, id)
			} else if rank := comm.Rank(); rank != r {
				t.Errorf("Goroutine %d, group %q: comm.Rank(): %d, groupMap[%[2]q][%[1]d]: %[4]d.", wr, id, rank, r)
			}
			r++
		}
	}
}

func TestController_Wait_BeforeLaunch(t *testing.T) {
	ctrl := New(0, func(world Communicator, commMap map[string]Communicator) {
		// Do nothing.
	}, nil)
	if r := ctrl.Wait(); r != -1 {
		t.Errorf("ctrl.Wait returns %d (not -1) before calling Launch.", r)
	}
}
