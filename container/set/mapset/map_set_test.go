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

package mapset_test

import (
	"fmt"
	"maps"
	"testing"

	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/container/set"
	"github.com/donyori/gogo/container/set/mapset"
	"github.com/donyori/gogo/fmtcoll"
)

type (
	IntSDA    = array.SliceDynamicArray[int]
	IntSDAPtr = *array.SliceDynamicArray[int]
)

var dataList = [][]int{
	nil, {},
	{0}, {0, 0},
	{0, 1}, {0, 1, 1},
	{0, 1, 2}, {0, 1, 2, 2},
	{0, 1, 2, 3}, {0, 1, 2, 3, 3},
	{0, 1, 2, 3, 4}, {0, 1, 2, 3, 4, 4},
	{0, 1, 1, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 5},
}

var dataSetList []map[int]struct{} // it is set in function init

func init() {
	dataSetList = make([]map[int]struct{}, len(dataList))
	for i := range dataSetList {
		dataSetList[i] = make(map[int]struct{}, len(dataList[i]))
		for _, x := range dataList[i] {
			dataSetList[i][x] = struct{}{}
		}
	}
}

func TestNew(t *testing.T) {
	var n int
	for _, data := range dataList {
		n += len(data) + 3
	}

	testCases := make([]struct {
		data     []int
		capacity int
		want     map[int]struct{}
	}, n)
	var idx int
	for i, data := range dataList {
		for c := -1; c <= len(data)+1; c++ {
			testCases[idx].data = data
			testCases[idx].capacity = c
			testCases[idx].want = dataSetList[i]
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&cap=%d", sliceToName(tc.data), tc.capacity),
			func(t *testing.T) {
				ms := mapset.New[int](tc.capacity, IntSDAPtr(&tc.data))
				if setWrong(ms, tc.want) {
					t.Errorf("got %s; want %s",
						setToString(ms), mapKeyToString(tc.want))
				}
			},
		)
	}
}

func TestMapSet_Len(t *testing.T) {
	for i, data := range dataList {
		want := len(dataSetList[i])
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			ms := mapset.New[int](0, IntSDAPtr(&data))
			if n := ms.Len(); n != want {
				t.Errorf("got %d; want %d", n, want)
			}
		})
	}
}

func TestMapSet_Range(t *testing.T) {
	for i, data := range dataList {
		counterMap := make(map[int]int, len(dataSetList[i]))
		for x := range dataSetList[i] {
			counterMap[x] = 1
		}
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			ms := mapset.New[int](0, IntSDAPtr(&data))
			ms.Range(func(x int) (cont bool) {
				counterMap[x]--
				return true
			})
			for x, ctr := range counterMap {
				if ctr > 0 {
					t.Errorf("insufficient accesses to %d", x)
				} else if ctr < 0 {
					t.Errorf("too many accesses to %d", x)
				}
			}
		})
	}
}

func TestMapSet_Filter(t *testing.T) {
	filterList := []func(x int) (keep bool){
		func(x int) (keep bool) {
			return x > 1
		},
		func(x int) (keep bool) {
			return x%2 == 0
		},
	}

	testCases := make([]struct {
		data      []int
		filterIdx int
		want      map[int]struct{}
	}, len(dataList)*len(filterList))
	var idx int
	for _, data := range dataList {
		for filterIdx, filter := range filterList {
			testCases[idx].data = data
			testCases[idx].filterIdx = filterIdx
			testCases[idx].want = make(map[int]struct{}, len(data))
			for _, x := range data {
				if filter(x) {
					testCases[idx].want[x] = struct{}{}
				}
			}
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&filterIdx=%d",
				sliceToName(tc.data), tc.filterIdx),
			func(t *testing.T) {
				ms := mapset.New[int](0, IntSDAPtr(&tc.data))
				ms.Filter(filterList[tc.filterIdx])
				if setWrong(ms, tc.want) {
					t.Errorf("got %s; want %s",
						setToString(ms), mapKeyToString(tc.want))
				}
			},
		)
	}
}

func TestMapSet_ContainsItem(t *testing.T) {
	const MaxX int = 6

	testCases := make([]struct {
		data []int
		x    int
		want bool
	}, len(dataList)*(MaxX+2))
	var idx int
	for i, data := range dataList {
		for x := -1; x <= MaxX; x++ {
			testCases[idx].data = data
			testCases[idx].x = x
			_, testCases[idx].want = dataSetList[i][x]
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&x=%d", sliceToName(tc.data), tc.x),
			func(t *testing.T) {
				ms := mapset.New[int](0, IntSDAPtr(&tc.data))
				if got := ms.ContainsItem(tc.x); got != tc.want {
					t.Errorf("got %t; want %t", got, tc.want)
				}
			},
		)
	}
}

func TestMapSet_ContainsSet(t *testing.T) {
	setList := []set.Set[int]{
		nil,
		mapset.New[int](0, nil),
		mapset.New[int](0, &IntSDA{0}),
		mapset.New[int](0, &IntSDA{0, 1}),
		mapset.New[int](0, &IntSDA{0, 1, 2}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4, 5}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4, 5, 6}),
	}

	testCases := make([]struct {
		data []int
		s    set.Set[int]
		want bool
	}, len(dataList)*len(setList))
	var idx int
	for i, data := range dataList {
		for _, s := range setList {
			testCases[idx].data = data
			testCases[idx].s = s
			want := true
			if s != nil {
				s.Range(func(x int) (cont bool) {
					if _, ok := dataSetList[i][x]; !ok {
						want = false
						return false
					}
					return true
				})
			}
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&s=%s",
				sliceToName(tc.data), setToString(tc.s)),
			func(t *testing.T) {
				ms := mapset.New[int](0, IntSDAPtr(&tc.data))
				if got := ms.ContainsSet(tc.s); got != tc.want {
					t.Errorf("got %t; want %t", got, tc.want)
				}
			},
		)
	}
}

func TestMapSet_ContainsAny(t *testing.T) {
	ctnrDataList := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3},
		{0, 1, 2, 3, 4},
		{0, 1, 2, 3, 4, 5},
		{0, 1, 2, 3, 4, 5, 6},
		{1, 2, 3, 4, 5, 6},
		{2, 3, 4, 5, 6},
		{3, 4, 5, 6},
		{4, 5, 6},
		{5, 6},
		{6},
	}

	testCases := make([]struct {
		data, c []int
		want    bool
	}, len(dataList)*len(ctnrDataList))
	var idx int
	for i, data := range dataList {
		for _, ctnr := range ctnrDataList {
			testCases[idx].data = data
			testCases[idx].c = ctnr
			for _, x := range ctnr {
				if _, ok := dataSetList[i][x]; ok {
					testCases[idx].want = true
					break
				}
			}
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&c=%s",
				sliceToName(tc.data), sliceToName(tc.c)),
			func(t *testing.T) {
				ms := mapset.New[int](0, IntSDAPtr(&tc.data))
				if got := ms.ContainsAny(IntSDAPtr(&tc.c)); got != tc.want {
					t.Errorf("got %t; want %t", got, tc.want)
				}
			},
		)
	}
}

func TestMapSet_Add(t *testing.T) {
	testCases := make([]struct {
		data, x []int
		want    map[int]struct{}
	}, len(dataList)*len(dataList))
	var idx int
	for i, data := range dataList {
		for _, x := range dataList {
			testCases[idx].data = data
			testCases[idx].x = x
			want := maps.Clone(dataSetList[i])
			for _, item := range x {
				want[item] = struct{}{}
			}
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&x=%s",
				sliceToName(tc.data), sliceToName(tc.x)),
			func(t *testing.T) {
				ms := mapset.New[int](0, IntSDAPtr(&tc.data))
				ms.Add(tc.x...)
				if setWrong(ms, tc.want) {
					t.Errorf("got %s; want %s",
						setToString(ms), mapKeyToString(tc.want))
				}
			},
		)
	}
}

func TestMapSet_Remove(t *testing.T) {
	testCases := make([]struct {
		data, x []int
		want    map[int]struct{}
	}, len(dataList)*len(dataList))
	var idx int
	for i, data := range dataList {
		for _, x := range dataList {
			testCases[idx].data = data
			testCases[idx].x = x
			want := maps.Clone(dataSetList[i])
			for _, item := range x {
				delete(want, item)
			}
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&x=%s",
				sliceToName(tc.data), sliceToName(tc.x)),
			func(t *testing.T) {
				ms := mapset.New[int](0, IntSDAPtr(&tc.data))
				ms.Remove(tc.x...)
				if setWrong(ms, tc.want) {
					t.Errorf("got %s; want %s",
						setToString(ms), mapKeyToString(tc.want))
				}
			},
		)
	}
}

func TestMapSet_Union(t *testing.T) {
	setList := []set.Set[int]{
		nil,
		mapset.New[int](0, nil),
		mapset.New[int](0, &IntSDA{0}),
		mapset.New[int](0, &IntSDA{0, 1}),
		mapset.New[int](0, &IntSDA{0, 1, 2}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4, 5}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4, 5, 6}),
	}

	testCases := make([]struct {
		data []int
		s    set.Set[int]
		want map[int]struct{}
	}, len(dataList)*len(setList))
	var idx int
	for i, data := range dataList {
		for _, s := range setList {
			testCases[idx].data = data
			testCases[idx].s = s
			want := maps.Clone(dataSetList[i])
			if s != nil {
				s.Range(func(x int) (cont bool) {
					want[x] = struct{}{}
					return true
				})
			}
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&s=%s",
				sliceToName(tc.data), setToString(tc.s)),
			func(t *testing.T) {
				ms := mapset.New[int](0, IntSDAPtr(&tc.data))
				ms.Union(tc.s)
				if setWrong(ms, tc.want) {
					t.Errorf("got %s; want %s",
						setToString(ms), mapKeyToString(tc.want))
				}
			},
		)
	}
}

func TestMapSet_Intersect(t *testing.T) {
	setList := []set.Set[int]{
		nil,
		mapset.New[int](0, nil),
		mapset.New[int](0, &IntSDA{0}),
		mapset.New[int](0, &IntSDA{0, 1}),
		mapset.New[int](0, &IntSDA{0, 1, 2}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4, 5}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4, 5, 6}),
	}

	testCases := make([]struct {
		data []int
		s    set.Set[int]
		want map[int]struct{}
	}, len(dataList)*len(setList))
	var idx int
	for i, data := range dataList {
		for _, s := range setList {
			testCases[idx].data = data
			testCases[idx].s = s
			var want map[int]struct{}
			if s != nil {
				want = maps.Clone(dataSetList[i])
				for x := range want {
					if !s.ContainsItem(x) {
						delete(want, x)
					}
				}
			} else {
				want = map[int]struct{}{}
			}
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&s=%s",
				sliceToName(tc.data), setToString(tc.s)),
			func(t *testing.T) {
				ms := mapset.New[int](0, IntSDAPtr(&tc.data))
				ms.Intersect(tc.s)
				if setWrong(ms, tc.want) {
					t.Errorf("got %s; want %s",
						setToString(ms), mapKeyToString(tc.want))
				}
			},
		)
	}
}

func TestMapSet_Subtract(t *testing.T) {
	setList := []set.Set[int]{
		nil,
		mapset.New[int](0, nil),
		mapset.New[int](0, &IntSDA{0}),
		mapset.New[int](0, &IntSDA{0, 1}),
		mapset.New[int](0, &IntSDA{0, 1, 2}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4, 5}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4, 5, 6}),
	}

	testCases := make([]struct {
		data []int
		s    set.Set[int]
		want map[int]struct{}
	}, len(dataList)*len(setList))
	var idx int
	for i, data := range dataList {
		for _, s := range setList {
			testCases[idx].data = data
			testCases[idx].s = s
			want := maps.Clone(dataSetList[i])
			if s != nil {
				s.Range(func(x int) (cont bool) {
					delete(want, x)
					return true
				})
			}
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&s=%s",
				sliceToName(tc.data), setToString(tc.s)),
			func(t *testing.T) {
				ms := mapset.New[int](0, IntSDAPtr(&tc.data))
				ms.Subtract(tc.s)
				if setWrong(ms, tc.want) {
					t.Errorf("got %s; want %s",
						setToString(ms), mapKeyToString(tc.want))
				}
			},
		)
	}
}

func TestMapSet_DisjunctiveUnion(t *testing.T) {
	setList := []set.Set[int]{
		nil,
		mapset.New[int](0, nil),
		mapset.New[int](0, &IntSDA{0}),
		mapset.New[int](0, &IntSDA{0, 1}),
		mapset.New[int](0, &IntSDA{0, 1, 2}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4, 5}),
		mapset.New[int](0, &IntSDA{0, 1, 2, 3, 4, 5, 6}),
	}

	testCases := make([]struct {
		data []int
		s    set.Set[int]
		want map[int]struct{}
	}, len(dataList)*len(setList))
	var idx int
	for i, data := range dataList {
		for _, s := range setList {
			testCases[idx].data = data
			testCases[idx].s = s
			var want map[int]struct{}
			if s != nil {
				want = make(map[int]struct{}, len(data)+s.Len())
				for _, x := range data {
					if !s.ContainsItem(x) {
						want[x] = struct{}{}
					}
				}
				s.Range(func(x int) (cont bool) {
					if _, ok := dataSetList[i][x]; !ok {
						want[x] = struct{}{}
					}
					return true
				})
			} else {
				want = dataSetList[i]
			}
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&s=%s",
				sliceToName(tc.data), setToString(tc.s)),
			func(t *testing.T) {
				ms := mapset.New[int](0, IntSDAPtr(&tc.data))
				ms.DisjunctiveUnion(tc.s)
				if setWrong(ms, tc.want) {
					t.Errorf("got %s; want %s",
						setToString(ms), mapKeyToString(tc.want))
				}
			},
		)
	}
}

func TestMapSet_Clear(t *testing.T) {
	for _, data := range dataList {
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			ms := mapset.New[int](0, IntSDAPtr(&data))
			ms.Clear()
			if setWrong(ms, map[int]struct{}{}) {
				t.Errorf("got %s; want {}", setToString(ms))
			}
		})
	}
}

func sliceToName[T any](s []T) string {
	return fmtcoll.MustFormatSliceToString(s, &fmtcoll.SequenceFormat[T]{
		CommonFormat: fmtcoll.CommonFormat{
			Separator:   ",",
			PrependType: true,
		},
		FormatItemFn: fmtcoll.FprintfToFormatFunc[T]("%v"),
	})
}

func setToString(s set.Set[int]) string {
	if s == nil {
		return "<nil>"
	}
	m := make(map[int]struct{}, s.Len())
	s.Range(func(x int) (cont bool) {
		m[x] = struct{}{}
		return true
	})
	return mapKeyToString(m)
}

func mapKeyToString[V any](m map[int]V) string {
	return fmtcoll.MustFormatMapToString(m, &fmtcoll.MapFormat[int, V]{
		CommonFormat: fmtcoll.CommonFormat{
			Separator: ",",
		},
		FormatKeyFn: fmtcoll.FprintfToFormatFunc[int]("%d"),
		CompareKeyValueFn: func(key1 int, _ V, key2 int, _ V) int {
			if key1 < key2 {
				return -1
			} else if key1 > key2 {
				return 1
			}
			return 0
		},
	})
}

func setWrong(s set.Set[int], want map[int]struct{}) bool {
	if s == nil {
		return want != nil
	} else if want == nil || s.Len() != len(want) {
		return true
	}
	counterMap := make(map[int]int, len(want))
	for x := range want {
		counterMap[x] = 1
	}
	s.Range(func(x int) (cont bool) {
		counterMap[x]--
		return true
	})
	for _, ctr := range counterMap {
		if ctr != 0 {
			return true
		}
	}
	return false
}
