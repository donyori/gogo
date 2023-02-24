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

import (
	"github.com/donyori/gogo/concurrency/framework"
	"github.com/donyori/gogo/container/sequence/array"
)

// chanDispQry is a channel dispatch query.
// It consists of the communicator sending the query and
// a counter specifying the communication uniquely.
type chanDispQry[Message any] struct {
	comm *communicator[Message] // Communicator that sends this query.
	ctr  int64                  // Counter to specify a communication uniquely.
}

// chanDispr is a channel dispatcher.
//
// It receives queries from communicators and
// sends channels or channel lists they need back to them.
//
// It is only for the cluster communications with multiple communicators.
// A communicator within a group (context) that has no other communicators
// must not use the channel dispatcher.
type chanDispr[Message any] struct {
	bcastChan   chan *chanDispQry[Message] // Channel for receiving channel dispatch queries for Broadcast.
	scatterChan chan *chanDispQry[Message] // Channel for receiving channel dispatch queries for Scatter.
	gatherChan  chan *chanDispQry[Message] // Channel for receiving channel dispatch queries for Gather.
}

// newChanDispr creates a new channel dispatcher.
// Only for function New.
func newChanDispr[Message any]() *chanDispr[Message] {
	return &chanDispr[Message]{
		bcastChan:   make(chan *chanDispQry[Message]),
		scatterChan: make(chan *chanDispQry[Message]),
		gatherChan:  make(chan *chanDispQry[Message]),
	}
}

// Run launches the channel dispatcher on current goroutine.
//
// quitDevice is the device to receive a quit signal.
// It should be obtained from Controller.
// The function panics if quitDevice is nil.
//
// finChan is a channel to broadcast a finish signal by closing the channel.
// It is closed at the end of this function.
// finChan is ignored if it is nil.
func (cd *chanDispr[Message]) Run(quitDevice framework.QuitDevice, finChan chan<- struct{}) {
	if finChan != nil {
		defer close(finChan)
	}
	quitChan := quitDevice.QuitChan()
	for {
		select {
		case <-quitChan:
			return
		case qry := <-cd.bcastChan:
			if qry != nil && cd.handleBcast(quitChan, qry) {
				return
			}
		case qry := <-cd.scatterChan:
			if qry != nil && cd.handleScatter(quitChan, qry) {
				return
			}
		case qry := <-cd.gatherChan:
			if qry != nil && cd.handleGather(quitChan, qry) {
				return
			}
		}
	}
}

// handleBcast deals with a channel dispatcher query for Broadcast.
//
// quitChan is a channel to receive a quit signal.
// It should be obtained from the quit device passed to the caller.
// qry is the channel dispatcher query received from bcastChan.
//
// It returns an indicator, which is true if and only if
// a quit signal is detected.
func (cd *chanDispr[Message]) handleBcast(quitChan <-chan struct{}, qry *chanDispQry[Message]) bool {
	ctx := qry.comm.ctx
	if ctx.bcastMap == nil {
		ctx.bcastMap = make(map[int64]*bcastChanCtr[Message])
	}
	cc := ctx.bcastMap[qry.ctr]
	if cc != nil {
		cc.ctr--
		if cc.ctr == 0 {
			delete(ctx.bcastMap, qry.ctr)
		}
	} else {
		n := len(ctx.comms) - 1
		// n > 0 because a communicator within a group (context) that has
		// no other communicators must not use the channel dispatcher.
		cc = &bcastChanCtr[Message]{
			c:   make(chan Message, n),
			ctr: n,
		}
		ctx.bcastMap[qry.ctr] = cc
	}
	select {
	case <-quitChan:
		return true
	case qry.comm.bcdc <- cc.c:
		return false
	}
}

// handleScatter deals with a channel dispatcher query for Scatter.
//
// quitChan is a channel to receive a quit signal.
// It should be obtained from the quit device passed to the caller.
// qry is the channel dispatcher query received from scatterChan.
//
// It returns an indicator, which is true if and only if
// a quit signal is detected.
func (cd *chanDispr[Message]) handleScatter(quitChan <-chan struct{}, qry *chanDispQry[Message]) bool {
	ctx := qry.comm.ctx
	if ctx.scatterMap == nil {
		ctx.scatterMap = make(map[int64]*scatterChanCtr[Message])
	}
	cc := ctx.scatterMap[qry.ctr]
	if cc != nil {
		cc.ctr--
		if cc.ctr == 0 {
			delete(ctx.scatterMap, qry.ctr)
		}
	} else {
		n := len(ctx.comms) - 1
		// n > 0 because a communicator within a group (context) that has
		// no other communicators must not use the channel dispatcher.
		cc = &scatterChanCtr[Message]{
			cs:  make([]chan array.Array[Message], n),
			ctr: n,
		}
		for i := range cc.cs {
			cc.cs[i] = make(chan array.Array[Message], 1)
		}
		ctx.scatterMap[qry.ctr] = cc
	}
	select {
	case <-quitChan:
		return true
	case qry.comm.scdc <- cc.cs:
		return false
	}
}

// handleGather deals with a channel dispatcher query for Gather.
//
// quitChan is a channel to receive a quit signal.
// It should be obtained from the quit device passed to the caller.
// qry is the channel dispatcher query received from gatherChan.
//
// It returns an indicator, which is true if and only if
// a quit signal is detected.
func (cd *chanDispr[Message]) handleGather(quitChan <-chan struct{}, qry *chanDispQry[Message]) bool {
	ctx := qry.comm.ctx
	if ctx.gatherMap == nil {
		ctx.gatherMap = make(map[int64]*gatherChanCtr[Message])
	}
	cc := ctx.gatherMap[qry.ctr]
	if cc != nil {
		cc.ctr--
		if cc.ctr == 0 {
			delete(ctx.gatherMap, qry.ctr)
		}
	} else {
		n := len(ctx.comms) - 1
		// n > 0 because a communicator within a group (context) that has
		// no other communicators must not use the channel dispatcher.
		cc = &gatherChanCtr[Message]{
			c:   make(chan *sndrMsg[Message], n),
			ctr: n,
		}
		ctx.gatherMap[qry.ctr] = cc
	}
	select {
	case <-quitChan:
		return true
	case qry.comm.gcdc <- cc.c:
		return false
	}
}
