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

	"github.com/donyori/gogo/errors"
)

// Response the requests "show w" and "show c" to w, where "show w" means
// showing the disclaimer of warranty and "show c" means showing the terms and
// conditions. If w is nil, it will use os.Stdout instead. It returns a boolean
// doResp to indicate whether the user input "show w" or "show c" (by checking
// os.Args directly). It also returns any encountered error during writing to w.
// Note that err must be nil if doResp is false.
func RespShowWC(w io.Writer) (doResp bool, err error) {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		return
	}
	var toPrint string
	switch os.Args[1] {
	case "show":
		if len(os.Args) != 3 {
			return
		}
		switch os.Args[2] {
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
