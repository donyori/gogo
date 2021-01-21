// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

import "github.com/donyori/gogo/concurrency/framework"

// chanCtr is a combination of channel or list of channels, and counter.
type chanCtr struct {
	Chan interface{}
	Ctr  int
}

// chanDispr is a channel dispatcher.
//
// Its underlying type is an array of channels for
// receiving the channel dispatch query.
type chanDispr [numCOp]chan *communicator

// newChanDispr creates a new channel dispatcher.
// Only for function New.
func newChanDispr() *chanDispr {
	cd := new(chanDispr)
	for i := range cd {
		cd[i] = make(chan *communicator)
	}
	return cd
}

// Run launches the channel dispatcher on current goroutine.
//
// quitDevice is the device to receive a quit signal.
// It should be obtained from Controller.
// The function will panic if quitDevice is nil.
//
// finChan is a channel to broadcast a finish signal by closing the channel.
// It will be closed at the end of this function.
// finChan will be ignored if it is nil.
func (cd *chanDispr) Run(quitDevice framework.QuitDevice, finChan chan<- struct{}) {
	if finChan != nil {
		defer close(finChan)
	}
	var (
		comm  *communicator
		op, n int
		ctr   int64
		ctx   *context
		m     map[int64]*chanCtr
		cc    *chanCtr
		cs    []chan interface{}
	)
	quitChan := quitDevice.QuitChan()
	for {
		comm, ctx, m, cc = nil, nil, nil, nil // Reset variables to enable GC to clear contexts that are no longer used.
		select {
		case <-quitChan:
			return
		case comm = <-cd[cOpBcast]:
			op = cOpBcast
		case comm = <-cd[cOpScatter]:
			op = cOpScatter
		case comm = <-cd[cOpGather]:
			op = cOpGather
		}
		if comm == nil {
			continue
		}
		ctr = comm.COpCtrs[op]
		comm.COpCtrs[op]++
		ctx = comm.Ctx
		m = ctx.ChanMaps[op]
		if m == nil {
			m = make(map[int64]*chanCtr)
			ctx.ChanMaps[op] = m
		}
		cc = m[ctr]
		if cc == nil {
			n = len(ctx.Comms) - 1
			if n > 0 {
				cc = &chanCtr{Ctr: n}
				switch op {
				case cOpBcast:
					cc.Chan = make(chan interface{}, n)
				case cOpScatter:
					cs = make([]chan interface{}, n)
					for i := range cs {
						cs[i] = make(chan interface{}, 1)
					}
					cc.Chan, cs = cs, nil // Reset cs to enable GC to clear contexts that are no longer used.
				case cOpGather:
					cc.Chan = make(chan *sndrMsg, n)
				default:
					continue // Ignore invalid operation.
				}
				m[ctr] = cc
			} else {
				cc = &chanCtr{Chan: make(chan interface{}, 1)}
				// Don't store cc into m when n is 0.
			}
		} else {
			cc.Ctr--
			if cc.Ctr == 0 {
				delete(m, ctr)
			}
		}
		select {
		case <-quitChan:
			return
		case comm.Cdc <- cc.Chan:
		}
	}
}
