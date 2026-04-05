// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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

package vseq_test

import (
	"fmt"
	"iter"
	"slices"
	"testing"

	"github.com/donyori/gogo/algorithm/traversal/vseq"
	"github.com/donyori/gogo/fmtcoll"
	"github.com/donyori/gogo/internal/unequal"
)

func TestNewVertexSequence(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			gotList := vs.ToList()
			if !slices.Equal(gotList, data) {
				t.Errorf("got list %s; want %s",
					verticesToName(gotList), verticesToName(data))
			}
		})
	}
}

func TestVertexSequence_Len(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			got := vs.Len()
			if got != len(data) {
				t.Errorf("got %d; want %d", got, len(data))
			}
		})
	}
}

func TestVertexSequence_Range(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			want := data
			if len(data) > 1 {
				want = data[:len(data)-1]
			}

			gotData := make([]int, 0, len(data))

			vs.Range(func(x int) (cont bool) {
				gotData = append(gotData, x)
				return len(gotData) < len(data)-1
			})

			if !slices.Equal(gotData, want) {
				t.Errorf("got %s; want %s",
					verticesToName(gotData), verticesToName(want))
			}
		})
	}
}

func TestVertexSequence_IterItems(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			seq := vs.IterItems()
			if seq == nil {
				t.Fatal("got nil iterator")
			} else if len(data) == 0 {
				for x := range seq {
					t.Error("yielded", x)
				}

				return
			}

			want := data
			if len(data) > 1 {
				want = data[:len(data)-1]
			}

			gotData := make([]int, 0, len(data))
			iterateAndCheck(t, "", len(data), seq, gotData, want)

			// Rewind the iterator and test it again.
			iterateAndCheck(t, "rewind, ", len(data), seq, gotData[:0], want)
		})
	}
}

func TestVertexSequence_RangeBackward(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			want := data
			if len(data) > 1 {
				want = make([]int, len(data)-1)
				for i := range want {
					want[i] = data[len(data)-1-i]
				}
			}

			gotData := make([]int, 0, len(data))

			vs.RangeBackward(func(v int) (cont bool) {
				gotData = append(gotData, v)
				return len(gotData) < len(data)-1
			})

			if !slices.Equal(gotData, want) {
				t.Errorf("got %s; want %s",
					verticesToName(gotData), verticesToName(want))
			}
		})
	}
}

func TestVertexSequence_IterItemsBackward(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			seq := vs.IterItemsBackward()
			if seq == nil {
				t.Fatal("got nil iterator")
			} else if len(data) == 0 {
				for x := range seq {
					t.Error("yielded", x)
				}

				return
			}

			want := data
			if len(data) > 1 {
				want = make([]int, len(data)-1)
				for i := range want {
					want[i] = data[len(data)-1-i]
				}
			}

			gotData := make([]int, 0, len(data))
			iterateAndCheck(t, "", len(data), seq, gotData, want)

			// Rewind the iterator and test it again.
			iterateAndCheck(t, "rewind, ", len(data), seq, gotData[:0], want)
		})
	}
}

// iterateAndCheck appends items got from seq to gotDataBuf
// and compares gotDataBuf with want.
// If n > 1, it gets at most n-1 items from seq and then breaks the iteration.
func iterateAndCheck(
	t *testing.T,
	logPrefix string,
	n int,
	seq iter.Seq[int],
	gotDataBuf []int,
	want []int,
) {
	t.Helper()

	for x := range seq {
		gotDataBuf = append(gotDataBuf, x)
		if len(gotDataBuf) >= n-1 {
			break
		}
	}

	if !slices.Equal(gotDataBuf, want) {
		t.Errorf("%sgot %s; want %s",
			logPrefix, verticesToName(gotDataBuf), verticesToName(want))
	}
}

func TestVertexSequence_IterIndexItems(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			seq2 := vs.IterIndexItems()
			if seq2 == nil {
				t.Fatal("got nil iterator")
			} else if len(data) == 0 {
				for i, x := range seq2 {
					t.Errorf("yielded %d: %d", i, x)
				}

				return
			}

			var want [][2]int
			if len(data) > 1 {
				want = make([][2]int, len(data)-1)
				for i := range want {
					want[i] = [2]int{i, data[i]}
				}
			} else {
				want = [][2]int{{0, data[0]}}
			}

			gotData := make([][2]int, 0, len(data))
			iterateAndCheck2(t, "", len(data), seq2, gotData, want)

			// Rewind the iterator and test it again.
			iterateAndCheck2(t, "rewind, ", len(data), seq2, gotData[:0], want)
		})
	}
}

func TestVertexSequence_IterIndexItemsBackward(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			seq2 := vs.IterIndexItemsBackward()
			if seq2 == nil {
				t.Fatal("got nil iterator")
			} else if len(data) == 0 {
				for i, x := range seq2 {
					t.Errorf("yielded %d: %d", i, x)
				}

				return
			}

			var want [][2]int
			if len(data) > 1 {
				want = make([][2]int, len(data)-1)
				for i := range want {
					want[i] = [2]int{len(data) - 1 - i, data[len(data)-1-i]}
				}
			} else {
				want = [][2]int{{0, data[0]}}
			}

			gotData := make([][2]int, 0, len(data))
			iterateAndCheck2(t, "", len(data), seq2, gotData, want)

			// Rewind the iterator and test it again.
			iterateAndCheck2(t, "rewind, ", len(data), seq2, gotData[:0], want)
		})
	}
}

// iterateAndCheck2 appends index-item pairs got from seq2 to gotDataBuf
// and compares gotDataBuf with want.
// If n > 1, it gets at most n-1 index-item pairs from seq2
// and then breaks the iteration.
func iterateAndCheck2(
	t *testing.T,
	logPrefix string,
	n int,
	seq2 iter.Seq2[int, int],
	gotDataBuf [][2]int,
	want [][2]int,
) {
	t.Helper()

	for i, x := range seq2 {
		gotDataBuf = append(gotDataBuf, [2]int{i, x})
		if len(gotDataBuf) >= n-1 {
			break
		}
	}

	if !slices.Equal(gotDataBuf, want) {
		t.Errorf("%sgot %v; want %v", logPrefix, gotDataBuf, want)
	}
}

func TestVertexSequence_Front(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			got := vs.Front()
			if got != data[0] {
				t.Errorf("got %d; want %d", got, data[0])
			}
		})
	}
}

func TestVertexSequence_Back(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			got := vs.Back()
			if got != data[len(data)-1] {
				t.Errorf("got %d; want %d", got, data[len(data)-1])
			}
		})
	}
}

func TestVertexSequence_ToList(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			var want []int
			if len(data) > 0 {
				want = data
			}

			got := vs.ToList()
			if unequal.Slice(got, want) {
				t.Errorf("got %s; want %s",
					verticesToName(got), verticesToName(want))
			}
		})
	}
}

func TestVertexSequence_Append(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		for _, v := range vertices {
			t.Run(
				fmt.Sprintf("data=%s&v=%s",
					verticesToName(data), verticesToName(v)),
				func(t *testing.T) {
					t.Parallel()

					vs := vseq.NewVertexSequence(data...)
					if vs == nil {
						t.Fatal("got nil vseq.VertexSequence[int]")
					}

					want := slices.Concat(data, v)
					got := vs.Append(v...)
					checkResultVertexSequence(t, got, want, "Append")
				},
			)
		}
	}
}

func TestVertexSequence_TruncateBack(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		t.Run("data="+verticesToName(data), func(t *testing.T) {
			t.Parallel()

			vs := vseq.NewVertexSequence(data...)
			if vs == nil {
				t.Fatal("got nil vseq.VertexSequence[int]")
			}

			var want []int
			if len(data) > 1 {
				want = data[:len(data)-1]
			}

			got := vs.TruncateBack()
			checkResultVertexSequence(t, got, want, "TruncateBack")
		})
	}
}

func TestVertexSequence_Truncate(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		for i := -2; i <= len(data)+2; i++ {
			t.Run(
				fmt.Sprintf("data=%s&i=%d", verticesToName(data), i),
				func(t *testing.T) {
					t.Parallel()

					vs := vseq.NewVertexSequence(data...)
					if vs == nil {
						t.Fatal("got nil vseq.VertexSequence[int]")
					}

					want := getTruncateWantList(data, i)
					got := vs.Truncate(i)
					checkResultVertexSequence(t, got, want, "Truncate")
				},
			)
		}
	}
}

// getTruncateWantList returns the wanted list for
// the subtests of TestVertexSequence_Truncate.
func getTruncateWantList(data []int, i int) []int {
	if i <= 0 || len(data) == 0 {
		return nil
	} else if i >= len(data) {
		return data
	}

	return data[:i]
}

func TestVertexSequence_ReplaceBack(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		for _, v := range vertices {
			t.Run(
				fmt.Sprintf("data=%s&v=%s",
					verticesToName(data), verticesToName(v)),
				func(t *testing.T) {
					t.Parallel()

					vs := vseq.NewVertexSequence(data...)
					if vs == nil {
						t.Fatal("got nil vseq.VertexSequence[int]")
					}

					var want []int
					if len(data) > 0 {
						want = slices.Concat(data[:len(data)-1], v)
					} else if len(v) > 0 {
						want = v
					}

					got := vs.ReplaceBack(v...)
					checkResultVertexSequence(t, got, want, "ReplaceBack")
				},
			)
		}
	}
}

func TestVertexSequence_Replace(t *testing.T) {
	t.Parallel()

	vertices := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 1, 4},
		{2, 1, 4},
	}

	for _, data := range vertices {
		for i := -2; i <= len(data); i++ {
			for _, v := range vertices {
				t.Run(
					fmt.Sprintf("data=%s&i=%d&v=%s",
						verticesToName(data), i, verticesToName(v)),
					func(t *testing.T) {
						t.Parallel()

						vs := vseq.NewVertexSequence(data...)
						if vs == nil {
							t.Fatal("got nil vseq.VertexSequence[int]")
						}

						want := getReplaceWantList(data, i, v)
						got := vs.Replace(i, v...)
						checkResultVertexSequence(t, got, want, "Replace")
					},
				)
			}
		}
	}
}

// getReplaceWantList returns the wanted list for
// the subtests of TestVertexSequence_Replace.
func getReplaceWantList(data []int, i int, v []int) []int {
	if i <= 0 || len(data) == 0 {
		if len(v) == 0 {
			return nil
		}

		return v
	} else if i >= len(data) {
		return slices.Concat(data, v)
	}

	return slices.Concat(data[:i], v)
}

// checkResultVertexSequence checks the vertex sequence returned by
// Append, TruncateBack, Truncate, ReplaceBack, or Replace.
//
// It tests the nilness of the vertex sequence.
// Then, it gets the data of the vertex sequence via ToList
// and compares the data with want.
func checkResultVertexSequence(
	t *testing.T,
	got vseq.VertexSequence[int],
	want []int,
	methodName string,
) {
	t.Helper()

	if got == nil {
		t.Errorf("vs.%s returned nil vseq.VertexSequence[int]", methodName)
		return
	}

	gotList := got.ToList()
	if unequal.Slice(gotList, want) {
		t.Errorf("got list %s; want %s",
			verticesToName(gotList), verticesToName(want))
	}
}

func verticesToName(v []int) string {
	return fmtcoll.MustFormatSliceToString(
		v,
		fmtcoll.NewDefaultSequenceFormat[int](),
	)
}
