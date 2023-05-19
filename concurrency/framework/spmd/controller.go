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

package spmd

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"sync"

	"github.com/donyori/gogo/concurrency"
	"github.com/donyori/gogo/concurrency/framework"
	"github.com/donyori/gogo/concurrency/framework/internal/quitdevice"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/errors"
)

// BusinessFunc is a function to achieve the user business.
//
// The first parameter is the communicator of the world group,
// which is the default group and includes all goroutines to process the job.
//
// The second parameter is the map of communicators of custom groups.
// The key of the map is the ID of the custom group.
// The value of the map is the corresponding communicator for this goroutine.
// If there is no custom group, commMap is nil.
type BusinessFunc[Message any] func(world Communicator[Message], commMap map[string]Communicator[Message])

// groupIDPattern is a regular expression pattern for verifying group ID.
var groupIDPattern = regexp.MustCompile(`[a-z0-9][a-z0-9_]*`)

// New creates a Controller for a new job.
//
// n is the number of goroutines to process the job.
// If n is non-positive, runtime.NumCPU() is used instead.
//
// biz is the business handler for the job.
// It panics if biz is nil.
//
// groupMap is the map of custom groups.
// The key is the ID of the group, which consists of lowercase English letters,
// numbers, and underscores, and cannot be empty.
// For custom groups, the ID cannot start with an underscore.
// (i.e., in regular expression: [a-z0-9][a-z0-9_]*).
// An illegal ID causes a panic.
// The value is a list of the world ranks of goroutines.
// Duplicate numbers are ignored.
// Out-of-range numbers (i.e., < 0 or >= n) cause a panic.
// A nil or empty group also causes a panic.
// Each group has its own communicator for each goroutine.
// The rank of the communicator depends on the order of the world ranks
// in the list.
// The client can get the communicators of custom groups via argument commMap
// of the business function biz.
func New[Message any](n int, biz BusinessFunc[Message], groupMap map[string][]int) framework.Controller {
	if biz == nil {
		panic(errors.AutoMsg("biz is nil"))
	} else if n <= 0 {
		n = runtime.NumCPU()
	}
	ctrl := &controller[Message]{
		qd:           quitdevice.NewQuitDevice(),
		cd:           newChanDispr[Message](),
		biz:          biz,
		lnchOi:       concurrency.NewOnceIndicator(),
		cdOi:         concurrency.NewOnceIndicator(),
		lnchCommMaps: make([]map[string]Communicator[Message], n),
	}
	worldRanks := make([]int, n)
	for i := range worldRanks {
		worldRanks[i] = i
	}
	ctrl.world = newContext(ctrl, "_world", worldRanks)
	for id, group := range groupMap {
		if !groupIDPattern.MatchString(id) {
			panic(errors.AutoMsg("group ID is illegal: " + id))
		} else if len(group) == 0 {
			panic(errors.AutoMsg("group is nil or empty"))
		}
		g, set := make(array.SliceDynamicArray[int], 0, len(group)), make(map[int]bool, len(group))
		// Deduplicate and check out-of-range items:
		for _, wr := range group {
			if wr < 0 || wr >= n {
				panic(errors.AutoMsg(fmt.Sprintf("world rank %d is out of range (n: %d)", wr, n)))
			} else if set[wr] {
				continue
			}
			set[wr] = true
			g.Push(wr)
		}
		g.Shrink()
		ctx := newContext(ctrl, id, g)
		for r, wr := range ctx.worldRanks {
			if ctrl.lnchCommMaps[wr] == nil {
				ctrl.lnchCommMaps[wr] = make(map[string]Communicator[Message])
			}
			ctrl.lnchCommMaps[wr][id] = ctx.comms[r]
		}
	}
	return ctrl
}

// Run creates a Controller with specified parameters, and then runs it.
// It returns the panic records of the Controller.
//
// The parameters are the same as those of function New.
func Run[Message any](n int, biz BusinessFunc[Message], groupMap map[string][]int) []framework.PanicRecord {
	ctrl := New(n, biz, groupMap)
	ctrl.Run()
	return ctrl.PanicRecords()
}

// controller is an implementation of interface Controller.
type controller[Message any] struct {
	qd    framework.QuitDevice // Quit device.
	world *context[Message]    // World context.
	cd    *chanDispr[Message]  // Channel dispatcher.

	biz    BusinessFunc[Message]     // Business function.
	pr     framework.PanicRecords    // Panic records.
	wg     sync.WaitGroup            // Wait group for the main process.
	lnchOi concurrency.OnceIndicator // For launching the job.
	cdOi   concurrency.OnceIndicator // For launching the channel dispatcher.
	cdFinC chan struct{}             // Channel for the finish signal of the channel dispatcher.

	// List of commMap used by method Launch,
	// will be nil after calling Launch.
	lnchCommMaps []map[string]Communicator[Message]
}

func (ctrl *controller[Message]) QuitChan() <-chan struct{} {
	return ctrl.qd.QuitChan()
}

func (ctrl *controller[Message]) IsQuit() bool {
	return ctrl.qd.IsQuit()
}

func (ctrl *controller[Message]) Quit() {
	ctrl.qd.Quit()
}

func (ctrl *controller[Message]) Launch() {
	ctrl.lnchOi.Do(func() {
		n, commMaps := len(ctrl.world.comms), ctrl.lnchCommMaps
		ctrl.wg.Add(n)
		for i := 0; i < n; i++ {
			go func(rank int) {
				defer func() {
					if e := recover(); e != nil {
						ctrl.qd.Quit()
						ctrl.pr.Append(framework.PanicRecord{
							Name:    strconv.Itoa(rank),
							Content: e,
						})
					}
					ctrl.wg.Done()
				}()
				ctrl.biz(ctrl.world.comms[rank], commMaps[rank])
			}(i)
		}
		ctrl.lnchCommMaps = nil
	})
}

func (ctrl *controller[Message]) Wait() int {
	if !ctrl.lnchOi.Test() {
		return -1
	}
	defer func() {
		ctrl.qd.Quit() // For cleanup possible daemon goroutines that wait for a quit signal to exit.
		if ctrl.cdOi.Test() {
			<-ctrl.cdFinC // Wait for the channel dispatcher to finish.
		}
	}()
	ctrl.wg.Wait()
	return ctrl.pr.Len()
}

func (ctrl *controller[Message]) Run() int {
	ctrl.Launch()
	return ctrl.Wait()
}

func (ctrl *controller[Message]) NumGoroutine() int {
	return len(ctrl.world.comms)
}

func (ctrl *controller[Message]) PanicRecords() []framework.PanicRecord {
	return ctrl.pr.List()
}

// launchChannelDispatcher launches a channel dispatcher in a daemon goroutine.
// This method takes effect only once.
func (ctrl *controller[Message]) launchChannelDispatcher() {
	ctrl.cdOi.Do(func() {
		ctrl.cdFinC = make(chan struct{})
		go func() {
			defer func() {
				if e := recover(); e != nil {
					ctrl.qd.Quit()
					ctrl.pr.Append(framework.PanicRecord{
						Name:    "channel_dispatcher",
						Content: e,
					})
				}
			}()
			ctrl.cd.Run(ctrl.qd, ctrl.cdFinC)
		}()
	})
}
