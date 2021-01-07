// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

package runtime

import "testing"

type TestExportedStruct struct{}

func (tes *TestExportedStruct) Foo() (pkg, fn string) {
	pkg, fn, _ = CallerPkgFunc(0)
	return pkg, fn
}

func (tes *TestExportedStruct) foo() (pkg, fn string) {
	pkg, fn, _ = CallerPkgFunc(0)
	return pkg, fn
}

type testLocalStruct struct{}

func (tls *testLocalStruct) Foo() (pkg, fn string) {
	pkg, fn, _ = CallerPkgFunc(0)
	return pkg, fn
}

func (tls *testLocalStruct) foo() (pkg, fn string) {
	pkg, fn, _ = CallerPkgFunc(0)
	return pkg, fn
}

var testGlobalTagPkgFuncs [][3]string

func init() {
	pkg, fn, _ := CallerPkgFunc(0)
	testGlobalTagPkgFuncs = append(testGlobalTagPkgFuncs, [3]string{"init-1", pkg, fn})
}

func init() {
	pkg, fn, _ := CallerPkgFunc(0)
	testGlobalTagPkgFuncs = append(testGlobalTagPkgFuncs, [3]string{"init-2", pkg, fn})
}

func _TestShowCallerPkgFunc(t *testing.T) {
	var records []struct {
		tag string
		pkg string
		fn  string
	}
	for _, elem := range testGlobalTagPkgFuncs {
		records = append(records, struct {
			tag string
			pkg string
			fn  string
		}{tag: elem[0], pkg: elem[1], fn: elem[2]})
	}
	pkg, fn, _ := CallerPkgFunc(0)
	records = append(records, struct {
		tag string
		pkg string
		fn  string
	}{tag: "TestFunc", pkg: pkg, fn: fn})
	func() {
		defer func() {
			pkg, fn, _ := CallerPkgFunc(0)
			records = append(records, struct {
				tag string
				pkg string
				fn  string
			}{tag: "inner-2", pkg: pkg, fn: fn})
		}()
		pkg, fn, _ := CallerPkgFunc(0)
		records = append(records, struct {
			tag string
			pkg string
			fn  string
		}{tag: "inner-1", pkg: pkg, fn: fn})
	}()
	tes := new(TestExportedStruct)
	pkg, fn = tes.Foo()
	records = append(records, struct {
		tag string
		pkg string
		fn  string
	}{tag: "TES.Foo", pkg: pkg, fn: fn})
	pkg, fn = tes.foo()
	records = append(records, struct {
		tag string
		pkg string
		fn  string
	}{tag: "TES.foo", pkg: pkg, fn: fn})
	tls := new(testLocalStruct)
	pkg, fn = tls.Foo()
	records = append(records, struct {
		tag string
		pkg string
		fn  string
	}{tag: "TLS.Foo", pkg: pkg, fn: fn})
	pkg, fn = tls.foo()
	records = append(records, struct {
		tag string
		pkg string
		fn  string
	}{tag: "TLS.foo", pkg: pkg, fn: fn})

	for i := range records {
		t.Logf("Record %d: %+v.", i, records[i])
	}
}
