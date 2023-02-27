// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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

import "github.com/donyori/gogo/errors"

// GoMap is a map wrapped on Go map.
// *GoMap implements the interface Map.
//
// The client can convert a Go map to GoMap by type conversion, e.g.:
//
//	GoMap[string, int](map[string]int{"A": 1})
//
// Or allocate a new GoMap by the map literal or
// the built-in function make, e.g.:
//
//	GoMap[string, int]{"A": 1}
//	make(GoMap[string, int])
type GoMap[Key comparable, Value any] map[Key]Value

var _ Map[string, any] = (*GoMap[string, any])(nil)

// Len returns the number of key-value pairs in the map.
//
// It returns 0 if the map is nil.
func (gm *GoMap[Key, Value]) Len() int {
	var n int
	if gm != nil {
		n = len(*gm)
	}
	return n
}

// Range accesses the key-value pairs in the map.
// Each key-value pair is accessed once.
// The order of the access is random.
//
// Its parameter handler is a function to deal with the key-value pair x
// in the map and report whether to continue to access the next key-value pair.
func (gm *GoMap[Key, Value]) Range(handler func(x Entry[Key, Value]) (cont bool)) {
	if gm != nil {
		for k, v := range *gm {
			if !handler(Entry[Key, Value]{Key: k, Value: v}) {
				return
			}
		}
	}
}

// Filter refines key-value pairs in the map.
//
// Its parameter filter is a function to report
// whether to keep the key-value pair x.
func (gm *GoMap[Key, Value]) Filter(filter func(x Entry[Key, Value]) (keep bool)) {
	if gm != nil {
		for k, v := range *gm {
			if !filter(Entry[Key, Value]{Key: k, Value: v}) {
				var zero Value
				(*gm)[k] = zero // avoid memory leak
				delete(*gm, k)
			}
		}
	}
}

// Get finds the value (if any) bound to the specified key
// and returns that value.
// If no value is found, it returns a zero value.
//
// It also returns an indicator present to report
// whether the value has been found.
func (gm *GoMap[Key, Value]) Get(key Key) (value Value, present bool) {
	if gm != nil {
		value, present = (*gm)[key]
	}
	return
}

// Set adds a new key-value pair to the map.
// Any existing mapping is overwritten.
//
// It panics if gm is nil.
func (gm *GoMap[Key, Value]) Set(key Key, value Value) {
	if gm == nil {
		panic(errors.AutoMsg(nilGoMapPointerPanicMessage))
	} else if *gm == nil {
		*gm = make(GoMap[Key, Value])
	}
	(*gm)[key] = value
}

// GetAndSet adds a new key-value pair to the map.
// Any existing mapping is overwritten.
//
// Unlike Set, GetAndSet returns the previous value (if any)
// bound to the key and an indicator present to report whether
// the key was present before calling GetAndSet.
//
// It panics if gm is nil.
func (gm *GoMap[Key, Value]) GetAndSet(key Key, value Value) (previous Value, present bool) {
	switch {
	case gm == nil:
		panic(errors.AutoMsg(nilGoMapPointerPanicMessage))
	case *gm != nil:
		previous, present = (*gm)[key]
	default:
		*gm = make(GoMap[Key, Value])
	}
	(*gm)[key] = value
	return
}

// SetMap adds the key-value pairs in m to this map.
// Any existing mapping is overwritten.
//
// It panics if m is not nil or empty and gm is nil.
func (gm *GoMap[Key, Value]) SetMap(m Map[Key, Value]) {
	if m == nil {
		return
	}
	n := m.Len()
	switch {
	case n == 0:
		return
	case gm == nil:
		panic(errors.AutoMsg(nilGoMapPointerPanicMessage))
	case *gm == nil:
		*gm = make(GoMap[Key, Value], n)
	}
	m.Range(func(entry Entry[Key, Value]) (cont bool) {
		(*gm)[entry.Key] = entry.Value
		return true
	})
}

// GetAndSetMap adds the key-value pairs in m to this map.
// Any existing mapping is overwritten.
//
// Unlike SetMap, GetAndSetMap returns the previous values (if any)
// bound to the keys in m in the form of Map.
// If m is nil or empty, or all keys in m are not present in this map,
// GetAndSetMap returns nil.
//
// It panics if m is not nil or empty and gm is nil.
func (gm *GoMap[Key, Value]) GetAndSetMap(m Map[Key, Value]) (previous Map[Key, Value]) {
	if m == nil {
		return
	}
	n := m.Len()
	switch {
	case n == 0:
		return
	case gm == nil:
		panic(errors.AutoMsg(nilGoMapPointerPanicMessage))
	case *gm == nil:
		*gm = make(GoMap[Key, Value], n)
		m.Range(func(entry Entry[Key, Value]) (cont bool) {
			(*gm)[entry.Key] = entry.Value
			return true
		})
		return
	}

	prev := new(GoMap[Key, Value])
	m.Range(func(entry Entry[Key, Value]) (cont bool) {
		v, ok := (*gm)[entry.Key]
		if ok {
			if *prev == nil {
				*prev = make(GoMap[Key, Value])
			}
			(*prev)[entry.Key] = v
		}
		(*gm)[entry.Key] = entry.Value
		return true
	})
	if len(*prev) > 0 {
		previous = prev
	}
	return
}

// Remove unmaps the specified keys from their values and
// removes the key-value pairs from the map.
//
// It does nothing for the keys that are not present in the map.
func (gm *GoMap[Key, Value]) Remove(key ...Key) {
	if gm != nil && len(*gm) > 0 {
		for _, k := range key {
			if _, ok := (*gm)[k]; ok {
				var zero Value
				(*gm)[k] = zero // avoid memory leak
				delete(*gm, k)
			}
		}
	}
}

// GetAndRemove unmaps the specified key from its value and
// removes the key-value pair from the map.
//
// Unlike Remove, GetAndRemove returns the previous value (if any)
// bound to the key and an indicator present to report whether
// the key was present before calling GetAndRemove.
func (gm *GoMap[Key, Value]) GetAndRemove(key Key) (previous Value, present bool) {
	if gm != nil && len(*gm) > 0 {
		previous, present = (*gm)[key]
		if present {
			var zero Value
			(*gm)[key] = zero // avoid memory leak
			delete(*gm, key)
		}
	}
	return
}

// Clear unmaps all keys from their values,
// removes all key-value pairs in the map,
// and asks to release the memory.
func (gm *GoMap[Key, Value]) Clear() {
	if gm != nil {
		*gm = nil
	}
}

const nilGoMapPointerPanicMessage = "*GoMap[...] is nil"
