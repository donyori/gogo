// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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
	"math/bits"
	"strings"

	"github.com/donyori/gogo/algorithm/mathalgo"
	"github.com/donyori/gogo/encoding/hex"
	"github.com/donyori/gogo/errors"
)

// Checksum calculates hash checksums of the specified file,
// and returns the result in hexadecimal representation and
// any read error encountered.
//
// If the file is a directory, Checksum reports ErrIsDir
// and returns nil checksums.
// (To test whether err is ErrIsDir, use function errors.Is.)
//
// To ensure that this function can work as expected,
// the input file must be ready to be read from the beginning and
// must not be operated by anyone else during the call to this function.
//
// closeFile indicates whether this function should close the file.
// If closeFile is false, the client is responsible for closing file after use.
// If closeFile is true, file is closed by this function.
//
// upper indicates whether to use uppercase in hexadecimal representation.
//
// newHash is a list of functions that create new hash functions
// (e.g., crypto/sha256.New, crypto.SHA256.New).
//
// The length of the returned checksums is the same as that of newHash.
// The hash result of newHash[i] is checksums[i], encoded in hexadecimal.
// In particular, if newHash[i] is nil or returns nil,
// checksums[i] is an empty string.
// If len(newHash) is 0, checksums is nil.
//
// This function panics if file is nil.
func Checksum(
	file fs.File,
	closeFile bool,
	upper bool,
	newHash ...func() hash.Hash,
) (checksums []string, err error) {
	if file == nil {
		panic(errors.AutoMsg("file is nil"))
	} else if closeFile {
		defer func(f fs.File) {
			_ = f.Close() // ignore error
		}(file)
	}

	err = checksumTestDir(file)
	if err != nil || len(newHash) == 0 {
		return nil, errors.AutoWrap(err)
	}

	checksums = make([]string, len(newHash))
	hs := make([]hash.Hash, len(newHash))
	ws := make([]io.Writer, 0, len(newHash))
	bs := make([]uint, 0, len(newHash))

	for i := range newHash {
		if newHash[i] != nil {
			hs[i] = newHash[i]()
			if hs[i] != nil {
				ws = append(ws, hs[i])
				bs = append(bs, uint(hs[i].BlockSize())) //gosec:disable G115 -- the block size is nonnegative
			}
		}
	}

	if len(ws) == 0 {
		return
	}

	w := ws[0]
	bufSize := bs[0]

	if len(ws) > 1 {
		w = io.MultiWriter(ws...)
		bufSize = mathalgo.LCM(bs...) // make bufSize a multiple of the block sizes
	}

	const NumBitOf4096 int = 13
	if bufSize == 0 {
		// Act as a safeguard for the hash.Hash
		// whose BlockSize returns 0.
		bufSize = 5120 // = (2^10) * 5
	} else if shift := NumBitOf4096 - bits.Len(bufSize); shift > 0 {
		bufSize <<= shift // make bufSize at least 4096
	}

	_, err = io.CopyBuffer(w, file, make([]byte, bufSize))
	if err != nil {
		return nil, errors.AutoWrap(err)
	}

	for i := range hs {
		if hs[i] != nil {
			checksums[i] = hex.EncodeToString(hs[i].Sum(nil), upper)
		}
	}

	return
}

// checksumTestDir is a subprocess of Checksum
// that tests whether the file is a directory.
func checksumTestDir(file fs.File) error {
	info, err := file.Stat()
	if err != nil {
		return errors.AutoWrap(err)
	} else if info.IsDir() {
		return errors.AutoWrap(ErrIsDir)
	}

	return nil
}

// ChecksumFromFS calculates hash checksums of the file opened from fsys
// by specified name, and returns the result in hexadecimal representation
// and any error encountered during opening and reading the file.
//
// If the file is a directory, ChecksumFromFS reports ErrIsDir
// and returns nil checksums.
// (To test whether err is ErrIsDir, use function errors.Is.)
//
// upper indicates whether to use uppercase in hexadecimal representation.
//
// newHash is a list of functions that create new hash functions
// (e.g., crypto/sha256.New, crypto.SHA256.New).
//
// The length of the returned checksums is the same as that of newHash.
// The hash result of newHash[i] is checksums[i], encoded in hexadecimal.
// In particular, if newHash[i] is nil or returns nil,
// checksums[i] is an empty string.
// If len(newHash) is 0, checksums is nil.
//
// This function panics if fsys is nil.
func ChecksumFromFS(
	fsys fs.FS,
	name string,
	upper bool,
	newHash ...func() hash.Hash,
) (checksums []string, err error) {
	if fsys == nil {
		panic(errors.AutoMsg("fsys is nil"))
	}

	f, err := fsys.Open(name)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}

	return Checksum(f, true, upper, newHash...)
}

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
	} else if strings.IndexFunc(prefixHex, func(r rune) bool {
		return r < '0' || r > '9' && r < 'A' || r > 'F' && r < 'a' || r > 'f'
	}) >= 0 {
		panic(errors.AutoMsg(fmt.Sprintf(
			"prefixHex (%q) is not hexadecimal", prefixHex)))
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
// If closeFile is true and file is not nil, file is closed by this function.
//
// It returns true if the file is not nil and can be read,
// and matches all HashVerifier in hv
// (nil and duplicate HashVerifier are ignored).
// In particular, it returns true if there is no non-nil HashVerifier in hv.
// In this case, the file is not read.
//
// Note that VerifyChecksum does not reset the hash state of anyone in hv.
// The client should use new HashVerifier returned by NewHashVerifier or
// call the Reset method of HashVerifier before calling this function if needed.
func VerifyChecksum(file fs.File, closeFile bool, hv ...HashVerifier) bool {
	if file == nil {
		return false
	} else if closeFile {
		defer func(f fs.File) {
			_ = f.Close() // ignore error
		}(file)
	}

	info, err := file.Stat()
	if err != nil || info.IsDir() {
		return false
	}

	hv = nonNilDeduplicatedHashVerifiers(hv)
	if len(hv) == 0 {
		return true
	}

	ws := make([]io.Writer, len(hv))
	bs := make([]uint, len(hv))

	for i := range hv {
		ws[i] = hv[i]
		bs[i] = uint(hv[i].BlockSize()) //gosec:disable G115 -- the block size is nonnegative
	}

	w := ws[0]
	bufSize := bs[0]

	if len(ws) > 1 {
		w = io.MultiWriter(ws...)
		bufSize = mathalgo.LCM(bs...) // make bufSize a multiple of the block sizes
	}

	const NumBitOf4096 int = 13
	if bufSize == 0 {
		// Act as a safeguard for the hash.Hash
		// whose BlockSize returns 0.
		bufSize = 5120 // = (2^10) * 5
	} else if shift := NumBitOf4096 - bits.Len(bufSize); shift > 0 {
		bufSize <<= shift // make bufSize at least 4096
	}

	_, err = io.CopyBuffer(w, file, make([]byte, bufSize))
	if err != nil {
		return false
	}

	for _, v := range hv {
		if !v.Match() {
			return false
		}
	}

	return true
}

// nonNilDeduplicatedHashVerifiers returns all
// non-nil deduplicated HashVerifier in hv.
//
// If there is no nil or duplicate HashVerifier in hv, it returns hv itself.
// Otherwise, it copies all non-nil deduplicated HashVerifier from hv to
// a new slice and returns that slice.
// The content of hv is not modified in both cases.
//
// If there is no non-nil HashVerifier in hv, it returns nil.
func nonNilDeduplicatedHashVerifiers(hv []HashVerifier) []HashVerifier {
	result := hv
	set := make(map[HashVerifier]struct{}, 1+len(hv))
	set[nil] = struct{}{}
	original := true

	for i, v := range hv {
		if original {
			if _, ok := set[v]; ok {
				result, original = make([]HashVerifier, i, len(hv)-1), false
				copy(result, hv[:i])
			} else {
				set[v] = struct{}{}
			}
		} else if _, ok := set[v]; !ok {
			result, set[v] = append(result, v), struct{}{}
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// VerifyChecksumFromFS verifies a file by hash checksum,
// where the file is opened from fsys by specified name.
//
// It returns true if fsys is not nil,
// and the file can be read and matches all HashVerifier in hv
// (nil and duplicate HashVerifier are ignored).
// In particular, it returns true if there is no non-nil HashVerifier
// in hv and the file can be opened for reading.
// In this case, the file is not read.
//
// Note that VerifyChecksumFromFS does not reset
// the hash state of anyone in hv.
// The client should use new HashVerifier returned by NewHashVerifier or
// call the Reset method of HashVerifier before calling this function if needed.
func VerifyChecksumFromFS(fsys fs.FS, name string, hv ...HashVerifier) bool {
	if fsys == nil {
		return false
	}

	f, err := fsys.Open(name)
	if err != nil {
		return false
	}

	return VerifyChecksum(f, true, hv...)
}
