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

// Context of the communicators.
// Each goroutine group has its own context.
type context struct {
	Id         string          // ID of the group.
	Ctrl       *controller     // Controller.
	Comms      []*communicator // List of communicators.
	WorldRanks []int           // List of world ranks of the goroutines, corresponding to Comms.

	// List of channel maps for cluster communication.
	// Only for chanDispr.
	ChanMaps [numCOp]map[int64]*chanCntr
}

// Create a new context.
// Only for function New.
func newContext(ctrl *controller, id string, worldRanks []int) *context {
	n := len(worldRanks)
	ctx := &context{
		Id:         id,
		Ctrl:       ctrl,
		Comms:      make([]*communicator, n),
		WorldRanks: make([]int, n),
	}
	copy(ctx.WorldRanks, worldRanks) // Keep a copy to avoid unexpected modifications.
	for i := 0; i < n; i++ {
		ctx.Comms[i] = newCommunicator(ctx, i)
	}
	return ctx
}
