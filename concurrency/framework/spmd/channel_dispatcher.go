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

import (
	"github.com/donyori/gogo/concurrency/framework"
	"github.com/donyori/gogo/container/sequence"
)

// chanDispQry is a channel dispatch query.
// It consists of the communicator sending the query and
// a counter specifying the communication uniquely.
type chanDispQry struct {
	Comm *communicator // Communicator that sends this query.
	Ctr  int64         // Counter to specify a communication uniquely.
}

// chanDispr is a channel dispatcher.
//
// It receives queries from communicators and
// sends channels or channel lists they need back to them.
//
// It is only for the cluster communications with multiple communicators.
// A communicator within a group (context) that has no other communicators
// must not use the channel dispatcher.
type chanDispr struct {
	BcastChan   chan *chanDispQry // Channel for receiving channel dispatch queries for Broadcast.
	ScatterChan chan *chanDispQry // Channel for receiving channel dispatch queries for Scatter.
	GatherChan  chan *chanDispQry // Channel for receiving channel dispatch queries for Gather.
}

// newChanDispr creates a new channel dispatcher.
// Only for function New.
func newChanDispr() *chanDispr {
	return &chanDispr{
		BcastChan:   make(chan *chanDispQry),
		ScatterChan: make(chan *chanDispQry),
		GatherChan:  make(chan *chanDispQry),
	}
}

// Run launches the channel dispatcher on current goroutine.
//
// quitDevice is the device to receive a quit signal.
// It should be obtained from Controller.
// The function will panic if quitDevice is nil.
//
// finChan is a channel to broadcast a finish signal by closing the channel.
// It will be closed at the end of this function.
// finChan will be ignored if it is nil.
func (cd *chanDispr) Run(quitDevice framework.QuitDevice, finChan chan<- struct{}) {
	if finChan != nil {
		defer close(finChan)
	}
	quitChan := quitDevice.QuitChan()
	for {
		select {
		case <-quitChan:
			return
		case qry := <-cd.BcastChan:
			if qry != nil && cd.handleBcast(quitChan, qry) {
				return
			}
		case qry := <-cd.ScatterChan:
			if qry != nil && cd.handleScatter(quitChan, qry) {
				return
			}
		case qry := <-cd.GatherChan:
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
// qry is the channel dispatcher query received from BcastChan.
//
// It returns an indicator, which is true if and only if
// a quit signal is detected.
func (cd *chanDispr) handleBcast(quitChan <-chan struct{}, qry *chanDispQry) bool {
	ctx := qry.Comm.Ctx
	if ctx.BcastMap == nil {
		ctx.BcastMap = make(map[int64]*bcastChanCtr)
	}
	cc := ctx.BcastMap[qry.Ctr]
	if cc != nil {
		cc.Ctr--
		if cc.Ctr == 0 {
			delete(ctx.BcastMap, qry.Ctr)
		}
	} else {
		n := len(ctx.Comms) - 1
		// n > 0 because a communicator within a group (context) that has
		// no other communicators must not use the channel dispatcher.
		cc = &bcastChanCtr{
			Chan: make(chan interface{}, n),
			Ctr:  n,
		}
		ctx.BcastMap[qry.Ctr] = cc
	}
	select {
	case <-quitChan:
		return true
	case qry.Comm.Bcdc <- cc.Chan:
		return false
	}
}

// handleScatter deals with a channel dispatcher query for Scatter.
//
// quitChan is a channel to receive a quit signal.
// It should be obtained from the quit device passed to the caller.
// qry is the channel dispatcher query received from ScatterChan.
//
// It returns an indicator, which is true if and only if
// a quit signal is detected.
func (cd *chanDispr) handleScatter(quitChan <-chan struct{}, qry *chanDispQry) bool {
	ctx := qry.Comm.Ctx
	if ctx.ScatterMap == nil {
		ctx.ScatterMap = make(map[int64]*scatterChanCtr)
	}
	cc := ctx.ScatterMap[qry.Ctr]
	if cc != nil {
		cc.Ctr--
		if cc.Ctr == 0 {
			delete(ctx.ScatterMap, qry.Ctr)
		}
	} else {
		n := len(ctx.Comms) - 1
		// n > 0 because a communicator within a group (context) that has
		// no other communicators must not use the channel dispatcher.
		cc = &scatterChanCtr{
			Chans: make([]chan sequence.Array, n),
			Ctr:   n,
		}
		for i := range cc.Chans {
			cc.Chans[i] = make(chan sequence.Array, 1)
		}
		ctx.ScatterMap[qry.Ctr] = cc
	}
	select {
	case <-quitChan:
		return true
	case qry.Comm.Scdc <- cc.Chans:
		return false
	}
}

// handleGather deals with a channel dispatcher query for Gather.
//
// quitChan is a channel to receive a quit signal.
// It should be obtained from the quit device passed to the caller.
// qry is the channel dispatcher query received from GatherChan.
//
// It returns an indicator, which is true if and only if
// a quit signal is detected.
func (cd *chanDispr) handleGather(quitChan <-chan struct{}, qry *chanDispQry) bool {
	ctx := qry.Comm.Ctx
	if ctx.GatherMap == nil {
		ctx.GatherMap = make(map[int64]*gatherChanCtr)
	}
	cc := ctx.GatherMap[qry.Ctr]
	if cc != nil {
		cc.Ctr--
		if cc.Ctr == 0 {
			delete(ctx.GatherMap, qry.Ctr)
		}
	} else {
		n := len(ctx.Comms) - 1
		// n > 0 because a communicator within a group (context) that has
		// no other communicators must not use the channel dispatcher.
		cc = &gatherChanCtr{
			Chan: make(chan *sndrMsg, n),
			Ctr:  n,
		}
		ctx.GatherMap[qry.Ctr] = cc
	}
	select {
	case <-quitChan:
		return true
	case qry.Comm.Gcdc <- cc.Chan:
		return false
	}
}
