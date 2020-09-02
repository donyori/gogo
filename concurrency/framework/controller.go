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

package framework

// A controller to launch, quit, and wait for the job.
//
// The use of all the frameworks under this package starts with creating
// a controller through their own New function.
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
	//
	// Note that it only includes the main goroutines to process the job.
	// Any possible daemon goroutines, auxiliary goroutines, or the goroutines
	// launched in the client's business functions are all excluded.
	NumGoroutine() int

	// Return the channel for the quit signal.
	// When the job is finished or quit, this channel will be closed
	// to broadcast the quit signal.
	QuitChan() <-chan struct{}

	// Return the panic records.
	PanicRecords() []PanicRec
}
