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

package spmd

import (
	"testing"

	"github.com/donyori/gogo/container/sequence"
)

func TestController_Start(t *testing.T) {
	n := 8
	groupMap := map[string][]int{
		"g1": {0, 1, 2, 3},
		"g2": {4, 5, 6, 7},
		"g3": {0, 2, 4, 6},
		"g4": {0, 1, 2, 3, 7, 6, 5, 4},
		"g5": {4, 1, 1, 1},
	}
	prs := Run(n, func(world Communicator, commMap map[string]Communicator) {
		wr := world.Rank()
		for id, comm := range commMap {
			if comm == nil {
				t.Errorf("Goroutine %d, group %q: comm is nil.", wr, id)
				continue
			}
			if r := comm.Rank(); wr != groupMap[id][r] {
				t.Errorf("Goroutine %d, group %q: comm.Rank(): %d, groupMap[%[2]q][%[1]d]: %[4]d.", wr, id, r, groupMap[id][wr])
			}
		}
		ctrl := world.(*communicator).ctx.Ctrl
		for id, group := range groupMap {
			g := sequence.IntDynamicArray(group)
			set := make(map[int]bool)
			g.Filter(func(x interface{}) (keep bool) {
				i := x.(int)
				if set[i] {
					return false
				}
				set[i] = true
				return true
			})
			for r, worldRank := range g {
				if wr == worldRank {
					if commMap[id] != ctrl.CtxMap[id].Comms[r] {
						t.Errorf("Goroutine %d, group %q, rank %d: communicators does not match.", wr, id, r)
					}
				}
			}
		}
	}, groupMap)
	if len(prs) > 0 {
		t.Errorf("Panic: %v.", prs)
	}
}
