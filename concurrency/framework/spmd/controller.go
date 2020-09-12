// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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
	"github.com/donyori/gogo/concurrency/framework/internal"
	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/errors"
)

// Business handler.
//
// The first argument is the communicator of the group World,
// which is the default group and contains all goroutines to process the job.
//
// The second argument is the map of communicators of custom groups.
// The key of the map is the ID of the custom group.
// The value of the map is the corresponding communicator for this goroutine.
// If there is no custom group, commMap is nil.
type BusinessFunc func(world Communicator, commMap map[string]Communicator)

// Regexp pattern for verifying group ID.
var groupIdPattern = regexp.MustCompile(`[a-z0-9][a-z0-9_]*`)

// Create a Controller for a new job.
//
// n is the number of goroutines to process the job.
// If n is non-positive, runtime.NumCPU() will be used instead.
//
// biz is the business handler for the job.
// It will panic if biz is nil.
//
// groupMap is the map of custom groups.
// The key is the ID of the group, which consists of lowercase English letters,
// numbers, and underscores, and cannot be empty.
// For custom groups, the ID cannot start with an underscore.
// (i.e., in regular expression: [a-z0-9][a-z0-9_]*).
// An illegal ID will cause a panic.
// The value is a list of the world ranks of goroutines.
// Duplicated numbers will be ignored.
// Out-of-range numbers (i.e., < 0 or >= n) will cause a panic.
// A nil or empty group will also cause a panic.
// Every group has its own communicator for each goroutine.
// The rank of the communicator depends on the order of the world ranks
// in the list.
// The client can get the communicators of custom groups via argument commMap
// of the business function biz.
func New(n int, biz BusinessFunc, groupMap map[string][]int) framework.Controller {
	if biz == nil {
		panic(errors.AutoMsg("biz is nil"))
	}
	if n <= 0 {
		n = runtime.NumCPU()
	}
	ctrl := &controller{
		Qd:           internal.NewQuitDevice(),
		Cd:           newChanDispr(),
		biz:          biz,
		lnchOi:       concurrency.NewOnceIndicator(),
		cdOi:         concurrency.NewOnceIndicator(),
		lnchCommMaps: make([]map[string]Communicator, n),
	}
	worldRanks := make([]int, n)
	for i := range worldRanks {
		worldRanks[i] = i
	}
	ctrl.World = newContext(ctrl, "_world", worldRanks)
	var (
		set map[int]bool
		g   sequence.IntDynamicArray
		ctx *context
	)
	for id, group := range groupMap {
		if !groupIdPattern.MatchString(id) {
			panic(errors.AutoMsg("group ID is illegal: " + id))
		}
		if len(group) == 0 {
			panic(errors.AutoMsg("group is nil or empty"))
		}
		set = make(map[int]bool)
		g = group
		g.Filter(func(x interface{}) (keep bool) {
			i := x.(int)
			if i < 0 || i >= n {
				panic(errors.AutoMsgWithStrategy(fmt.Sprintf("world rank %d is out of range (n: %d)", i, n), -1, 2))
			}
			if set[i] {
				return false
			}
			set[i] = true
			return true
		})
		ctx = newContext(ctrl, id, g)
		for r, wr := range ctx.WorldRanks {
			if ctrl.lnchCommMaps[wr] == nil {
				ctrl.lnchCommMaps[wr] = make(map[string]Communicator)
			}
			ctrl.lnchCommMaps[wr][id] = ctx.Comms[r]
		}
	}
	return ctrl
}

// Create a Controller with given parameters, and then run it.
// It returns the panic records of the Controller.
//
// The arguments are the same as those of function New.
func Run(n int, biz BusinessFunc, groupMap map[string][]int) []framework.PanicRec {
	ctrl := New(n, biz, groupMap)
	ctrl.Run()
	return ctrl.PanicRecords()
}

// An implementation of interface Controller.
type controller struct {
	Qd    framework.QuitDevice // Quit device.
	World *context             // World context.
	Cd    *chanDispr           // Channel dispatcher.

	biz    BusinessFunc              // Business function.
	pr     framework.PanicRecords    // Panic records.
	wg     sync.WaitGroup            // Wait group for the main process.
	lnchOi concurrency.OnceIndicator // For launching the job.
	cdOi   concurrency.OnceIndicator // For launching the channel dispatcher.
	cdFinC chan struct{}             // Channel for the finish signal of the channel dispatcher.

	// List of commMap used by method Launch,
	// will be nil after calling Launch.
	lnchCommMaps []map[string]Communicator
}

func (ctrl *controller) QuitChan() <-chan struct{} {
	return ctrl.Qd.QuitChan()
}

func (ctrl *controller) IsQuit() bool {
	return ctrl.Qd.IsQuit()
}

func (ctrl *controller) Quit() {
	ctrl.Qd.Quit()
}

func (ctrl *controller) Launch() {
	ctrl.lnchOi.Do(func() {
		n := len(ctrl.World.Comms)
		commMaps := ctrl.lnchCommMaps
		ctrl.wg.Add(n)
		for i := 0; i < n; i++ {
			go func(rank int) {
				defer func() {
					if r := recover(); r != nil {
						ctrl.Qd.Quit()
						ctrl.pr.Append(framework.PanicRec{
							Name:    strconv.Itoa(rank),
							Content: r,
						})
					}
					ctrl.wg.Done()
				}()
				ctrl.biz(ctrl.World.Comms[rank], commMaps[rank])
			}(i)
		}
		ctrl.lnchCommMaps = nil
	})
}

func (ctrl *controller) Wait() int {
	if !ctrl.lnchOi.Test() {
		return -1
	}
	defer func() {
		ctrl.Qd.Quit() // For cleanup possible daemon goroutines that wait for a quit signal to exit.
		if ctrl.cdOi.Test() {
			<-ctrl.cdFinC // Wait for the channel dispatcher to finish.
		}
	}()
	ctrl.wg.Wait()
	return ctrl.pr.Len()
}

func (ctrl *controller) Run() int {
	ctrl.Launch()
	return ctrl.Wait()
}

func (ctrl *controller) NumGoroutine() int {
	return len(ctrl.World.Comms)
}

func (ctrl *controller) PanicRecords() []framework.PanicRec {
	return ctrl.pr.List()
}

// Launch channel dispatcher in a daemon goroutine.
// This method takes effect only once.
func (ctrl *controller) launchChannelDispatcher() {
	ctrl.cdOi.Do(func() {
		ctrl.cdFinC = make(chan struct{})
		go func() {
			defer func() {
				if r := recover(); r != nil {
					ctrl.Qd.Quit()
					ctrl.pr.Append(framework.PanicRec{
						Name:    "channel_dispatcher",
						Content: r,
					})
				}
			}()
			ctrl.Cd.Run(ctrl.Qd, ctrl.cdFinC)
		}()
	})
}
