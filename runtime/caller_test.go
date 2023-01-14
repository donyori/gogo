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
	"github.com/donyori/gogo/runtime/internal/testing/dotpkg.v1"
	"github.com/donyori/gogo/runtime/internal/testing/dotpkg.v1/subpkg"
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
	wantPkg string
	wantFn  string
	pkg     string
	fn      string
}

func TestCallerPkgFunc(t *testing.T) {
	const WantPkg = "github.com/donyori/gogo/runtime_test"

	var records []callerPkgFuncRecord
	for _, elem := range globalTagPkgFuncs {
		records = append(records, callerPkgFuncRecord{
			wantPkg: WantPkg,
			wantFn:  elem[0],
			pkg:     elem[1],
			fn:      elem[2],
		})
	}
	pkg, fn, _ := runtime.CallerPkgFunc(0)
	records = append(records, callerPkgFuncRecord{
		wantPkg: WantPkg,
		wantFn:  "TestCallerPkgFunc",
		pkg:     pkg,
		fn:      fn,
	})
	func() {
		defer func() {
			pkg, fn, _ := runtime.CallerPkgFunc(0)
			records = append(records, callerPkgFuncRecord{
				wantPkg: WantPkg,
				wantFn:  "TestCallerPkgFunc.func1.1",
				pkg:     pkg,
				fn:      fn,
			})
		}()
		pkg, fn, _ := runtime.CallerPkgFunc(0)
		records = append(records, callerPkgFuncRecord{
			wantPkg: WantPkg,
			wantFn:  "TestCallerPkgFunc.func1",
			pkg:     pkg,
			fn:      fn,
		})
	}()
	tes := new(ExportedStruct)
	pkg, fn = tes.Foo()
	records = append(records, callerPkgFuncRecord{
		wantPkg: WantPkg,
		wantFn:  "(*ExportedStruct).Foo",
		pkg:     pkg,
		fn:      fn,
	})
	pkg, fn = tes.foo()
	records = append(records, callerPkgFuncRecord{
		wantPkg: WantPkg,
		wantFn:  "(*ExportedStruct).foo",
		pkg:     pkg,
		fn:      fn,
	})
	tls := new(localStruct)
	pkg, fn = tls.Foo()
	records = append(records, callerPkgFuncRecord{
		wantPkg: WantPkg,
		wantFn:  "(*localStruct).Foo",
		pkg:     pkg,
		fn:      fn,
	})
	pkg, fn = tls.foo()
	records = append(records, callerPkgFuncRecord{
		wantPkg: WantPkg,
		wantFn:  "(*localStruct).foo",
		pkg:     pkg,
		fn:      fn,
	})

	for _, rec := range records {
		if rec.pkg != rec.wantPkg || rec.fn != rec.wantFn {
			t.Errorf("got pkg: %s, fn: %s; want pkg: %s, fn: %s", rec.pkg, rec.fn, rec.wantPkg, rec.wantFn)
		}
	}
}

func TestCallerPkgFunc_DotPkg(t *testing.T) {
	var records []callerPkgFuncRecord
	dotpkg.Do(func() {
		pkg, fn, _ := runtime.CallerPkgFunc(1)
		records = append(records, callerPkgFuncRecord{
			wantPkg: "github.com/donyori/gogo/runtime/internal/testing/dotpkg%2ev1",
			wantFn:  "Do",
			pkg:     pkg,
			fn:      fn,
		})
	})
	subpkg.Do(func() {
		pkg, fn, _ := runtime.CallerPkgFunc(1)
		records = append(records, callerPkgFuncRecord{
			wantPkg: "github.com/donyori/gogo/runtime/internal/testing/dotpkg.v1/subpkg",
			wantFn:  "Do",
			pkg:     pkg,
			fn:      fn,
		})
	})

	for _, rec := range records {
		if rec.pkg != rec.wantPkg || rec.fn != rec.wantFn {
			t.Errorf("got pkg: %s, fn: %s; want pkg: %s, fn: %s", rec.pkg, rec.fn, rec.wantPkg, rec.wantFn)
		}
	}
}
