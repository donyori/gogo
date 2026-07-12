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
	"fmt"
	"sync"
	"testing"

	"github.com/donyori/gogo/concurrency"
)

func TestOnceCanceler(t *testing.T) {
	t.Parallel()

	const N int = 10

	canceler := concurrency.NewCanceler()
	testCancelerCAndCanceled(t, "before calling Cancel, ", canceler, false)

	var wg sync.WaitGroup
	for rank := range N {
		wg.Go(func() {
			canceler.Cancel()
			testCancelerCAndCanceled(
				t,
				fmt.Sprintf("goroutine %d, after calling Cancel, ", rank),
				canceler,
				true,
			)
		})
	}

	wg.Wait()
}

func TestContextCanceler_CancelByCanceler(t *testing.T) {
	t.Parallel()

	testContextCancelerFunc(t, 0)
}

func TestContextCanceler_CancelByCancelFunc(t *testing.T) {
	t.Parallel()

	testContextCancelerFunc(t, 1)
}

func TestContextCanceler_CancelByBothCancelerAndCancelFunc(t *testing.T) {
	t.Parallel()

	testContextCancelerFunc(t, 2)
}

// testContextCancelerFunc is the common code for testing contextCanceler.
//
// cancelManner indicates how to cancel the contextCanceler, as follows:
//   - 0: by Canceler.Cancel
//   - 1: by context.CancelFunc
//   - 2: half by Canceler.Cancel, half by context.CancelFunc
func testContextCancelerFunc(t *testing.T, cancelManner int) {
	t.Helper()

	if cancelManner < 0 || cancelManner > 2 {
		t.Errorf("unknown cancelManner %d; want 0, 1, or 2", cancelManner)
		return
	}

	const N int = 10

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	canceler := concurrency.NewCancelerFromContext(ctx, cancel)
	logPrefix := "before calling Cancel, "
	testCancelerCAndCanceled(t, logPrefix, canceler, false)
	testContextDoneErrAndCause(t, ctx, logPrefix, false, nil)

	var wg sync.WaitGroup
	for rank := range N {
		wg.Go(func() {
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

			logPrefix := fmt.Sprintf(
				"goroutine %d, after calling Cancel, ",
				rank,
			)
			testCancelerCAndCanceled(t, logPrefix, canceler, true)
			testContextDoneErrAndCause(
				t,
				ctx,
				logPrefix,
				true,
				context.Canceled,
			)
		})
	}

	wg.Wait()
}

func TestCauseContextCanceler_CancelByCanceler(t *testing.T) {
	t.Parallel()

	testCauseContextCancelerFunc(t, 0)
}

func TestCauseContextCanceler_CancelByCancelFunc(t *testing.T) {
	t.Parallel()

	testCauseContextCancelerFunc(t, 1)
}

func TestCauseContextCanceler_CancelByBothCancelerAndCancelFunc(t *testing.T) {
	t.Parallel()

	testCauseContextCancelerFunc(t, 2)
}

// testCauseContextCancelerFunc is the common code
// for testing causeContextCanceler.
//
// cancelManner indicates how to cancel the contextCanceler, as follows:
//   - 0: by Canceler.Cancel
//   - 1: by context.CancelFunc
//   - 2: half by Canceler.Cancel, half by context.CancelFunc
func testCauseContextCancelerFunc(t *testing.T, cancelManner int) {
	t.Helper()

	if cancelManner < 0 || cancelManner > 2 {
		t.Errorf("unknown cancelManner %d; want 0, 1, or 2", cancelManner)
		return
	}

	const N int = 10

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(concurrency.ErrCanceled)

	canceler := concurrency.NewCancelerFromCauseContext(ctx, cancel)
	logPrefix := "before calling Cancel, "
	testCancelerCAndCanceled(t, logPrefix, canceler, false)
	testContextDoneErrAndCause(t, ctx, logPrefix, false, nil)

	var wg sync.WaitGroup
	for rank := range N {
		wg.Go(func() {
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

			logPrefix := fmt.Sprintf(
				"goroutine %d, after calling Cancel, ",
				rank,
			)
			testCancelerCAndCanceled(t, logPrefix, canceler, true)
			testContextDoneErrAndCause(
				t,
				ctx,
				logPrefix,
				true,
				concurrency.ErrCanceled,
			)
		})
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
	t.Helper()

	c := canceler.C()
	switch {
	case c == nil:
		t.Errorf("%sCanceler.C returned nil", logPrefix)
	case concurrency.TestChannel(c):
		if !wantCanceled {
			t.Errorf("%sCanceler.C is closed", logPrefix)
		}
	case wantCanceled:
		t.Errorf("%sCanceler.C is not closed", logPrefix)
	}

	if gotCanceled := canceler.Canceled(); gotCanceled != wantCanceled {
		t.Errorf("%sgot Canceler.Canceled %t; want %t",
			logPrefix, gotCanceled, wantCanceled)
	}
}

//revive:disable:context-as-argument
// t (of type *testing.T) takes priority over ctx (of type context.Context)
// for the following test function.

// testContextDoneErrAndCause checks ctx.Done(), ctx.Err(),
// and context.Cause(ctx).
func testContextDoneErrAndCause(
	t *testing.T,
	ctx context.Context,
	logPrefix string,
	wantDone bool,
	wantCause error,
) {
	t.Helper()

	if concurrency.TestDone(ctx) {
		if !wantDone {
			t.Errorf("%sContext.Done is closed", logPrefix)
		}
	} else if wantDone {
		t.Errorf("%sContext.Done is not closed", logPrefix)
	}

	var wantErr error
	if wantCause != nil {
		wantErr = context.Canceled
	}

	err := ctx.Err()
	if !errors.Is(err, wantErr) {
		t.Errorf("%sgot Context.Err %v; want %v", logPrefix, err, wantErr)
	}

	err = context.Cause(ctx)
	if !errors.Is(err, wantCause) {
		t.Errorf("%sgot cause %v; want %v", logPrefix, err, wantCause)
	}
}

//revive:enable:context-as-argument
