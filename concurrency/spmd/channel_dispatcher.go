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

// Combination of channel and counter.
type chanCntr struct {
	Chan chan interface{}
	Cntr int
}

// Channel dispatcher.
type chanDispr struct {
	QueryChans [numCOp]chan *communicator // List of channels for receiving the channel dispatch query.
}

// Create a new channel dispatcher.
// Only for function New.
func newChanDispr() *chanDispr {
	cd := new(chanDispr)
	for i := range cd.QueryChans {
		cd.QueryChans[i] = make(chan *communicator)
	}
	return cd
}

// Launch the channel dispatcher on current goroutine.
// The parameter quitChan should be obtained from Controller.
func (cd *chanDispr) Run(quitChan <-chan struct{}) {
	var (
		comm  *communicator
		op, n int
		cntr  int64
		ctx   *context
		m     map[int64]*chanCntr
		cc    *chanCntr
	)
	for {
		select {
		case <-quitChan:
			return
		case comm = <-cd.QueryChans[cOpBcast]:
			op = cOpBcast
		case comm = <-cd.QueryChans[cOpScatter]:
			op = cOpScatter
		case comm = <-cd.QueryChans[cOpGather]:
			op = cOpGather
		}
		if comm == nil {
			continue
		}
		cntr = comm.COpCntrs[op]
		comm.COpCntrs[op]++
		ctx = comm.Ctx
		m = ctx.ChanMaps[op]
		if m == nil {
			m = make(map[int64]*chanCntr)
			ctx.ChanMaps[op] = m
		}
		cc = m[cntr]
		if cc == nil {
			n = len(ctx.Comms) - 1
			if n > 0 {
				cc = &chanCntr{
					Chan: make(chan interface{}, n),
					Cntr: n,
				}
				m[cntr] = cc
			} else {
				cc = &chanCntr{Chan: make(chan interface{}, 1)}
				// Don't store cc into m when n is 0.
			}
		} else {
			cc.Cntr--
			if cc.Cntr == 0 {
				delete(m, cntr)
			}
		}
		select {
		case <-quitChan:
			return
		case comm.Cdc <- cc.Chan:
		}
	}
}
