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

import (
	"fmt"
	"time"

	"github.com/donyori/gogo/concurrency/framework"
	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/errors"
)

// Communicator is a device for goroutines to communicate with each other.
type Communicator interface {
	framework.QuitDevice

	// Rank returns the rank (from 0 to NumGoroutine()-1) of current goroutine
	// in the goroutine group of this communicator.
	Rank() int

	// NumGoroutine returns the number of goroutines in the goroutine group
	// of this communicator.
	NumGoroutine() int

	// Send sends the message msg to another goroutine.
	//
	// The method will block until the destination goroutine receives
	// the message successfully, or a quit signal is detected.
	// It will panic if the destination goroutine is the sender itself.
	//
	// dst is the rank of the destination goroutine.
	// msg is the message to be sent.
	//
	// It returns true if the communication succeeds,
	// otherwise (e.g., a quit signal is detected), false.
	Send(dst int, msg interface{}) bool

	// Receive receives a message from another goroutine.
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

	// SendPublic sends the message msg to the public channel of this group.
	// All other goroutines can get this message from the public channel.
	//
	// The method will block until a goroutine receives this message
	// successfully, or a quit signal is detected.
	// It will panic if there is only one goroutine in this group.
	//
	// It returns the rank of the receiver.
	// If a quit signal is detected, it returns -1.
	SendPublic(msg interface{}) int

	// ReceivePublic receives a message from the public channel of this group.
	//
	// The method will block until it receives a message successfully,
	// or a quit signal is detected.
	// It will panic if there is only one goroutine in this group.
	//
	// It returns the rank of the sender and the received message.
	// If a quit signal is detected, it returns src to -1 and msg to nil.
	ReceivePublic() (src int, msg interface{})

	// SendAny sends the message msg to any other goroutines.
	// The destination goroutine is unspecified.
	// All other goroutines can receive this message from their own channels or
	// the public channel of this group.
	// Note that when there are two or more goroutines ready for receiving this
	// message, this method cannot guarantee that the first ready goroutine
	// will receive the message first.
	// Which message will receive it depends on the implementation.
	//
	// The method will block until a goroutine receives this message
	// successfully, or a quit signal is detected.
	// It will panic if there is only one goroutine in this group.
	//
	// It returns the rank of the receiver.
	// If a quit signal is detected, it returns -1.
	//
	// To get higher performance, you should try to avoid using this method.
	// If you know which goroutine will receive the message,
	// use the method Send instead.
	// If you just want to send the message to another goroutine and
	// you are sure that the receiver won't wait for this message through
	// the method Receive, use the method SendPublic instead.
	SendAny(msg interface{}) int

	// ReceiveAny receives a message from any other goroutines.
	// The source goroutine is unspecified.
	// It can receive a message from its own channel or
	// the public channel of this group.
	// Note that when there are two or more messages sent to this goroutine or
	// the public channel, this method cannot guarantee that the first message
	// will be received first.
	// Which message will be received depends on the implementation.
	//
	// The method will block until it receives a message successfully,
	// or a quit signal is detected.
	// It will panic if there is only one goroutine in this group.
	//
	// It returns the rank of the sender and the received message.
	// If a quit signal is detected, it returns src to -1 and msg to nil.
	//
	// To get higher performance, you should try to avoid using this method.
	// If you know which goroutine will send the message,
	// use the method Receive instead.
	// If you just want to receive a message from another goroutine and
	// you are sure that the sender won't send the message through
	// the method Send, use the method ReceivePublic instead.
	ReceiveAny() (src int, msg interface{})

	// Barrier blocks until all other goroutines in this group call
	// method Barrier of their own communicators, or a quit signal is detected.
	//
	// It returns true if all other goroutines call methods Barrier
	// successfully, otherwise (e.g., a quit signal is detected), false.
	//
	// It is used to make all goroutines synchronous,
	// i.e., consistent in the execution progress.
	Barrier() bool

	// Broadcast sends the message x from the root to others in this group.
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

	// Scatter equally divides the message x of the root into n parts,
	// where n = NumGoroutine(), and then scatters them
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

	// Gather collects messages from all goroutines (including the root)
	// in this group to the root.
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

// communicator is an implementation of interface Communicator.
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

// sndrMsg is a combination of the sender's rank and message.
type sndrMsg struct {
	Sndr int         // Rank of the sender.
	Msg  interface{} // Message content.
}

// sndrMsgRxc is a combination of the sender's rank, message, and a channel
// for reporting the receiver's rank.
type sndrMsgRxc struct {
	sndrMsg
	RxC chan int
}

// newCommunicator creates a new communicator.
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

// QuitChan returns the channel for the quit signal.
// When the job is finished or quit, this channel will be closed
// to broadcast the quit signal.
func (comm *communicator) QuitChan() <-chan struct{} {
	return comm.Ctx.Ctrl.Qd.QuitChan()
}

// IsQuit detects the quit signal on the quit channel.
// It returns true if a quit signal is detected, and false otherwise.
func (comm *communicator) IsQuit() bool {
	return comm.Ctx.Ctrl.Qd.IsQuit()
}

// Quit broadcasts a quit signal to quit the job.
//
// This method will NOT wait until the job ends.
func (comm *communicator) Quit() {
	comm.Ctx.Ctrl.Qd.Quit()
}

// Rank returns the rank (from 0 to NumGoroutine()-1) of current goroutine
// in the goroutine group of this communicator.
func (comm *communicator) Rank() int {
	return comm.rank
}

// NumGoroutine returns the number of goroutines in the goroutine group
// of this communicator.
func (comm *communicator) NumGoroutine() int {
	return len(comm.Ctx.Comms)
}

// Send sends the message msg to another goroutine.
//
// The method will block until the destination goroutine receives
// the message successfully, or a quit signal is detected.
// It will panic if the destination goroutine is the sender itself.
//
// dst is the rank of the destination goroutine.
// msg is the message to be sent.
//
// It returns true if the communication succeeds,
// otherwise (e.g., a quit signal is detected), false.
func (comm *communicator) Send(dst int, msg interface{}) bool {
	if comm.rank == dst {
		panic(errors.AutoMsg("dst is the sender itself"))
	}
	idx := comm.rank
	if idx > dst {
		idx--
	}
	select {
	case <-comm.Ctx.Ctrl.Qd.QuitChan():
		return false
	case comm.Ctx.Comms[dst].pcs[idx] <- msg:
		return true
	}
}

// Receive receives a message from another goroutine.
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
func (comm *communicator) Receive(src int) (msg interface{}, ok bool) {
	if comm.rank == src {
		panic(errors.AutoMsg("src is the receiver itself"))
	}
	idx := src
	if idx > comm.rank {
		idx--
	}
	select {
	case <-comm.Ctx.Ctrl.Qd.QuitChan():
	case msg = <-comm.pcs[idx]:
		ok = true
	}
	return
}

// SendPublic sends the message msg to the public channel of this group.
// All other goroutines can get this message from the public channel.
//
// The method will block until a goroutine receives this message
// successfully, or a quit signal is detected.
// It will panic if there is only one goroutine in this group.
//
// It returns the rank of the receiver.
// If a quit signal is detected, it returns -1.
func (comm *communicator) SendPublic(msg interface{}) int {
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
	quitChan := comm.Ctx.Ctrl.Qd.QuitChan()
	select {
	case <-quitChan:
		return -1
	case comm.Ctx.PubC <- m:
	}
	select {
	case <-quitChan:
		return -1
	case dst := <-rxC:
		return dst
	}
}

// ReceivePublic receives a message from the public channel of this group.
//
// The method will block until it receives a message successfully,
// or a quit signal is detected.
// It will panic if there is only one goroutine in this group.
//
// It returns the rank of the sender and the received message.
// If a quit signal is detected, it returns src to -1 and msg to nil.
func (comm *communicator) ReceivePublic() (src int, msg interface{}) {
	if len(comm.Ctx.Comms) == 1 {
		panic(errors.AutoMsg("only one goroutine in this group"))
	}
	quitChan := comm.Ctx.Ctrl.Qd.QuitChan()
	select {
	case <-quitChan:
		return -1, nil
	case m := <-comm.Ctx.PubC:
		select {
		case <-quitChan:
			return -1, nil
		case m.RxC <- comm.rank:
			return m.Sndr, m.Msg
		}
	}
}

// SendAny sends the message msg to any other goroutines.
// The destination goroutine is unspecified.
// All other goroutines can receive this message from their own channels or
// the public channel of this group.
// Note that when there are two or more goroutines ready for receiving this
// message, this method cannot guarantee that the first ready goroutine
// will receive the message first.
// Which message will receive it depends on the implementation.
//
// The method will block until a goroutine receives this message
// successfully, or a quit signal is detected.
// It will panic if there is only one goroutine in this group.
//
// It returns the rank of the receiver.
// If a quit signal is detected, it returns -1.
//
// To get higher performance, you should try to avoid using this method.
// If you know which goroutine will receive the message,
// use the method Send instead.
// If you just want to send the message to another goroutine and
// you are sure that the receiver won't wait for this message through
// the method Receive, use the method SendPublic instead.
func (comm *communicator) SendAny(msg interface{}) int {
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
	n := len(comm.Ctx.Comms)
	quitChan := comm.Ctx.Ctrl.Qd.QuitChan()
	var dst, idx int
	poll := func() int {
		for dst = 0; dst < n; dst++ {
			if dst == comm.rank {
				dst++
			}
			if comm.rank > dst {
				idx = comm.rank - 1
			} else {
				idx = comm.rank
			}
			select {
			case <-quitChan:
				return -1
			case comm.Ctx.PubC <- m:
				select {
				case <-quitChan:
					return -1
				case dst := <-rxC:
					return dst
				}
			case comm.Ctx.Comms[dst].pcs[idx] <- msg:
				return dst
			default:
			}
		}
		return -2
	}
	pollR := poll()
	if pollR > -2 {
		return pollR
	}
	ticker := time.NewTicker(time.Duration(n-1)*50 + 100)
	defer ticker.Stop()
	for {
		select {
		case <-quitChan:
			return -1
		case comm.Ctx.PubC <- m:
			select {
			case <-quitChan:
				return -1
			case dst := <-rxC:
				return dst
			}
		case <-ticker.C:
			pollR = poll()
			if pollR > -2 {
				return pollR
			}
		}
	}
}

// ReceiveAny receives a message from any other goroutines.
// The source goroutine is unspecified.
// It can receive a message from its own channel or
// the public channel of this group.
// Note that when there are two or more messages sent to this goroutine or
// the public channel, this method cannot guarantee that the first message
// will be received first.
// Which message will be received depends on the implementation.
//
// The method will block until it receives a message successfully,
// or a quit signal is detected.
// It will panic if there is only one goroutine in this group.
//
// It returns the rank of the sender and the received message.
// If a quit signal is detected, it returns src to -1 and msg to nil.
//
// To get higher performance, you should try to avoid using this method.
// If you know which goroutine will send the message,
// use the method Receive instead.
// If you just want to receive a message from another goroutine and
// you are sure that the sender won't send the message through
// the method Send, use the method ReceivePublic instead.
func (comm *communicator) ReceiveAny() (src int, msg interface{}) {
	if len(comm.Ctx.Comms) == 1 {
		panic(errors.AutoMsg("only one goroutine in this group"))
	}
	i, n := 0, len(comm.pcs)
	quitChan := comm.Ctx.Ctrl.Qd.QuitChan()
	for {
		select {
		case <-quitChan:
			return -1, nil
		case m := <-comm.Ctx.PubC:
			select {
			case <-quitChan:
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

// Barrier blocks until all other goroutines in this group call
// method Barrier of their own communicators, or a quit signal is detected.
//
// It returns true if all other goroutines call methods Barrier
// successfully, otherwise (e.g., a quit signal is detected), false.
//
// It is used to make all goroutines synchronous,
// i.e., consistent in the execution progress.
func (comm *communicator) Barrier() bool {
	n := len(comm.Ctx.Comms)
	if n <= 1 {
		return !comm.Ctx.Ctrl.Qd.IsQuit()
	}
	var c chan struct{}
	quitChan := comm.Ctx.Ctrl.Qd.QuitChan()
	if comm.rank == 0 {
		// The first goroutine makes the signal channel.
		c = make(chan struct{})
	} else {
		// Other goroutines receive the signal channel from their respective previous goroutines.
		select {
		case <-quitChan:
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
		case <-quitChan:
			return false
		case comm.Ctx.Comms[comm.rank+1].bc <- c: // Send it to the receiver's channel.
		}
		// Then listen on the signal channel.
		select {
		case <-quitChan:
			return false
		case <-c:
		}
	}
	return true
}

// Broadcast sends the message x from the root to others in this group.
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
func (comm *communicator) Broadcast(root int, x interface{}) (msg interface{}, ok bool) {
	if comm.checkRootAndN(root) {
		ok = !comm.Ctx.Ctrl.Qd.IsQuit()
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
	quitChan := comm.Ctx.Ctrl.Qd.QuitChan()
	if comm.rank != root {
		select {
		case <-quitChan:
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
			case <-quitChan:
				return nil, false
			case c <- x:
			}
		}
		msg = x
	}
	return
}

// Scatter equally divides the message x of the root into n parts,
// where n = NumGoroutine(), and then scatters them
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
func (comm *communicator) Scatter(root int, x sequence.Sequence) (msg sequence.Array, ok bool) {
	if comm.checkRootAndN(root) {
		ok = !comm.Ctx.Ctrl.Qd.IsQuit()
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
	quitChan := comm.Ctx.Ctrl.Qd.QuitChan()
	if comm.rank != root {
		idx := comm.rank
		if idx > root {
			idx--
		}
		select {
		case <-quitChan:
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
				case <-quitChan:
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

// Gather collects messages from all goroutines (including the root)
// in this group to the root.
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
func (comm *communicator) Gather(root int, msg interface{}) (x []interface{}, ok bool) {
	if comm.checkRootAndN(root) {
		ok = !comm.Ctx.Ctrl.Qd.IsQuit()
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
	quitChan := comm.Ctx.Ctrl.Qd.QuitChan()
	if comm.rank != root {
		select {
		case <-quitChan:
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
			case <-quitChan:
				return nil, false
			case m := <-c:
				x[m.Sndr] = m.Msg
			}
		}
	}
	return
}

// checkRootAndN panics if root is out of range.
// It returns true iff comm.NumGoroutine() <= 1.
func (comm *communicator) checkRootAndN(root int) bool {
	n := len(comm.Ctx.Comms)
	if root < 0 || root >= n {
		panic(errors.AutoMsgWithStrategy(fmt.Sprintf("root %d is out of range (n: %d)", root, n), -1, 1))
	}
	return n <= 1
}

// queryChannels acquires channels from the channel dispatcher.
//
// It returns the channel, or list of channels, as interface{},
// and an indicator ok. ok is false iff a quit signal is detected.
func (comm *communicator) queryChannels(op int) (chanItf interface{}, ok bool) {
	comm.Ctx.Ctrl.launchChannelDispatcher()
	quitChan := comm.Ctx.Ctrl.Qd.QuitChan()
	select {
	case <-quitChan:
		return
	case comm.Ctx.Ctrl.Cd[op] <- comm:
	}
	select {
	case <-quitChan:
	case chanItf = <-comm.Cdc:
		ok = true
	}
	return
}
