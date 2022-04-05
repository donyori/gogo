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

package spmd

import "github.com/donyori/gogo/container/sequence"

// bcastChanCtr is a combination of a channel used in Broadcast and a counter.
type bcastChanCtr struct {
	Chan chan interface{}
	Ctr  int // Counter for the number of remaining uses.
}

// scatterChanCtr is a combination of a channel list used in Scatter
// and a counter.
type scatterChanCtr struct {
	Chans []chan sequence.Array
	Ctr   int // Counter for the number of remaining uses.
}

// gatherChanCtr is a combination of a channel used in Gather and a counter.
type gatherChanCtr struct {
	Chan chan *sndrMsg
	Ctr  int // Counter for the number of remaining uses.
}

// context is the environment of the communicators.
// Each goroutine group has its own context.
// Each communicator belongs to only one group (context).
type context struct {
	Id         string           // ID of the group.
	Ctrl       *controller      // Controller.
	Comms      []*communicator  // List of communicators.
	WorldRanks []int            // List of world ranks of the goroutines, corresponding to Comms.
	PubC       chan *sndrMsgRxc // Public channel used by communicators.

	BcastMap   map[int64]*bcastChanCtr   // Channel map for Broadcast, maintained by the channel dispatcher, initially nil.
	ScatterMap map[int64]*scatterChanCtr // Channel list map for Scatter, maintained by the channel dispatcher, initially nil.
	GatherMap  map[int64]*gatherChanCtr  // Channel map for Gather, maintained by the channel dispatcher, initially nil.
}

// newContext creates a new context.
// Only for function New.
func newContext(ctrl *controller, id string, worldRanks []int) *context {
	n := len(worldRanks)
	ctx := &context{
		Id:         id,
		Ctrl:       ctrl,
		Comms:      make([]*communicator, n),
		WorldRanks: make([]int, n),
		PubC:       make(chan *sndrMsgRxc),
	}
	copy(ctx.WorldRanks, worldRanks) // Keep a copy to avoid unexpected modifications.
	for i := 0; i < n; i++ {
		ctx.Comms[i] = newCommunicator(ctx, i)
	}
	return ctx
}
