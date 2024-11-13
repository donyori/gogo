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

package mapping_test

import (
	"fmt"
	"maps"
	"testing"

	"github.com/donyori/gogo/container/mapping"
	"github.com/donyori/gogo/fmtcoll"
)

type (
	SIM  = map[string]int
	SIGM = mapping.GoMap[string, int]
)

func TestGoMap_Len(t *testing.T) {
	testCases := []struct {
		gm   *SIGM
		want int
	}{
		{nil, 0},
		{new(SIGM), 0},
		{&SIGM{}, 0},
		{&SIGM{"A": 1}, 1},
	}

	for _, tc := range testCases {
		t.Run("gm="+gmPtrToName(tc.gm), func(t *testing.T) {
			if n := tc.gm.Len(); n != tc.want {
				t.Errorf("got %d; want %d", n, tc.want)
			}
		})
	}
}

func TestGoMap_Range(t *testing.T) {
	gm := SIGM{"A": 1, "B": 2, "C": 3, "a": -1, "b": -2, "c": -3}
	want := SIM{"A": 1, "B": 2, "C": 3, "a": -1, "b": -2, "c": -3}
	m := make(SIM, len(gm))
	gm.Range(func(x mapping.Entry[string, int]) (cont bool) {
		m[x.Key] = x.Value
		return true
	})
	if mapWrong(m, want) {
		t.Errorf("got %s; want %s", mapToString(m), mapToString(want))
	}
}

func TestGoMap_Range_NilAndEmpty(t *testing.T) {
	gms := []*SIGM{nil, new(SIGM), {}}
	for _, gm := range gms {
		t.Run("gm="+gmPtrToName(gm), func(t *testing.T) {
			gm.Range(func(x mapping.Entry[string, int]) (cont bool) {
				t.Error("handler was called, x:", x)
				return true
			})
		})
	}
}

func TestGoMap_Clear(t *testing.T) {
	dataList := []SIM{
		nil,
		{},
		{"A": 1},
		{"A": 1, "B": 2},
		{"A": 1, "B": 2, "C": 3},
	}
	for _, data := range dataList {
		m := maps.Clone(data)
		gm := (*SIGM)(&m)
		t.Run("gm="+gmPtrToName(gm), func(t *testing.T) {
			gm.Clear()
			if gm == nil || *gm != nil {
				t.Errorf("got %s; want <nil>", gmPtrToName(gm))
			}
		})
	}

	var nilGM *SIGM
	t.Run("gm="+gmPtrToName(nilGM), func(t *testing.T) {
		nilGM.Clear()
		if nilGM != nil {
			t.Errorf("got %s; want <nil>", gmPtrToName(nilGM))
		}
	})
}

func TestGoMap_RemoveAll(t *testing.T) {
	dataList := []SIM{
		nil,
		{},
		{"A": 1},
		{"A": 1, "B": 2},
		{"A": 1, "B": 2, "C": 3},
	}
	for _, data := range dataList {
		m := maps.Clone(data)
		gm := (*SIGM)(&m)
		t.Run("gm="+gmPtrToName(gm), func(t *testing.T) {
			gm.RemoveAll()
			if m != nil {
				if gm == nil || *gm == nil || len(*gm) != 0 {
					t.Errorf("got %s; want {}", gmPtrToName(gm))
				}
			} else if gm == nil || *gm != nil {
				t.Errorf("got %s; want <nil>", gmPtrToName(gm))
			}
		})
	}

	var nilGM *SIGM
	t.Run("gm="+gmPtrToName(nilGM), func(t *testing.T) {
		nilGM.RemoveAll()
		if nilGM != nil {
			t.Errorf("got %s; want <nil>", gmPtrToName(nilGM))
		}
	})
}

func TestGoMap_Filter(t *testing.T) {
	filterList := []func(x mapping.Entry[string, int]) (keep bool){
		func(x mapping.Entry[string, int]) (keep bool) {
			return len(x.Key) == 1 && x.Key[0] >= 'A' && x.Key[0] <= 'Z'
		},
		func(x mapping.Entry[string, int]) (keep bool) {
			return x.Value < 0
		},
	}
	data := SIM{"A": 1, "B": 2, "C": 3, "a": -1, "b": -2, "c": -3}
	testCases := []struct {
		filterIdx int
		want      SIM
	}{
		{0, SIM{"A": 1, "B": 2, "C": 3}},
		{1, SIM{"a": -1, "b": -2, "c": -3}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("filterIdx=%d", tc.filterIdx), func(t *testing.T) {
			gm := SIGM(maps.Clone(data))
			gm.Filter(filterList[tc.filterIdx])
			if goMapWrong(&gm, tc.want) {
				t.Errorf("got %s; want %s",
					mapToString(gm), mapToString(tc.want))
			}
		})
	}
}

func TestGoMap_Filter_NilAndEmpty(t *testing.T) {
	gms := []*SIGM{nil, new(SIGM), {}}
	for _, gm := range gms {
		t.Run("gm="+gmPtrToName(gm), func(t *testing.T) {
			gm.Filter(func(x mapping.Entry[string, int]) (keep bool) {
				t.Error("handler was called, x:", x)
				return true
			})
		})
	}
}

func TestGoMap_Get(t *testing.T) {
	data := SIM{"A": 1, "B": 2, "C": 3, "a": -1, "b": -2, "c": -3}
	const Absent = "absent"
	testCases := []struct {
		data  SIM // if data is nil, gm is (*SIGM)(nil)
		k     string
		wantV int
		wantP bool
	}{
		{nil, "", 0, false},
		{nil, "A", 0, false},
		{nil, "B", 0, false},
		{nil, "C", 0, false},
		{nil, "a", 0, false},
		{nil, "b", 0, false},
		{nil, "c", 0, false},
		{nil, Absent, 0, false},

		{data, "", 0, false},
		{data, "A", 1, true},
		{data, "B", 2, true},
		{data, "C", 3, true},
		{data, "a", -1, true},
		{data, "b", -2, true},
		{data, "c", -3, true},
		{data, Absent, 0, false},
	}

	for _, tc := range testCases {
		var gm *SIGM
		if tc.data != nil {
			m := maps.Clone(tc.data)
			gm = (*SIGM)(&m)
		}
		t.Run(
			fmt.Sprintf("gm=%s&key=%+q", gmPtrToName(gm), tc.k),
			func(t *testing.T) {
				v, p := gm.Get(tc.k)
				if v != tc.wantV || p != tc.wantP {
					t.Errorf("got (%d, %t); want (%d, %t)",
						v, p, tc.wantV, tc.wantP)
				}
			},
		)
	}
}

func TestGoMap_Set(t *testing.T) {
	data := SIM{"A": 1, "B": 2, "C": 3}
	testCases := []struct {
		data SIM // if data is nil, gm is new(SIGM)
		k    string
		v    int
		want SIM
	}{
		{nil, "A", 2, SIM{"A": 2}},
		{nil, "B", 3, SIM{"B": 3}},
		{nil, "C", 4, SIM{"C": 4}},
		{nil, "a", -1, SIM{"a": -1}},

		{data, "A", 2, SIM{"A": 2, "B": 2, "C": 3}},
		{data, "B", 3, SIM{"A": 1, "B": 3, "C": 3}},
		{data, "C", 4, SIM{"A": 1, "B": 2, "C": 4}},
		{data, "a", -1, SIM{"A": 1, "B": 2, "C": 3, "a": -1}},
	}

	for _, tc := range testCases {
		var gm *SIGM
		if tc.data != nil {
			m := maps.Clone(tc.data)
			gm = (*SIGM)(&m)
		} else {
			gm = new(SIGM)
		}
		t.Run(
			fmt.Sprintf("gm=%s&key=%+q&value=%d", gmPtrToName(gm), tc.k, tc.v),
			func(t *testing.T) {
				gm.Set(tc.k, tc.v)
				if goMapWrong(gm, tc.want) {
					t.Errorf("got %s; want %s",
						gmPtrToName(gm), mapToString(tc.want))
				}
			},
		)
	}
}

func TestGoMap_Set_Panic(t *testing.T) {
	defer func() {
		_ = recover()
	}()
	var gm *SIGM
	gm.Set("A", 1) // want panic here
	t.Error("want panic but not")
}

func TestGoMap_GetAndSet(t *testing.T) {
	data := SIM{"A": 1, "B": 2, "C": 3}
	testCases := []struct {
		data  SIM // if data is nil, gm is new(SIGM)
		k     string
		v     int
		wantM SIM
		wantV int
		wantP bool
	}{
		{nil, "A", 2, SIM{"A": 2}, 0, false},
		{nil, "B", 3, SIM{"B": 3}, 0, false},
		{nil, "C", 4, SIM{"C": 4}, 0, false},
		{nil, "a", -1, SIM{"a": -1}, 0, false},

		{data, "A", 2, SIM{"A": 2, "B": 2, "C": 3}, 1, true},
		{data, "B", 3, SIM{"A": 1, "B": 3, "C": 3}, 2, true},
		{data, "C", 4, SIM{"A": 1, "B": 2, "C": 4}, 3, true},
		{data, "a", -1, SIM{"A": 1, "B": 2, "C": 3, "a": -1}, 0, false},
	}

	for _, tc := range testCases {
		var gm *SIGM
		if tc.data != nil {
			m := maps.Clone(tc.data)
			gm = (*SIGM)(&m)
		} else {
			gm = new(SIGM)
		}
		t.Run(
			fmt.Sprintf("gm=%s&key=%+q&value=%d", gmPtrToName(gm), tc.k, tc.v),
			func(t *testing.T) {
				v, p := gm.GetAndSet(tc.k, tc.v)
				if goMapWrong(gm, tc.wantM) {
					t.Errorf("got map %s; want %s",
						gmPtrToName(gm), mapToString(tc.wantM))
				}
				if v != tc.wantV || p != tc.wantP {
					t.Errorf("got (%d, %t); want (%d, %t)",
						v, p, tc.wantV, tc.wantP)
				}
			},
		)
	}
}

func TestGoMap_GetAndSet_Panic(t *testing.T) {
	defer func() {
		_ = recover()
	}()
	var gm *SIGM
	v, p := gm.GetAndSet("A", 1) // want panic here
	t.Errorf("want panic but got (%d, %t)", v, p)
}

func TestGoMap_SetMap(t *testing.T) {
	data := SIM{"A": 1, "B": 2, "C": 3}
	m1 := new(SIGM)
	m2 := &SIGM{"A": 2, "B": 3}
	m3 := &SIGM{"B": 3, "D": 4, "a": -1}
	testCases := []struct {
		data SIM // if data is nil, gm is new(SIGM)
		m    *SIGM
		want SIM
	}{
		{nil, nil, SIM{}},
		{nil, m1, SIM{}},
		{nil, m2, SIM{"A": 2, "B": 3}},
		{nil, m3, SIM{"B": 3, "D": 4, "a": -1}},

		{data, nil, SIM{"A": 1, "B": 2, "C": 3}},
		{data, m1, SIM{"A": 1, "B": 2, "C": 3}},
		{data, m2, SIM{"A": 2, "B": 3, "C": 3}},
		{data, m3, SIM{"A": 1, "B": 3, "C": 3, "D": 4, "a": -1}},
	}

	for _, tc := range testCases {
		var gm *SIGM
		if tc.data != nil {
			m := maps.Clone(tc.data)
			gm = (*SIGM)(&m)
		} else {
			gm = new(SIGM)
		}
		t.Run(
			fmt.Sprintf("gm=%s&m=%s", gmPtrToName(gm), gmPtrToName(tc.m)),
			func(t *testing.T) {
				gm.SetMap(tc.m)
				if goMapWrong(gm, tc.want) {
					t.Errorf("got %s; want %s",
						gmPtrToName(gm), mapToString(tc.want))
				}
			},
		)
	}
}

func TestGoMap_SetMap_Panic(t *testing.T) {
	m1 := new(SIGM)
	m2 := &SIGM{"A": 2, "B": 3}
	testCases := []struct {
		m         *SIGM
		wantPanic bool
	}{
		{nil, false},
		{m1, false},
		{m2, true},
	}

	for _, tc := range testCases {
		var gm *SIGM
		t.Run(
			fmt.Sprintf("gm=%s&m=%s", gmPtrToName(gm), gmPtrToName(tc.m)),
			func(t *testing.T) {
				defer func() {
					e := recover()
					if tc.wantPanic {
						if e == nil {
							t.Error("want panic but not")
						}
					} else if e != nil {
						t.Error("panic -", e)
					}
				}()
				gm.SetMap(tc.m)
			},
		)
	}
}

func TestGoMap_GetAndSetMap(t *testing.T) {
	data := SIM{"A": 1, "B": 2, "C": 3}
	m1 := new(SIGM)
	m2 := &SIGM{"A": 2, "B": 3}
	m3 := &SIGM{"B": 3, "D": 4, "a": -1}
	testCases := []struct {
		data  SIM // if data is nil, gm is new(SIGM)
		m     *SIGM
		wantM SIM
		wantP SIM
	}{
		{nil, nil, SIM{}, nil},
		{nil, m1, SIM{}, nil},
		{nil, m2, SIM{"A": 2, "B": 3}, nil},
		{nil, m3, SIM{"B": 3, "D": 4, "a": -1}, nil},

		{data, nil, SIM{"A": 1, "B": 2, "C": 3}, nil},
		{data, m1, SIM{"A": 1, "B": 2, "C": 3}, nil},
		{data, m2, SIM{"A": 2, "B": 3, "C": 3}, SIM{"A": 1, "B": 2}},
		{data, m3, SIM{"A": 1, "B": 3, "C": 3, "D": 4, "a": -1}, SIM{"B": 2}},
	}

	for _, tc := range testCases {
		var gm *SIGM
		if tc.data != nil {
			m := maps.Clone(tc.data)
			gm = (*SIGM)(&m)
		} else {
			gm = new(SIGM)
		}
		t.Run(
			fmt.Sprintf("gm=%s&m=%s", gmPtrToName(gm), gmPtrToName(tc.m)),
			func(t *testing.T) {
				p := gm.GetAndSetMap(tc.m)
				if goMapWrong(gm, tc.wantM) {
					t.Errorf("got map %s; want %s",
						gmPtrToName(gm), mapToString(tc.wantM))
				}
				if mapItfWrong(p, tc.wantP) {
					t.Errorf("got %s; want %s",
						mapItfToString(p), mapToString(tc.wantP))
				}
			},
		)
	}
}

func TestGoMap_GetAndSetMap_Panic(t *testing.T) {
	m1 := new(SIGM)
	m2 := &SIGM{"A": 2, "B": 3}
	testCases := []struct {
		m         *SIGM
		wantPanic bool
	}{
		{nil, false},
		{m1, false},
		{m2, true},
	}

	for _, tc := range testCases {
		var gm *SIGM
		t.Run(
			fmt.Sprintf("gm=%s&m=%s", gmPtrToName(gm), gmPtrToName(tc.m)),
			func(t *testing.T) {
				defer func() {
					if e := recover(); !tc.wantPanic && e != nil {
						t.Error("panic -", e)
					}
				}()
				p := gm.GetAndSetMap(tc.m)
				if tc.wantPanic {
					t.Error("want panic but got", mapItfToString(p))
				} else if p != nil {
					t.Errorf("got %s; want <nil>", mapItfToString(p))
				}
			},
		)
	}
}

func TestGoMap_Remove(t *testing.T) {
	data := SIM{"A": 1, "B": 2, "C": 3}
	testCases := []struct {
		data SIM // if data is nil, gm is (*SIGM)(nil)
		ks   []string
		want SIM
	}{
		{nil, nil, nil},
		{nil, []string{}, nil},
		{nil, []string{""}, nil},
		{nil, []string{"A"}, nil},
		{nil, []string{"a"}, nil},
		{nil, []string{"", "A", "D"}, nil},
		{nil, []string{"A", "B"}, nil},
		{nil, []string{"a", "b"}, nil},
		{nil, []string{"", "A", "B", "D"}, nil},
		{nil, []string{"A", "B", "C"}, nil},
		{nil, []string{"a", "b", "c"}, nil},
		{nil, []string{"", "A", "B", "C", "D"}, nil},

		{data, nil, SIM{"A": 1, "B": 2, "C": 3}},
		{data, []string{}, SIM{"A": 1, "B": 2, "C": 3}},
		{data, []string{""}, SIM{"A": 1, "B": 2, "C": 3}},
		{data, []string{"A"}, SIM{"B": 2, "C": 3}},
		{data, []string{"a"}, SIM{"A": 1, "B": 2, "C": 3}},
		{data, []string{"", "A", "D"}, SIM{"B": 2, "C": 3}},
		{data, []string{"A", "B"}, SIM{"C": 3}},
		{data, []string{"a", "b"}, SIM{"A": 1, "B": 2, "C": 3}},
		{data, []string{"", "A", "B", "D"}, SIM{"C": 3}},
		{data, []string{"A", "B", "C"}, SIM{}},
		{data, []string{"a", "b", "c"}, SIM{"A": 1, "B": 2, "C": 3}},
		{data, []string{"", "A", "B", "C", "D"}, SIM{}},
	}

	for _, tc := range testCases {
		var gm *SIGM
		if tc.data != nil {
			m := maps.Clone(tc.data)
			gm = (*SIGM)(&m)
		}
		t.Run(
			fmt.Sprintf("gm=%s&key=%s", gmPtrToName(gm), keysToName(tc.ks)),
			func(t *testing.T) {
				gm.Remove(tc.ks...)
				if goMapWrong(gm, tc.want) {
					t.Errorf("got %s; want %s",
						gmPtrToName(gm), mapToString(tc.want))
				}
			},
		)
	}
}

func TestGoMap_GetAndRemove(t *testing.T) {
	data := SIM{"A": 1, "B": 2, "C": 3}
	testCases := []struct {
		data  SIM // if data is nil, gm is (*SIGM)(nil)
		k     string
		wantM SIM
		wantV int
		wantP bool
	}{
		{nil, "", nil, 0, false},
		{nil, "A", nil, 0, false},
		{nil, "B", nil, 0, false},
		{nil, "C", nil, 0, false},
		{nil, "a", nil, 0, false},

		{data, "", SIM{"A": 1, "B": 2, "C": 3}, 0, false},
		{data, "A", SIM{"B": 2, "C": 3}, 1, true},
		{data, "B", SIM{"A": 1, "C": 3}, 2, true},
		{data, "C", SIM{"A": 1, "B": 2}, 3, true},
		{data, "a", SIM{"A": 1, "B": 2, "C": 3}, 0, false},
	}

	for _, tc := range testCases {
		var gm *SIGM
		if tc.data != nil {
			m := maps.Clone(tc.data)
			gm = (*SIGM)(&m)
		}
		t.Run(
			fmt.Sprintf("gm=%s&key=%+q", gmPtrToName(gm), tc.k),
			func(t *testing.T) {
				v, p := gm.GetAndRemove(tc.k)
				if goMapWrong(gm, tc.wantM) {
					t.Errorf("got map %s; want %s",
						gmPtrToName(gm), mapToString(tc.wantM))
				}
				if v != tc.wantV || p != tc.wantP {
					t.Errorf("got (%d, %t); want (%d, %t)",
						v, p, tc.wantV, tc.wantP)
				}
			},
		)
	}
}

func keysToName(keys []string) string {
	return fmtcoll.MustFormatSliceToString(
		keys,
		&fmtcoll.SequenceFormat[string]{
			CommonFormat: fmtcoll.CommonFormat{
				Separator: ",",
			},
			FormatItemFn: fmtcoll.FprintfToFormatFunc[string]("%+q"),
		},
	)
}

func gmPtrToName(p *SIGM) string {
	if p == nil {
		return fmt.Sprintf("(%T)<nil>", p)
	}
	return mapToString(*p)
}

func mapToString(m SIM) string {
	return fmtcoll.MustFormatMapToString(m, &fmtcoll.MapFormat[string, int]{
		CommonFormat: fmtcoll.CommonFormat{
			Separator: ",",
		},
		FormatKeyFn:   fmtcoll.FprintfToFormatFunc[string]("%+q"),
		FormatValueFn: fmtcoll.FprintfToFormatFunc[int]("%d"),
		CompareKeyValueFn: func(
			key1 string,
			value1 int,
			key2 string,
			value2 int,
		) int {
			switch {
			case key1 < key2:
				return -1
			case key1 > key2:
				return 1
			case value1 < value2:
				return -1
			case value1 > value2:
				return 1
			}
			return 0
		},
	})
}

func mapItfToString(m mapping.Map[string, int]) string {
	return fmtcoll.MustFormatGogoMapToString(m, &fmtcoll.MapFormat[string, int]{
		CommonFormat: fmtcoll.CommonFormat{
			Separator:   ",",
			PrependType: true,
		},
		FormatKeyFn:   fmtcoll.FprintfToFormatFunc[string]("%+q"),
		FormatValueFn: fmtcoll.FprintfToFormatFunc[int]("%d"),
		CompareKeyValueFn: func(
			key1 string,
			value1 int,
			key2 string,
			value2 int,
		) int {
			switch {
			case key1 < key2:
				return -1
			case key1 > key2:
				return 1
			case value1 < value2:
				return -1
			case value1 > value2:
				return 1
			}
			return 0
		},
	})
}

func mapWrong(m, want SIM) bool {
	return (m == nil) != (want == nil) || !maps.Equal(m, want)
}

func goMapWrong(gm *SIGM, want SIM) bool {
	if gm == nil {
		return want != nil
	}
	return mapItfWrong(gm, want)
}

func mapItfWrong(m mapping.Map[string, int], want SIM) bool {
	if m == nil {
		return want != nil
	} else if want == nil || m.Len() != len(want) {
		return true
	}
	ok := true
	m.Range(func(x mapping.Entry[string, int]) (cont bool) {
		var v int
		v, ok = want[x.Key]
		if ok {
			ok = x.Value == v
		}
		return ok
	})
	return !ok
}
