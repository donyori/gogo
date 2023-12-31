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

package spmd

import "github.com/donyori/gogo/concurrency/framework"

// Export for testing only.

type ControllerImpl[Message any] struct {
	*controller[Message]
}

// WrapController wraps a framework.Controller returned by
// github.com/donyori/gogo/concurrency/framework/spmd.New
// to access its unexported fields for testing.
//
// It panics if ctrl is not the result of
// github.com/donyori/gogo/concurrency/framework/spmd.New.
func WrapController[Message any](ctrl framework.Controller) *ControllerImpl[Message] {
	if ctrl == nil {
		return nil
	}
	return &ControllerImpl[Message]{ctrl.(*controller[Message])}
}

func (ctrl *ControllerImpl[Message]) GetLnchCommMaps() []map[string]Communicator[Message] {
	if ctrl == nil {
		return nil
	}
	return ctrl.lnchCommMaps
}

func (ctrl *ControllerImpl[Message]) GetWorldBcastMapLen() int {
	if ctrl == nil {
		return 0
	}
	return len(ctrl.world.bcastMap)
}

func (ctrl *ControllerImpl[Message]) GetWorldScatterMapLen() int {
	if ctrl == nil {
		return 0
	}
	return len(ctrl.world.scatterMap)
}

func (ctrl *ControllerImpl[Message]) GetWorldGatherMapLen() int {
	if ctrl == nil {
		return 0
	}
	return len(ctrl.world.gatherMap)
}
