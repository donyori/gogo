// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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

// Package spmd provides a basic framework for SPMD
// (single program, multiple data) style programming.
//
// To use this framework, you should start with the function New to create
// a controller and custom your business function biz.
// And then call the Launch method to run the job.
// Finally, call the Wait method to wait for the job to finish.
// Or simply, you can start with the function Run, which combines New, Launch,
// and Wait together.
//
// In the business function biz,
// you can communicate with other goroutines via Communicator.
package spmd
