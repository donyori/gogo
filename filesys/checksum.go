// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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

package filesys

import (
	"fmt"
	"hash"
	"io"
	"io/fs"
	"strings"

	"github.com/donyori/gogo/encoding/hex"
	"github.com/donyori/gogo/errors"
)

// HashVerifier extends hash.Hash by adding a Match method to report
// whether the hash result matches a pre-specified prefix.
type HashVerifier interface {
	hash.Hash

	// Match reports whether the current hash (as returned by Sum(nil))
	// matches the pre-specified prefix.
	Match() bool
}

// hashVerifier is an implementation of interface HashVerifier.
type hashVerifier struct {
	h         hash.Hash
	prefixHex string
}

// NewHashVerifier creates a new HashVerifier with specified arguments.
//
// newHash is a function that creates a new hash function
// (e.g., crypto/sha256.New, crypto.SHA256.New).
//
// prefixHex is the prefix of the expected hash result,
// in hexadecimal representation.
//
// It panics if newHash is nil or returns nil, or prefixHex is not hexadecimal.
func NewHashVerifier(newHash func() hash.Hash, prefixHex string) HashVerifier {
	if newHash == nil {
		panic(errors.AutoMsg("newHash is nil"))
	}
	if strings.IndexFunc(prefixHex, func(r rune) bool {
		return r < '0' || r > '9' && r < 'A' || r > 'F' && r < 'a' || r > 'f'
	}) >= 0 {
		panic(errors.AutoMsg(fmt.Sprintf("prefixHex (%q) is not hexadecimal", prefixHex)))
	}
	h := newHash()
	if h == nil {
		panic(errors.AutoMsg("newHash returns nil"))
	}
	return &hashVerifier{
		h:         h,
		prefixHex: prefixHex,
	}
}

func (hv *hashVerifier) Write(p []byte) (n int, err error) {
	n, err = hv.h.Write(p)
	return n, errors.AutoWrap(err)
}

func (hv *hashVerifier) Sum(b []byte) []byte {
	return hv.h.Sum(b)
}

func (hv *hashVerifier) Reset() {
	hv.h.Reset()
}

func (hv *hashVerifier) Size() int {
	return hv.h.Size()
}

func (hv *hashVerifier) BlockSize() int {
	return hv.h.BlockSize()
}

// Match reports whether the current hash (as returned by Sum(nil))
// matches the pre-specified prefix.
func (hv *hashVerifier) Match() bool {
	return len(hv.prefixHex) <= hex.EncodedLen(hv.h.Size()) &&
		hex.CanEncodeToPrefix(hv.h.Sum(nil), hv.prefixHex)
}

// VerifyChecksum verifies a file by hash checksum.
//
// To ensure that this function can work as expected,
// the input file must be ready to be read from the beginning and
// must not be operated by anyone else during the call to this function.
//
// closeFile indicates whether this function should close the file.
// If closeFile is false, the client is responsible for closing file after use.
// If closeFile is true and file is not nil,
// file will be closed by this function.
//
// It returns true if the file is not nil and can be read,
// and matches all HashVerifier in hvs
// (nil and duplicate HashVerifier will be ignored).
// In particular, it returns true if hvs has no non-nil HashVerifier.
// In this case, the file will not be read.
//
// Note that VerifyChecksum will not reset the hash state of anyone in hvs.
// The client should use new HashVerifier returned by NewHashVerifier or
// call the Reset method of HashVerifier before calling this function if needed.
func VerifyChecksum(file fs.File, closeFile bool, hvs ...HashVerifier) bool {
	if file == nil {
		return false
	}
	if closeFile {
		defer func(f fs.File) {
			_ = f.Close() // ignore error
		}(file)
	}
	hvs = nonNilDeduplicatedHashVerifiers(hvs)
	if len(hvs) == 0 {
		return true
	}
	ws := make([]io.Writer, len(hvs))
	for i := range hvs {
		ws[i] = hvs[i]
	}
	w := ws[0]
	if len(ws) > 1 {
		w = io.MultiWriter(ws...)
	}
	_, err := io.Copy(w, file)
	if err != nil {
		return false
	}
	for _, hv := range hvs {
		if !hv.Match() {
			return false
		}
	}
	return true
}

// VerifyChecksumFromFS verifies a file by hash checksum,
// where the file is opened from fsys by the specified name.
//
// It returns true if fsys is not nil,
// and the file can be read and matches all HashVerifier in hvs
// (nil and duplicate HashVerifier will be ignored).
// In particular, it returns true if hvs has no non-nil HashVerifier and
// the file can be opened for reading.
// In this case, the file will not be read.
//
// Note that VerifyChecksumFromFS will not reset
// the hash state of anyone in hvs.
// The client should use new HashVerifier returned by NewHashVerifier or
// call the Reset method of HashVerifier before calling this function if needed.
func VerifyChecksumFromFS(fsys fs.FS, name string, hvs ...HashVerifier) bool {
	if fsys == nil {
		return false
	}
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	return VerifyChecksum(f, true, hvs...)
}

// nonNilDeduplicatedHashVerifiers returns all
// non-nil deduplicated HashVerifier in hvs.
//
// If hvs has no nil or duplicate HashVerifier, it returns hvs itself.
// Otherwise, it copies all non-nil deduplicated HashVerifier from hvs to
// a new slice and returns that slice.
// The content of hvs will not be modified in both cases.
//
// If hvs has no non-nil HashVerifier, it returns nil.
func nonNilDeduplicatedHashVerifiers(hvs []HashVerifier) []HashVerifier {
	result := hvs
	set := make(map[HashVerifier]bool, len(hvs))
	original := true
	for i, hv := range hvs {
		if original {
			if hv == nil || set[hv] {
				result, original = make([]HashVerifier, i, len(hvs)-1), false
				copy(result, hvs[:i])
			} else {
				set[hv] = true
			}
		} else if hv != nil && !set[hv] {
			result, set[hv] = append(result, hv), true
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
