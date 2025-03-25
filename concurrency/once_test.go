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
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/donyori/gogo/concurrency"
)

func TestOnce_Do_Once(t *testing.T) {
	const N int = 10
	var x atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
	})
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			once.Do()
			if gotX := x.Load(); gotX != 1 {
				t.Errorf("goroutine %d, got x %d; want 1", rank, gotX)
			}
		}(i)
	}
	wg.Wait()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("main goroutine, got x %d; want 1", gotX)
	}
}

func TestOnce_DoRecover_Once(t *testing.T) {
	const N int = 10
	var x, calledCtr atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
	})
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			called, panicValue := once.DoRecover()
			if called {
				calledCtr.Add(1)
			}
			if panicValue != nil {
				t.Errorf("goroutine %d, got panicValue %v", rank, panicValue)
			}
			if gotX := x.Load(); gotX != 1 {
				t.Errorf("goroutine %d, got x %d; want 1", rank, gotX)
			}
		}(i)
	}
	wg.Wait()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("main goroutine, got x %d; want 1", gotX)
	}
	if ctr := calledCtr.Load(); ctr != 1 {
		t.Errorf("got true called %d; want 1", ctr)
	}
}

func TestOnce_Do_DoRecover_Once(t *testing.T) {
	const N int = 20
	var x atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
	})
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			if rank&1 == 0 {
				once.Do()
			} else {
				_, panicValue := once.DoRecover()
				if panicValue != nil {
					t.Errorf("goroutine %d, got panicValue %v",
						rank, panicValue)
				}
			}
			if gotX := x.Load(); gotX != 1 {
				t.Errorf("goroutine %d, got x %d; want 1", rank, gotX)
			}
		}(i)
	}
	wg.Wait()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("main goroutine, got x %d; want 1", gotX)
	}
}

func TestOnce_Do_Panic(t *testing.T) {
	const N int = 10
	var panicMsg any = "test panic message"
	var x atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
		panic(panicMsg)
	})
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			defer func() {
				if e := recover(); e != panicMsg {
					t.Errorf("goroutine %d, got panic %v; want %v",
						rank, e, panicMsg)
				}
			}()
			once.Do()
		}(i)
	}
	wg.Wait()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("main goroutine, got x %d; want 1", gotX)
	}
}

func TestOnce_DoRecover_Panic(t *testing.T) {
	const N int = 10
	var panicMsg any = "test panic message"
	var x, calledCtr atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
		panic(panicMsg)
	})
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			defer func() {
				if e := recover(); e != nil {
					t.Errorf("goroutine %d, got panic %v", rank, e)
				}
			}()
			called, panicValue := once.DoRecover()
			if called {
				calledCtr.Add(1)
			}
			if panicValue != panicMsg {
				t.Errorf("goroutine %d, got panicValue %v; want %v",
					rank, panicValue, panicMsg)
			}
			if gotX := x.Load(); gotX != 1 {
				t.Errorf("goroutine %d, got x %d; want 1", rank, gotX)
			}
		}(i)
	}
	wg.Wait()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("main goroutine, got x %d; want 1", gotX)
	}
	if ctr := calledCtr.Load(); ctr != 1 {
		t.Errorf("got true called %d; want 1", ctr)
	}
}

func TestOnce_DoPanicThenDoRecover(t *testing.T) {
	var panicMsg any = "test panic message"
	var x atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
		panic(panicMsg)
	})

	func() {
		defer func() {
			if e := recover(); e != panicMsg {
				t.Errorf("after calling Do, got panic %v; want %v",
					e, panicMsg)
			}
		}()
		once.Do()
	}()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("after calling Do, got x %d; want 1", gotX)
	}

	called, panicValue := once.DoRecover()
	if called {
		t.Error("got called true; want false")
	}
	if panicValue != panicMsg {
		t.Errorf("got panicValue %v; want %v", panicValue, panicMsg)
	}
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("after calling DoRecover, got x %d; want 1", gotX)
	}
}

func TestOnce_DoRecoverPanicThenDo(t *testing.T) {
	var panicMsg any = "test panic message"
	var x atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
		panic(panicMsg)
	})

	called, panicValue := once.DoRecover()
	if !called {
		t.Error("got called false; want true")
	}
	if panicValue != panicMsg {
		t.Errorf("got panicValue %v; want %v", panicValue, panicMsg)
	}
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("after calling DoRecover, got x %d; want 1", gotX)
	}

	func() {
		defer func() {
			if e := recover(); e != panicMsg {
				t.Errorf("after calling Do, got panic %v; want %v",
					e, panicMsg)
			}
		}()
		once.Do()
	}()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("after calling Do, got x %d; want 1", gotX)
	}
}

func TestOnce_Do_PanicTraceback(t *testing.T) {
	// Test that on the first invocation of Once.Do,
	// which calls the specified function for the first time,
	// the stack trace goes all the way to the origin of the panic.
	var panicMsg any = "test panic message"
	var x atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
		panic(panicMsg)
	})
	func() {
		defer func() {
			if e := recover(); e != panicMsg {
				t.Errorf("got panic %v; want %v", e, panicMsg)
				return
			}
			stack := string(debug.Stack())
			want := "concurrency_test.TestOnce_Do_PanicTraceback.func1"
			if !strings.Contains(stack, want) {
				t.Errorf("want stack containing %q; got\n%s", want, stack)
			}
		}()
		once.Do()
	}()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("got x %d; want 1", gotX)
	}
}

func TestOnce_Do_Goexit(t *testing.T) {
	const N int = 10
	var x atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
		runtime.Goexit()
	})
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			defer func() {
				if e := recover(); e != nil {
					t.Errorf("goroutine %d, got panic %v", rank, e)
				}
			}()
			once.Do()
			t.Errorf("goroutine %d was not interrupted", rank)
		}(i)
	}
	wg.Wait()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("got x %d; want 1", gotX)
	}
}

func TestOnce_DoRecover_Goexit(t *testing.T) {
	const N int = 10
	var x atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
		runtime.Goexit()
	})
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			defer func() {
				if e := recover(); e != nil {
					t.Errorf("goroutine %d, got panic %v", rank, e)
				}
			}()
			once.DoRecover()
			t.Errorf("goroutine %d was not interrupted", rank)
		}(i)
	}
	wg.Wait()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("got x %d; want 1", gotX)
	}
}

func TestOnce_Do_DoRecover_Goexit(t *testing.T) {
	const N int = 20
	var x atomic.Int32
	once := concurrency.NewOnce(func() {
		x.Add(1)
		runtime.Goexit()
	})
	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func(rank int) {
			defer wg.Done()
			defer func() {
				if e := recover(); e != nil {
					t.Errorf("goroutine %d, got panic %v", rank, e)
				}
			}()
			if rank&1 == 0 {
				once.Do()
			} else {
				once.DoRecover()
			}
			t.Errorf("goroutine %d was not interrupted", rank)
		}(i)
	}
	wg.Wait()
	if gotX := x.Load(); gotX != 1 {
		t.Errorf("got x %d; want 1", gotX)
	}
}

func TestOnce_C(t *testing.T) {
	testOnceCAndDoneFunc(t, func(t *testing.T, once concurrency.Once) bool {
		c := once.C()
		if c == nil {
			t.Error("Once.C returned nil")
			return false
		}
		select {
		case <-c:
			return true
		default:
			return false
		}
	})
}

func TestOnce_Done(t *testing.T) {
	testOnceCAndDoneFunc(t, func(t *testing.T, once concurrency.Once) bool {
		return once.Done()
	})
}

func TestOnce_PanicValue(t *testing.T) {
	const (
		EmptyF  = "emptyF"
		PanicF  = "panicF"
		GoexitF = "goexitF"
		NilF    = "<nil>"
	)
	var panicMsg any = "test panic message"
	testCases := []struct {
		fName           string
		useDo           bool
		wantInterrupted bool
		wantPanicValue  any
	}{
		{EmptyF, true, false, nil},
		{EmptyF, false, false, nil},
		{PanicF, true, true, panicMsg},
		{PanicF, false, true, panicMsg},
		{GoexitF, true, true, nil},
		{GoexitF, false, true, nil},
		{NilF, true, false, nil},
		{NilF, false, false, nil},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("f=%s&useDo=%t", tc.fName, tc.useDo),
			func(t *testing.T) {
				var f func()
				switch tc.fName {
				case EmptyF:
					f = func() {}
				case PanicF:
					f = func() {
						panic(panicMsg)
					}
				case GoexitF:
					f = runtime.Goexit
				case NilF:
					// Do nothing here.
				default:
					t.Fatalf("unknown fName %q", tc.fName)
				}
				methodName := "DoRecover"
				if tc.useDo {
					methodName = "Do"
				}

				once := concurrency.NewOnce(f)
				interrupted, panicValue := once.PanicValue()
				if interrupted {
					t.Errorf("before calling %s, got interrupted true; want false",
						methodName)
				}
				if panicValue != nil {
					t.Errorf("before calling %s, got panicValue %v; want <nil>",
						methodName, panicValue)
				}

				// Launch a new goroutine to call Do or DoRecover
				// because runtime.Goexit will break the test goroutine.
				doneC := make(chan struct{})
				go func(doneC chan<- struct{}) {
					defer close(doneC)
					if tc.useDo {
						defer func() {
							if e := recover(); e != tc.wantPanicValue {
								t.Errorf("got panic %v; want %v",
									e, tc.wantPanicValue)
							}
						}()
						once.Do()
					} else {
						called, panicValue := once.DoRecover()
						if !called {
							t.Error("got called false; want true")
						}
						if panicValue != tc.wantPanicValue {
							t.Errorf("got panicValue %v; want %v",
								panicValue, tc.wantPanicValue)
						}
					}
				}(doneC)
				<-doneC

				interrupted, panicValue = once.PanicValue()
				if interrupted != tc.wantInterrupted {
					t.Errorf("after calling %s, got interrupted %t; want %t",
						methodName, interrupted, tc.wantInterrupted)
				}
				if panicValue != tc.wantPanicValue {
					t.Errorf("after calling %s, got panicValue %v; want %v",
						methodName, panicValue, tc.wantPanicValue)
				}
			},
		)
	}
}

// testOnceCAndDoneFunc is the common code for testing Once.C and Once.Done.
//
// To test Once.C, set doneFunc as follows:
//
//	func(t *testing.T, once concurrency.Once) bool {
//		c := once.C()
//		if c == nil {
//			t.Error("Once.C returned nil")
//			return false
//		}
//		select {
//		case <-c:
//			return true
//		default:
//			return false
//		}
//	}
//
// To test Once.Done, set doneFunc as follows:
//
//	func(t *testing.T, once concurrency.Once) bool {
//		return once.Done()
//	}
func testOnceCAndDoneFunc(
	t *testing.T,
	doneFunc func(t *testing.T, once concurrency.Once) bool,
) {
	const (
		EmptyF  = "emptyF"
		PanicF  = "panicF"
		GoexitF = "goexitF"
		NilF    = "<nil>"
	)
	testCases := []struct {
		fName string
		useDo bool
	}{
		{EmptyF, true},
		{EmptyF, false},
		{PanicF, true},
		{PanicF, false},
		{GoexitF, true},
		{GoexitF, false},
		{NilF, true},
		{NilF, false},
	}

	var panicMsg any = "test panic message"
	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("f=%s&useDo=%t", tc.fName, tc.useDo),
			func(t *testing.T) {
				var f func()
				var wantPanicValue any
				switch tc.fName {
				case EmptyF:
					f = func() {}
				case PanicF:
					f = func() {
						panic(panicMsg)
					}
					wantPanicValue = panicMsg
				case GoexitF:
					f = runtime.Goexit
				case NilF:
					// Do nothing here.
				default:
					t.Fatalf("unknown fName %q", tc.fName)
				}

				// Launch a new goroutine because
				// runtime.Goexit will break the test goroutine.
				var numCallToDoneFunc atomic.Int32
				doneC := make(chan struct{})
				go goroutineTestOnceCAndDoneFunc(
					doneC,
					t,
					doneFunc,
					tc.useDo,
					f,
					wantPanicValue,
					&numCallToDoneFunc,
				)
				<-doneC
				if n := numCallToDoneFunc.Load(); n != 2 {
					t.Errorf("the number of calls to doneFunc is %d; want 2", n)
				}
			},
		)
	}
}

// goroutineTestOnceCAndDoneFunc is the process of the goroutine
// launched by testOnceCAndDoneFunc.
func goroutineTestOnceCAndDoneFunc(
	doneC chan<- struct{},
	t *testing.T,
	doneFunc func(t *testing.T, once concurrency.Once) bool,
	useDo bool,
	f func(),
	wantPanicValue any,
	numCallToDoneFunc *atomic.Int32,
) {
	defer close(doneC)
	once := concurrency.NewOnce(f)
	methodName := "DoRecover"
	if useDo {
		methodName = "Do"
		defer func() {
			if e := recover(); e != wantPanicValue {
				t.Errorf("got panic %v; want %v",
					e, wantPanicValue)
			}
		}()
	}
	done := doneFunc(t, once)
	numCallToDoneFunc.Add(1)
	if !t.Failed() && done {
		t.Errorf("doneFunc returned true before calling %s; want false",
			methodName)
	}
	defer func() {
		done = doneFunc(t, once)
		numCallToDoneFunc.Add(1)
		if !t.Failed() && !done {
			t.Errorf("doneFunc returned false after calling %s; want true",
				methodName)
		}
	}()
	if useDo {
		once.Do()
	} else {
		called, panicValue := once.DoRecover()
		if !called {
			t.Error("got called false; want true")
		}
		if panicValue != wantPanicValue {
			t.Errorf("got panicValue %v; want %v",
				panicValue, wantPanicValue)
		}
	}
}
