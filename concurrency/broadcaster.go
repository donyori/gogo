// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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
	"context"

	"github.com/donyori/gogo/errors"
)

// Broadcaster is a device to broadcast messages to its subscribers.
// It is used to send messages from one sender to multiple receivers.
//
// The sender should call the method Broadcast to send a message to
// all receivers (i.e., subscribers).
// The broadcaster may launch a daemon goroutine after creation.
// Therefore, when there are no more messages to send,
// the sender should call the method Close as soon as possible
// to notify all receivers and release resources associated with the broadcaster
// (e.g., stopping the daemon goroutine).
//
// The receiver should first call the method Subscribe to
// subscribe to the broadcaster and acquire a channel for receiving messages.
// Then, the receiver can receive messages by listening to the channel.
// Finally, when the receiver no longer needs to receive messages from
// the broadcaster, it should call the method Unsubscribe as soon as possible
// to notify the sender and avoid blocking the sender.
type Broadcaster[Message any] interface {
	// Close closes the broadcaster.
	// Close does not interrupt a running broadcast and
	// takes effect after the broadcast ends.
	//
	// All channels assigned by the broadcaster are closed to notify
	// its subscribers that there are no more messages.
	//
	// After closing the broadcaster,
	// all subsequent calls to the method Broadcast
	// do nothing and report ErrBroadcasterClosed.
	//
	// Close can take effect only once for one instance.
	// After the first call, subsequent calls to Close do nothing.
	//
	// Close should only be called by the sender.
	//
	// When there are no more messages to send,
	// the sender should call Close as soon as possible to notify all receivers
	// and release resources associated with the broadcaster
	// (e.g., stopping the daemon goroutine).
	Close()

	// CloseC returns a channel for the close of the broadcaster.
	// The channel is closed when the broadcaster closes.
	//
	// Successive calls to CloseC return the same value.
	CloseC() <-chan struct{}

	// Closed reports whether the broadcaster is closed.
	Closed() bool

	// Broadcast sends the specified message to all subscribers.
	//
	// It returns the number of subscribers who have received or buffered
	// this message, and any error encountered.
	//
	// It blocks until all subscribers receive or buffer the message,
	// or an error occurs.
	//
	// It does nothing and reports ErrBroadcasterClosed
	// if the broadcaster is closed.
	Broadcast(msg Message) (n int, err error)

	// BroadcastWithContext is like Broadcast
	// but works in concert with a specified context.
	//
	// In particular, if the specified context is done
	// (canceled, deadline expired, or timeout elapsed),
	// BroadcastWithContext stops early and reports ErrContextDone.
	//
	// BroadcastWithContext panics if the specified context is nil.
	BroadcastWithContext(ctx context.Context, msg Message) (n int, err error)

	// Subscribe subscribes to the broadcaster to acquire a channel
	// for receiving messages from the sender.
	//
	// If Subscribe is called while the broadcaster is broadcasting a message,
	// that message is also sent to the channel returned.
	//
	// bufSize is the buffer size of the returned channel.
	// 0 for no buffer.
	// Negative values for using the default buffer size.
	//
	// Subscribe returns the channel for receiving messages from the sender
	// and any error encountered.
	// It returns a non-nil closed unbuffered channel and
	// reports ErrBroadcasterClosed if the broadcaster is closed.
	//
	// After a successful subscription (i.e., the returned error is nil),
	// the subscriber is responsible for calling Unsubscribe as soon as
	// it no longer needs to receive messages from the broadcaster
	// to avoid blocking the sender.
	Subscribe(bufSize int) (c <-chan Message, err error)

	// Unsubscribe unsubscribes from the broadcaster to
	// unregister the specified channel returned by the method Subscribe
	// and stop receiving messages.
	//
	// c is the channel returned by the method Subscribe.
	// Unsubscribe does nothing and reports ErrUnregisteredChannel
	// if c is unregistered with the broadcaster
	// (i.e., c has already been unregistered by the method Unsubscribe
	// or is not obtained from the method Subscribe).
	//
	// When a receiver no longer needs to receive messages from the broadcaster,
	// it should call Unsubscribe as soon as possible
	// to notify the sender and avoid blocking the sender.
	Unsubscribe(c <-chan Message) error
}

// ErrBroadcasterClosed is an error indicating that the broadcaster is closed.
//
// The client should use errors.Is to test whether
// an error is ErrBroadcasterClosed.
var ErrBroadcasterClosed = errors.AutoNewCustom(
	"broadcaster is closed",
	errors.PrependFullPkgName,
	0,
)

// ErrContextDone is an error indicating that the context is done
// (canceled, deadline expired, or timeout elapsed).
//
// The client should use errors.Is to test whether an error is ErrContextDone.
var ErrContextDone = errors.AutoNewCustom(
	"context is done (canceled, deadline expired, or timeout elapsed)",
	errors.PrependFullPkgName,
	0,
)

// ErrUnregisteredChannel is an error indicating that
// the channel is unregistered with the broadcaster, i.e.,
// the channel has already been unregistered by the method Unsubscribe
// or is not obtained from the method Subscribe.
//
// The client should use errors.Is to test whether
// an error is ErrUnregisteredChannel.
var ErrUnregisteredChannel = errors.AutoNewCustom(
	"the channel is unregistered with the broadcaster",
	errors.PrependFullPkgName,
	0,
)

// NewBroadcaster creates a new Broadcaster.
//
// dfltBufSize is the default buffer size of channels created by
// the new broadcaster and returned by its method Subscribe.
// Nonpositive values for no buffer.
//
// A daemon goroutine is launched after the broadcaster is created.
// The broadcaster should be closed as soon as
// there are no more messages to send,
// so that resources associated with the broadcaster can be released and
// the daemon goroutine can finish.
func NewBroadcaster[Message any](dfltBufSize int) Broadcaster[Message] {
	return newBroadcaster[Message](nil, dfltBufSize)
}

// NewBroadcasterWithContext is like NewBroadcaster,
// but the created broadcaster works in concert with the specified context.
// The broadcaster stops any broadcasts and closes when the specified context
// is done (canceled, deadline expired, or timeout elapsed).
//
// NewBroadcasterWithContext panics if the specified context is nil.
func NewBroadcasterWithContext[Message any](
	ctx context.Context,
	dfltBufSize int,
) Broadcaster[Message] {
	if ctx == nil {
		panic(errors.AutoMsg("the provided context is nil"))
	}

	derivedCtx, cancel := context.WithCancel(ctx)

	return newBroadcaster[Message](
		NewCancelerFromContext(derivedCtx, cancel),
		dfltBufSize,
	)
}

// broadcasterBroadcastRequest is a request used by the broadcaster to
// broadcast a message, sent from the sender (by calling the method Broadcast)
// to the daemon goroutine.
// It consists of a send-only channel to send the result
// (i.e., the number of subscribers who have received or buffered the message),
// the context of the broadcast, and the message to be sent.
type broadcasterBroadcastRequest[Message any] struct {
	// Channel to send the number of subscribers
	// who have received or buffered the message.
	c chan<- int

	// Context of the broadcast (must be non-nil).
	ctx context.Context

	// Message to be sent.
	msg Message
}

// broadcasterSubscribeRequest is a request used by the broadcaster to
// subscribe, sent from a receiver (by calling the method Subscribe)
// to the daemon goroutine.
// It consists of a send-only channel to send the subscribe result channel
// and the buffer size of the subscribe result channel,
// which is the argument of the method Subscribe.
type broadcasterSubscribeRequest[Message any] struct {
	// Channel to send the subscribe result channel.
	c chan<- <-chan Message

	// Buffer size of the subscribe result channel,
	// which is the argument of the method Subscribe.
	bufSize int
}

// broadcasterUnsubscribeRequest is a request used by the broadcaster to
// unsubscribe, sent from a receiver (by calling the method Unsubscribe)
// to the daemon goroutine.
// It consists of a send-only channel to send the unsubscribe result
// (i.e., any error encountered) and the channel to be unregistered,
// which is the argument of the method Unsubscribe.
type broadcasterUnsubscribeRequest[Message any] struct {
	// Channel to send the unsubscribe result (i.e., any error encountered).
	ec chan<- error

	// Channel to be unregistered,
	// which is the argument of the method Unsubscribe.
	c <-chan Message
}

// broadcaster is an implementation of interface Broadcaster.
type broadcaster[Message any] struct {
	// Canceler for the daemon goroutine.
	c Canceler

	// Map from receive-only channels to send-only channels,
	// for sending messages to subscribers.
	cm map[<-chan Message]chan<- Message

	// Close channel, to send close requests
	// from the sender to the daemon goroutine.
	cc chan struct{}

	// Broadcast channel, to send broadcast requests
	// from the sender to the daemon goroutine.
	bc chan *broadcasterBroadcastRequest[Message]

	// Subscribe channel, to send subscribe requests
	// from receivers to the daemon goroutine.
	sc chan *broadcasterSubscribeRequest[Message]

	// Unsubscribe channel, to send unsubscribe requests
	// from receivers to the daemon goroutine.
	uc chan *broadcasterUnsubscribeRequest[Message]

	// Default buffer size of channels returned by the method Subscribe.
	dbs int
}

// newBroadcaster creates a new broadcaster with the specified
// default buffer size and uses the specified canceler for the daemon goroutine.
//
// If the specified canceler is nil, newBroadcaster creates a new canceler by
// github.com/donyori/gogo/concurrency.NewCanceler.
//
// dfltBufSize is the default buffer size of channels created by
// the new broadcaster and returned by its method Subscribe.
// Nonpositive values for no buffer.
//
// A daemon goroutine is launched after the broadcaster is created.
// The broadcaster should be closed as soon as
// there are no more messages to send,
// so that resources associated with the broadcaster can be released and
// the daemon goroutine can finish.
func newBroadcaster[Message any](
	canceler Canceler,
	dfltBufSize int,
) *broadcaster[Message] {
	if canceler == nil {
		canceler = NewCanceler()
	}

	if dfltBufSize < 0 {
		dfltBufSize = 0
	}

	bcast := &broadcaster[Message]{
		c:   canceler,
		cm:  make(map[<-chan Message]chan<- Message),
		cc:  make(chan struct{}),
		bc:  make(chan *broadcasterBroadcastRequest[Message]),
		sc:  make(chan *broadcasterSubscribeRequest[Message]),
		uc:  make(chan *broadcasterUnsubscribeRequest[Message]),
		dbs: dfltBufSize,
	}
	go bcast.daemonProc() // launch the daemon goroutine

	return bcast
}

func (bcast *broadcaster[Message]) Close() {
	closeC := bcast.c.C()
	select {
	case <-closeC:
	case bcast.cc <- struct{}{}:
		<-closeC
	}
}

func (bcast *broadcaster[Message]) CloseC() <-chan struct{} {
	return bcast.c.C()
}

func (bcast *broadcaster[Message]) Closed() bool {
	return bcast.c.Canceled()
}

func (bcast *broadcaster[Message]) Broadcast(msg Message) (n int, err error) {
	n, err = bcast.broadcast(nonContext, msg)
	return n, errors.AutoWrap(err)
}

func (bcast *broadcaster[Message]) BroadcastWithContext(
	ctx context.Context,
	msg Message,
) (n int, err error) {
	n, err = bcast.broadcast(ctx, msg)
	return n, errors.AutoWrap(err)
}

// broadcast is the common broadcast process.
func (bcast *broadcaster[Message]) broadcast(
	ctx context.Context,
	msg Message,
) (n int, err error) {
	if ctx == nil {
		panic(errors.AutoMsg("the provided context is nil"))
	}

	doneC, closeC, nc := ctx.Done(), bcast.c.C(), make(chan int)
	req := &broadcasterBroadcastRequest[Message]{
		c:   nc,
		ctx: ctx,
		msg: msg,
	}

	select {
	case <-doneC:
		err = errors.AutoWrap(ErrContextDone)
	case <-closeC:
		err = errors.AutoWrap(ErrBroadcasterClosed)
	case bcast.bc <- req:
		n = <-nc

		select {
		case <-doneC:
			err = errors.AutoWrap(ErrContextDone)
		case <-closeC:
			err = errors.AutoWrap(ErrBroadcasterClosed)
		default:
		}
	}

	return
}

func (bcast *broadcaster[Message]) Subscribe(
	bufSize int,
) (c <-chan Message, err error) {
	closeC, cc := bcast.c.C(), make(chan (<-chan Message))
	req := &broadcasterSubscribeRequest[Message]{
		c:       cc,
		bufSize: bufSize,
	}

	select {
	case <-closeC:
		err = errors.AutoWrap(ErrBroadcasterClosed)
	case bcast.sc <- req:
		c = <-cc
		switch {
		case c != nil:
			return
		case TestChannel(closeC):
			err = errors.AutoWrap(ErrBroadcasterClosed)
		default:
			err = errors.AutoNew(
				"failed to acquire a channel from the broadcaster")
		}
	}

	closedResultC := make(chan Message)
	close(closedResultC)
	c = closedResultC

	return
}

func (bcast *broadcaster[Message]) Unsubscribe(c <-chan Message) error {
	var err error

	ec := make(chan error)
	req := &broadcasterUnsubscribeRequest[Message]{
		ec: ec,
		c:  c,
	}

	select {
	case <-bcast.c.C():
		err = errors.AutoWrap(ErrUnregisteredChannel)
	case bcast.uc <- req:
		err = errors.AutoWrap(<-ec)
	}

	return err
}

// daemonProc is the main process of the daemon goroutine.
func (bcast *broadcaster[Message]) daemonProc() {
	defer bcast.daemonCleanup()

	for {
		select {
		case <-bcast.c.C():
			return
		case <-bcast.cc:
			return
		case br := <-bcast.bc:
			bcast.daemonBroadcast(br)
		case sr := <-bcast.sc:
			bcast.daemonSubscribe(sr, nil)
		case ur := <-bcast.uc:
			bcast.daemonUnsubscribe(ur, nil)
		}
	}
}

// daemonCleanup is the cleanup process of the daemon goroutine.
// It closes the broadcaster.
func (bcast *broadcaster[Message]) daemonCleanup() {
	bcast.c.Cancel()

	for _, c := range bcast.cm {
		close(c)
	}

	bcast.cm = nil
}

// daemonBroadcast handles the broadcast request.
// It should only be called by the daemon goroutine of the broadcaster.
func (bcast *broadcaster[Message]) daemonBroadcast(
	req *broadcasterBroadcastRequest[Message],
) {
	var n int
	defer func(resultC chan<- int) {
		// The result must be sent, even if daemonBroadcast panics.
		resultC <- n
	}(req.c)

	bcast.daemonBroadcastMain(req, &n)
}

// daemonBroadcastMain is the main body of daemonBroadcast.
// It handles the broadcast request without sending the result n back.
//
// daemonBroadcastMain uses a pointer to n (i.e., nPtr)
// instead of returning n to update n in time,
// so that the caller can get the correct n even if daemonBroadcastMain panics.
func (bcast *broadcaster[Message]) daemonBroadcastMain(
	req *broadcasterBroadcastRequest[Message],
	nPtr *int,
) {
	if len(bcast.cm) == 0 { // there are no subscribers
		return
	}

	doneC, closeC := req.ctx.Done(), bcast.c.C()

	toBeSent := make(map[chan<- Message]struct{}, len(bcast.cm))
	for _, c := range bcast.cm {
		toBeSent[c] = struct{}{}
	}

	for len(toBeSent) > 0 {
		listenToOtherChannels := true
		for listenToOtherChannels {
			select {
			case <-doneC:
				return
			case <-closeC:
				return
			case sr := <-bcast.sc:
				bcast.daemonSubscribe(sr, toBeSent)
			case ur := <-bcast.uc:
				bcast.daemonUnsubscribe(ur, toBeSent)
			default:
				listenToOtherChannels = false
			}
		}

		for c := range toBeSent {
			select {
			case <-doneC:
				return
			case <-closeC:
				return
			case c <- req.msg:
				*nPtr++

				delete(toBeSent, c)
			default:
				// The current channel c is not ready to receive the message.
				// Go to the next channel.
			}
		}
	}
}

// daemonSubscribe handles the subscribe request.
// It should only be called by the daemon goroutine of the broadcaster.
//
// toBeSent is the set of channels through which a message is to be sent.
// It is used to work in concert with a broadcast
// and can be nil without a broadcast.
func (bcast *broadcaster[Message]) daemonSubscribe(
	req *broadcasterSubscribeRequest[Message],
	toBeSent map[chan<- Message]struct{},
) {
	var c chan Message
	defer func(resultC chan<- <-chan Message) {
		// The result must be sent, even if daemonSubscribe panics.
		resultC <- c
	}(req.c)

	if !TestC(bcast.c) {
		bufSize := req.bufSize
		if bufSize < 0 {
			bufSize = bcast.dbs
		}

		c = make(chan Message, bufSize)
		bcast.cm[c] = c

		if toBeSent != nil {
			toBeSent[c] = struct{}{}
		}
	}
}

// daemonUnsubscribe handles the unsubscribe request.
// It should only be called by the daemon goroutine of the broadcaster.
//
// toBeSent is the set of channels through which a message is to be sent.
// It is used to work in concert with a broadcast
// and can be nil without a broadcast.
func (bcast *broadcaster[Message]) daemonUnsubscribe(
	req *broadcasterUnsubscribeRequest[Message],
	toBeSent map[chan<- Message]struct{},
) {
	var err error
	defer func(resultC chan<- error) {
		// The result must be sent, even if daemonUnsubscribe panics.
		resultC <- err
	}(req.ec)

	outC := bcast.cm[req.c]
	if outC != nil {
		delete(toBeSent, outC) // it is safe even if toBeSent is nil
		delete(bcast.cm, req.c)
		close(outC)
	} else {
		err = errors.AutoWrap(ErrUnregisteredChannel)
	}
}

// nonContext is used by the method Broadcast
// to indicate that it is non-context version.
// It is not nil to pass the staticcheck check SA1012.
var nonContext = context.Background()
