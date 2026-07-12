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
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency"
	"github.com/donyori/gogo/internal/unequal"
)

func TestNewBroadcaster(t *testing.T) {
	t.Parallel()

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	bcast.Close()
}

func TestBroadcaster_CloseMethods(t *testing.T) {
	t.Parallel()

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)

	c := bcast.CloseC()
	if c == nil {
		t.Fatal("got nil close channel")
	}

	if concurrency.TestChannel(c) {
		t.Error("<-bcast.CloseC() is selected before closing the broadcaster")
	}

	if bcast.Closed() {
		t.Error("bcast.Closed() returned true before closing the broadcaster")
	}

	bcast.Close()

	c2 := bcast.CloseC()
	if c2 != c {
		t.Error("got different close channels " +
			"before and after closing the broadcaster")
	}

	if !concurrency.TestChannel(c) {
		t.Error("<-bcast.CloseC() is not selected " +
			"after closing the broadcaster")
	}

	if !bcast.Closed() {
		t.Error("bcast.Closed() returned false after closing the broadcaster")
	}
}

func TestBroadcaster_Broadcast_NoSubscriber(t *testing.T) {
	t.Parallel()

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)

	n, err := bcast.Broadcast(1)

	if n != 0 {
		t.Errorf("got n %d; want 0", n)
	}

	if err != nil {
		t.Error(err)
	}
}

func TestBroadcaster_Broadcast_AfterClose(t *testing.T) {
	t.Parallel()

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)

	c, err := bcast.Subscribe(-1)
	if err != nil {
		t.Fatal("bcast.Subscribe(-1):", err)
	} else if c == nil {
		t.Fatal("bcast.Subscribe(-1): got nil channel")
	}

	t.Cleanup(func() {
		_ = bcast.Unsubscribe(c) // ignore error
	})

	bcast.Close()
	n, err := bcast.Broadcast(1)

	if n != 0 {
		t.Errorf("bcast.Broadcast(1): got n %d; want 0", n)
	}

	if !errors.Is(err, concurrency.ErrBroadcasterClosed) {
		t.Errorf("bcast.Broadcast(1): got error %v; "+
			"want concurrency.ErrBroadcasterClosed (%v)",
			err, concurrency.ErrBroadcasterClosed)
	}
}

func TestBroadcaster_Subscribe(t *testing.T) {
	t.Parallel()

	const DefaultBufferSize int = 1

	bcast := concurrency.NewBroadcaster[int](DefaultBufferSize)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)

	var cs []<-chan int

	t.Cleanup(func() {
		for _, c := range cs {
			_ = bcast.Unsubscribe(c) // ignore error
		}
	})

	for bufSize := -2; bufSize <= 2; bufSize++ {
		c, err := bcast.Subscribe(bufSize)
		if c != nil {
			cs = append(cs, c)
		}

		testBroadcasterSubscribeCheckResult(
			t,
			DefaultBufferSize,
			bufSize,
			c,
			err,
		)
	}
}

func testBroadcasterSubscribeCheckResult(
	t *testing.T,
	dfltBufSize int,
	bufSize int,
	c <-chan int,
	err error,
) {
	t.Helper()

	if c == nil {
		t.Errorf("bcast.Subscribe(%d): got nil channel", bufSize)
	} else {
		wantBufSize := bufSize
		if wantBufSize < 0 {
			wantBufSize = max(0, dfltBufSize)
		}

		capacity := cap(c)
		if capacity != wantBufSize {
			t.Errorf("bcast.Subscribe(%d): "+
				"got channel buffer capacity %d; want %d",
				bufSize, capacity, wantBufSize)
		}
	}

	if err != nil {
		t.Errorf("bcast.Subscribe(%d): %v", bufSize, err)
	}
}

func TestBroadcaster_Subscribe_AfterClose(t *testing.T) {
	t.Parallel()

	bcast := concurrency.NewBroadcaster[int](1)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)
	bcast.Close()

	var cs []<-chan int

	t.Cleanup(func() {
		for _, c := range cs {
			_ = bcast.Unsubscribe(c) // ignore error
		}
	})

	for bufSize := -2; bufSize <= 2; bufSize++ {
		c, err := bcast.Subscribe(bufSize)
		if c != nil {
			cs = append(cs, c)
		}

		testBroadcasterSubscribeAfterCloseCheckResult(t, bufSize, c, err)
	}
}

func testBroadcasterSubscribeAfterCloseCheckResult(
	t *testing.T,
	bufSize int,
	c <-chan int,
	err error,
) {
	t.Helper()

	if c == nil {
		t.Errorf("bcast.Subscribe(%d): got nil channel", bufSize)
	} else {
		select {
		case msg, ok := <-c:
			if ok {
				t.Errorf("bcast.Subscribe(%d): "+
					"got unclosed channel and received message %d",
					bufSize, msg)
			}
		default:
			t.Errorf("bcast.Subscribe(%d): got unclosed channel", bufSize)
		}

		if capacity := cap(c); capacity != 0 {
			t.Errorf("bcast.Subscribe(%d): "+
				"got channel buffer capacity %d; want 0",
				bufSize, capacity)
		}
	}

	if !errors.Is(err, concurrency.ErrBroadcasterClosed) {
		t.Errorf("bcast.Subscribe(%d): got error %v; "+
			"want concurrency.ErrBroadcasterClosed (%v)",
			bufSize, err, concurrency.ErrBroadcasterClosed)
	}
}

func TestBroadcaster_Unsubscribe(t *testing.T) {
	t.Parallel()

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)

	c, err := bcast.Subscribe(-1)
	if err != nil {
		t.Fatal("bcast.Subscribe(-1):", err)
	} else if c == nil {
		t.Fatal("bcast.Subscribe(-1): got nil channel")
	}

	err = bcast.Unsubscribe(c)
	if err != nil {
		t.Error("1st bcast.Unsubscribe(c):", err)
	}
}

func TestBroadcaster_Unsubscribe_Twice(t *testing.T) {
	t.Parallel()

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)

	c, err := bcast.Subscribe(-1)
	if err != nil {
		t.Fatal("bcast.Subscribe(-1):", err)
	} else if c == nil {
		t.Fatal("bcast.Subscribe(-1): got nil channel")
	}

	err = bcast.Unsubscribe(c)
	if err != nil {
		t.Error("1st bcast.Unsubscribe(c):", err)
	}

	err = bcast.Unsubscribe(c)
	if !errors.Is(err, concurrency.ErrUnregisteredChannel) {
		t.Errorf("2nd bcast.Unsubscribe(c): got error %v; "+
			"want concurrency.ErrUnregisteredChannel (%v)",
			err, concurrency.ErrUnregisteredChannel)
	}
}

func TestBroadcaster_Unsubscribe_UnregisteredC(t *testing.T) {
	t.Parallel()

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)

	err := bcast.Unsubscribe(make(chan int))
	if !errors.Is(err, concurrency.ErrUnregisteredChannel) {
		t.Errorf("bcast.Unsubscribe(make(chan int)): got error %v; "+
			"want concurrency.ErrUnregisteredChannel (%v)",
			err, concurrency.ErrUnregisteredChannel)
	}

	err = bcast.Unsubscribe(nil)
	if !errors.Is(err, concurrency.ErrUnregisteredChannel) {
		t.Errorf("bcast.Unsubscribe(nil): got error %v; "+
			"want concurrency.ErrUnregisteredChannel (%v)",
			err, concurrency.ErrUnregisteredChannel)
	}
}

func TestBroadcaster_Unsubscribe_AfterClose(t *testing.T) {
	t.Parallel()

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)

	c, err := bcast.Subscribe(-1)
	if err != nil {
		t.Fatal("bcast.Subscribe(-1):", err)
	} else if c == nil {
		t.Fatal("bcast.Subscribe(-1): got nil channel")
	}

	bcast.Close()

	err = bcast.Unsubscribe(c)
	if !errors.Is(err, concurrency.ErrUnregisteredChannel) {
		t.Errorf("bcast.Unsubscribe(c): got error %v; "+
			"want concurrency.ErrUnregisteredChannel (%v)",
			err, concurrency.ErrUnregisteredChannel)
	}

	err = bcast.Unsubscribe(make(chan int))
	if !errors.Is(err, concurrency.ErrUnregisteredChannel) {
		t.Errorf("bcast.Unsubscribe(make(chan int)): got error %v; "+
			"want concurrency.ErrUnregisteredChannel (%v)",
			err, concurrency.ErrUnregisteredChannel)
	}

	err = bcast.Unsubscribe(nil)
	if !errors.Is(err, concurrency.ErrUnregisteredChannel) {
		t.Errorf("bcast.Unsubscribe(nil): got error %v; "+
			"want concurrency.ErrUnregisteredChannel (%v)",
			err, concurrency.ErrUnregisteredChannel)
	}
}

func TestBroadcaster_BroadcastAndReceiveAll(t *testing.T) {
	t.Parallel()

	const (
		NumMessage  int = 10
		NumReceiver int = 32
	)

	data := make([]int, NumMessage)
	for i := range data {
		data[i] = i + 1
	}

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)

	// This barrier is for synchronizing goroutines.
	// Calling barrier.Done(i) means this goroutine is ready to delay
	// a random duration and then receive the i-th message.
	barrier := NewBarrier(NumMessage, NumReceiver)

	// delayDurations is for making goroutines receive messages asynchronously,
	// i.e., the goroutines receive the same message from the broadcaster
	// at different times.
	delayDurations := getRandomDurations2D(
		NumReceiver,
		NumMessage,
		[32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456")),
		time.Microsecond*10*time.Duration(NumReceiver),
	)

	// The context is used to early-stop receiver goroutines
	// if any receiver goroutine fails or panics.
	recvCtx, recvCancel := context.WithCancel(context.Background())
	t.Cleanup(recvCancel)

	var wg sync.WaitGroup
	t.Cleanup(wg.Wait)

	for rank := range NumReceiver {
		wg.Go(func() {
			goroutineTestBroadcasterNormalReceiver(
				t,
				recvCtx,
				recvCancel,
				data,
				bcast,
				barrier,
				delayDurations,
				rank,
			)
		})
	}

	// Ensure that all receivers have subscribed to the broadcaster.
	barrier.Wait(0)

	for _, msg := range data {
		if concurrency.TestDone(recvCtx) {
			break
		}

		n, err := bcast.Broadcast(msg)

		if n != NumReceiver {
			t.Errorf("bcast.Broadcast(%d): got n %d; want %d",
				msg, n, NumReceiver)
		}

		if err != nil {
			t.Errorf("bcast.Broadcast(%d): %v", msg, err)
		}
	}

	bcast.Close()
}

func TestBroadcaster_BroadcastAndReceivePart(t *testing.T) {
	t.Parallel()

	const (
		NumMessage int = 10

		NumNormalReceiver      int = 16
		NumUnsubscribeReceiver int = NumMessage + 1

		NumReceiver int = NumNormalReceiver + NumUnsubscribeReceiver
	)

	data := make([]int, NumMessage)
	for i := range data {
		data[i] = i + 1
	}

	bcast := concurrency.NewBroadcaster[int](0)
	if bcast == nil {
		t.Fatal("got nil broadcaster")
	}

	t.Cleanup(bcast.Close)

	// This barrier is for synchronizing goroutines.
	// Calling barrier.Done(i) means this goroutine is ready to delay
	// a random duration and then receive the i-th message.
	barrier := NewBarrier(NumMessage, NumReceiver)

	// delayDurations is for making goroutines receive messages asynchronously,
	// i.e., the goroutines receive the same message from the broadcaster
	// at different times.
	delayDurations := getRandomDurations2D(
		NumReceiver,
		NumMessage,
		[32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456")),
		time.Microsecond*10*time.Duration(NumReceiver),
	)

	// The context is used to early-stop receiver goroutines
	// if any receiver goroutine fails or panics.
	recvCtx, recvCancel := context.WithCancel(context.Background())
	t.Cleanup(recvCancel)

	var normalWG, unsubscribeWG sync.WaitGroup
	t.Cleanup(normalWG.Wait)
	t.Cleanup(unsubscribeWG.Wait)

	for rank := range NumNormalReceiver {
		normalWG.Go(func() {
			goroutineTestBroadcasterNormalReceiver(
				t,
				recvCtx,
				recvCancel,
				data,
				bcast,
				barrier,
				delayDurations,
				rank,
			)
		})
	}

	for i := range NumUnsubscribeReceiver {
		unsubscribeWG.Go(func() {
			// Rank starts with NumNormalReceiver
			// to fit barrier and delayDurations.
			goroutineTestBroadcasterUnsubscribeReceiver(
				t,
				recvCtx,
				recvCancel,
				len(data),
				data[:i],
				bcast,
				barrier,
				delayDurations,
				NumNormalReceiver+i,
			)
		})
	}

	// Ensure that all receivers have subscribed to the broadcaster.
	barrier.Wait(0)

	for i, msg := range data {
		if concurrency.TestDone(recvCtx) {
			break
		}

		n, err := bcast.Broadcast(msg)

		if n != NumReceiver-1-i {
			t.Errorf("bcast.Broadcast(%d): got n %d; want %d",
				msg, n, NumReceiver-1-i)
		}

		if err != nil {
			t.Errorf("bcast.Broadcast(%d): %v", msg, err)
		}
	}

	// Wait for unsubscribe receivers before closing the broadcaster to avoid
	// unexpected errors when receivers unsubscribe from the broadcaster.
	unsubscribeWG.Wait()
	bcast.Close()
}

// Barrier is a simple device to synchronize goroutines multiple times.
// It consists of multiple wait groups, called layers.
type Barrier struct {
	wgs []*sync.WaitGroup
}

// NewBarrier creates a new Barrier with the specified number of layers
// for the specified number of goroutines.
func NewBarrier(numLayer, numGoroutine int) *Barrier {
	wgs := make([]*sync.WaitGroup, numLayer)
	for i := range wgs {
		wgs[i] = new(sync.WaitGroup)
		wgs[i].Add(numGoroutine)
	}

	return &Barrier{wgs: wgs}
}

// Done decreases the task counter of the specified layer by one.
func (b *Barrier) Done(layer int) {
	b.wgs[layer].Done()
}

// Wait blocks until the task counter of the specified layer is zero.
func (b *Barrier) Wait(layer int) {
	b.wgs[layer].Wait()
}

// getRandomDurations2D returns random durations as a 2-D array
// (similar to [m][n]time.Duration).
//
// The random values are generated by ChaCha8 seeded with the specified seed,
// from 0 (inclusive) to the specified max duration (exclusive).
// The specified max duration must be nonnegative.
func getRandomDurations2D(
	m int,
	n int,
	chaCha8Seed [32]byte,
	maxDuration time.Duration,
) [][]time.Duration {
	random := rand.New(rand.NewChaCha8(chaCha8Seed)) //gosec:disable G404 -- math/rand/v2 is reproducible

	durations := make([][]time.Duration, m)
	for i := range durations {
		ds := make([]time.Duration, n)
		for j := range ds {
			ds[j] = time.Duration(random.Int64N(int64(maxDuration)))
		}

		durations[i] = ds
	}

	return durations
}

//revive:disable:context-as-argument
// t (of type *testing.T) takes priority over ctx (of type context.Context)
// for the following test functions.

// goroutineTestBroadcasterNormalReceiver is the main process of
// a goroutine for a normal receiver that subscribes to the broadcaster
// and receives messages from it.
// Each message is received after a specified delay.
func goroutineTestBroadcasterNormalReceiver( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	ctx context.Context,
	cancel context.CancelFunc,
	wantMsgs []int,
	bcast concurrency.Broadcaster[int],
	barrier *Barrier,
	delayDurations [][]time.Duration,
	rank int,
) {
	defer func() {
		if e := recover(); e != nil {
			t.Errorf("normal receiver goroutine %d, panic: %v", rank, e)
			cancel()
		}
	}()

	lastLayer := -1
	defer func(barrier *Barrier) {
		for lastLayer+1 < len(wantMsgs) {
			lastLayer++
			barrier.Done(lastLayer) // avoid blocking other goroutines
		}
	}(barrier)

	c, err := bcast.Subscribe(-1)
	if err != nil {
		t.Errorf("normal receiver goroutine %d, bcast.Subscribe(-1): %v",
			rank, err)
		cancel()

		return
	} else if c == nil {
		t.Errorf("normal receiver goroutine %d, "+
			"bcast.Subscribe(-1): got nil channel",
			rank)
		cancel()

		return
	}
	defer func(c <-chan int) {
		_ = bcast.Unsubscribe(c) // ignore error
	}(c)

	n := testBroadcasterReceiverMainLoopAndCheckMessages(
		t,
		ctx,
		cancel,
		wantMsgs,
		barrier,
		delayDurations,
		&lastLayer,
		c,
		rank,
		"normal receiver",
	)

	doneC := ctx.Done()
	if concurrency.TestChannel(doneC) {
		return
	}

	select {
	case <-doneC:
	case msg, ok := <-c:
		if ok {
			t.Errorf("normal receiver goroutine %d, "+
				"received No.%d message %d, which is more than %d",
				rank, n+1, msg, len(wantMsgs))
			cancel()
		}
	}
}

// goroutineTestBroadcasterUnsubscribeReceiver is the main process of
// a goroutine for an unsubscribe receiver that subscribes to the broadcaster,
// receives (part of) messages from it, and then unsubscribes.
// Each message is received after a specified delay.
func goroutineTestBroadcasterUnsubscribeReceiver( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	ctx context.Context,
	cancel context.CancelFunc,
	numAllMessage int,
	wantMsgs []int,
	bcast concurrency.Broadcaster[int],
	barrier *Barrier,
	delayDurations [][]time.Duration,
	rank int,
) {
	defer func() {
		if e := recover(); e != nil {
			t.Errorf("unsubscribe receiver goroutine %d, panic: %v", rank, e)
			cancel()
		}
	}()

	lastLayer := -1
	decreaseRemainingBarrierLayers := func(barrier *Barrier) {
		for lastLayer+1 < numAllMessage {
			lastLayer++
			barrier.Done(lastLayer) // avoid blocking other goroutines
		}
	}

	defer decreaseRemainingBarrierLayers(barrier)

	c, err := bcast.Subscribe(-1)
	if err != nil {
		t.Errorf("unsubscribe receiver goroutine %d, bcast.Subscribe(-1): %v",
			rank, err)
		cancel()

		return
	} else if c == nil {
		t.Errorf("unsubscribe receiver goroutine %d, "+
			"bcast.Subscribe(-1): got nil channel",
			rank)
		cancel()

		return
	}
	defer func(c <-chan int) {
		_ = bcast.Unsubscribe(c) // ignore error
	}(c)

	n := testBroadcasterReceiverMainLoopAndCheckMessages(
		t,
		ctx,
		cancel,
		wantMsgs,
		barrier,
		delayDurations,
		&lastLayer,
		c,
		rank,
		"unsubscribe receiver",
	)

	doneC := ctx.Done()
	if concurrency.TestChannel(doneC) {
		return
	}

	err = bcast.Unsubscribe(c)
	if err != nil {
		t.Errorf("unsubscribe receiver goroutine %d, bcast.Unsubscribe(c): %v",
			rank, err)
		cancel()

		return
	}

	decreaseRemainingBarrierLayers(barrier)

	select {
	case <-doneC:
	case msg, ok := <-c:
		if ok {
			t.Errorf("unsubscribe receiver goroutine %d, "+
				"received No.%d message %d, which is more than %d",
				rank, n+1, msg, len(wantMsgs))
			cancel()
		}
	}
}

// testBroadcasterReceiverMainLoopAndCheckMessages is the common subprocess of
// goroutineTestBroadcasterNormalReceiver and
// goroutineTestBroadcasterUnsubscribeReceiver.
// It receives at most len(wantMsgs) messages from c and
// compares them with the wanted messages.
//
// receiverLogPrefix should be "normal receiver"
// for goroutineTestBroadcasterNormalReceiver and
// "unsubscribe receiver" for goroutineTestBroadcasterUnsubscribeReceiver.
//
// testBroadcasterReceiverMainLoopAndCheckMessages returns
// the number of messages received.
func testBroadcasterReceiverMainLoopAndCheckMessages(
	t *testing.T,
	ctx context.Context,
	cancel context.CancelFunc,
	wantMsgs []int,
	barrier *Barrier,
	delayDurations [][]time.Duration,
	lastLayerPtr *int,
	c <-chan int,
	rank int,
	receiverLogPrefix string,
) int {
	t.Helper()

	if wantMsgs == nil {
		wantMsgs = []int{}
	}

	recvMsgs := make([]int, 0, len(wantMsgs))
	for len(recvMsgs) < len(wantMsgs) {
		*lastLayerPtr++
		barrier.Done(*lastLayerPtr)
		barrier.Wait(*lastLayerPtr)
		time.Sleep(delayDurations[rank][len(recvMsgs)])

		select {
		case <-ctx.Done():
			return len(recvMsgs)
		case msg, ok := <-c:
			if !ok {
				t.Errorf("%s goroutine %d, channel closed "+
					"after receiving %d message(s); want %d messages",
					receiverLogPrefix, rank, len(recvMsgs), len(wantMsgs))
				cancel()

				return len(recvMsgs)
			}

			recvMsgs = append(recvMsgs, msg)
		}
	}

	if unequal.Slice(recvMsgs, wantMsgs) {
		t.Errorf("%s goroutine %d, received %d; want %d",
			receiverLogPrefix, rank, recvMsgs, wantMsgs)
		cancel()
	}

	return len(recvMsgs)
}

//revive:enable:context-as-argument
