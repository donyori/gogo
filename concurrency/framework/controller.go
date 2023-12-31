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

package framework

import "github.com/donyori/gogo/concurrency"

// Controller is a device to launch, cancel, and wait for the job.
//
// The use of all the frameworks under this package starts with creating
// a controller through their own New function.
type Controller interface {
	// Canceler returns a concurrency.Canceler for the job.
	//
	// Calls to Canceler always return the same non-nil value.
	Canceler() concurrency.Canceler

	// Launch starts the job.
	//
	// This method will not wait until the job ends.
	// Use method Wait if you want to wait for that.
	//
	// Note that Launch can take effect only once for one instance.
	// After the first call, subsequent calls to Launch do nothing.
	// To do the same job again, create a new Controller
	// with the same parameters.
	Launch()

	// Wait waits for the job to finish or cancel.
	// It returns the number of panic goroutines.
	//
	// If the job is not launched, it does nothing and returns -1.
	Wait() int

	// Run launches the job and waits for it.
	// It returns the number of panic goroutines.
	Run() int

	// NumGoroutine returns the number of goroutines to process this job.
	//
	// Note that it only includes the main goroutines to process the job.
	// Any possible control goroutines, daemon goroutines, auxiliary goroutines,
	// or the goroutines launched in the client's business functions
	// are all excluded.
	NumGoroutine() int

	// PanicRecords returns a list of the panic records.
	PanicRecords() []PanicRecord
}
