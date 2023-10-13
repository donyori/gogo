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
	"fmt"
	"time"

	"github.com/donyori/gogo/concurrency"
	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/errors"
)

// Communicator is a device for goroutines to communicate with each other.
type Communicator[Message any] interface {
	// Canceler returns a concurrency.Canceler for the job.
	//
	// Calls to Canceler always return the same non-nil value.
	Canceler() concurrency.Canceler

	// Rank returns the rank (from 0 to NumGoroutine()-1) of current goroutine
	// in the goroutine group of this communicator.
	Rank() int

	// NumGoroutine returns the number of goroutines in the goroutine group
	// of this communicator.
	NumGoroutine() int

	// Send sends the message msg to another goroutine.
	//
	// The method blocks until the destination goroutine receives
	// the message successfully, or a cancellation signal is detected.
	// It panics if the destination goroutine is the sender itself.
	//
	// dst is the rank of the destination goroutine.
	// msg is the message to be sent.
	//
	// It returns true if the communication succeeds,
	// otherwise (e.g., a cancellation signal is detected), false.
	Send(dst int, msg Message) bool

	// Receive receives a message from another goroutine.
	//
	// The method blocks until it receives the message from
	// the source goroutine successfully, or a cancellation signal is detected.
	// It panics if the source goroutine is the receiver itself.
	//
	// src is the rank of the source goroutine.
	//
	// It returns the message received from the source, and an indicator ok.
	// ok is true if the communication succeeds,
	// otherwise (e.g., a cancellation signal is detected), false.
	Receive(src int) (msg Message, ok bool)

	// SendPublic sends the message msg to the public channel of this group.
	// All other goroutines can get this message from the public channel.
	//
	// The method blocks until a goroutine receives this message
	// successfully, or a cancellation signal is detected.
	// It panics if there is only one goroutine in this group.
	//
	// It returns the rank of the receiver.
	// If a cancellation signal is detected, it returns -1.
	SendPublic(msg Message) int

	// ReceivePublic receives a message from the public channel of this group.
	//
	// The method blocks until it receives a message successfully,
	// or a cancellation signal is detected.
	// It panics if there is only one goroutine in this group.
	//
	// It returns the rank of the sender and the received message.
	// If a cancellation signal is detected,
	// it returns src to -1 and msg to zero value.
	ReceivePublic() (src int, msg Message)

	// SendAny sends the message msg to any other goroutines.
	// The destination goroutine is unspecified.
	// All other goroutines can receive this message from their own channels or
	// the public channel of this group.
	// Note that when there are two or more goroutines ready for receiving this
	// message, this method cannot guarantee that the first ready goroutine
	// receives the message first.
	// Which goroutine receives it depends on the implementation.
	//
	// The method blocks until a goroutine receives this message
	// successfully, or a cancellation signal is detected.
	// It panics if there is only one goroutine in this group.
	//
	// It returns the rank of the receiver.
	// If a cancellation signal is detected, it returns -1.
	//
	// To get higher performance, you should try to avoid using this method.
	// If you know which goroutine receives the message,
	// use the method Send instead.
	// If you just want to send the message to another goroutine and
	// you are sure that the receiver does not wait for this message through
	// the method Receive, use the method SendPublic instead.
	SendAny(msg Message) int

	// ReceiveAny receives a message from any other goroutines.
	// The source goroutine is unspecified.
	// It can receive a message from its own channel or
	// the public channel of this group.
	// Note that when there are two or more messages sent to this goroutine or
	// the public channel, this method cannot guarantee that the first message
	// is received first.
	// Which message is received depends on the implementation.
	//
	// The method blocks until it receives a message successfully,
	// or a cancellation signal is detected.
	// It panics if there is only one goroutine in this group.
	//
	// It returns the rank of the sender and the received message.
	// If a cancellation signal is detected,
	// it returns src to -1 and msg to zero value.
	//
	// To get higher performance, you should try to avoid using this method.
	// If you know which goroutine sends the message,
	// use the method Receive instead.
	// If you just want to receive a message from another goroutine,
	// and you are sure that the sender does not send the message through
	// the method Send, use the method ReceivePublic instead.
	ReceiveAny() (src int, msg Message)

	// Barrier blocks until all other goroutines in this group call
	// method Barrier of their own communicators,
	// or a cancellation signal is detected.
	//
	// It returns true if all other goroutines call methods Barrier
	// successfully, otherwise (e.g., a cancellation signal is detected), false.
	//
	// It is used to make all goroutines synchronous,
	// i.e., consistent in the execution progress.
	Barrier() bool

	// Broadcast sends the message x from the root to others in this group.
	//
	// The method does not wait for all goroutines to finish the broadcast.
	// To synchronize all goroutines, use method Barrier.
	//
	// root is the rank of the sender goroutine in this group.
	// It panics if root is out of range.
	//
	// For the root, x is the message to be broadcast.
	// For others, x can be anything (including zero value) and is ignored.
	//
	// It returns the message to be broadcast (equals to x of the root)
	// and an indicator ok.
	// ok is false if and only if a cancellation signal is detected.
	Broadcast(root int, x Message) (msg Message, ok bool)

	// Scatter equally divides the message x of the root into n parts,
	// where n = NumGoroutine(), and then scatters them
	// to all goroutines (including the root) in this group
	// in turn according to their ranks.
	//
	// The method does not wait for all goroutines to finish the scattering.
	// To synchronize all goroutines, use method Barrier.
	//
	// root is the rank of the sender goroutine in this group.
	// It panics if root is out of range.
	//
	// For the root, x is the message to be scattered.
	// For others, x can be anything (including nil) and is ignored.
	//
	// It returns the received message and an indicator ok.
	// ok is false if and only if a cancellation signal is detected.
	Scatter(root int, x sequence.Sequence[Message]) (
		msg array.Array[Message], ok bool)

	// Gather collects messages from all goroutines (including the root)
	// in this group to the root.
	//
	// The method does not wait for all goroutines to finish the gathering.
	// To synchronize all goroutines, use method Barrier.
	//
	// root is the rank of the receiver goroutine in this group.
	// It panics if root is out of range.
	//
	// msg is the message sent to the root.
	//
	// It returns the gathered messages as a list x, and an indicator ok.
	// For the root, x is the list of messages ordered by
	// the ranks of sender goroutines.
	// For others, x is nil.
	// ok is false if and only if a cancellation signal is detected.
	Gather(root int, msg Message) (x []Message, ok bool)
}

// communicator is an implementation of interface Communicator.
type communicator[Message any] struct {
	ctx  *context[Message]                // Context of the goroutine group.
	bcdc chan chan Message                // Channel for receiving channels from the channel dispatcher for method Broadcast.
	scdc chan []chan array.Array[Message] // Channel for receiving channel lists from the channel dispatcher for method Scatter.
	gcdc chan chan *sndrMsg[Message]      // Channel for receiving channels from the channel dispatcher for method Gather.

	rank int                // The rank of current goroutine.
	pcs  []chan Message     // List of channels for point-to-point communication.
	bc   chan chan struct{} // Channel for method Barrier.
	bCtr int64              // Counter to specify a Broadcast communication uniquely.
	sCtr int64              // Counter to specify a Scatter communication uniquely.
	gCtr int64              // Counter to specify a Gather communication uniquely.
}

// sndrMsg is a combination of the sender's rank and message.
type sndrMsg[Message any] struct {
	sndr int     // Rank of the sender.
	msg  Message // Message content.
}

// sndrMsgRxc is a combination of the sender's rank, message, and a channel
// for reporting the receiver's rank.
type sndrMsgRxc[Message any] struct {
	sndrMsg[Message]
	rxC chan int
}

// newCommunicator creates a new communicator.
// Only for function newContext.
func newCommunicator[Message any](
	ctx *context[Message],
	rank int,
) *communicator[Message] {
	comm := &communicator[Message]{
		ctx:  ctx,
		bcdc: make(chan chan Message, 1),
		scdc: make(chan []chan array.Array[Message], 1),
		gcdc: make(chan chan *sndrMsg[Message], 1),
		rank: rank,
		pcs:  make([]chan Message, len(ctx.worldRanks)-1),
	}
	if rank > 0 {
		comm.bc = make(chan chan struct{}, 1)
	}
	for i := range comm.pcs {
		comm.pcs[i] = make(chan Message) // no buffer here because the point-to-point sending requires blocking
	}
	return comm
}

func (comm *communicator[Message]) Canceler() concurrency.Canceler {
	return comm.ctx.ctrl.c
}

func (comm *communicator[Message]) Rank() int {
	return comm.rank
}

func (comm *communicator[Message]) NumGoroutine() int {
	return len(comm.ctx.comms)
}

func (comm *communicator[Message]) Send(dst int, msg Message) bool {
	if comm.rank == dst {
		panic(errors.AutoMsg("dst is the sender itself"))
	}
	idx := comm.rank
	if idx > dst {
		idx--
	}
	select {
	case <-comm.ctx.ctrl.c.C():
		return false
	case comm.ctx.comms[dst].pcs[idx] <- msg:
		return true
	}
}

func (comm *communicator[Message]) Receive(src int) (msg Message, ok bool) {
	if comm.rank == src {
		panic(errors.AutoMsg("src is the receiver itself"))
	}
	idx := src
	if idx > comm.rank {
		idx--
	}
	select {
	case <-comm.ctx.ctrl.c.C():
	case msg = <-comm.pcs[idx]:
		ok = true
	}
	return
}

func (comm *communicator[Message]) SendPublic(msg Message) int {
	if len(comm.ctx.comms) == 1 {
		panic(errors.AutoMsg("only one goroutine in this group"))
	}
	rxC := make(chan int, 1)
	m := &sndrMsgRxc[Message]{
		sndrMsg: sndrMsg[Message]{
			sndr: comm.rank,
			msg:  msg,
		},
		rxC: rxC,
	}
	cancelChan := comm.ctx.ctrl.c.C()
	select {
	case <-cancelChan:
		return -1
	case comm.ctx.pubC <- m:
	}
	select {
	case <-cancelChan:
		return -1
	case dst := <-rxC:
		return dst
	}
}

func (comm *communicator[Message]) ReceivePublic() (src int, msg Message) {
	if len(comm.ctx.comms) == 1 {
		panic(errors.AutoMsg("only one goroutine in this group"))
	}
	cancelChan := comm.ctx.ctrl.c.C()
	select {
	case <-cancelChan:
		src = -1
		return
	case m := <-comm.ctx.pubC:
		select {
		case <-cancelChan:
			src = -1
			return
		case m.rxC <- comm.rank:
			return m.sndr, m.msg
		}
	}
}

func (comm *communicator[Message]) SendAny(msg Message) int {
	if len(comm.ctx.comms) == 1 {
		panic(errors.AutoMsg("only one goroutine in this group"))
	}
	rxC := make(chan int, 1)
	m := &sndrMsgRxc[Message]{
		sndrMsg: sndrMsg[Message]{
			sndr: comm.rank,
			msg:  msg,
		},
		rxC: rxC,
	}
	cancelChan, n := comm.ctx.ctrl.c.C(), len(comm.ctx.comms)
	poll := func() int {
		for dst := 0; dst < n; dst++ {
			if dst == comm.rank {
				dst++
			}
			var idx int
			if comm.rank > dst {
				idx = comm.rank - 1
			} else {
				idx = comm.rank
			}
			select {
			case <-cancelChan:
				return -1
			case comm.ctx.pubC <- m:
				select {
				case <-cancelChan:
					return -1
				case dst := <-rxC:
					return dst
				}
			case comm.ctx.comms[dst].pcs[idx] <- msg:
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
		case <-cancelChan:
			return -1
		case comm.ctx.pubC <- m:
			select {
			case <-cancelChan:
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

func (comm *communicator[Message]) ReceiveAny() (src int, msg Message) {
	if len(comm.ctx.comms) == 1 {
		panic(errors.AutoMsg("only one goroutine in this group"))
	}
	i, n := 0, len(comm.pcs)
	cancelChan := comm.ctx.ctrl.c.C()
	for {
		select {
		case <-cancelChan:
			src = -1
			return
		case m := <-comm.ctx.pubC:
			select {
			case <-cancelChan:
				src = -1
				return
			case m.rxC <- comm.rank:
				return m.sndr, m.msg
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

func (comm *communicator[Message]) Barrier() bool {
	n := len(comm.ctx.comms)
	if n <= 1 {
		return !comm.ctx.ctrl.c.Canceled()
	}
	var c chan struct{}
	cancelChan := comm.ctx.ctrl.c.C()
	if comm.rank == 0 {
		// The first goroutine makes the signal channel.
		c = make(chan struct{})
	} else {
		// Other goroutines receive the signal channel from their respective previous goroutines.
		select {
		case <-cancelChan:
			return false
		case c = <-comm.ctx.comms[comm.rank].bc: // listen on its own channel, not the sender's!
		}
	}
	if comm.rank == n-1 {
		// The last goroutine close the signal channel to broadcast
		// the information that all goroutines call method Barrier successfully.
		close(c)
	} else {
		// Other goroutines send the signal channel to their respective next goroutines.
		select {
		case <-cancelChan:
			return false
		case comm.ctx.comms[comm.rank+1].bc <- c: // send it to the receiver's channel
		}
		// Then listen on the signal channel.
		select {
		case <-cancelChan:
			return false
		case <-c:
		}
	}
	return true
}

func (comm *communicator[Message]) Broadcast(root int, x Message) (
	msg Message, ok bool) {
	if comm.checkRootAndN(root) {
		// No other goroutines in this group.
		ok = !comm.ctx.ctrl.c.Canceled()
		if ok {
			msg = x
		}
		return
	}

	comm.ctx.ctrl.LaunchChannelDispatcher()
	qry := &chanDispQry[Message]{
		comm: comm,
		ctr:  comm.bCtr,
	}
	comm.bCtr++ // update counter before communication
	cancelChan := comm.ctx.ctrl.c.C()
	// Send channel dispatch query:
	select {
	case <-cancelChan:
		return
	case comm.ctx.ctrl.cd.bcastChan <- qry:
	}
	// Wait for channel dispatcher:
	var c chan Message
	select {
	case <-cancelChan:
		return
	case c = <-comm.bcdc:
	}

	if comm.rank != root {
		select {
		case <-cancelChan:
			return
		case msg = <-c:
		}
	} else {
		for i, n := 1, len(comm.ctx.comms); i < n; i++ {
			select {
			case <-cancelChan:
				return
			case c <- x:
			}
		}
		msg = x
	}
	return msg, true
}

func (comm *communicator[Message]) Scatter(
	root int,
	x sequence.Sequence[Message],
) (msg array.Array[Message], ok bool) {
	if comm.checkRootAndN(root) {
		// No other goroutines in this group.
		ok = !comm.ctx.ctrl.c.Canceled()
		if ok && x != nil {
			if a, b := any(x).(array.Array[Message]); b {
				msg = a
			} else {
				sda := make(array.SliceDynamicArray[Message], 0, x.Len())
				sda.Append(x)
				msg = &sda
			}
		}
		return
	}

	comm.ctx.ctrl.LaunchChannelDispatcher()
	qry := &chanDispQry[Message]{
		comm: comm,
		ctr:  comm.sCtr,
	}
	comm.sCtr++ // update counter before communication
	cancelChan := comm.ctx.ctrl.c.C()
	// Send channel dispatch query:
	select {
	case <-cancelChan:
		return
	case comm.ctx.ctrl.cd.scatterChan <- qry:
	}
	// Wait for channel dispatcher:
	var cs []chan array.Array[Message]
	select {
	case <-cancelChan:
		return
	case cs = <-comm.scdc:
	}

	if comm.rank != root {
		idx := comm.rank
		if idx > root {
			idx--
		}
		select {
		case <-cancelChan:
			return
		case msg = <-cs[idx]:
		}
	} else {
		if x == nil || x.Len() <= 0 {
			for _, c := range cs {
				close(c)
			}
			return nil, true
		}

		size := x.Len()
		a, b := any(x).(array.Array[Message])
		if !b {
			sda := make(array.SliceDynamicArray[Message], 0, size)
			sda.Append(x)
			a = &sda
		}
		n, idx, cIdx := len(comm.ctx.comms), 0, 0
		q, r := size/n, size%n
		chunkLen := q + 1
		for i := 0; i < n; i++ {
			if i == r {
				chunkLen--
			}
			chunk := a.Slice(idx, idx+chunkLen)
			idx += chunkLen
			if i != root {
				// Send this chunk to the target goroutine.
				select {
				case <-cancelChan:
					return nil, false
				case cs[cIdx] <- chunk:
					cIdx++
				}
			} else {
				// This chunk is for the root itself.
				msg = chunk
			}
		}
	}
	return msg, true
}

func (comm *communicator[Message]) Gather(root int, msg Message) (
	x []Message, ok bool) {
	if comm.checkRootAndN(root) {
		// No other goroutines in this group.
		ok = !comm.ctx.ctrl.c.Canceled()
		if ok {
			x = []Message{msg}
		}
		return
	}

	comm.ctx.ctrl.LaunchChannelDispatcher()
	qry := &chanDispQry[Message]{
		comm: comm,
		ctr:  comm.gCtr,
	}
	comm.gCtr++ // update counter before communication
	cancelChan := comm.ctx.ctrl.c.C()
	// Send channel dispatch query:
	select {
	case <-cancelChan:
		return
	case comm.ctx.ctrl.cd.gatherChan <- qry:
	}
	// Wait for channel dispatcher:
	var c chan *sndrMsg[Message]
	select {
	case <-cancelChan:
		return
	case c = <-comm.gcdc:
	}

	if comm.rank != root {
		select {
		case <-cancelChan:
			return
		case c <- &sndrMsg[Message]{comm.rank, msg}:
		}
	} else {
		x = make([]Message, len(comm.ctx.comms))
		x[comm.rank] = msg
		for i, n := 1, len(x); i < n; i++ {
			select {
			case <-cancelChan:
				return nil, false
			case m := <-c:
				x[m.sndr] = m.msg
			}
		}
	}
	return x, true
}

// checkRootAndN panics if root is out of range.
// It returns true if and only if comm.NumGoroutine() <= 1.
func (comm *communicator[Message]) checkRootAndN(root int) bool {
	n := len(comm.ctx.comms)
	if root < 0 || root >= n {
		panic(errors.AutoMsgCustom(
			fmt.Sprintf("root %d is out of range (n: %d)", root, n), -1, 1))
	}
	return n <= 1
}
