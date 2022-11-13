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

package local

import (
	"hash"
	"io"
	"net/http"
	"strconv"

	"github.com/donyori/gogo/encoding/hex"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/filesys"
)

// HttpDownload downloads a file from the specified URL via HTTP Get,
// and saves as a local file specified by filename,
// with specified permission perm.
//
// This function will not create any directory.
// The client is responsible for creating necessary directories.
//
// The client can specify checksums cs to verify the downloaded file.
// A damaged file will be removed and ErrVerificationFail will be returned.
//
// It reports an error and downloads nothing if anyone of cs contains
// a nil NewHash or an empty ExpHex.
func HttpDownload(url, filename string, perm filesys.FileMode, cs ...filesys.HashChecksum) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.AutoWrap(err)
	}
	return errors.AutoWrap(httpRequestDownload(req, filename, perm, cs...))
}

// HttpCustomDownload downloads a file from the specified HTTP request req,
// and saves as a local file specified by filename,
// with specified permission perm.
//
// The client can create the custom request req using function http.NewRequest.
//
// This function will not create any directory.
// The client is responsible for creating necessary directories.
//
// The client can specify checksums cs to verify the downloaded file.
// A damaged file will be removed and ErrVerificationFail will be returned.
//
// It reports an error and downloads nothing if anyone of cs contains
// a nil NewHash or an empty ExpHex.
func HttpCustomDownload(req *http.Request, filename string, perm filesys.FileMode, cs ...filesys.HashChecksum) error {
	if req == nil {
		return errors.AutoNew("req is nil")
	}
	return errors.AutoWrap(httpRequestDownload(req, filename, perm, cs...))
}

// HttpUpdate is the update mode of function HttpDownload.
//
// It verifies the file with specified filename using function VerifyChecksum.
// If the verification is passed, it does nothing and returns (false, nil).
// Otherwise, it calls function HttpDownload to download the file.
//
// It returns an indicator updated and any error encountered.
// updated is true if and only if this function has created or edited the file.
func HttpUpdate(url, filename string, perm filesys.FileMode, cs ...filesys.HashChecksum) (updated bool, err error) {
	if VerifyChecksum(filename, cs...) {
		return
	}
	err = errors.AutoWrap(HttpDownload(url, filename, perm, cs...))
	return err == nil, err
}

// HttpCustomUpdate is the update mode of function HttpCustomDownload.
//
// It verifies the file with specified filename using function VerifyChecksum.
// If the verification is passed, it does nothing and returns (false, nil).
// Otherwise, it calls function HttpCustomDownload to download the file.
//
// It returns an indicator updated and any error encountered.
// updated is true if and only if this function has created or edited the file.
func HttpCustomUpdate(req *http.Request, filename string, perm filesys.FileMode, cs ...filesys.HashChecksum) (updated bool, err error) {
	if VerifyChecksum(filename, cs...) {
		return
	}
	err = errors.AutoWrap(HttpCustomDownload(req, filename, perm, cs...))
	return err == nil, err
}

// httpRequestDownload downloads a file from the specified HTTP request req,
// and saves as a local file specified by filename,
// with specified permission perm.
//
// This function will not create any directory.
// The client is responsible for creating necessary directories.
//
// The client can specify checksums cs to verify the downloaded file.
// A damaged file will be removed and ErrVerificationFail will be returned.
//
// It reports an error and downloads nothing if anyone of cs contains
// a nil NewHash or an empty ExpHex.
//
// Caller should guarantee that req != nil.
func httpRequestDownload(req *http.Request, filename string, perm filesys.FileMode, cs ...filesys.HashChecksum) (err error) {
	var hs []hash.Hash
	var ws []io.Writer
	if len(cs) > 0 {
		hs = make([]hash.Hash, len(cs))
		ws = make([]io.Writer, len(cs))
		for i := range cs {
			if cs[i].NewHash == nil {
				return errors.AutoNew("specified hash is nil")
			}
			if cs[i].ExpHex == "" {
				return errors.AutoNew("specified expected checksum is empty")
			}
			hs[i] = cs[i].NewHash()
			ws[i] = hs[i]
		}
	}
	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return errors.AutoWrap(err)
	}
	defer func(body io.ReadCloser) {
		if err1 := body.Close(); err1 != nil {
			err = errors.AutoWrapSkip(errors.Combine(err, err1), 1) // skip = 1 to skip the inner function
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		errMsg := resp.Status
		if errMsg == "" {
			errMsg = "status code: " + strconv.Itoa(resp.StatusCode)
		}
		return errors.AutoNew("response status is not OK when downloading " + req.URL.String() + ": " + errMsg)
	}
	w, err := Write(filename, perm, &WriteOptions{
		Raw:    true,
		Backup: true,
		VerifyFn: func() bool {
			if err != nil {
				return false
			}
			for i := range cs {
				if !hex.CanEncodeToString(hs[i].Sum(nil), cs[i].ExpHex) {
					return false
				}
			}
			return true
		},
	}, ws...)
	if err != nil {
		return errors.AutoWrap(err)
	}
	defer func(w Writer) {
		if err1 := w.Close(); err1 != nil {
			err = errors.AutoWrapSkip(errors.Combine(err, err1), 1) // skip = 1 to skip the inner function
		}
	}(w)
	_, err = io.Copy(w, resp.Body)
	return errors.AutoWrap(err)
}
