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

package file

import (
	"hash"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/donyori/gogo/encoding/hex"
	"github.com/donyori/gogo/errors"
)

// HttpDownload downloads a file from the specified URL via HTTP Get,
// and saves as a local file specified by filename,
// with specified permission perm.
//
// The client can specify checksums cs to verify the downloaded file.
// A damaged file will be removed and ErrVerificationFail will be returned.
//
// It reports an error and downloads nothing if anyone of cs contains
// a nil HashGen or an empty HexExpSum.
func HttpDownload(url, filename string, perm os.FileMode, cs ...Checksum) (err error) {
	var hashes []hash.Hash
	var ws []io.Writer
	if len(cs) > 0 {
		hashes = make([]hash.Hash, len(cs))
		ws = make([]io.Writer, len(cs))
		for i := range cs {
			if cs[i].HashGen == nil {
				return errors.AutoNew("specified hash generator is nil")
			}
			if cs[i].HexExpSum == "" {
				return errors.AutoNew("specified expected checksum is empty")
			}
			hashes[i] = cs[i].HashGen()
			ws[i] = hashes[i]
		}
	}
	var client http.Client
	resp, err := client.Get(url)
	if err != nil {
		return errors.AutoWrap(err)
	}
	defer func() {
		err1 := resp.Body.Close()
		if err1 != nil {
			err = errors.AutoWrap(errors.Combine(err, err1))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		errMsg := resp.Status
		if errMsg == "" {
			errMsg = "status code: " + strconv.Itoa(resp.StatusCode)
		}
		return errors.AutoNew("response status is not OK when downloading " + url + ": " + errMsg)
	}
	w, err := Write(filename, perm, &WriteOption{
		Raw:      true,
		Backup:   true,
		MakeDirs: true,
		VerifyFn: func() bool {
			if err != nil {
				return false
			}
			for i := range cs {
				sum := hex.EncodeToString(hashes[i].Sum(nil), false)
				if sum != strings.ToLower(cs[i].HexExpSum) {
					return false
				}
			}
			return true
		},
	}, ws...)
	if err != nil {
		return errors.AutoWrap(err)
	}
	defer func() {
		err1 := w.Close()
		if err1 != nil {
			err = errors.AutoWrap(errors.Combine(err, err1))
		}
	}()
	_, err = io.Copy(w, resp.Body)
	return errors.AutoWrap(err)
}

// HttpUpdate is the update mode of function HttpDownload.
//
// It verifies the file with specified filename using function VerifyChecksum.
// If the verification is passed, it does nothing and returns (false, nil).
// Otherwise, it calls function HttpDownload to download the file.
//
// It returns an indicator updated and any error encountered.
// updated is true if and only if this function has created or edited the file.
func HttpUpdate(url, filename string, perm os.FileMode, cs ...Checksum) (updated bool, err error) {
	if VerifyChecksum(filename, cs...) {
		return
	}
	err = errors.AutoWrap(HttpDownload(url, filename, perm, cs...))
	return err == nil, err
}
