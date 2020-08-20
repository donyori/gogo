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
	"fmt"

	"github.com/donyori/gogo/errors"
)

// A communicator for goroutines to communicate with each other.
type Communicator interface {
	// Return the rank (from 0 to NumGoroutine()-1) of current goroutine
	// in the goroutine group of this communicator.
	Rank() int

	// Return the number of goroutines in the goroutine group
	// of this communicator.
	NumGoroutine() int

	// Return the channel for the quit signal.
	QuitChan() <-chan struct{}

	// Detect the quit signal on the quit channel.
	// It returns true if a quit signal is detected, and false otherwise.
	IsQuit() bool

	// Broadcast a quit signal to quit the job.
	Quit()

	// Block until all other goroutines in this group call method Barrier
	// of their own communicators, or a quit signal is detected.
	//
	// It returns true if all other goroutines call methods Barrier
	// successfully, otherwise (a quit signal is detected), false.
	//
	// It is used to make all goroutines synchronous,
	// i.e., consistent in the execution progress.
	Barrier() bool

	// Broadcast the message x from the root to others in this group.
	// The method will not wait for all goroutines to finish the broadcast.
	// To synchronize all goroutines, use method Barrier.
	//
	// root is the rank of the sender goroutine in this group.
	// It will panic if root is out of range.
	//
	// For the root, x is the message to be broadcast.
	// For others, x can be anything (including nil) and will be ignored.
	//
	// It returns the message to be broadcast (equals to x of the root) and
	// an indicator ok. ok is false iff a quit signal is detected.
	Broadcast(root int, x interface{}) (msg interface{}, ok bool)
}

// Constants representing cluster communication operations of Communicator.
const (
	cOpBcast int = iota
	cOpScatter
	cOpGather

	// The number of cluster communication operations.
	numCOp
)

// An implementation of interface Communicator.
type communicator struct {
	Ctx *context              // Context of the goroutine group.
	Cdc chan chan interface{} // Channel for receiving channel form the channel dispatcher.
	// Counters for cluster communication operations.
	// Only for method chanDispr.Run.
	COpCntrs [numCOp]int64

	rank int                // The rank of current goroutine.
	bc   chan chan struct{} // A channel used for method Barrier.
}

// Create a new communicator.
// Only for function newContext.
func newCommunicator(ctx *context, rank int) *communicator {
	comm := &communicator{
		Ctx:  ctx,
		Cdc:  make(chan chan interface{}),
		rank: rank,
	}
	if rank > 0 {
		comm.bc = make(chan chan struct{})
	}
	return comm
}

func (comm *communicator) Rank() int {
	return comm.rank
}

func (comm *communicator) NumGoroutine() int {
	return len(comm.Ctx.Comms)
}

func (comm *communicator) QuitChan() <-chan struct{} {
	return comm.Ctx.Ctrl.QuitC
}

func (comm *communicator) IsQuit() bool {
	select {
	case <-comm.Ctx.Ctrl.QuitC:
		return true
	default:
		return false
	}
}

func (comm *communicator) Quit() {
	comm.Ctx.Ctrl.Quit()
}

func (comm *communicator) Barrier() bool {
	n := len(comm.Ctx.Comms)
	if n <= 1 {
		return !comm.IsQuit()
	}
	var c chan struct{}
	if comm.rank == 0 {
		// The first goroutine makes the signal channel.
		c = make(chan struct{})
	} else {
		// Other goroutines receive the signal channel from their respective previous goroutines.
		select {
		case <-comm.Ctx.Ctrl.QuitC:
			return false
		case c = <-comm.Ctx.Comms[comm.rank].bc: // Listen on its own channel, not the sender's!
		}
	}
	if comm.rank == n-1 {
		// The last goroutine close the signal channel to broadcast
		// the information that all goroutines call method Barrier successfully.
		close(c)
	} else {
		// Other goroutines send the signal channel to their respective next goroutines.
		select {
		case <-comm.Ctx.Ctrl.QuitC:
			return false
		case comm.Ctx.Comms[comm.rank+1].bc <- c: // Send it to the receiver's channel.
		}
		// Then listen on the signal channel.
		select {
		case <-comm.Ctx.Ctrl.QuitC:
			return false
		case <-c:
		}
	}
	return true
}

func (comm *communicator) Broadcast(root int, x interface{}) (msg interface{}, ok bool) {
	n := len(comm.Ctx.Comms)
	if root < 0 || root >= n {
		panic(errors.AutoMsg(fmt.Sprintf("root %d is out of range (n: %d)", root, n)))
	}
	if n <= 1 {
		return x, !comm.IsQuit()
	}
	c, ok := comm.queryChannel(cOpBcast)
	if !ok {
		return
	}
	if comm.rank != root {
		select {
		case <-comm.Ctx.Ctrl.QuitC:
			return nil, false
		case msg = <-c:
		}
	} else {
		for i := 1; i < n; i++ {
			select {
			case <-comm.Ctx.Ctrl.QuitC:
				return nil, false
			case c <- x:
			}
		}
		msg = x
	}
	return
}

func (comm *communicator) queryChannel(op int) (c chan interface{}, ok bool) {
	select {
	case <-comm.Ctx.Ctrl.QuitC:
		return
	case comm.Ctx.Ctrl.Cd.QueryChans[op] <- comm:
	}
	select {
	case <-comm.Ctx.Ctrl.QuitC:
	case c = <-comm.Cdc:
		ok = true
	}
	return
}
