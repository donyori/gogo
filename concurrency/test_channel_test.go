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
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency"
)

func TestTestChannel(t *testing.T) {
	t.Parallel()

	// For test case #2:
	unclosedUnbufferedC := make(<-chan int)

	// For test case #3:
	unclosedBufferedC := make(chan int, 1)
	unclosedBufferedC <- 1

	// For test case #4:
	closedUnbufferedC := make(chan int)
	close(closedUnbufferedC)

	// For test case #5:
	closedBufferedC := make(chan int, 1)
	closedBufferedC <- 1

	close(closedBufferedC)

	// For test case #6:
	sendingC := make(chan int)
	readyC, doneC := make(chan struct{}), make(chan struct{})
	sendingCSetup := func() {
		go func(readyC, doneC chan<- struct{}, c chan<- int) {
			defer close(doneC)

			close(readyC)

			c <- 1
		}(readyC, doneC, sendingC)

		<-readyC
	}
	sendingCCleanup := func() {
		<-doneC
	}

	testCases := []struct {
		name    string
		c       <-chan int
		setup   func()
		cleanup func()
		want    bool
	}{
		{"nilC", nil, nil, nil, false},
		{"unclosedUnbufferedC", unclosedUnbufferedC, nil, nil, false},
		{"unclosedBufferedC", unclosedBufferedC, nil, nil, true},
		{"closedUnbufferedC", closedUnbufferedC, nil, nil, true},
		{"closedBufferedC", closedBufferedC, nil, nil, true},
		{"sendingC", sendingC, sendingCSetup, sendingCCleanup, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.setup != nil {
				tc.setup()
			}

			if tc.cleanup != nil {
				t.Cleanup(tc.cleanup)
			}

			got := concurrency.TestChannel(tc.c)
			if got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

func TestTestF(t *testing.T) {
	t.Parallel()

	// For test case #2:
	returnNilF := func() <-chan int {
		return nil
	}

	// For test case #3:
	unclosedUnbufferedF := func() <-chan int {
		return make(<-chan int)
	}

	// For test case #4:
	unclosedBufferedF := func() <-chan int {
		c := make(chan int, 1)
		c <- 1

		return c
	}

	// For test case #5:
	closedUnbufferedF := func() <-chan int {
		c := make(chan int)
		close(c)

		return c
	}

	// For test case #6:
	closedBufferedF := func() <-chan int {
		c := make(chan int, 1)
		c <- 1

		close(c)

		return c
	}

	// For test case #7:
	doneC := make(chan struct{})
	sendingF := func() <-chan int {
		c := make(chan int)
		readyC := make(chan struct{})

		go func(readyC, doneC chan<- struct{}, c chan<- int) {
			defer close(doneC)

			close(readyC)

			c <- 1
		}(readyC, doneC, c)

		<-readyC

		return c
	}
	sendingFCleanup := func() {
		<-doneC
	}

	testCases := []struct {
		name    string
		f       func() <-chan int
		cleanup func()
		want    bool
	}{
		{"nilF", nil, nil, false},
		{"returnNilF", returnNilF, nil, false},
		{"unclosedUnbufferedF", unclosedUnbufferedF, nil, false},
		{"unclosedBufferedF", unclosedBufferedF, nil, true},
		{"closedUnbufferedF", closedUnbufferedF, nil, true},
		{"closedBufferedF", closedBufferedF, nil, true},
		{"sendingF", sendingF, sendingFCleanup, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.cleanup != nil {
				t.Cleanup(tc.cleanup)
			}

			got := concurrency.TestF(tc.f)
			if got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

// interfaceCImpl implements interface{ C() <-chan T }
// for testing concurrency.TestC.
type interfaceCImpl[T any] struct {
	f       func() <-chan T
	cleanup func()
}

func (itf *interfaceCImpl[T]) C() <-chan T {
	return itf.f()
}

func TestTestC(t *testing.T) {
	t.Parallel()

	// For test case #2:
	returnNilItf := &interfaceCImpl[int]{
		f: func() <-chan int {
			return nil
		},
	}

	// For test case #3:
	unclosedUnbufferedItf := &interfaceCImpl[int]{
		f: func() <-chan int {
			return make(<-chan int)
		},
	}

	// For test case #4:
	unclosedBufferedItf := &interfaceCImpl[int]{
		f: func() <-chan int {
			c := make(chan int, 1)
			c <- 1

			return c
		},
	}

	// For test case #5:
	closedUnbufferedItf := &interfaceCImpl[int]{
		f: func() <-chan int {
			c := make(chan int)
			close(c)

			return c
		},
	}

	// For test case #6:
	closedBufferedItf := &interfaceCImpl[int]{
		f: func() <-chan int {
			c := make(chan int, 1)
			c <- 1

			close(c)

			return c
		},
	}

	// For test case #7:
	doneC := make(chan struct{})
	sendingItf := &interfaceCImpl[int]{
		f: func() <-chan int {
			c := make(chan int)
			readyC := make(chan struct{})

			go func(readyC, doneC chan<- struct{}, c chan<- int) {
				defer close(doneC)

				close(readyC)

				c <- 1
			}(readyC, doneC, c)

			<-readyC

			return c
		},
		cleanup: func() {
			<-doneC
		},
	}

	testCases := []struct {
		name string
		itf  *interfaceCImpl[int]
		want bool
	}{
		{"nilItf", nil, false},
		{"returnNilItf", returnNilItf, false},
		{"unclosedUnbufferedItf", unclosedUnbufferedItf, false},
		{"unclosedBufferedItf", unclosedBufferedItf, true},
		{"closedUnbufferedItf", closedUnbufferedItf, true},
		{"closedBufferedItf", closedBufferedItf, true},
		{"sendingItf", sendingItf, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.itf != nil && tc.itf.cleanup != nil {
				t.Cleanup(tc.itf.cleanup)
			}

			var itf interface{ C() <-chan int }
			if tc.itf != nil {
				itf = tc.itf
			}

			got := concurrency.TestC(itf)
			if got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

// interfaceDoneImpl implements interface{ Done() <-chan T }
// for testing concurrency.TestDone.
type interfaceDoneImpl[T any] struct {
	f       func() <-chan T
	cleanup func()
}

func (itf *interfaceDoneImpl[T]) Done() <-chan T {
	return itf.f()
}

func TestTestDone(t *testing.T) {
	t.Parallel()

	// For test case #2:
	returnNilItf := &interfaceDoneImpl[int]{
		f: func() <-chan int {
			return nil
		},
	}

	// For test case #3:
	unclosedUnbufferedItf := &interfaceDoneImpl[int]{
		f: func() <-chan int {
			return make(<-chan int)
		},
	}

	// For test case #4:
	unclosedBufferedItf := &interfaceDoneImpl[int]{
		f: func() <-chan int {
			c := make(chan int, 1)
			c <- 1

			return c
		},
	}

	// For test case #5:
	closedUnbufferedItf := &interfaceDoneImpl[int]{
		f: func() <-chan int {
			c := make(chan int)
			close(c)

			return c
		},
	}

	// For test case #6:
	closedBufferedItf := &interfaceDoneImpl[int]{
		f: func() <-chan int {
			c := make(chan int, 1)
			c <- 1

			close(c)

			return c
		},
	}

	// For test case #7:
	doneC := make(chan struct{})
	sendingItf := &interfaceDoneImpl[int]{
		f: func() <-chan int {
			c := make(chan int)
			readyC := make(chan struct{})

			go func(readyC, doneC chan<- struct{}, c chan<- int) {
				defer close(doneC)

				close(readyC)

				c <- 1
			}(readyC, doneC, c)

			<-readyC

			return c
		},
		cleanup: func() {
			<-doneC
		},
	}

	testCases := []struct {
		name string
		itf  *interfaceDoneImpl[int]
		want bool
	}{
		{"nilItf", nil, false},
		{"returnNilItf", returnNilItf, false},
		{"unclosedUnbufferedItf", unclosedUnbufferedItf, false},
		{"unclosedBufferedItf", unclosedBufferedItf, true},
		{"closedUnbufferedItf", closedUnbufferedItf, true},
		{"closedBufferedItf", closedBufferedItf, true},
		{"sendingItf", sendingItf, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.itf != nil && tc.itf.cleanup != nil {
				t.Cleanup(tc.itf.cleanup)
			}

			var itf interface{ Done() <-chan int }
			if tc.itf != nil {
				itf = tc.itf
			}

			got := concurrency.TestDone(itf)
			if got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

func TestTestDone_Context(t *testing.T) {
	t.Parallel()

	// For test case #2:
	backgroundContext := context.Background()

	// For test case #3:
	uncanceledContext, uncanceledCancel := context.WithCancel(
		context.Background(),
	)
	t.Cleanup(uncanceledCancel)

	// For test case #4:
	canceledContext, canceledCancel := context.WithCancel(context.Background())
	canceledCancel()

	// For test case #5:
	notTimeoutContext, notTimeoutCancel := context.WithTimeout(
		context.Background(),
		time.Minute,
	)
	t.Cleanup(notTimeoutCancel)

	// For test case #6:
	timeoutContext, timeoutCancel := context.WithTimeout(
		context.Background(),
		time.Microsecond,
	)
	t.Cleanup(timeoutCancel)

	testCases := []struct {
		name        string
		ctx         context.Context
		cancel      context.CancelFunc
		waitForDone bool
		want        bool
	}{
		{
			"nilContext",
			nil,
			nil,
			false,
			false,
		},
		{
			"backgroundContext",
			backgroundContext,
			nil,
			false,
			false,
		},
		{
			"uncanceledContext",
			uncanceledContext,
			uncanceledCancel,
			false,
			false,
		},
		{
			"canceledContext",
			canceledContext,
			canceledCancel,
			false,
			true,
		},
		{
			"notTimeoutContext",
			notTimeoutContext,
			notTimeoutCancel,
			false,
			false,
		},
		{
			"timeoutContext",
			timeoutContext,
			timeoutCancel,
			true,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.cancel != nil {
				t.Cleanup(tc.cancel)
			}

			if tc.waitForDone && tc.ctx != nil {
				<-tc.ctx.Done()
			}

			got := concurrency.TestDone(tc.ctx)
			if got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}
