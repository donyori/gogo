// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/donyori/gogo/concurrency"
)

func TestRecorder(t *testing.T) {
	rec := concurrency.NewRecorder[int](0)
	if rec == nil {
		t.Fatal("got nil Recorder")
	}
	r := rec.Reader()
	if _, ok := any(r).(concurrency.Recorder[int]); ok {
		t.Error("the RecordReader can be converted to a Recorder")
	}

	data := [][]int{
		nil,
		{},
		{0},
		{1, 2, 3},
		{4, 5, 6},
		{7, 8},
		{9},
		nil,
	}
	const NumReader int = 3
	const NumRecorder int = 4
	// readyCsList: to broadcast signals signifying that the reader is ready
	// (sender: reader, receiver: recorder).
	var readyCsList [NumReader][]chan struct{}
	for i := range readyCsList {
		readyCsList[i] = make([]chan struct{}, len(data))
		for j := range readyCsList[i] {
			readyCsList[i][j] = make(chan struct{})
		}
	}
	// recordedCsList: to broadcast signals signifying that the message has been
	// written to the Recorder (sender: recorder, receiver: reader).
	var recordedCsList [NumRecorder][]chan struct{}
	for i := range recordedCsList {
		recordedCsList[i] = make([]chan struct{}, len(data))
		for j := range recordedCsList[i] {
			recordedCsList[i][j] = make(chan struct{})
		}
	}
	var wg sync.WaitGroup
	wg.Add(NumReader + NumRecorder)

	for i := range NumReader {
		go goroutineTestRecorderReader(
			&wg,
			i,
			t,
			r,
			NumRecorder,
			data,
			readyCsList[:],
			recordedCsList[:],
		)
	}

	for i := range NumRecorder {
		go func(rank int) {
			defer wg.Done()
			for round, x := range data {
				for i := range NumReader {
					<-readyCsList[i][round]
				}
				rec.Record(x...)
				close(recordedCsList[rank][round])
			}
		}(i)
	}

	wg.Wait()
}

// goroutineTestRecorderReader is the process of the goroutine
// launched by TestRecorder to test concurrency.RecordReader.
func goroutineTestRecorderReader(
	wg *sync.WaitGroup,
	rank int,
	t *testing.T,
	r concurrency.RecordReader[int],
	numRecorder int,
	data [][]int,
	readyCsList [][]chan struct{},
	recordedCsList [][]chan struct{},
) {
	defer wg.Done()

	check := func(round int, wantLastX int, wantLastOK bool, wantAll []int) {
		s := "initially"
		if round >= 0 {
			s = "round " + strconv.Itoa(round)
		}
		if got := r.Len(); got != len(wantAll) {
			t.Errorf("reader %d, %s - got Len %d; want %d",
				rank, s, got, len(wantAll))
		}
		if gotX, gotOK := r.Last(); gotX != wantLastX || gotOK != wantLastOK {
			t.Errorf("reader %d, %s - got Last (%d, %t); want (%d, %t)",
				rank, s, gotX, gotOK, wantLastX, wantLastOK)
		}
		if got := r.All(); (got == nil) != (wantAll == nil) ||
			!slices.Equal(got, wantAll) {
			switch {
			case got == nil:
				// wantAll is non-nil.
				t.Errorf("reader %d, %s - got All <nil>; want %v",
					rank, s, wantAll)
			case wantAll == nil:
				// got is non-nil.
				t.Errorf("reader %d, %s - got All %v; want <nil>",
					rank, s, got)
			default:
				// Both got and wantAll are non-nil.
				t.Errorf("reader %d, %s - got All %v; want %v",
					rank, s, got, wantAll)
			}
		}
	}

	check(-1, 0, false, nil)

	var wantAll []int
	for round, x := range data {
		if len(x) > 0 {
			for range numRecorder {
				wantAll = append(wantAll, x...)
			}
		}
		var wantLastX int
		var wantLastOK bool
		if len(wantAll) > 0 {
			wantLastX, wantLastOK = wantAll[len(wantAll)-1], true
		}

		close(readyCsList[rank][round])
		for i := range numRecorder {
			<-recordedCsList[i][round]
		}

		check(round, wantLastX, wantLastOK, wantAll)
	}
}
