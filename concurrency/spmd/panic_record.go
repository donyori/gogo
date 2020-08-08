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
	"sync"
)

// Panic record, including the rank of the goroutine
// and the panic content (i.e., the parameter passed to function panic).
type PanicRec struct {
	Rank    int         // The rank of the goroutine.
	Content interface{} // The parameter passed to function panic.
}

func (pr PanicRec) String() string {
	if pr.Content == nil {
		return "no panic"
	}
	return fmt.Sprintf("panic on Goroutine %d: %v", pr.Rank, pr.Content)
}

func (pr PanicRec) Error() string {
	return pr.String()
}

// Panic records.
// It is safe for concurrent use by multiple goroutines.
type panicRecords struct {
	recs []PanicRec   // List of panic records.
	lock sync.RWMutex // Lock for concurrent use.
}

func (pr *panicRecords) Len() int {
	pr.lock.RLock()
	defer pr.lock.RUnlock()
	return len(pr.recs)
}

func (pr *panicRecords) List() []PanicRec {
	pr.lock.RLock()
	defer pr.lock.RUnlock()
	return append(pr.recs[:0:0], pr.recs...) // Return a copy of pr.recs.
}

func (pr *panicRecords) Append(panicRecs ...PanicRec) {
	pr.lock.Lock()
	defer pr.lock.Unlock()
	pr.recs = append(pr.recs, panicRecs...)
}
