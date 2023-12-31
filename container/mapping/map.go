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

package mapping

import (
	"fmt"

	"github.com/donyori/gogo/container"
)

// Entry represents an entry of Map.
// It is a key-value pair.
type Entry[Key, Value any] struct {
	Key   Key
	Value Value
}

// String formats the entry in the form of
//
//	<key> ": " <value>
//
// For example, the result of
//
//	Entry[string, int]{Key: "A", Value: 1}.String()
//
// is "A: 1".
//
// It is equivalent to fmt.Sprintf("%v: %v", entry.Key, entry.Value).
func (entry Entry[Key, Value]) String() string {
	return fmt.Sprintf("%v: %v", entry.Key, entry.Value)
}

// Map is an interface representing a map
// (also known as an associative array, symbol table, or dictionary).
//
// Note that the entries passed to the methods Range and Filter
// are only used to carry key-value pairs to the client and
// are not owned by the map.
// The modifications to the entries should not affect the map.
type Map[Key, Value any] interface {
	container.Container[Entry[Key, Value]]
	container.Filter[Entry[Key, Value]]

	// Get finds the value (if any) bound to the specified key
	// and returns that value.
	// If no value is found, it returns a zero value.
	//
	// It also returns an indicator present to report
	// whether the value has been found.
	Get(key Key) (value Value, present bool)

	// Set adds a new key-value pair to the map.
	// Any existing mapping is overwritten.
	Set(key Key, value Value)

	// GetAndSet adds a new key-value pair to the map.
	// Any existing mapping is overwritten.
	//
	// Unlike Set, GetAndSet returns the previous value (if any)
	// bound to the key and an indicator present to report whether
	// the key was present before calling GetAndSet.
	GetAndSet(key Key, value Value) (previous Value, present bool)

	// SetMap adds the key-value pairs in m to this map.
	// Any existing mapping is overwritten.
	SetMap(m Map[Key, Value])

	// GetAndSetMap adds the key-value pairs in m to this map.
	// Any existing mapping is overwritten.
	//
	// Unlike SetMap, GetAndSetMap returns the previous values (if any)
	// bound to the keys in m in the form of Map.
	// If m is nil or empty, or all keys in m are not present in this map,
	// GetAndSetMap returns nil.
	GetAndSetMap(m Map[Key, Value]) (previous Map[Key, Value])

	// Remove unmaps the specified keys from their values and
	// removes the key-value pairs from the map.
	//
	// It does nothing for the keys that are not present in the map.
	Remove(key ...Key)

	// GetAndRemove unmaps the specified key from its value and
	// removes the key-value pair from the map.
	//
	// Unlike Remove, GetAndRemove returns the previous value (if any)
	// bound to the key and an indicator present to report whether
	// the key was present before calling GetAndRemove.
	GetAndRemove(key Key) (previous Value, present bool)

	// Clear unmaps all keys from their values,
	// removes all key-value pairs in the map,
	// and asks to release the memory.
	Clear()
}
