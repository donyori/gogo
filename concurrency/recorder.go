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

package concurrency

import "sync"

// RecordReader is a device to read messages recorded by a Recorder.
type RecordReader[Message any] interface {
	// Len returns the number of messages recorded.
	Len() int

	// Last returns the last recorded message.
	// It also returns an indicator to report whether the message exists.
	Last() (x Message, ok bool)

	// All returns all messages in recording order.
	All() []Message
}

// Recorder is a device to record messages.
//
// Recorder can be used to log information in some goroutines
// and retrieve it in other goroutines later.
//
// Recorder does not allow removing messages recorded.
type Recorder[Message any] interface {
	RecordReader[Message]

	// Reader returns a RecordReader that
	// reads messages recorded by this recorder.
	//
	// The returned RecordReader should not be able to
	// be converted to a Recorder and should not provide
	// any way to record messages or obtain the Recorder.
	Reader() RecordReader[Message]

	// Record logs the messages into this recorder.
	Record(x ...Message)
}

// NewRecorder creates a new Recorder.
//
// capacity is the number of messages the Recorder can hold initially.
// If capacity is nonpositive, the Recorder does not reserve initial space.
func NewRecorder[Message any](capacity int) Recorder[Message] {
	rec := new(recorder[Message])
	if capacity > 0 {
		rec.msgList = make([]Message, 0, capacity)
	}
	return rec
}

// recorder is an implementation of interface Recorder.
type recorder[Message any] struct {
	msgList []Message    // List of messages.
	lock    sync.RWMutex // Lock to protect msgList.
}

func (rec *recorder[Message]) Len() int {
	rec.lock.RLock()
	defer rec.lock.RUnlock()
	return len(rec.msgList)
}

func (rec *recorder[Message]) Last() (x Message, ok bool) {
	rec.lock.RLock()
	defer rec.lock.RUnlock()
	if len(rec.msgList) > 0 {
		x, ok = rec.msgList[len(rec.msgList)-1], true
	}
	return
}

func (rec *recorder[Message]) All() []Message {
	rec.lock.RLock()
	defer rec.lock.RUnlock()
	var ms []Message
	if len(rec.msgList) > 0 {
		ms = make([]Message, len(rec.msgList))
		copy(ms, rec.msgList)
	}
	return ms
}

func (rec *recorder[Message]) Reader() RecordReader[Message] {
	return &recordReader[Message]{rec: rec}
}

func (rec *recorder[Message]) Record(x ...Message) {
	if len(x) == 0 {
		return
	}
	rec.lock.Lock()
	defer rec.lock.Unlock()
	rec.msgList = append(rec.msgList, x...)
}

// recordReader is an implementation of interface RecordReader.
type recordReader[Message any] struct {
	rec *recorder[Message]
}

func (r *recordReader[Message]) Len() int {
	return r.rec.Len()
}

func (r *recordReader[Message]) Last() (x Message, ok bool) {
	return r.rec.Last()
}

func (r *recordReader[Message]) All() []Message {
	return r.rec.All()
}
