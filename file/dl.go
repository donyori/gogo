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

// A combination of a hash algorithm and an expected checksum.
type Checksum struct {
	// A function to generate a hasher. E.g. crypto/sha256.New.
	HashGen func() hash.Hash

	// Expected checksum, encoding to hexadecimal representation.
	HexExpSum string
}

// Download a file via HTTP/HTTPS Get,
// and save as the local file specified by filename, with given permission perm.
//
// The client can specify checksums to verify the downloaded file.
// A damaged file will be removed and ErrVerificationFail will be returned.
func HttpDownload(url, filename string, perm os.FileMode, chksums ...Checksum) error {
	var hashes []hash.Hash
	if len(chksums) > 0 {
		hashes = make([]hash.Hash, len(chksums))
		for i := range chksums {
			if chksums[i].HashGen == nil {
				return errors.AutoNew("given hash generator is nil")
			}
			if chksums[i].HexExpSum == "" {
				return errors.AutoNew("given expected checksum is empty")
			}
			hashes[i] = chksums[i].HashGen()
		}
	}
	resp, err := http.Get(url)
	if err != nil {
		return errors.AutoWrap(err)
	}
	defer resp.Body.Close() // ignore error
	if resp.StatusCode != http.StatusOK {
		errMsg := resp.Status
		if errMsg == "" {
			errMsg = "status code: " + strconv.Itoa(resp.StatusCode)
		}
		return errors.AutoNew("response status is not OK when downloading " + url + ": " + errMsg)
	}
	fw, err := New(filename, perm, &WriteOption{
		Raw:      true,
		Backup:   true,
		MakeDirs: true,
		VerifyFn: func() bool {
			if err != nil {
				return false
			}
			for i := range chksums {
				sum := hex.EncodeToString(hashes[i].Sum(nil), false)
				if sum != strings.ToLower(chksums[i].HexExpSum) {
					return false
				}
			}
			return true
		},
	})
	if err != nil {
		return errors.AutoWrap(err)
	}
	defer fw.Close() // ignore error
	var w io.Writer = fw
	if len(hashes) > 0 {
		ws := make([]io.Writer, 1+len(hashes))
		ws[0] = w
		for i := range hashes {
			ws[i+1] = hashes[i]
		}
		w = io.MultiWriter(ws...)
	}
	_, err = io.Copy(w, resp.Body)
	return errors.AutoWrap(err)
}
