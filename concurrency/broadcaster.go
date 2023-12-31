// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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

package concurrency

import (
	"sync/atomic"

	"github.com/donyori/gogo/errors"
)

// Broadcaster is a device to broadcast messages to its subscriber.
// It is used to send messages from one sender to multiple receivers.
//
// The sender should call the method Broadcast to send a message to
// all receivers (aka subscribers).
// When there are no more messages to send, the sender can call
// the method Close to notify all receivers.
//
// The receiver should first call the method Subscribe to acquire a channel
// for receiving messages.
// Then, the receiver can get messages by listening to the channel.
// Finally, when the receiver no longer needs to get messages from this
// broadcaster, it can call the method Unsubscribe.
type Broadcaster[Message any] interface {
	// Closed reports whether the broadcaster is closed.
	Closed() bool

	// Broadcast sends the message x to all subscribers.
	//
	// The method blocks until all subscribers receive
	// or buffer the message.
	//
	// It panics if the broadcaster is closed.
	Broadcast(x Message)

	// Close closes the broadcaster.
	//
	// All channels assigned by the broadcaster are closed to notify
	// its subscribers that there are no more messages.
	//
	// After calling the method Close,
	// all subsequent calls to the method Broadcast panic.
	//
	// This method can take effect only once for one instance.
	// After the first call, subsequent calls to Close do nothing.
	//
	// This method should only be used by the sender.
	Close()

	// Subscribe subscribes the broadcaster to get messages from the sender.
	//
	// bufSize is the buffer size of the returned channel.
	// 0 for no buffer.
	// Negative values for using the default buffer size.
	//
	// It returns a channel for receiving messages from the sender.
	// If the broadcaster is closed, it returns nil.
	Subscribe(bufSize int) <-chan Message

	// Unsubscribe unsubscribes the broadcaster to stop receiving messages.
	//
	// c is the channel acquired by the method Subscribe.
	// It panics if c is not gotten from the method Subscribe,
	// or c has already unsubscribed, unless the broadcaster is closed.
	// Note that this method may read messages from the channel c.
	// Setting c to a channel not assigned by this broadcaster
	// may cause data loss.
	//
	// After calling Unsubscribe, the channel c is closed and drained
	// by the broadcaster.
	//
	// It returns all buffered and unreceived messages on the channel c
	// before unsubscribing, in order of the broadcaster sending them.
	Unsubscribe(c <-chan Message) []Message
}

// NewBroadcaster creates a new Broadcaster.
//
// dfltBufSize is the default buffer size for the new broadcaster.
// Non-positive values for no buffer.
func NewBroadcaster[Message any](dfltBufSize int) Broadcaster[Message] {
	if dfltBufSize < 0 {
		dfltBufSize = 0
	}
	b := &broadcaster[Message]{
		cm:  make(map[<-chan Message]chan<- Message),
		m:   NewMutex(),
		dbs: dfltBufSize,
	}
	b.co = NewOnce(b.closeProc)
	return b
}

// broadcaster is an implementation of interface Broadcaster.
type broadcaster[Message any] struct {
	// Map from receive-only channels to send-only channels,
	// for sending messages to subscribers.
	cm map[<-chan Message]chan<- Message

	m   Mutex        // Lock for the field cm.
	co  Once         // For closing the broadcaster.
	bs  atomic.Int32 // Broadcast semaphore.
	dbs int          // Default buffer size.
}

func (b *broadcaster[Message]) Closed() bool {
	if b.bs.Load() > 0 {
		// The broadcaster is executing the method Broadcast,
		// which means it is not closed.
		return false
	}
	// Acquire the lock first to avoid the case that the method Close
	// is executing but not finished, and b.co.Done() returns false.
	b.m.Lock()
	defer b.m.Unlock()
	return b.co.Done()
}

func (b *broadcaster[Message]) Broadcast(x Message) {
	b.m.Lock()
	defer b.m.Unlock()
	if b.co.Done() {
		panic(errors.AutoMsg("broadcaster is closed"))
	}
	b.bs.Add(1)
	defer b.bs.Add(-1)
	var unsent []chan<- Message
	for _, c := range b.cm {
		select {
		case c <- x:
		default:
			unsent = append(unsent, c)
		}
	}
	if unsent == nil {
		return
	}

	// Implement a singly linked list using slice.
	// When sending x to a channel successfully, remove the channel from
	// the list by modifying "head" and "next".
	// When the list is empty, the broadcast finishes.
	n := len(unsent)
	next, head := make([]int, n), 0
	for i := range next {
		next[i] = i + 1
	}
	for head < n {
		for i, last := head, -1; i < n; i = next[i] {
			select {
			case unsent[i] <- x:
				if last >= 0 {
					next[last] = next[i]
				} else {
					head = next[i]
				}
			default:
				last = i
			}
		}
	}
}

func (b *broadcaster[Message]) Close() {
	b.co.Do()
}

func (b *broadcaster[Message]) Subscribe(bufSize int) <-chan Message {
	if bufSize < 0 {
		bufSize = b.dbs
	}
	b.m.Lock()
	defer b.m.Unlock()
	if b.co.Done() {
		return nil
	}
	c := make(chan Message, bufSize)

	b.cm[c] = c
	// If a compiler bug occurs on the above statement, try this:
	//
	//	var inC <-chan Message = c
	//	b.cm[inC] = c

	return c
}

func (b *broadcaster[Message]) Unsubscribe(c <-chan Message) []Message {
	var r []Message
	inC, unlocked := c, true
	for unlocked {
		select {
		case <-b.m.C():
			unlocked = false
		case msg, ok := <-inC:
			if !ok {
				inC = nil // disable the channel
				break
			}
			r = append(r, msg)
		}
	}
	defer b.m.Unlock()
	if b.co.Done() {
		return r
	}
	outC := b.cm[c]
	if outC == nil {
		panic(errors.AutoMsg(
			"c is not gotten from this broadcaster or has already unsubscribed"))
	}
	close(outC)
	delete(b.cm, c)
	for msg := range c {
		r = append(r, msg)
	}
	return r
}

// closeProc is the process of closing the broadcaster.
// It is invoked by b.co.Do.
func (b *broadcaster[Message]) closeProc() {
	b.m.Lock()
	defer b.m.Unlock()
	for _, c := range b.cm {
		close(c)
	}
	b.cm = nil
}
