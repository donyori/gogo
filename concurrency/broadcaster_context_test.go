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

package concurrency_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency"
)

func TestNewBroadcasterWithContext(t *testing.T) {
	t.Parallel()

	bcast := concurrency.NewBroadcasterWithContext[int](context.Background(), 0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	bcast.Close()
}

func TestBroadcaster_CancelContext_WithCancel(t *testing.T) {
	t.Parallel()

	testBroadcasterCancelContextFunc(
		t,
		func() (ctx context.Context, cancel context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
		func(_ context.Context, cancel context.CancelFunc) {
			time.Sleep(time.Millisecond)
			cancel()
		},
	)
}

func TestBroadcaster_CancelContext_WithTimeout(t *testing.T) {
	t.Parallel()

	testBroadcasterCancelContextFunc(
		t,
		func() (ctx context.Context, cancel context.CancelFunc) {
			return context.WithTimeout(
				context.Background(),
				time.Millisecond*10,
			)
		},
		nil,
	)
}

// testBroadcasterCancelContextFunc is the common framework
// for testing a broadcaster with context by canceling the context.
func testBroadcasterCancelContextFunc(
	t *testing.T,
	getContext func() (ctx context.Context, cancel context.CancelFunc),
	manuallyCancel func(ctx context.Context, cancel context.CancelFunc),
) {
	t.Helper()

	ctx, cancel := getContext()
	if ctx == nil {
		t.Error("got nil context")
		return
	} else if cancel == nil {
		t.Error("got nil cancel function")
		return
	}

	t.Cleanup(cancel)

	bcast := concurrency.NewBroadcasterWithContext[int](ctx, 0)
	if bcast == nil {
		t.Error("got nil broadcaster")
		return
	}

	t.Cleanup(bcast.Close)

	if manuallyCancel != nil {
		manuallyCancel(ctx, cancel)
	} else {
		<-ctx.Done()
	}

	if !bcast.Closed() {
		t.Error("broadcaster was not closed after the context was done")
	}
}

func TestBroadcaster_CancelContextDuringBroadcast_WithCancel(t *testing.T) {
	t.Parallel()

	testBroadcasterCancelContextDuringBroadcastFunc(
		t,
		func() (ctx context.Context, cancel context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
		func(_ context.Context, cancel context.CancelFunc) {
			time.Sleep(time.Millisecond * 300)
			cancel()
		},
	)
}

func TestBroadcaster_CancelContextDuringBroadcast_WithTimeout(t *testing.T) {
	t.Parallel()

	testBroadcasterCancelContextDuringBroadcastFunc(
		t,
		func() (ctx context.Context, cancel context.CancelFunc) {
			return context.WithTimeout(
				context.Background(),
				time.Millisecond*300,
			)
		},
		nil,
	)
}

// nextBroadcastMessage returns the next message to broadcast.
func nextBroadcastMessage(currentMsg int) int {
	const Mask int = 1<<10 - 1
	return (currentMsg + 1) & Mask
}

// testBroadcasterCancelContextDuringBroadcastFunc is the common framework
// for testing a broadcaster with context
// by canceling the context during a broadcast.
func testBroadcasterCancelContextDuringBroadcastFunc(
	t *testing.T,
	getContext func() (ctx context.Context, cancel context.CancelFunc),
	manuallyCancel func(ctx context.Context, cancel context.CancelFunc),
) {
	t.Helper()

	const NumReceiver int = 32

	ctx, cancel := getContext()
	if ctx == nil {
		t.Error("got nil context")
		return
	} else if cancel == nil {
		t.Error("got nil cancel function")
		return
	}

	t.Cleanup(cancel)

	bcast := concurrency.NewBroadcasterWithContext[int](ctx, 0)
	if bcast == nil {
		t.Error("got nil broadcaster")
		return
	}

	t.Cleanup(bcast.Close)

	// The context is used to early-stop receiver goroutines
	// if any receiver goroutine fails or panics.
	recvCtx, recvCancel := context.WithCancel(context.Background())
	t.Cleanup(recvCancel)

	var wg, readyWG sync.WaitGroup
	t.Cleanup(wg.Wait)
	readyWG.Add(NumReceiver)

	for rank := range NumReceiver {
		wg.Go(func() {
			goroutineTestBroadcasterUntilCloseReceiver(
				t,
				recvCtx,
				recvCancel,
				bcast,
				&readyWG,
				rank,
			)
		})
	}

	wg.Go(func() {
		goroutineTestBroadcasterUntilCloseSender(
			t,
			recvCtx,
			NumReceiver,
			bcast,
			&readyWG,
		)
	})

	if manuallyCancel != nil {
		manuallyCancel(ctx, cancel)
	} else {
		<-ctx.Done()
	}

	wg.Wait()

	if !bcast.Closed() {
		t.Error("broadcaster was not closed after the context was done")
	}
}

//revive:disable:context-as-argument
// t (of type *testing.T) takes priority over ctx (of type context.Context)
// for the following test functions.

// goroutineTestBroadcasterUntilCloseSender is the main process of
// a goroutine for a sender that broadcasts messages to receivers
// through the broadcaster until the broadcaster closes.
func goroutineTestBroadcasterUntilCloseSender( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	ctx context.Context,
	numReceiver int,
	bcast concurrency.Broadcaster[int],
	readyWG *sync.WaitGroup,
) {
	defer func() {
		if e := recover(); e != nil {
			t.Error("until close sender goroutine, panic:", e)
		}
	}()

	readyWG.Wait()

	numMsg, msg := 0, -1

	for !bcast.Closed() {
		if concurrency.TestDone(ctx) {
			return
		}

		numMsg++
		msg = nextBroadcastMessage(msg)

		n, err := bcast.Broadcast(msg)
		if errors.Is(err, concurrency.ErrBroadcasterClosed) {
			if n < 0 || n > numReceiver {
				t.Errorf("bcast.Broadcast(%d) (No.%d message): "+
					"got n %d out of range [0, %d] when the broadcaster closed",
					msg, numMsg, n, numReceiver)
			}

			return
		}

		if n != numReceiver {
			t.Errorf("bcast.Broadcast(%d) (No.%d message): got n %d; want %d",
				msg, numMsg, n, numReceiver)
		}

		if err != nil {
			t.Errorf("bcast.Broadcast(%d) (No.%d message): %v",
				msg, numMsg, err)
		}
	}
}

// goroutineTestBroadcasterUntilCloseReceiver is the main process of
// a goroutine for a receiver that subscribes to the broadcaster and
// receives messages from it until the broadcaster closes.
func goroutineTestBroadcasterUntilCloseReceiver( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	ctx context.Context,
	cancel context.CancelFunc,
	bcast concurrency.Broadcaster[int],
	readyWG *sync.WaitGroup,
	rank int,
) {
	defer func() {
		if e := recover(); e != nil {
			t.Errorf("until close receiver goroutine %d, panic: %v", rank, e)
			cancel()
		}
	}()

	readyOnce := sync.OnceFunc(readyWG.Done)
	defer readyOnce() // avoid blocking other goroutines

	c, err := bcast.Subscribe(-1)
	if err != nil {
		t.Errorf("until close receiver goroutine %d, bcast.Subscribe(-1): %v",
			rank, err)
		cancel()

		return
	} else if c == nil {
		t.Errorf("until close receiver goroutine %d, "+
			"bcast.Subscribe(-1): got nil channel",
			rank)
		cancel()

		return
	}
	defer func(c <-chan int) {
		_ = bcast.Unsubscribe(c) // ignore error
	}(c)

	readyOnce()

	n, wantMsg := 0, -1

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c:
			if !ok {
				return
			}

			n++
			wantMsg = nextBroadcastMessage(wantMsg)

			if msg != wantMsg {
				t.Errorf("until close receiver goroutine %d, "+
					"received No.%d message %d; want %d",
					rank, n, msg, wantMsg)
			}
		}
	}
}

//revive:enable:context-as-argument

func TestBroadcaster_BroadcastWithContext_CancelContext_WithCancel(
	t *testing.T,
) {
	t.Parallel()

	testBroadcasterBroadcastWithContextCancelContextFunc(
		t,
		func() (ctx context.Context, cancel context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
		func(_ context.Context, cancel context.CancelFunc) {
			time.Sleep(time.Millisecond * 30)
			cancel()
		},
	)
}

func TestBroadcaster_BroadcastWithContext_CancelContext_WithTimeout(
	t *testing.T,
) {
	t.Parallel()

	testBroadcasterBroadcastWithContextCancelContextFunc(
		t,
		func() (ctx context.Context, cancel context.CancelFunc) {
			return context.WithTimeout(
				context.Background(),
				time.Millisecond*30,
			)
		},
		nil,
	)
}

// testBroadcasterBroadcastWithContextCancelContextFunc is the common framework
// for testing the method BroadcastWithContext by canceling the context.
func testBroadcasterBroadcastWithContextCancelContextFunc(
	t *testing.T,
	getContext func() (ctx context.Context, cancel context.CancelFunc),
	manuallyCancel func(ctx context.Context, cancel context.CancelFunc),
) {
	t.Helper()

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Error("got nil broadcaster")
		return
	}

	t.Cleanup(bcast.Close)

	// Acquire a channel by bcast.Subscribe,
	// but never receive messages from it to block the broadcast.
	c, err := bcast.Subscribe(-1)
	if err != nil {
		t.Error("bcast.Subscribe(-1):", err)
		return
	} else if c == nil {
		t.Error("bcast.Subscribe(-1): got nil channel")
		return
	}

	t.Cleanup(func() {
		_ = bcast.Unsubscribe(c) // ignore error
	})

	ctx, cancel := getContext()
	if ctx == nil {
		t.Error("got nil context")
		return
	} else if cancel == nil {
		t.Error("got nil cancel function")
		return
	}

	t.Cleanup(cancel)

	var wg sync.WaitGroup
	t.Cleanup(wg.Wait)
	wg.Go(func() {
		if manuallyCancel != nil {
			manuallyCancel(ctx, cancel)
		} else {
			<-ctx.Done()
		}
	})

	n, err := bcast.BroadcastWithContext(ctx, 1)

	if n != 0 {
		t.Errorf("bcast.BroadcastWithContext(ctx, 1): got n %d; want 0", n)
	}

	if !errors.Is(err, concurrency.ErrContextDone) {
		t.Errorf("bcast.BroadcastWithContext(ctx, 1): got error %v; "+
			"want concurrency.ErrContextDone (%v)",
			err, concurrency.ErrContextDone)
	}
}
