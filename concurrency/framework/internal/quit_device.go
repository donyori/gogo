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

package internal

import "github.com/donyori/gogo/concurrency"

// An implementation of
// github.com/donyori/gogo/concurrency/framework.QuitDevice.
type QuitDevice struct {
	Oi concurrency.OnceIndicator // QuitChan() ~ Oi.C(), IsQuit() ~ Oi.Test(), Quit() ~ Oi.Do(nil)
}

// Create a new quit device.
func NewQuitDevice() *QuitDevice {
	return &QuitDevice{concurrency.NewOnceIndicator()}
}

func (qd *QuitDevice) QuitChan() <-chan struct{} {
	return qd.Oi.C()
}

func (qd *QuitDevice) IsQuit() bool {
	return qd.Oi.Test()
}

func (qd *QuitDevice) Quit() {
	qd.Oi.Do(nil)
}
