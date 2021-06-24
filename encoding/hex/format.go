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

package hex

import (
	"fmt"
	"io"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/inout"
)

// FormatConfig is a configuration for hexadecimal formatting.
type FormatConfig struct {
	Upper bool   // Use uppercase or not.
	Sep   string // Separator.

	// Block size, i.e., the number of bytes (before hexadecimal encoding)
	// between every two separators.
	// Non-positive values for no separators.
	BlockLen int
}

// formatCfgNotValid returns true if formatting with cfg
// won't insert any separators.
func (cfg *FormatConfig) formatCfgNotValid() bool {
	return cfg == nil || cfg.Sep == "" || cfg.BlockLen <= 0
}

// FormattedLen returns the length of formatting n source bytes with cfg.
func FormattedLen(n int, cfg *FormatConfig) int {
	if n == 0 {
		return 0
	}
	if cfg.formatCfgNotValid() {
		return EncodedLen(n)
	}
	return (n-1)/cfg.BlockLen*(cfg.BlockLen*2+len(cfg.Sep)) + ((n-1)%cfg.BlockLen+1)*2
}

// FormattedLen64 returns the length of formatting n source bytes with cfg.
func FormattedLen64(n int64, cfg *FormatConfig) int64 {
	if n == 0 {
		return 0
	}
	if cfg.formatCfgNotValid() {
		return EncodedLen64(n)
	}
	blockLen := int64(cfg.BlockLen)
	return (n-1)/blockLen*(blockLen*2+int64(len(cfg.Sep))) + ((n-1)%blockLen+1)*2
}

// Format outputs hexadecimal representation of src
// in the format specified by cfg to dst.
//
// It panics if dst doesn't have enough space to hold the formatting result.
// The client should guarantee that len(dst) >= FormattedLen(len(src), cfg).
//
// It returns the number of bytes written into dst,
// exactly FormattedLen(len(src), cfg).
//
// Format(dst, src, nil) is equivalent to Encode(dst, src)
// in official package encoding/hex.
func Format(dst, src []byte, cfg *FormatConfig) int {
	if reqLen := FormattedLen(len(src), cfg); reqLen > len(dst) {
		panic(errors.AutoMsg(fmt.Sprintf("dst is too small, length: %d, required: %d", len(dst), reqLen)))
	}
	return format(dst, src, cfg)
}

// FormatToString returns hexadecimal representation of src
// in the format specified by cfg.
//
// FormatToString(src, nil) is equivalent to EncodeToString(src)
// in official package encoding/hex.
func FormatToString(src []byte, cfg *FormatConfig) string {
	dst := make([]byte, FormattedLen(len(src), cfg))
	format(dst, src, cfg)
	return string(dst)
}

// FormatTo outputs hexadecimal representation of src
// in the format specified by cfg to w.
//
// It returns the number of bytes formatted from src,
// and any write error encountered.
func FormatTo(w io.Writer, src []byte, cfg *FormatConfig) (n int, err error) {
	dst := make([]byte, FormattedLen(len(src), cfg))
	format(dst, src, cfg)
	n, err = w.Write(dst)
	if n == len(dst) {
		n = len(src)
	} else {
		n = ParsedLen(n, cfg)
	}
	return n, errors.AutoWrap(err)
}

// Formatter is a device to write hexadecimal representation of input data,
// in the specified format, to the destination writer.
//
// It combines io.Writer, io.ByteWriter, io.ReaderFrom.
// All the methods format input data and output to the destination writer.
//
// It contains the method Close and the method Flush.
// The client should flush the formatter after use,
// and close it when it is no longer needed.
type Formatter interface {
	io.Writer
	io.ByteWriter
	io.ReaderFrom
	io.Closer
	inout.Flusher

	// FormatDst returns the destination writer of this formatter.
	// It returns nil if the formatter is closed successfully.
	FormatDst() io.Writer

	// FormatCfg returns a copy of the configuration for hexadecimal formatting
	// used by this formatter.
	FormatCfg() *FormatConfig
}

// formatter is an implementation of interface Formatter.
type formatter struct {
	w       io.Writer
	cfg     FormatConfig
	err     error
	bufp    *[]byte // Pointer of the buffer to store formatted characters.
	idx     int     // Index of unused buffer.
	written int     // Index of already written to w.
	sepCd   int     // Countdown for writing a separator, negative if formatCfgNotValid(cfg).
}

// NewFormatter creates a formatter to write hexadecimal characters
// with separators to w.
// The format is specified by cfg.
//
// It panics if w is nil.
//
// Note that the created formatter will keep a copy of cfg,
// so you can't change its config after creating it.
func NewFormatter(w io.Writer, cfg *FormatConfig) Formatter {
	if w == nil {
		panic(errors.AutoMsg("w is nil"))
	}
	f := &formatter{w: w}
	if cfg != nil {
		f.cfg = *cfg
	}
	if f.cfg.formatCfgNotValid() {
		f.sepCd = -1
	} else {
		f.sepCd = f.cfg.BlockLen
	}
	return f
}

// Write writes hexadecimal representation of p,
// in the specified format, to the destination writer.
//
// It conforms to interface io.Writer.
func (f *formatter) Write(p []byte) (n int, err error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.bufp == nil {
		f.bufp = formatBufferPool.Get().(*[]byte)
		f.idx, f.written = 0, 0
	}
	ht := lowercaseHexTable
	if f.cfg.Upper {
		ht = uppercaseHexTable
	}
	n, err = f.write(ht, p)
	f.err = errors.AutoWrap(err)
	return n, f.err
}

// WriteByte writes hexadecimal representation of c,
// in the specified format, to the destination writer.
//
// It conforms to interface io.ByteWriter.
func (f *formatter) WriteByte(c byte) error {
	if f.err != nil {
		return f.err
	}
	if f.bufp == nil {
		f.bufp = formatBufferPool.Get().(*[]byte)
		f.idx, f.written = 0, 0
	}
	ht := lowercaseHexTable
	if f.cfg.Upper {
		ht = uppercaseHexTable
	}
	f.err = errors.AutoWrap(f.writeByte(ht, c))
	return f.err
}

// ReadFrom writes hexadecimal representation of data read from r,
// in the specified format, to the destination writer.
//
// It conforms to interface io.ReaderFrom.
func (f *formatter) ReadFrom(r io.Reader) (n int64, err error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.bufp == nil {
		f.bufp = formatBufferPool.Get().(*[]byte)
		f.idx, f.written = 0, 0
	}
	ht := lowercaseHexTable
	if f.cfg.Upper {
		ht = uppercaseHexTable
	}
	bufp := sourceBufferPool.Get().(*[]byte)
	defer sourceBufferPool.Put(bufp)
	buf := *bufp
	for {
		readLen, readErr := r.Read(buf)
		var writeErr error
		if readLen > 0 {
			n += int64(readLen)
			_, writeErr = f.write(ht, buf[:readLen])
		}
		err = readErr
		if errors.Is(err, io.EOF) {
			err = nil
		}
		if readErr != nil {
			if err != nil {
				return n, errors.AutoWrap(err) // don't record a read error to f.err
			}
			f.err = errors.AutoWrap(writeErr)
			return n, f.err
		} else if writeErr != nil {
			f.err = errors.AutoWrap(writeErr)
			return n, f.err
		}
	}
}

// Close outputs all buffered content to the destination writer
// and then closes the formatter.
//
// It returns any write error encountered.
//
// All operations, except the method Close, on a closed formatter
// will report the error ErrWriterClosed.
// Calling the method Close on a closed formatter will perform nothing
// and report no error.
func (f *formatter) Close() error {
	if errors.Is(f.err, inout.ErrWriterClosed) {
		return nil
	}
	if f.err != nil {
		return f.err
	}
	err := f.flush()
	if err != nil {
		f.err = errors.AutoWrap(err)
		return f.err
	}
	f.w = nil
	f.err = errors.AutoWrap(inout.ErrWriterClosed)
	f.sepCd = 0
	return nil
}

// Flush outputs all buffered content to the destination writer.
//
// It returns any write error encountered.
func (f *formatter) Flush() error {
	if f.err != nil {
		return f.err
	}
	f.err = errors.AutoWrap(f.flush())
	return f.err
}

// FormatDst returns the destination writer of this formatter.
// It returns nil if the formatter is closed successfully.
func (f *formatter) FormatDst() io.Writer {
	return f.w
}

// FormatCfg returns a copy of the configuration for hexadecimal formatting
// used by this formatter.
func (f *formatter) FormatCfg() *FormatConfig {
	cfg := new(FormatConfig)
	*cfg = f.cfg
	return cfg
}

// flush writes all data in the buffer to f.w.
// If no error occurs during writing to f.w,
// it puts the buffer to formatBufferPool.
//
// It returns any write error encountered.
//
// Caller should guarantee that f != nil and f.w != nil.
func (f *formatter) flush() error {
	if f.bufp == nil {
		return nil
	}
	n, err := f.w.Write((*f.bufp)[f.written:f.idx])
	f.written += n
	if err == nil {
		formatBufferPool.Put(f.bufp)
		f.bufp = nil
		f.idx, f.written = 0, 0
	}
	return err
}

// flushAndGetBuffer calls the method flush first.
// If no error occurs, it gets a new buffer from formatBufferPool.
//
// It returns any write error encountered.
//
// Caller should guarantee that f != nil and f.w != nil.
func (f *formatter) flushAndGetBuffer() error {
	err := f.flush()
	if err == nil && f.bufp == nil {
		f.bufp = formatBufferPool.Get().(*[]byte)
	}
	return err
}

// writeByte writes hexadecimal representation of c,
// in the specified format, to the destination writer.
//
// Caller should guarantee that f != nil, f.w != nil and f.bufp != nil.
func (f *formatter) writeByte(ht string, b byte) error {
	buf := *f.bufp
	if f.sepCd == 0 {
		if len(buf)-f.idx < len(f.cfg.Sep) {
			err := f.flushAndGetBuffer()
			if err != nil {
				return err
			}
		}
		f.idx += copy(buf[f.idx:], f.cfg.Sep)
		f.sepCd = f.cfg.BlockLen
	}
	if len(buf)-f.idx < 2 {
		err := f.flushAndGetBuffer()
		if err != nil {
			return err
		}
	}
	buf[f.idx] = ht[b>>4]
	buf[f.idx+1] = ht[b&0x0f]
	f.idx += 2
	if f.sepCd > 0 {
		f.sepCd--
	}
	return nil
}

// write writes hexadecimal representation of p,
// in the specified format, to the destination writer.
//
// Caller should guarantee that f != nil, f.w != nil and f.bufp != nil.
func (f *formatter) write(ht string, p []byte) (n int, err error) {
	for _, b := range p {
		err = f.writeByte(ht, b)
		if err != nil {
			return
		}
		n++
	}
	return
}

// format is an implementation of function Format,
// without checking the length of dst.
//
// Caller should guarantee that len(dst) >= FormattedLen(len(src), cfg).
func format(dst, src []byte, cfg *FormatConfig) int {
	if cfg.formatCfgNotValid() {
		var upper bool
		if cfg != nil {
			upper = cfg.Upper
		}
		return Encode(dst, src, upper)
	}
	ht := lowercaseHexTable
	if cfg.Upper {
		ht = uppercaseHexTable
	}
	var n int
	for i, b := range src {
		if n > 0 && i%cfg.BlockLen == 0 {
			n += copy(dst[n:], cfg.Sep)
		}
		dst[n] = ht[b>>4]
		dst[n+1] = ht[b&0x0f]
		n += 2
	}
	return n
}
