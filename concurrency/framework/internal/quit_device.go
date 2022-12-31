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

package internal

import (
	"github.com/donyori/gogo/concurrency"
	"github.com/donyori/gogo/errors"
)

// QuitDevice is an implementation of interface framework.QuitDevice.
type QuitDevice struct {
	oi concurrency.OnceIndicator // QuitChan() ~ oi.C(), IsQuit() ~ oi.Test(), Quit() ~ oi.Do(nil)
}

// NewQuitDevice creates a new quit device.
func NewQuitDevice() *QuitDevice {
	return &QuitDevice{oi: concurrency.NewOnceIndicator()}
}

// QuitChan returns the channel for the quit signal.
// When the job is finished or quit, this channel will be closed
// to broadcast the quit signal.
func (qd *QuitDevice) QuitChan() <-chan struct{} {
	if qd == nil {
		panic(errors.AutoMsg("*QuitDevice is nil"))
	}
	return qd.oi.C()
}

// IsQuit detects the quit signal on the quit channel.
// It returns true if a quit signal is detected, and false otherwise.
func (qd *QuitDevice) IsQuit() bool {
	if qd == nil {
		panic(errors.AutoMsg("*QuitDevice is nil"))
	}
	return qd.oi.Test()
}

// Quit broadcasts a quit signal to quit the job.
//
// This method will NOT wait until the job ends.
func (qd *QuitDevice) Quit() {
	if qd == nil {
		panic(errors.AutoMsg("*QuitDevice is nil"))
	}
	qd.oi.Do(nil)
}
