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

package spmd

import "github.com/donyori/gogo/container/sequence/array"

// bcastChanCtr is a combination of a channel used in Broadcast and a counter.
type bcastChanCtr[Message any] struct {
	c   chan Message
	ctr int // Counter for the number of remaining uses.
}

// scatterChanCtr is a combination of a channel list
// used in Scatter and a counter.
type scatterChanCtr[Message any] struct {
	cs  []chan array.Array[Message]
	ctr int // Counter for the number of remaining uses.
}

// gatherChanCtr is a combination of a channel used in Gather and a counter.
type gatherChanCtr[Message any] struct {
	c   chan *sndrMsg[Message]
	ctr int // Counter for the number of remaining uses.
}

// context is the environment of the communicators.
// Each goroutine group has its own context.
// Each communicator belongs to only one group (context).
type context[Message any] struct {
	id         string                    // ID of the group.
	ctrl       *controller[Message]      // Controller.
	comms      []*communicator[Message]  // List of communicators.
	worldRanks []int                     // List of world ranks of the goroutines, corresponding to comms.
	pubC       chan *sndrMsgRxc[Message] // Public channel used by communicators.

	bcastMap   map[int64]*bcastChanCtr[Message]   // Channel map for Broadcast, maintained by the channel dispatcher, initially nil.
	scatterMap map[int64]*scatterChanCtr[Message] // Channel list map for Scatter, maintained by the channel dispatcher, initially nil.
	gatherMap  map[int64]*gatherChanCtr[Message]  // Channel map for Gather, maintained by the channel dispatcher, initially nil.
}

// newContext creates a new context.
// Only for function New.
//
// The caller (function New) must guarantee that
// worldRanks is non-nil and non-empty,
// has no duplicates and no out-of-range items,
// and cannot be modified by others (such as the caller of function New).
func newContext[Message any](
	ctrl *controller[Message],
	id string,
	worldRanks []int,
) *context[Message] {
	n := len(worldRanks)
	ctx := &context[Message]{
		id:         id,
		ctrl:       ctrl,
		comms:      make([]*communicator[Message], n),
		worldRanks: worldRanks,
		pubC:       make(chan *sndrMsgRxc[Message]),
	}
	for i := 0; i < n; i++ {
		ctx.comms[i] = newCommunicator(ctx, i)
	}
	return ctx
}
