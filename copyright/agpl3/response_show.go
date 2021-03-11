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

package agpl3

import (
	"fmt"
	"io"
	"os"

	"github.com/donyori/gogo/errors"
)

// RespShowWC responses the requests "show w" and "show c" to w,
// where "show w" means showing the disclaimer of warranty and
// "show c" means showing the terms and conditions.
//
// If w is nil, it will use os.Stdout instead.
// input is the user's input arguments.
// If input is nil, it will use os.Args[1:] instead (If user inputs nothing,
// set input to []string{} but not nil).
//
// It returns a Boolean doResp to indicate whether
// the user's input is one of "show w" and "show c".
// It also returns any write error encountered.
// Note that err must be nil if doResp is false.
func RespShowWC(w io.Writer, input []string) (doResp bool, err error) {
	if input == nil {
		input = os.Args[1:]
	}
	if len(input) < 1 || len(input) > 2 {
		return
	}
	var toPrint string
	switch input[0] {
	case "show":
		if len(input) != 2 {
			return
		}
		switch input[1] {
		case "w":
			toPrint = DisclaimerOfWarranty
		case "c":
			toPrint = TermsAndConditions
		default:
			return
		}
	case "show w":
		toPrint = DisclaimerOfWarranty
	case "show c":
		toPrint = TermsAndConditions
	default:
		return
	}
	if w == nil {
		w = os.Stdout
	}
	_, err = fmt.Fprintln(w, "  "+toPrint)
	return true, errors.AutoWrap(err)
}
