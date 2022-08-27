// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/container/set"
	"github.com/donyori/gogo/container/set/mapset"
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

var dataSetList []map[int]bool // It will be set in function init.

func init() {
	dataSetList = make([]map[int]bool, len(dataList))
	for i := range dataSetList {
		dataSetList[i] = make(map[int]bool, len(dataList[i]))
		for _, x := range dataList[i] {
			dataSetList[i][x] = true
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
		want     map[int]bool
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
		t.Run(fmt.Sprintf("data=%s&cap=%d", sliceToName(tc.data), tc.capacity), func(t *testing.T) {
			ms := mapset.New[int](tc.capacity, array.SliceDynamicArray[int](tc.data))
			if setWrong(ms, tc.want) {
				t.Errorf("got %s; want %s", setToString(ms), mapKeyToString(tc.want))
			}
		})
	}
}

func TestMapSet_Len(t *testing.T) {
	for i, data := range dataList {
		want := len(dataSetList[i])
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](data))
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
			ms := mapset.New[int](0, array.SliceDynamicArray[int](data))
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
			testCases[idx].want = dataSetList[i][x]
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("data=%s&x=%d", sliceToName(tc.data), tc.x), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](tc.data))
			if got := ms.ContainsItem(tc.x); got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

func TestMapSet_ContainsSet(t *testing.T) {
	setList := []set.Set[int]{
		nil, mapset.New[int](0, nil),
		mapset.New[int](0, array.SliceDynamicArray[int]{0}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4, 5}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4, 5, 6}),
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
					if !dataSetList[i][x] {
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
		t.Run(fmt.Sprintf("data=%s&s=%s", sliceToName(tc.data), setToString(tc.s)), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](tc.data))
			if got := ms.ContainsSet(tc.s); got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

func TestMapSet_ContainsAny(t *testing.T) {
	ctnrDataList := [][]int{
		nil, {},
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
				if dataSetList[i][x] {
					testCases[idx].want = true
					break
				}
			}
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("data=%s&c=%s", sliceToName(tc.data), sliceToName(tc.c)), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](tc.data))
			if got := ms.ContainsAny(array.SliceDynamicArray[int](tc.c)); got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

func TestMapSet_Add(t *testing.T) {
	testCases := make([]struct {
		data, x []int
		want    map[int]bool
	}, len(dataList)*len(dataList))
	var idx int
	for i, data := range dataList {
		for _, x := range dataList {
			testCases[idx].data = data
			testCases[idx].x = x
			want := copyMap(dataSetList[i])
			for _, item := range x {
				want[item] = true
			}
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("data=%s&x=%s", sliceToName(tc.data), sliceToName(tc.x)), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](tc.data))
			ms.Add(tc.x...)
			if setWrong(ms, tc.want) {
				t.Errorf("got %s; want %s", setToString(ms), mapKeyToString(tc.want))
			}
		})
	}
}

func TestMapSet_Remove(t *testing.T) {
	testCases := make([]struct {
		data, x []int
		want    map[int]bool
	}, len(dataList)*len(dataList))
	var idx int
	for i, data := range dataList {
		for _, x := range dataList {
			testCases[idx].data = data
			testCases[idx].x = x
			want := copyMap(dataSetList[i])
			for _, item := range x {
				delete(want, item)
			}
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("data=%s&x=%s", sliceToName(tc.data), sliceToName(tc.x)), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](tc.data))
			ms.Remove(tc.x...)
			if setWrong(ms, tc.want) {
				t.Errorf("got %s; want %s", setToString(ms), mapKeyToString(tc.want))
			}
		})
	}
}

func TestMapSet_Union(t *testing.T) {
	setList := []set.Set[int]{
		nil, mapset.New[int](0, nil),
		mapset.New[int](0, array.SliceDynamicArray[int]{0}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4, 5}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4, 5, 6}),
	}

	testCases := make([]struct {
		data []int
		s    set.Set[int]
		want map[int]bool
	}, len(dataList)*len(setList))
	var idx int
	for i, data := range dataList {
		for _, s := range setList {
			testCases[idx].data = data
			testCases[idx].s = s
			want := copyMap(dataSetList[i])
			if s != nil {
				s.Range(func(x int) (cont bool) {
					want[x] = true
					return true
				})
			}
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("data=%s&s=%s", sliceToName(tc.data), setToString(tc.s)), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](tc.data))
			ms.Union(tc.s)
			if setWrong(ms, tc.want) {
				t.Errorf("got %s; want %s", setToString(ms), mapKeyToString(tc.want))
			}
		})
	}
}

func TestMapSet_Intersect(t *testing.T) {
	setList := []set.Set[int]{
		nil, mapset.New[int](0, nil),
		mapset.New[int](0, array.SliceDynamicArray[int]{0}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4, 5}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4, 5, 6}),
	}

	testCases := make([]struct {
		data []int
		s    set.Set[int]
		want map[int]bool
	}, len(dataList)*len(setList))
	var idx int
	for i, data := range dataList {
		for _, s := range setList {
			testCases[idx].data = data
			testCases[idx].s = s
			if s != nil {
				want := copyMap(dataSetList[i])
				for x := range want {
					if !s.ContainsItem(x) {
						delete(want, x)
					}
				}
				testCases[idx].want = want
			}
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("data=%s&s=%s", sliceToName(tc.data), setToString(tc.s)), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](tc.data))
			ms.Intersect(tc.s)
			if setWrong(ms, tc.want) {
				t.Errorf("got %s; want %s", setToString(ms), mapKeyToString(tc.want))
			}
		})
	}
}

func TestMapSet_Subtract(t *testing.T) {
	setList := []set.Set[int]{
		nil, mapset.New[int](0, nil),
		mapset.New[int](0, array.SliceDynamicArray[int]{0}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4, 5}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4, 5, 6}),
	}

	testCases := make([]struct {
		data []int
		s    set.Set[int]
		want map[int]bool
	}, len(dataList)*len(setList))
	var idx int
	for i, data := range dataList {
		for _, s := range setList {
			testCases[idx].data = data
			testCases[idx].s = s
			want := copyMap(dataSetList[i])
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
		t.Run(fmt.Sprintf("data=%s&s=%s", sliceToName(tc.data), setToString(tc.s)), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](tc.data))
			ms.Subtract(tc.s)
			if setWrong(ms, tc.want) {
				t.Errorf("got %s; want %s", setToString(ms), mapKeyToString(tc.want))
			}
		})
	}
}

func TestMapSet_DisjunctiveUnion(t *testing.T) {
	setList := []set.Set[int]{
		nil, mapset.New[int](0, nil),
		mapset.New[int](0, array.SliceDynamicArray[int]{0}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4, 5}),
		mapset.New[int](0, array.SliceDynamicArray[int]{0, 1, 2, 3, 4, 5, 6}),
	}

	testCases := make([]struct {
		data []int
		s    set.Set[int]
		want map[int]bool
	}, len(dataList)*len(setList))
	var idx int
	for i, data := range dataList {
		for _, s := range setList {
			testCases[idx].data = data
			testCases[idx].s = s
			var want map[int]bool
			if s != nil {
				want = make(map[int]bool, len(data)+s.Len())
				for _, x := range data {
					if !s.ContainsItem(x) {
						want[x] = true
					}
				}
				s.Range(func(x int) (cont bool) {
					if !dataSetList[i][x] {
						want[x] = true
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
		t.Run(fmt.Sprintf("data=%s&s=%s", sliceToName(tc.data), setToString(tc.s)), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](tc.data))
			ms.DisjunctiveUnion(tc.s)
			if setWrong(ms, tc.want) {
				t.Errorf("got %s; want %s", setToString(ms), mapKeyToString(tc.want))
			}
		})
	}
}

func TestMapSet_Clear(t *testing.T) {
	for _, data := range dataList {
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			ms := mapset.New[int](0, array.SliceDynamicArray[int](data))
			ms.Clear()
			if setWrong(ms, nil) {
				t.Errorf("got %s; want {}", setToString(ms))
			}
		})
	}
}

func sliceToName[T any](s []T) string {
	typeStr := fmt.Sprintf("(%T)", s)
	if s == nil {
		return typeStr + "<nil>"
	}
	var b strings.Builder
	b.Grow(len(typeStr) + len(s)*3 + 2)
	b.WriteString(typeStr)
	b.WriteByte('[')
	for i, x := range s {
		if i > 0 {
			b.WriteByte(',')
		}
		_, _ = fmt.Fprintf(&b, "%v", x) // ignore error as error is always nil
	}
	b.WriteByte(']')
	return b.String()
}

func setWrong(s set.Set[int], want map[int]bool) bool {
	if s == nil {
		return len(want) != 0
	}
	if s.Len() != len(want) {
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

func setToString(s set.Set[int]) string {
	if s == nil {
		return "<nil>"
	}
	data := make([]int, 0, s.Len())
	s.Range(func(x int) (cont bool) {
		data = append(data, x)
		return true
	})
	return setOrMapKeyToString(data)
}

func mapKeyToString(m map[int]bool) string {
	if m == nil {
		return "<nil>"
	}
	data := make([]int, 0, len(m))
	for x := range m {
		data = append(data, x)
	}
	return setOrMapKeyToString(data)
}

func setOrMapKeyToString(data []int) string {
	sort.Ints(data)
	var b strings.Builder
	b.Grow(len(data)*2 + 2)
	b.WriteByte('{')
	for i, x := range data {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(x))
	}
	b.WriteByte('}')
	return b.String()
}

func copyMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}
	t := make(map[K]V, len(m))
	for k, v := range m {
		t[k] = v
	}
	return t
}
