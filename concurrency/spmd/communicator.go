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

// A communicator for goroutines to communicate with each other.
type Communicator interface {
	// Return the rank (from 0 to NumGoroutine()-1) of current goroutine
	// for this job.
	Rank() int

	// Return the number of goroutines to process this job.
	NumGoroutine() int

	// Return the channel for the quit signal.
	QuitChan() <-chan struct{}

	// Detect the quit signal on the quit channel.
	// It returns true if a quit signal is detected, and false otherwise.
	IsQuit() bool

	// Broadcast a quit signal to quit the job.
	Quit()

	// Block until all other goroutines for this job call method Barrier
	// of their own communicators, or a quit signal is detected.
	//
	// It returns true if all other goroutines call methods Barrier
	// successfully, otherwise (a quit signal is detected), false.
	//
	// It is used to make all goroutines synchronous,
	// i.e., consistent in the execution progress.
	Barrier() bool
}

// An implementation of interface Communicator.
type communicator struct {
	Ctx *world // Context of the job.

	rank int                // The rank of current goroutine.
	bc   chan chan struct{} // A channel used for method Barrier.
}

// Create a new communicator.
// Only for function New.
func newCommunicator(w *world, rank int) *communicator {
	comm := &communicator{
		Ctx:  w,
		rank: rank,
	}
	if rank > 0 {
		comm.bc = make(chan chan struct{})
	}
	return comm
}

func (comm *communicator) Rank() int {
	return comm.rank
}

func (comm *communicator) NumGoroutine() int {
	return comm.Ctx.NumGoroutine()
}

func (comm *communicator) QuitChan() <-chan struct{} {
	return comm.Ctx.QuitChan
}

func (comm *communicator) IsQuit() bool {
	select {
	case <-comm.Ctx.QuitChan:
		return true
	default:
		return false
	}
}

func (comm *communicator) Quit() {
	comm.Ctx.Quit()
}

func (comm *communicator) Barrier() bool {
	n := comm.Ctx.NumGoroutine()
	if n <= 1 {
		return !comm.IsQuit()
	}
	var c chan struct{}
	if comm.rank == 0 {
		// The first goroutine makes the signal channel.
		c = make(chan struct{})
	} else {
		// Other goroutines receive the signal channel from their respective previous goroutines.
		select {
		case <-comm.Ctx.QuitChan:
			return false
		case c = <-comm.Ctx.Comms[comm.rank].bc: // Listen on its own channel, not the sender's!
		}
	}
	if comm.rank == n-1 {
		// The last goroutine close the signal channel to broadcast
		// the information that all goroutines call method Barrier successfully.
		close(c)
	} else {
		// Other goroutines send the signal channel to their respective next goroutines.
		select {
		case <-comm.Ctx.QuitChan:
			return false
		case comm.Ctx.Comms[comm.rank+1].bc <- c: // Send it to the receiver's channel.
		}
		// Then listen on the signal channel.
		select {
		case <-comm.Ctx.QuitChan:
			return false
		case <-c:
		}
	}
	return true
}
