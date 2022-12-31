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

package runtime_test

import (
	"testing"

	"github.com/donyori/gogo/runtime"
)

type ExportedStruct struct{}

func (es *ExportedStruct) Foo() (pkg, fn string) {
	pkg, fn, _ = runtime.CallerPkgFunc(0)
	return pkg, fn
}

func (es *ExportedStruct) foo() (pkg, fn string) {
	pkg, fn, _ = runtime.CallerPkgFunc(0)
	return pkg, fn
}

type localStruct struct{}

func (ls *localStruct) Foo() (pkg, fn string) {
	pkg, fn, _ = runtime.CallerPkgFunc(0)
	return pkg, fn
}

func (ls *localStruct) foo() (pkg, fn string) {
	pkg, fn, _ = runtime.CallerPkgFunc(0)
	return pkg, fn
}

var globalTagPkgFuncs [][3]string

func init() {
	pkg, fn, _ := runtime.CallerPkgFunc(0)
	globalTagPkgFuncs = append(globalTagPkgFuncs, [3]string{"init.0", pkg, fn})
}

func init() {
	pkg, fn, _ := runtime.CallerPkgFunc(0)
	globalTagPkgFuncs = append(globalTagPkgFuncs, [3]string{"init.1", pkg, fn})
}

type callerPkgFuncRecord struct {
	expFn string
	pkg   string
	fn    string
}

func TestCallerPkgFunc(t *testing.T) {
	var records []callerPkgFuncRecord
	for _, elem := range globalTagPkgFuncs {
		records = append(records, callerPkgFuncRecord{elem[0], elem[1], elem[2]})
	}
	pkg, fn, _ := runtime.CallerPkgFunc(0)
	records = append(records, callerPkgFuncRecord{"TestCallerPkgFunc", pkg, fn})
	func() {
		defer func() {
			pkg, fn, _ := runtime.CallerPkgFunc(0)
			records = append(records, callerPkgFuncRecord{"TestCallerPkgFunc.func1.1", pkg, fn})
		}()
		pkg, fn, _ := runtime.CallerPkgFunc(0)
		records = append(records, callerPkgFuncRecord{"TestCallerPkgFunc.func1", pkg, fn})
	}()
	tes := new(ExportedStruct)
	pkg, fn = tes.Foo()
	records = append(records, callerPkgFuncRecord{"(*ExportedStruct).Foo", pkg, fn})
	pkg, fn = tes.foo()
	records = append(records, callerPkgFuncRecord{"(*ExportedStruct).foo", pkg, fn})
	tls := new(localStruct)
	pkg, fn = tls.Foo()
	records = append(records, callerPkgFuncRecord{"(*localStruct).Foo", pkg, fn})
	pkg, fn = tls.foo()
	records = append(records, callerPkgFuncRecord{"(*localStruct).foo", pkg, fn})

	for _, rec := range records {
		if rec.pkg != "github.com/donyori/gogo/runtime_test" || rec.fn != rec.expFn {
			t.Errorf("got pkg: %s, fn: %s\nwant pkg: github.com/donyori/gogo/runtime_test, fn: %s", rec.pkg, rec.fn, rec.expFn)
		}
	}
}
