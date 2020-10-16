// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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
	"strconv"
	"time"

	"github.com/donyori/gogo/errors"
)

// Layout for formatting notice. The verbs in the first line stand for
// <program>, <year>, and <name of author> respectively.
const noticeLayout = "    %s  Copyright (C) %s  %s\n" +
	"    This program comes with ABSOLUTELY NO WARRANTY; for details use `%[1]s show w'.\n" +
	"    This is free software, and you are welcome to redistribute it\n" +
	"    under certain conditions; use `%[1]s show c' for details.\n"

// Layout for formatting notice, with additional program source information.
const noticeWithSourceLayout = noticeLayout + "    Program source: <%[4]s>.\n"

// An error for that the author is missing.
var ErrAuthorMissing = errors.New("author is missing")

// Print a short notice to w, typically in the terminal interaction. program is
// the name of the program. year is the publish year of your software. author
// is the name of the author. source is the URL of the source of your software.
// If w is nil, it will use os.Stdout instead. If program is empty, it will use
// os.Args[0] instead. If year is empty, it will use current year instead. If
// author is empty, it will return ErrAuthorMissing. If source is empty, the
// part of the program source will be discarded.
func PrintNotice(w io.Writer, program, year, author, source string) error {
	if author == "" {
		return errors.AutoWrap(ErrAuthorMissing)
	}
	if w == nil {
		w = os.Stdout
	}
	if program == "" {
		program = os.Args[0]
	}
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}
	var err error
	if source != "" {
		_, err = fmt.Fprintf(w, noticeWithSourceLayout, program, year, author, source)
	} else {
		_, err = fmt.Fprintf(w, noticeLayout, program, year, author)
	}
	return errors.AutoWrap(err)
}
