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
	// The method will block until all subscribers receive
	// or buffer the message.
	//
	// It will panic if the broadcaster is closed.
	Broadcast(x Message)

	// Close closes the broadcaster.
	//
	// All channels assigned by the broadcaster will be closed to notify
	// its subscribers that there are no more messages.
	//
	// After calling the method Close, all subsequent calls to
	// the method Broadcast will panic.
	//
	// This method can take effect only once for one instance.
	// All calls after the first call will perform nothing.
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
	// It will panic if c is not gotten from the method Subscribe,
	// or c has already unsubscribed, unless the broadcaster is closed.
	// Note that this method may read messages from the channel c.
	// Set c to a channel not assigned by this broadcaster may cause data loss.
	//
	// After calling Unsubscribe, the channel c will be closed and drained
	// by the broadcaster.
	//
	// It returns all buffered and unreceived messages on the channel c
	// before unsubscribing, in order of the broadcaster sending them.
	Unsubscribe(c <-chan Message) []Message
}

// NewBroadcaster creates a new instance of interface Broadcaster.
//
// dfltBufSize is the default buffer size for the new broadcaster.
// Non-positive values for no buffer.
func NewBroadcaster[Message any](dfltBufSize int) Broadcaster[Message] {
	if dfltBufSize < 0 {
		dfltBufSize = 0
	}
	return &broadcaster[Message]{
		cm:  make(map[<-chan Message]chan<- Message),
		coi: NewOnceIndicator(),
		m:   NewMutex(),
		dbs: dfltBufSize,
	}
}

// broadcaster is an implementation of interface Broadcaster.
type broadcaster[Message any] struct {
	// Map from receive-only channels to send-only channels,
	// for sending messages to subscribers.
	cm map[<-chan Message]chan<- Message

	coi OnceIndicator // For closing the broadcaster.
	m   Mutex         // Lock for the field cm.
	bs  int32         // Broadcast semaphore.
	dbs int           // Default buffer size.
}

func (b *broadcaster[Message]) Closed() bool {
	if atomic.LoadInt32(&b.bs) > 0 {
		// The broadcaster is executing the method Broadcast,
		// which means it is not closed.
		return false
	}
	// Acquire the lock first to avoid the case that the method Close
	// is executing but not finished, and b.coi.Test() returns false.
	b.m.Lock()
	defer b.m.Unlock()
	return b.coi.Test()
}

func (b *broadcaster[Message]) Broadcast(x Message) {
	b.m.Lock()
	defer b.m.Unlock()
	if b.coi.Test() {
		panic(errors.AutoMsg("broadcaster is closed"))
	}
	atomic.AddInt32(&b.bs, 1)
	defer atomic.AddInt32(&b.bs, -1)
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
	b.coi.Do(func() {
		b.m.Lock()
		defer b.m.Unlock()
		for _, c := range b.cm {
			close(c)
		}
		b.cm = nil
	})
}

func (b *broadcaster[Message]) Subscribe(bufSize int) <-chan Message {
	if bufSize < 0 {
		bufSize = b.dbs
	}
	b.m.Lock()
	defer b.m.Unlock()
	if b.coi.Test() {
		return nil
	}
	c := make(chan Message, bufSize)

	b.cm[c] = c
	// If a compiler bug occurs on the above statement, try this:
	//  var inC <-chan Message = c
	//  b.cm[inC] = c

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
				inC = nil // Disable the channel.
				break
			}
			r = append(r, msg)
		}
	}
	defer b.m.Unlock()
	if b.coi.Test() {
		return r
	}
	outC := b.cm[c]
	if outC == nil {
		panic(errors.AutoMsg("c is not gotten from this broadcaster or has already unsubscribed"))
	}
	close(outC)
	delete(b.cm, c)
	for msg := range c {
		r = append(r, msg)
	}
	return r
}
