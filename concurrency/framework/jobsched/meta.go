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

package jobsched

import "time"

// NoProperty is an alias for struct{}.
//
// It is used to instantiate Meta without custom properties:
//
//	Meta[NoProperty]
type NoProperty = struct{}

// Meta contains the meta information of a job,
// including its priority, creation time, and any custom properties.
type Meta[Properties any] struct {
	// The priority of the job, a non-negative integer.
	//
	// The greater the value, the higher the priority.
	// The default value 0 corresponds to the lowest priority.
	Priority uint

	// The creation time of the job.
	//
	// A zero-value CreationTime will be set to time.Now() by the framework
	// before adding to a job queue.
	CreationTime time.Time

	// Custom properties, used to customize job scheduling algorithm.
	//
	// If no custom property is required, set the type parameter to NoProperty.
	Custom Properties
}

// MetaJob combines the job and its meta information.
type MetaJob[Job, Properties any] struct {
	Meta Meta[Properties]
	Job  Job
}
