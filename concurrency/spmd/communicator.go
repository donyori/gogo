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

	"github.com/donyori/gogo/container/sequence"
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
	//
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

	// Equally divide the message x of the root into n parts,
	// where n = NumGoroutine(), and then scatter them
	// to all goroutines (including the root) in this group
	// in turn according to their ranks.
	//
	// The method will not wait for all goroutines to finish the scattering.
	// To synchronize all goroutines, use method Barrier.
	//
	// root is the rank of the sender goroutine in this group.
	// It will panic if root is out of range.
	//
	// For the root, x is the message to be scattered.
	// For others, x can be anything (including nil) and will be ignored.
	//
	// It returns the received message and an indicator ok.
	// ok is false iff a quit signal is detected.
	Scatter(root int, x sequence.Sequence) (msg sequence.Array, ok bool)

	// Gather messages from all goroutines (including the root) in this group
	// to the root.
	//
	// The method will not wait for all goroutines to finish the gathering.
	// To synchronize all goroutines, use method Barrier.
	//
	// root is the rank of the receiver goroutine in this group.
	// It will panic if root is out of range.
	//
	// msg is the message to be sent to the root.
	//
	// It returns the gathered messages as a list x, and an indicator ok.
	// For the root, x is the list of messages ordered by
	// the ranks of sender goroutines.
	// For others, x is nil.
	// ok is false iff a quit signal is detected.
	Gather(root int, msg interface{}) (x []interface{}, ok bool)
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
	Ctx *context         // Context of the goroutine group.
	Cdc chan interface{} // Channel for receiving channels form the channel dispatcher.
	// Counters for cluster communication operations.
	// Only for method chanDispr.Run.
	COpCntrs [numCOp]int64

	rank int                // The rank of current goroutine.
	bc   chan chan struct{} // A channel used for method Barrier.
}

// Combination of sender's rank and message.
type sndrMsg struct {
	Sndr int         // Rank of the sender.
	Msg  interface{} // Message content.
}

// Create a new communicator.
// Only for function newContext.
func newCommunicator(ctx *context, rank int) *communicator {
	comm := &communicator{
		Ctx:  ctx,
		Cdc:  make(chan interface{}, 1),
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
	if comm.checkRootAndN(root) {
		return x, !comm.IsQuit()
	}
	chanItf, ok := comm.queryChannels(cOpBcast)
	if !ok {
		return
	}
	c := chanItf.(chan interface{})
	if comm.rank != root {
		select {
		case <-comm.Ctx.Ctrl.QuitC:
			return nil, false
		case msg = <-c:
		}
	} else {
		if x == nil {
			close(c)
			return
		}
		for i, n := 1, len(comm.Ctx.Comms); i < n; i++ {
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

func (comm *communicator) Scatter(root int, x sequence.Sequence) (msg sequence.Array, ok bool) {
	if comm.checkRootAndN(root) {
		if x != nil {
			if a, b := x.(sequence.Array); b {
				msg = a
			} else {
				da := sequence.NewGeneralDynamicArray(x.Len())
				da.Append(x)
				msg = da
			}
		}
		return msg, !comm.IsQuit()
	}
	chanItf, ok := comm.queryChannels(cOpScatter)
	if !ok {
		return
	}
	cs := chanItf.([]chan interface{})
	if comm.rank != root {
		idx := comm.rank
		if idx > root {
			idx--
		}
		select {
		case <-comm.Ctx.Ctrl.QuitC:
			return nil, false
		case itf := <-cs[idx]:
			if itf != nil {
				msg = itf.(sequence.Array)
			}
		}
	} else {
		if x == nil || x.Len() <= 0 {
			for _, c := range cs {
				close(c)
			}
			return
		}
		size := x.Len()
		a, b := x.(sequence.Array)
		if !b {
			da := sequence.NewGeneralDynamicArray(size)
			da.Append(x)
			a = da
		}
		n, idx, cIdx := len(comm.Ctx.Comms), 0, 0
		q, r := size/n, size%n
		inc := q + 1
		var chunk sequence.Array
		for i := 0; i < n; i++ {
			if i == r {
				inc--
			}
			chunk = a.Slice(idx, idx+inc)
			idx += inc
			if i != root {
				select {
				case <-comm.Ctx.Ctrl.QuitC:
					return nil, false
				case cs[cIdx] <- chunk:
					cIdx++
				}
			} else {
				msg = chunk
			}
		}
	}
	return
}

func (comm *communicator) Gather(root int, msg interface{}) (x []interface{}, ok bool) {
	if comm.checkRootAndN(root) {
		return []interface{}{msg}, !comm.IsQuit()
	}
	chanItf, ok := comm.queryChannels(cOpGather)
	if !ok {
		return
	}
	c := chanItf.(chan *sndrMsg)
	if comm.rank != root {
		select {
		case <-comm.Ctx.Ctrl.QuitC:
			return nil, false
		case c <- &sndrMsg{
			Sndr: comm.rank,
			Msg:  msg,
		}:
		}
	} else {
		x = make([]interface{}, len(comm.Ctx.Comms))
		x[comm.rank] = msg
		for i, n := 1, len(x); i < n; i++ {
			select {
			case <-comm.Ctx.Ctrl.QuitC:
				return nil, false
			case m := <-c:
				x[m.Sndr] = m.Msg
			}
		}
	}
	return
}

// It panics if root is out of range.
// It returns true iff comm.NumGoroutine() <= 1.
func (comm *communicator) checkRootAndN(root int) bool {
	n := len(comm.Ctx.Comms)
	if root < 0 || root >= n {
		panic(errors.AutoMsgWithStrategy(fmt.Sprintf("root %d is out of range (n: %d)", root, n), -1, 1))
	}
	return n <= 1
}

// Query channels from the channel dispatcher.
//
// It returns the channel, or list of channels, as interface{},
// and an indicator ok. ok is false iff a quit signal is detected.
func (comm *communicator) queryChannels(op int) (chanItf interface{}, ok bool) {
	select {
	case <-comm.Ctx.Ctrl.QuitC:
		return
	case comm.Ctx.Ctrl.Cd.QueryChans[op] <- comm:
	}
	select {
	case <-comm.Ctx.Ctrl.QuitC:
	case chanItf = <-comm.Cdc:
		ok = true
	}
	return
}
