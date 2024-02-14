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

package hex_test

import (
	"bytes"
	stdhex "encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/donyori/gogo/encoding/hex"
)

func TestEncode_CompareWithOfficial(t *testing.T) {
	srcs := [][]byte{nil, {}, []byte("Hello world! 你好，世界！")}
	dst := make([]byte, stdhex.EncodedLen(len(srcs[len(srcs)-1])))
	stdDst := make([]byte, len(dst))
	for _, src := range srcs {
		var srcName string
		if src == nil {
			srcName = "<nil>"
		} else {
			srcName = strconv.QuoteToASCII(string(src))
		}
		t.Run("src="+srcName, func(t *testing.T) {
			n := hex.Encode(dst, src, false)
			if n2 := stdhex.Encode(stdDst, src); n != n2 {
				t.Fatalf("got n %d; want %d", n, n2)
			}
			if !bytes.Equal(dst[:n], stdDst[:n]) {
				t.Errorf("got %q; want %q", dst[:n], stdDst[:n])
			}
		})
	}
}

func TestEncodedLen(t *testing.T) {
	for _, tc := range testEncodeCases {
		if tc.upper { // only use the lower cases to avoid redundant sources
			continue
		}
		t.Run("src="+tc.srcName, func(t *testing.T) {
			t.Run("type=int", func(t *testing.T) {
				n := hex.EncodedLen(len(tc.srcStr))
				if n != len(tc.dstStr) {
					t.Errorf("got %d; want %d", n, len(tc.dstStr))
				}
			})
			t.Run("type=int64", func(t *testing.T) {
				n := hex.EncodedLen(int64(len(tc.srcStr)))
				if n != int64(len(tc.dstStr)) {
					t.Errorf("got %d; want %d", n, len(tc.dstStr))
				}
			})
		})
	}
}

func TestEncode(t *testing.T) {
	dst := make([]byte, testEncodeCasesDstMaxLen+1024)
	for _, tc := range testEncodeCases {
		t.Run(
			fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper),
			func(t *testing.T) {
				t.Run("type=[]byte", func(t *testing.T) {
					n := hex.Encode(dst, tc.srcBytes, tc.upper)
					if string(dst[:n]) != tc.dstStr {
						t.Errorf("got %q; want %q", dst[:n], tc.dstStr)
					}
				})
				t.Run("type=string", func(t *testing.T) {
					n := hex.Encode(dst, tc.srcStr, tc.upper)
					if string(dst[:n]) != tc.dstStr {
						t.Errorf("got %q; want %q", dst[:n], tc.dstStr)
					}
				})
			},
		)
	}
}

func TestAppendEncode(t *testing.T) {
	testCases := []struct {
		name string
		p    []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"nonempty", []byte("Append")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, etc := range testEncodeCases {
				t.Run(
					fmt.Sprintf("src=%s&upper=%t", etc.srcName, etc.upper),
					func(t *testing.T) {
						t.Run("type=[]byte", func(t *testing.T) {
							dst := slices.Clone(tc.p)
							want := string(tc.p) + etc.dstStr
							got := hex.AppendEncode(
								dst, etc.srcBytes, etc.upper)
							if len(want) == 0 && (got == nil) != (tc.p == nil) {
								if got == nil {
									t.Errorf("got <nil>; want %q", want)
								} else {
									t.Errorf("got %q; want <nil>", got)
								}
							} else if string(got) != want {
								t.Errorf("got %q; want %q", got, want)
							}
						})
						t.Run("type=string", func(t *testing.T) {
							dst := slices.Clone(tc.p)
							want := string(tc.p) + etc.dstStr
							got := hex.AppendEncode(dst, etc.srcStr, etc.upper)
							if len(want) == 0 && (got == nil) != (tc.p == nil) {
								if got == nil {
									t.Errorf("got <nil>; want %q", want)
								} else {
									t.Errorf("got %q; want <nil>", got)
								}
							} else if string(got) != want {
								t.Errorf("got %q; want %q", got, want)
							}
						})
					},
				)
			}
		})
	}
}

func TestEncodeToString(t *testing.T) {
	for _, tc := range testEncodeCases {
		t.Run(
			fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper),
			func(t *testing.T) {
				t.Run("type=[]byte", func(t *testing.T) {
					s := hex.EncodeToString(tc.srcBytes, tc.upper)
					if s != tc.dstStr {
						t.Errorf("got %q; want %q", s, tc.dstStr)
					}
				})
				t.Run("type=string", func(t *testing.T) {
					s := hex.EncodeToString(tc.srcStr, tc.upper)
					if s != tc.dstStr {
						t.Errorf("got %q; want %q", s, tc.dstStr)
					}
				})
			},
		)
	}
}

func TestEncoder_Write(t *testing.T) {
	buf := make([]byte, testEncodeCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := hex.NewEncoder(w, true)
	lowerEncoder := hex.NewEncoder(w, false)
	for _, tc := range testEncodeCases {
		t.Run(
			fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper),
			func(t *testing.T) {
				w.Reset()
				var encoder hex.Encoder
				if tc.upper {
					encoder = upperEncoder
				} else {
					encoder = lowerEncoder
				}
				n, err := encoder.Write(tc.srcBytes)
				if err != nil {
					t.Fatal(err)
				}
				n = hex.EncodedLen(n)
				if string(buf[:n]) != tc.dstStr {
					t.Errorf("got %q; want %q", buf[:n], tc.dstStr)
				}
			},
		)
	}
}

func TestEncoder_WriteByte(t *testing.T) {
	buf := make([]byte, testEncodeCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := hex.NewEncoder(w, true)
	lowerEncoder := hex.NewEncoder(w, false)
	for _, tc := range testEncodeCases {
		t.Run(
			fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper),
			func(t *testing.T) {
				w.Reset()
				var encoder hex.Encoder
				if tc.upper {
					encoder = upperEncoder
				} else {
					encoder = lowerEncoder
				}
				var n int
				for _, b := range tc.srcBytes {
					err := encoder.WriteByte(b)
					if err != nil {
						t.Fatalf("WriteByte(%q) - %v", b, err)
					}
					n++
				}
				n = hex.EncodedLen(n)
				if string(buf[:n]) != tc.dstStr {
					t.Errorf("got %q; want %q", buf[:n], tc.dstStr)
				}
			},
		)
	}
}

func TestEncoder_WriteString(t *testing.T) {
	buf := make([]byte, testEncodeCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := hex.NewEncoder(w, true)
	lowerEncoder := hex.NewEncoder(w, false)
	for _, tc := range testEncodeCases {
		t.Run(
			fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper),
			func(t *testing.T) {
				w.Reset()
				var encoder hex.Encoder
				if tc.upper {
					encoder = upperEncoder
				} else {
					encoder = lowerEncoder
				}
				n, err := encoder.WriteString(tc.srcStr)
				if err != nil {
					t.Fatal(err)
				}
				n = hex.EncodedLen(n)
				if string(buf[:n]) != tc.dstStr {
					t.Errorf("got %q; want %q", buf[:n], tc.dstStr)
				}
			},
		)
	}
}

func TestEncoder_ReadFrom(t *testing.T) {
	buf := make([]byte, testEncodeCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := hex.NewEncoder(w, true)
	lowerEncoder := hex.NewEncoder(w, false)
	for _, tc := range testEncodeCases {
		t.Run(
			fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper),
			func(t *testing.T) {
				w.Reset()
				var encoder hex.Encoder
				if tc.upper {
					encoder = upperEncoder
				} else {
					encoder = lowerEncoder
				}
				n, err := encoder.ReadFrom(strings.NewReader(tc.srcStr))
				if err != nil {
					t.Fatal(err)
				}
				n = hex.EncodedLen(n)
				if string(buf[:n]) != tc.dstStr {
					t.Errorf("got %q; want %q", buf[:n], tc.dstStr)
				}
			},
		)
	}
}
