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

package concurrency_test

import (
	"slices"
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

	for i := 0; i < NumReader; i++ {
		go func(rank int) {
			defer wg.Done()

			if got := r.Len(); got != 0 {
				t.Errorf("reader %d, initially - got Len %d; want 0",
					rank, got)
			}
			if gotX, gotOK := r.Last(); gotX != 0 || gotOK {
				t.Errorf("reader %d, initially - got Last (%d, %t); want (0, false)",
					rank, gotX, gotOK)
			}
			if got := r.All(); got != nil {
				t.Errorf("reader %d, initially - got All %v; want <nil>",
					rank, got)
			}

			var wantAll []int
			for round, x := range data {
				if len(x) > 0 {
					for k := 0; k < NumRecorder; k++ {
						wantAll = append(wantAll, x...)
					}
				}
				var wantLastX int
				var wantLastOK bool
				if len(wantAll) > 0 {
					wantLastX, wantLastOK = wantAll[len(wantAll)-1], true
				}

				close(readyCsList[rank][round])
				for k := 0; k < NumRecorder; k++ {
					<-recordedCsList[k][round]
				}

				if got := r.Len(); got != len(wantAll) {
					t.Errorf("reader %d, round %d - got Len %d; want %d",
						rank, round, got, len(wantAll))
				}
				if gotX, gotOK := r.Last(); gotX != wantLastX ||
					gotOK != wantLastOK {
					t.Errorf("reader %d, round %d - got Last (%d, %t); want (%d, %t)",
						rank, round, gotX, gotOK, wantLastX, wantLastOK)
				}
				if got := r.All(); (got == nil) != (wantAll == nil) ||
					!slices.Equal(got, wantAll) {
					switch {
					case got == nil:
						// wantAll is non-nil.
						t.Errorf("reader %d, round %d - got All <nil>; want %v",
							rank, round, wantAll)
					case wantAll == nil:
						// got is non-nil.
						t.Errorf("reader %d, round %d - got All %v; want <nil>",
							rank, round, got)
					default:
						// Both got and wantAll are non-nil.
						t.Errorf("reader %d, round %d - got All %v; want %v",
							rank, round, got, wantAll)
					}
				}
			}
		}(i)
	}

	for i := 0; i < NumRecorder; i++ {
		go func(rank int) {
			defer wg.Done()
			for round, x := range data {
				for k := 0; k < NumReader; k++ {
					<-readyCsList[k][round]
				}
				rec.Record(x...)
				close(recordedCsList[rank][round])
			}
		}(i)
	}

	wg.Wait()
}
