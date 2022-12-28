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

package agpl3_test

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/donyori/gogo/copyright/agpl3"
)

const CopyrightNoticePatternLayout = "^    %s  Copyright [(]C[)] %v  %s\n" +
	"    This program comes with ABSOLUTELY NO WARRANTY; for details use `%[1]s show w'.\n" +
	"    This is free software, and you are welcome to redistribute it\n" +
	"    under certain conditions; use `%[1]s show c' for details.\n$"

const CopyrightNoticeWithSourcePatternLayout = "^    %s  Copyright [(]C[)] %v  %s\n" +
	"    This program comes with ABSOLUTELY NO WARRANTY; for details use `%[1]s show w'.\n" +
	"    This is free software, and you are welcome to redistribute it\n" +
	"    under certain conditions; use `%[1]s show c' for details.\n" +
	"    Program source: <%[4]s>.\n$"

func TestPrintCopyrightNotice(t *testing.T) {
	const (
		program = "testgogo"
		author  = "donyori"
		source  = "https://github.com/donyori/gogo"
	)
	nowYear := time.Now().Year()
	year := fmt.Sprintf("2019-%d", nowYear)

	testCases := []struct {
		program, year, author, source string
		wantNoticePattern             string
		wantErr                       error
	}{
		{
			wantErr: agpl3.ErrAuthorMissing,
		},
		{
			program: program,
			year:    year,
			source:  source,
			wantErr: agpl3.ErrAuthorMissing,
		},
		{
			author:            author,
			wantNoticePattern: fmt.Sprintf(CopyrightNoticePatternLayout, "(.)+", nowYear, author),
		},
		{
			author:            author,
			source:            source,
			wantNoticePattern: fmt.Sprintf(CopyrightNoticeWithSourcePatternLayout, "(.)+", nowYear, author, source),
		},
		{
			program:           program,
			year:              year,
			author:            author,
			wantNoticePattern: fmt.Sprintf(CopyrightNoticePatternLayout, program, year, author),
		},
		{
			program:           program,
			year:              year,
			author:            author,
			source:            source,
			wantNoticePattern: fmt.Sprintf(CopyrightNoticeWithSourcePatternLayout, program, year, author, source),
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("program=%q&year=%q&author=%q&source=%q",
			tc.program, tc.year, tc.author, tc.source), func(t *testing.T) {
			p, err := regexp.Compile(tc.wantNoticePattern)
			if err != nil {
				t.Fatalf("cannot compile regular expression %q", tc.wantNoticePattern)
			}
			w := new(strings.Builder)
			n, err := agpl3.PrintCopyrightNotice(w, tc.program, tc.year, tc.author, tc.source)
			if wLen := w.Len(); n != wLen {
				t.Errorf("n, got %d; want %d", n, wLen)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("err, got %v; want %v", err, tc.wantErr)
			}
			if n == 0 && tc.wantNoticePattern == "" {
				return
			}
			matches := p.FindStringSubmatch(w.String())
			if matches == nil {
				t.Fatalf("output cannot match regular expression, output:\n%s\nregular expression:\n%s", w.String(), tc.wantNoticePattern)
			}
			if tc.program == "" {
				if len(matches) != 4 {
					t.Errorf("len(matches) %d; want 4", len(matches))
				}
				for i := 2; i < len(matches); i++ {
					if matches[i-1] != matches[i] {
						t.Errorf("different program names at %d (%s) and %d (%s)", i-1, matches[i-1], i, matches[i])
					}
				}
			}
		})
	}
}

func TestResponseShowWC(t *testing.T) {
	testCases := []struct {
		args       []string
		wantOutput string
	}{
		{[]string{}, ""},
		{[]string{"show a"}, ""},
		{[]string{"show", "a"}, ""},
		{[]string{"show w"}, DisclaimerOfWarranty},
		{[]string{"show", "w"}, DisclaimerOfWarranty},
		{[]string{"show\tw"}, DisclaimerOfWarranty},
		{[]string{"  show   w\t", "", "   "}, DisclaimerOfWarranty},
		{[]string{"show c"}, TermsAndConditions},
		{[]string{"show", "c"}, TermsAndConditions},
		{[]string{"show\tc"}, TermsAndConditions},
		{[]string{"  show   c\t", "", "   "}, TermsAndConditions},
		{[]string{"show w c"}, ""},
		{[]string{"show", "w", "c"}, ""},
	}

	for _, tc := range testCases {
		if tc.wantOutput != "" {
			tc.wantOutput = "  " + tc.wantOutput + "\n"
		}
		t.Run("args="+argsToName(tc.args), func(t *testing.T) {
			w := new(strings.Builder)
			n, err := agpl3.ResponseShowWC(w, tc.args...)
			if err != nil {
				t.Fatal(err)
			}
			if got := w.String(); got != tc.wantOutput {
				t.Errorf("got:\n%s\nwant:\n%s", got, tc.wantOutput)
			}
			if n != w.Len() || n != len(tc.wantOutput) {
				t.Errorf("n, got %d; w.Len() is %d; want %d", n, w.Len(), len(tc.wantOutput))
			}
		})
	}
}

func argsToName(args []string) string {
	if args == nil {
		return "<nil>"
	}
	var b strings.Builder
	b.WriteByte('[')
	for i, arg := range args {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Quote(arg))
	}
	b.WriteByte(']')
	return b.String()
}
