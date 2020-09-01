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

	// Send the message msg to another goroutine.
	//
	// The method will block until the destination goroutine receives
	// the message successfully, or a quit signal is detected.
	// It will panic if the destination goroutine is the sender itself.
	//
	// dest is the rank of the destination goroutine.
	// msg is the message to be sent.
	//
	// It returns true if the communication succeeds,
	// otherwise (e.g., a quit signal is detected), false.
	Send(dest int, msg interface{}) bool

	// Receive a message from another goroutine.
	//
	// The method will block until it receives the message from
	// the source goroutine successfully, or a quit signal is detected.
	// It will panic if the source goroutine is the receiver itself.
	//
	// src is the rank of the source goroutine.
	//
	// It returns the message received from the source, and an indicator ok.
	// ok is true if the communication succeeds,
	// otherwise (e.g., a quit signal is detected), false.
	Receive(src int) (msg interface{}, ok bool)

	// Send the message msg to any other goroutine.
	// The destination goroutine is unspecified.
	// It will send the message to the first one ready for receiving it.
	//
	// The method will block until a goroutine receives the message
	// successfully, or a quit signal is detected.
	// It will panic if there is only one goroutine in this group.
	//
	// It returns the rank of the receiver.
	// If a quit signal is detected, it returns -1.
	SendToAny(msg interface{}) int

	// Receive a message from any other goroutine.
	// It can receive the message sent by both the method Send and SendToAny.
	// Note that when there are two or more messages sent to this goroutine,
	// this method does not guarantee that the first message will be received
	// first. Which message will be received depends on the implementation.
	//
	// The method will block until a message is received successfully,
	// or a quit signal is detected.
	// It will panic if there is only one goroutine in this group.
	//
	// It returns the rank of the sender and the received message.
	// If a quit signal is detected, it returns src to -1 and msg to nil.
	ReceiveFromAny() (src int, msg interface{})

	// Perform similar to the method ReceiveFromAny.
	// But it can only receive the message sent by the method SendToAny.
	// The returned values, and other matters, are all the same as
	// that of the method ReceiveFromAny.
	ReceiveOnlyAny() (src int, msg interface{})

	// Block until all other goroutines in this group call method Barrier
	// of their own communicators, or a quit signal is detected.
	//
	// It returns true if all other goroutines call methods Barrier
	// successfully, otherwise (e.g., a quit signal is detected), false.
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
	pcs  []chan interface{} // List of channels for point-to-point communication.
	bc   chan chan struct{} // Channel for method Barrier.
}

// Combination of sender's rank and message.
type sndrMsg struct {
	Sndr int         // Rank of the sender.
	Msg  interface{} // Message content.
}

// Combination of sender's rank, message, and a channel for
// reporting the receiver's rank.
type sndrMsgRxc struct {
	sndrMsg
	RxC chan int
}

// Create a new communicator.
// Only for function newContext.
func newCommunicator(ctx *context, rank int) *communicator {
	comm := &communicator{
		Ctx:  ctx,
		Cdc:  make(chan interface{}, 1),
		rank: rank,
		pcs:  make([]chan interface{}, len(ctx.WorldRanks)-1),
	}
	if rank > 0 {
		comm.bc = make(chan chan struct{})
	}
	for i := range comm.pcs {
		comm.pcs[i] = make(chan interface{})
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

func (comm *communicator) Send(dest int, msg interface{}) bool {
	if comm.rank == dest {
		panic(errors.AutoMsg("dest is the sender itself"))
	}
	idx := comm.rank
	if idx > dest {
		idx--
	}
	select {
	case <-comm.Ctx.Ctrl.QuitC:
		return false
	case comm.Ctx.Comms[dest].pcs[idx] <- msg:
		return true
	}
}

func (comm *communicator) Receive(src int) (msg interface{}, ok bool) {
	if comm.rank == src {
		panic(errors.AutoMsg("src is the receiver itself"))
	}
	idx := src
	if idx > comm.rank {
		idx--
	}
	select {
	case <-comm.Ctx.Ctrl.QuitC:
	case msg = <-comm.pcs[idx]:
		ok = true
	}
	return
}

func (comm *communicator) SendToAny(msg interface{}) int {
	if len(comm.Ctx.Comms) == 1 {
		panic(errors.AutoMsg("only one goroutine in this group"))
	}
	rxC := make(chan int)
	m := &sndrMsgRxc{
		sndrMsg: sndrMsg{
			Sndr: comm.rank,
			Msg:  msg,
		},
		RxC: rxC,
	}
	select {
	case <-comm.Ctx.Ctrl.QuitC:
		return -1
	case comm.Ctx.PubC <- m:
	}
	select {
	case <-comm.Ctx.Ctrl.QuitC:
		return -1
	case dest := <-rxC:
		return dest
	}
}

func (comm *communicator) ReceiveFromAny() (src int, msg interface{}) {
	if len(comm.Ctx.Comms) == 1 {
		panic(errors.AutoMsg("only one goroutine in this group"))
	}
	i, n := 0, len(comm.pcs)
	for {
		select {
		case <-comm.Ctx.Ctrl.QuitC:
			return -1, nil
		case m := <-comm.Ctx.PubC:
			select {
			case <-comm.Ctx.Ctrl.QuitC:
				return -1, nil
			case m.RxC <- comm.rank:
				return m.Sndr, m.Msg
			}
		case msg = <-comm.pcs[i]:
			src = i
			if i >= comm.rank {
				src++
			}
			return
		default:
			i = (i + 1) % n
		}
	}
}

func (comm *communicator) ReceiveOnlyAny() (src int, msg interface{}) {
	if len(comm.Ctx.Comms) == 1 {
		panic(errors.AutoMsg("only one goroutine in this group"))
	}
	select {
	case <-comm.Ctx.Ctrl.QuitC:
		return -1, nil
	case m := <-comm.Ctx.PubC:
		select {
		case <-comm.Ctx.Ctrl.QuitC:
			return -1, nil
		case m.RxC <- comm.rank:
			return m.Sndr, m.Msg
		}
	}
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
		ok = !comm.IsQuit()
		if ok {
			msg = x
		}
		return
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
		ok = !comm.IsQuit()
		if ok && x != nil {
			if a, b := x.(sequence.Array); b {
				msg = a
			} else {
				da := sequence.NewGeneralDynamicArray(x.Len())
				da.Append(x)
				msg = da
			}
		}
		return
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
		ok = !comm.IsQuit()
		if ok {
			x = []interface{}{msg}
		}
		return
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
	comm.Ctx.Ctrl.launchChannelDispatcher()
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
