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

package local

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/donyori/gogo/errors"
)

// ErrPatternHasPathSeparator is an error indicating that the pattern,
// prefix, or suffix contains a path separator.
var ErrPatternHasPathSeparator = errors.AutoNewWithStrategy("pattern/prefix/suffix contains path separator",
	errors.PrefixFullPkgName, 0)

// Tmp creates and opens a new temporary file in the directory dir,
// with specified permission perm (before umask), for reading and writing.
//
// The filename is generated by concatenating prefix,
// a random string, and suffix.
// Both prefix and suffix must not contains a path separator.
// If prefix or suffix contains a path separator,
// it returns a nil f and an error ErrPatternHasPathSeparator.
// (To test whether err is ErrPatternHasPathSeparator, use function errors.Is.)
//
// If dir is empty, it will use the default directory for temporary files
// (as returned by os.TempDir) instead.
//
// Calling this function simultaneously will not choose the same file.
//
// The client can use f.Name() to find the path of the file.
// The client is responsible for removing the file when no longer needed.
func Tmp(dir, prefix, suffix string, perm fs.FileMode) (f *os.File, err error) {
	err = checkTmpPrefixAndSuffix(prefix, suffix)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	if dir == "" {
		dir = os.TempDir()
	}
	prefix = filepath.Join(dir, prefix)
	r := uint32(time.Now().UnixNano() + int64(os.Getpid())*1000)
	for try := 0; try < 100; try++ {
		f, err = os.OpenFile(prefix+strconv.FormatUint(uint64(r), 36)+suffix,
			os.O_RDWR|os.O_CREATE|os.O_EXCL, perm)
		if err == nil || !errors.Is(err, os.ErrExist) {
			return f, errors.AutoWrap(err)
		}
		r = r*1664525 + 1013904223 // constants from Numerical Recipes, for a linear congruential generator
	}
	return f, errors.AutoWrap(err)
}

// TmpDir creates a new temporary directory in the directory dir,
// with specified permission perm (before umask),
// and returns the path of the new directory.
//
// The new directory's name is generated by concatenating prefix,
// a random string, and suffix.
// Both prefix and suffix must not contains a path separator.
// If prefix or suffix contains a path separator,
// it returns an empty name and an error ErrPatternHasPathSeparator.
// (To test whether err is ErrPatternHasPathSeparator, use function errors.Is.)
//
// If dir is empty, it will use the default directory for temporary files
// (as returned by os.TempDir) instead.
//
// Calling this function simultaneously will not choose the same directory.
//
// The client is responsible for removing the directory when no longer needed.
func TmpDir(dir, prefix, suffix string, perm fs.FileMode) (name string, err error) {
	err = checkTmpPrefixAndSuffix(prefix, suffix)
	if err != nil {
		return "", errors.AutoWrap(err)
	}
	if dir == "" {
		dir = os.TempDir()
	}
	prefix = filepath.Join(dir, prefix)
	r := uint32(time.Now().UnixNano() + int64(os.Getpid())*1000)
	for try := 0; try < 100; try++ {
		name = prefix + strconv.FormatUint(uint64(r), 36) + suffix
		err = os.Mkdir(name, perm)
		if err == nil {
			return name, errors.AutoWrap(err)
		}
		if errors.Is(err, os.ErrNotExist) {
			// It may because dir doesn't exist. Try to report dir doesn't exist.
			if _, err := os.Lstat(dir); errors.Is(err, os.ErrNotExist) {
				return "", errors.AutoWrap(err)
			}
		}
		if !errors.Is(err, os.ErrExist) {
			return "", errors.AutoWrap(err)
		}
		r = r*1664525 + 1013904223 // constants from Numerical Recipes, for a linear congruential generator
	}
	return "", errors.AutoWrap(err)
}

// checkTmpPrefixAndSuffix checks whether prefix or suffix has a path separator.
//
// If a path separator is in prefix or suffix,
// it returns ErrPatternHasPathSeparator.
func checkTmpPrefixAndSuffix(prefix, suffix string) error {
	for i := 0; i < len(prefix); i++ {
		if os.IsPathSeparator(prefix[i]) {
			return ErrPatternHasPathSeparator // Don't wrap the error here.
		}
	}
	for i := 0; i < len(suffix); i++ {
		if os.IsPathSeparator(suffix[i]) {
			return ErrPatternHasPathSeparator // Don't wrap the error here.
		}
	}
	return nil
}