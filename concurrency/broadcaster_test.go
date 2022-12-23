// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency"
)

func TestBroadcaster_Broadcast(t *testing.T) {
	const NumMessage = 10
	const NumGoroutine = 32
	var data [NumMessage]int
	for i := range data {
		data[i] = i + 1
	}

	bcast := concurrency.NewBroadcaster[int](0)

	// Barriers to synchronize goroutines.
	// Calling barriers[i].Done() means this goroutine is ready to delay
	// a random duration and then receive the i-th message.
	var barriers [NumMessage + 1]sync.WaitGroup
	for i := 0; i <= NumMessage; i++ {
		barriers[i].Add(NumGoroutine)
	}
	var barriersWaitOnce [NumMessage]sync.Once

	random := rand.New(rand.NewSource(10)) // use a fixed seed for debugging
	var delayDurations [NumGoroutine][NumMessage - 1]time.Duration
	for i := 0; i < NumGoroutine; i++ {
		for k := 0; k < NumMessage-1; k++ {
			delayDurations[i][k] = time.Microsecond * time.Duration(random.Intn(NumGoroutine))
		}
	}

	var wg sync.WaitGroup
	wg.Add(NumGoroutine)
	for i := 0; i < NumGoroutine; i++ {
		go func(rank int) {
			defer wg.Done()
			var recv [NumMessage]int
			c, idx := bcast.Subscribe(-1), 0
			barriers[idx].Done()
			for msg := range c {
				if idx >= NumMessage {
					panic(fmt.Sprintf("goroutine %d, got messages more than %d", rank, NumMessage))
				}
				recv[idx] = msg
				idx++
				barriers[idx].Done()

				barriersWaitOnce[idx-1].Do(func() {
					barriers[idx].Wait()
				})

				if idx < NumMessage-1 {
					time.Sleep(delayDurations[rank][idx-1])
				}
			}
			if recv != data {
				t.Errorf("goroutine %d, recv %v; want %v", rank, recv, data)
			}
		}(i)
	}
	barriers[0].Wait() // Ensure that all subscribers have entered the for loop.
	for _, msg := range data {
		bcast.Broadcast(msg)
	}
	bcast.Close()
	wg.Wait()

	defer func() {
		if e := recover(); e != nil {
			msg, ok := e.(string)
			if !ok || !strings.HasSuffix(msg, "broadcaster is closed") {
				t.Error(e)
			}
		}
	}()
	bcast.Broadcast(0) // want panic here
	t.Error("no panic when calling Broadcast after the broadcaster closed")
}

func TestBroadcaster_Unsubscribe(t *testing.T) {
	const NumMessage = 40
	var data [NumMessage]int
	for i := range data {
		data[i] = i
	}
	bcast := concurrency.NewBroadcaster[int](0)
	c := bcast.Subscribe(NumMessage / 4)
	var wg sync.WaitGroup
	wg.Add(4) // 3 receivers + 1 sender
	var ready sync.WaitGroup
	ready.Add(3)
	for i := 0; i < 3; i++ {
		// Receive all messages normally, to ensure that Unsubscribe doesn't
		// affect other subscribers.
		go func(rank int) {
			defer wg.Done()
			var recv [NumMessage]int
			c, idx := bcast.Subscribe(-1), 0
			ready.Done()
			for msg := range c {
				if idx >= NumMessage {
					panic(fmt.Sprintf("goroutine %d, got messages more than %d", rank, NumMessage))
				}
				recv[idx] = msg
				idx++
				time.Sleep(time.Microsecond * time.Duration(rank)) // let the next reception not start at the same time
			}
			if recv != data {
				t.Errorf("goroutine %d, recv %v; want %v", rank, recv, data)
			}
		}(i)
	}
	go func() {
		defer wg.Done()
		defer bcast.Close()
		ready.Wait()
		for _, msg := range data {
			bcast.Broadcast(msg)
		}
	}()
	defer wg.Wait()
	var recv [NumMessage]int
	stop := NumMessage / 2
	for i := 0; i < stop; i++ {
		msg, ok := <-c
		if !ok {
			t.Errorf("c closed early, received messages %v", recv[:i])
			return
		}
		recv[i] = msg
	}
	unread := bcast.Unsubscribe(c)
	if len(unread) > cap(c)+1 || len(unread) > NumMessage-stop {
		t.Errorf("too many unread messages, len(unread) = %d, cap(c) = %d, NumMessage - stop = %d",
			len(unread), cap(c), NumMessage-stop)
	}
	for i := range unread {
		recv[i+stop] = unread[i]
	}
	for i, n := 0, stop+len(unread); i < n; i++ {
		if recv[i] != data[i] {
			t.Errorf("recv[:%d] = %v; want (data[:%[1]d]) %[3]v", n, recv[:n], data[:n])
			break
		}
	}
}

func TestBroadcaster_Unsubscribe_IllegalC(t *testing.T) {
	bcast := concurrency.NewBroadcaster[int](0)
	bcast.Subscribe(-1)
	defer func() {
		if e := recover(); e != nil {
			msg, ok := e.(string)
			if !ok || !strings.HasSuffix(msg, "c is not gotten from this broadcaster or has already unsubscribed") {
				t.Error(e)
			}
		}
	}()
	bcast.Unsubscribe(make(chan int)) // want panic here
	t.Error("no panic when calling Unsubscribe with a channel not assigned by the broadcaster")
}

func TestBroadcaster_Unsubscribe_AfterClose(t *testing.T) {
	bcast := concurrency.NewBroadcaster[int](0)
	c := bcast.Subscribe(-1)
	bcast.Close()
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic when calling Unsubscribe after the broadcaster closed,", e)
		}
	}()
	bcast.Unsubscribe(c)
	bcast.Unsubscribe(make(chan int))
}
