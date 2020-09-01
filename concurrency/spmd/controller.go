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
	"sync"

	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/errors"
)

// A controller to launch, quit and wait for the job.
type Controller interface {
	// Launch the job.
	//
	// This method will NOT wait until the job ends.
	// Use method Wait if you want to wait for that.
	//
	// Note that Launch can take effect only once.
	// To do the same job again, create a new Controller
	// with the same parameters.
	Launch()

	// Quit the job.
	//
	// This method will NOT wait until the job ends.
	// Use method Wait if you want to wait for that.
	Quit()

	// Wait for the job to finish or quit.
	// It returns the number of panic goroutines.
	Wait() int

	// Launch the job and wait for it.
	// It returns the number of panic goroutines.
	Run() int

	// Return the number of goroutines to process this job.
	NumGoroutine() int

	// Return the channel for the quit signal.
	QuitChan() <-chan struct{}

	// Return the panic records.
	PanicRecords() []PanicRec
}

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
var groupIdPattern = regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9_]*`)

// Create a Controller for a new job.
//
// n is the number of goroutines to process the job.
// If n is non-positive, runtime.NumCPU() will be used instead.
//
// biz is the business handler for the job.
// It will panic if biz is nil.
//
// groupMap is the map of custom groups.
// The key is the ID of the group, which consists of English letters, numbers,
// and underscores, and cannot be empty or start with an underscore
// (in regular expression: [A-Za-z0-9][A-Za-z0-9_]*).
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
func New(n int, biz BusinessFunc, groupMap map[string][]int) Controller {
	if biz == nil {
		panic(errors.AutoMsg("biz is nil"))
	}
	if n <= 0 {
		n = runtime.NumCPU()
	}
	ctrl := &controller{
		QuitC:        make(chan struct{}),
		Cd:           newChanDispr(),
		biz:          biz,
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

// Create a Controller with n, biz, and groupMap, and then run it.
// It returns the panic records of the Controller.
//
// The arguments n, biz, and groupMap are the same as those of function New.
func Run(n int, biz BusinessFunc, groupMap map[string][]int) []PanicRec {
	ctrl := New(n, biz, groupMap)
	ctrl.Run()
	return ctrl.PanicRecords()
}

// An implementation of interface Controller.
type controller struct {
	QuitC chan struct{} // A channel to broadcast the quit signal.
	World *context      // World context.
	Cd    *chanDispr    // Channel dispatcher.

	biz      BusinessFunc   // Business function.
	pr       panicRecords   // Panic records.
	wg       sync.WaitGroup // A wait group for the main process.
	lnchOnce sync.Once      // For launching the job.
	quitOnce sync.Once      // For closing QuitC.
	cdOnce   sync.Once      // For launching the channel dispatcher.
	// List of commMap used by method Launch,
	// will be nil after calling Launch.
	lnchCommMaps []map[string]Communicator
}

func (ctrl *controller) Launch() {
	ctrl.lnchOnce.Do(func() {
		n := len(ctrl.World.Comms)
		commMaps := ctrl.lnchCommMaps
		ctrl.wg.Add(n)
		for i := 0; i < n; i++ {
			go func(rank int) {
				defer func() {
					if r := recover(); r != nil {
						ctrl.Quit()
						ctrl.pr.Append(PanicRec{
							Rank:    rank,
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

func (ctrl *controller) Quit() {
	ctrl.quitOnce.Do(func() {
		close(ctrl.QuitC)
	})
}

func (ctrl *controller) Wait() int {
	defer ctrl.Quit() // For cleanup possible daemon goroutines that wait for a quit signal to exit.
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

func (ctrl *controller) QuitChan() <-chan struct{} {
	return ctrl.QuitC
}

func (ctrl *controller) PanicRecords() []PanicRec {
	return ctrl.pr.List()
}

// Launch channel dispatcher in a daemon goroutine.
// This method takes effect only once.
func (ctrl *controller) launchChannelDispatcher() {
	ctrl.cdOnce.Do(func() {
		go ctrl.Cd.Run(ctrl.QuitC)
	})
}
