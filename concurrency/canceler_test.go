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

package concurrency_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/donyori/gogo/concurrency"
)

func TestOnceCanceler(t *testing.T) {
	const N int = 10
	canceler := concurrency.NewCanceler()
	testCancelerCAndCanceled(t, "before calling Cancel, ", canceler, false)
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			canceler.Cancel()
			testCancelerCAndCanceled(
				t,
				fmt.Sprintf("goroutine %d, after calling Cancel, ", rank),
				canceler,
				true,
			)
		}(i)
	}
	wg.Wait()
}

func TestContextCanceler_CancelByCanceler(t *testing.T) {
	testContextCancelerFunc(t, 0)
}

func TestContextCanceler_CancelByCancelFunc(t *testing.T) {
	testContextCancelerFunc(t, 1)
}

func TestContextCanceler_CancelByBothCancelerAndCancelFunc(t *testing.T) {
	testContextCancelerFunc(t, 2)
}

// testContextCancelerFunc is the common code for testing contextCanceler.
//
// cancelManner indicates how to cancel the contextCanceler, as follows:
//   - 0: by Canceler.Cancel
//   - 1: by context.CancelFunc
//   - 2: half by Canceler.Cancel, half by context.CancelFunc
func testContextCancelerFunc(t *testing.T, cancelManner int) {
	if cancelManner < 0 || cancelManner > 2 {
		t.Fatalf("unknown cancelManner %d; want 0, 1, or 2", cancelManner)
	}
	const N int = 10
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	canceler := concurrency.NewCancelerFromContext(ctx, cancel)
	logPrefix := "before calling Cancel, "
	testCancelerCAndCanceled(t, logPrefix, canceler, false)
	testContextDoneErrAndCause(t, logPrefix, ctx, false, nil)
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			switch cancelManner {
			case 0:
				canceler.Cancel()
			case 1:
				cancel()
			case 2:
				if rank&1 == 0 {
					canceler.Cancel()
				} else {
					cancel()
				}
			}
			logPrefix := fmt.Sprintf("goroutine %d, after calling Cancel, ",
				rank)
			testCancelerCAndCanceled(t, logPrefix, canceler, true)
			testContextDoneErrAndCause(
				t, logPrefix, ctx, true, context.Canceled)
		}(i)
	}
	wg.Wait()
}

func TestContextCauseCanceler_CancelByCanceler(t *testing.T) {
	testContextCauseCancelerFunc(t, 0)
}

func TestContextCauseCanceler_CancelByCancelFunc(t *testing.T) {
	testContextCauseCancelerFunc(t, 1)
}

func TestContextCauseCanceler_CancelByBothCancelerAndCancelFunc(t *testing.T) {
	testContextCauseCancelerFunc(t, 2)
}

// testContextCauseCancelerFunc is the common code
// for testing contextCauseCanceler.
//
// cancelManner indicates how to cancel the contextCanceler, as follows:
//   - 0: by Canceler.Cancel
//   - 1: by context.CancelFunc
//   - 2: half by Canceler.Cancel, half by context.CancelFunc
func testContextCauseCancelerFunc(t *testing.T, cancelManner int) {
	if cancelManner < 0 || cancelManner > 2 {
		t.Fatalf("unknown cancelManner %d; want 0, 1, or 2", cancelManner)
	}
	const N int = 10
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(concurrency.ErrCanceled)
	canceler := concurrency.NewCancelerFromContextCause(ctx, cancel)
	logPrefix := "before calling Cancel, "
	testCancelerCAndCanceled(t, logPrefix, canceler, false)
	testContextDoneErrAndCause(t, logPrefix, ctx, false, nil)
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			switch cancelManner {
			case 0:
				canceler.Cancel()
			case 1:
				cancel(concurrency.ErrCanceled)
			case 2:
				if rank&1 == 0 {
					canceler.Cancel()
				} else {
					cancel(concurrency.ErrCanceled)
				}
			}
			logPrefix := fmt.Sprintf("goroutine %d, after calling Cancel, ",
				rank)
			testCancelerCAndCanceled(t, logPrefix, canceler, true)
			testContextDoneErrAndCause(
				t, logPrefix, ctx, true, concurrency.ErrCanceled)
		}(i)
	}
	wg.Wait()
}

// testCancelerCAndCanceled checks canceler.C() and canceler.Canceled().
func testCancelerCAndCanceled(
	t *testing.T,
	logPrefix string,
	canceler concurrency.Canceler,
	wantCanceled bool,
) {
	c := canceler.C()
	if c != nil {
		select {
		case <-c:
			if !wantCanceled {
				t.Errorf("%sCanceler.C is closed", logPrefix)
			}
		default:
			if wantCanceled {
				t.Errorf("%sCanceler.C is not closed", logPrefix)
			}
		}
	} else {
		t.Errorf("%sCanceler.C returned nil", logPrefix)
	}
	if gotCanceled := canceler.Canceled(); gotCanceled != wantCanceled {
		t.Errorf("%sgot Canceler.Canceled %t; want %t",
			logPrefix, gotCanceled, wantCanceled)
	}
}

// testContextDoneErrAndCause checks ctx.Done(), ctx.Err(),
// and context.Cause(ctx).
func testContextDoneErrAndCause(
	t *testing.T,
	logPrefix string,
	ctx context.Context,
	wantDone bool,
	wantCause error,
) {
	select {
	case <-ctx.Done():
		if !wantDone {
			t.Errorf("%sContext.Done is closed", logPrefix)
		}
	default:
		if wantDone {
			t.Errorf("%sContext.Done is not closed", logPrefix)
		}
	}
	var wantErr error
	if wantCause != nil {
		wantErr = context.Canceled
	}
	if err := ctx.Err(); !errors.Is(err, wantErr) {
		t.Errorf("%sgot Context.Err %v; want %v", logPrefix, err, wantErr)
	}
	if cause := context.Cause(ctx); !errors.Is(cause, wantCause) {
		t.Errorf("%sgot cause %v; want %v", logPrefix, cause, wantCause)
	}
}
