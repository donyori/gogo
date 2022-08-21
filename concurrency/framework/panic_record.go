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

package framework

import (
	"fmt"
	"sync"
)

// PanicRecord is a panic record, including the name of the goroutine
// and the panic content (i.e., the argument passed to function panic).
type PanicRecord struct {
	Name    string // Name of the goroutine.
	Content any    // The argument passed to function panic.
}

// Error formats the panic record into a string
// and reports it as an error message.
func (pr PanicRecord) Error() string {
	if pr.Content == nil {
		return "no panic"
	}
	return fmt.Sprintf("panic on goroutine %s: %v", pr.Name, pr.Content)
}

// PanicRecords are panic records, used by the framework codes.
//
// It is safe for concurrent use by multiple goroutines.
type PanicRecords struct {
	recs []PanicRecord // List of panic records.
	lock sync.RWMutex  // Lock for concurrent use.
}

// Len returns the number of records.
func (pr *PanicRecords) Len() int {
	pr.lock.RLock()
	defer pr.lock.RUnlock()
	return len(pr.recs)
}

// List copies and returns the panic records as a slice of PanicRecord.
// It returns nil if there is no panic record.
func (pr *PanicRecords) List() []PanicRecord {
	pr.lock.RLock()
	defer pr.lock.RUnlock()
	if len(pr.recs) == 0 {
		return nil
	}
	recs := make([]PanicRecord, len(pr.recs))
	copy(recs, pr.recs)
	return recs
}

// Append adds new panic records to the back of its panic record list.
func (pr *PanicRecords) Append(panicRec ...PanicRecord) {
	if len(panicRec) == 0 {
		return
	}
	pr.lock.Lock()
	defer pr.lock.Unlock()
	pr.recs = append(pr.recs, panicRec...)
}
