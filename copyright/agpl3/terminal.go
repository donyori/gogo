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

package agpl3

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/donyori/gogo/errors"
)

// terminalNoticeLayout is a layout for formatting a terminal copyright notice.
//
// It takes 3 arguments, <program>, <year>, and <name of author>.
const terminalNoticeLayout = "" +
	"    %s  Copyright (C) %s  %s\n" +
	"    This program comes with ABSOLUTELY NO WARRANTY; for details use `%[1]s show w'.\n" +
	"    This is free software, and you are welcome to redistribute it\n" +
	"    under certain conditions; use `%[1]s show c' for details.\n"

// terminalNoticeWithSourceLayout is a layout for formatting
// a terminal copyright notice with additional program source information.
//
// It takes 4 arguments, <program>, <year>, <name of author>, and <source>.
const terminalNoticeWithSourceLayout = terminalNoticeLayout +
	"    Program source: <%[4]s>.\n"

// ErrAuthorMissing is an error for that the author is missing.
//
// The client should use errors.Is to test whether an error is ErrAuthorMissing.
var ErrAuthorMissing = errors.AutoNewCustom(
	"author is missing",
	errors.PrependFullPkgName,
	0,
)

// PrintCopyrightNotice prints a short copyright notice to w,
// typically for terminal interaction.
//
// If w is nil, it uses os.Stdout instead.
//
// program is the name of the program.
// If program is empty, it uses the base name without the extension
// of os.Args[0] instead.
//
// year is the publishing year of the software.
// If year is empty, it uses current year instead.
//
// author is the name of the author.
// If author is empty, it returns (0, ErrAuthorMissing).
//
// source is the URL of the source of the software.
// If source is empty, the program source part is discarded.
//
// It returns the number of bytes written to w and any error encountered.
func PrintCopyrightNotice(w io.Writer, program, year, author, source string) (
	n int, err error) {
	if author == "" {
		return 0, errors.AutoWrap(ErrAuthorMissing)
	}
	if w == nil {
		w = os.Stdout
	}
	if program == "" && os.Args[0] != "" {
		program = filepath.Base(os.Args[0])
		program = program[:len(program)-len(filepath.Ext(program))]
	}
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}
	layout := terminalNoticeWithSourceLayout
	args := []any{program, year, author, source}
	if source == "" {
		layout, args = terminalNoticeLayout, args[:3]
	}
	n, err = fmt.Fprintf(w, layout, args...)
	return n, errors.AutoWrap(err)
}

// ResponseShowWC responses the user requests "show w" and "show c" to w,
// where "show w" means printing the disclaimer of warranty and
// "show c" means printing the terms and conditions.
//
// If w is nil, it uses os.Stdout instead.
//
// args is a list of arguments input by the user.
// If args is nil, it uses os.Args[1:] instead.
// If the user inputs nothing, set args to []string{} instead of nil.
//
// It returns the number of bytes written to w and any write error encountered.
// If args is neither "show w" nor "show c",
// ResponseShowWC does nothing and returns (0, nil).
func ResponseShowWC(w io.Writer, args ...string) (n int, err error) {
	if args == nil {
		args = os.Args[1:]
	}
	fields := make([]string, 0, 2)
	for _, arg := range args {
		fs := strings.Fields(arg)
		if len(fields)+len(fs) > 2 {
			return
		}
		for i := range fs {
			fs[i] = strings.ToLower(fs[i])
		}
		fields = append(fields, fs...)
	}
	if len(fields) != 2 || fields[0] != "show" {
		return
	}
	var toPrint string
	switch fields[1] {
	case "w":
		toPrint = License[DisclaimerOfWarrantyBegin:DisclaimerOfWarrantyEnd]
	case "c":
		toPrint = License[TermsAndConditionsBegin:TermsAndConditionsEnd]
	default:
		return
	}
	if w == nil {
		w = os.Stdout
	}
	n, err = fmt.Fprintln(w, "  "+toPrint)
	return n, errors.AutoWrap(err)
}
