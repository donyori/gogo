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

package hex

import "io"

// Configuration for hexadecimal formatting.
type FormatConfig struct {
	Upper bool   // Use upper case or not.
	Sep   string // Separator.

	// Block size, i.e., the number of bytes (before hexadecimal encoding)
	// between every two separators. 0 and negative values for no separators.
	BlockLen int
}

func (fc *FormatConfig) notValid() bool {
	return fc == nil || fc.Sep == "" || fc.BlockLen <= 0
}

// Return the length of formatting with cfg of n source bytes.
func FormattedLen(n int, cfg *FormatConfig) int {
	if n == 0 {
		return 0
	}
	if cfg.notValid() {
		return EncodedLen(n)
	}
	return (n-1)/cfg.BlockLen*(cfg.BlockLen*2+len(cfg.Sep)) + ((n-1)%cfg.BlockLen+1)*2
}

// Return the length of formatting with cfg of n source bytes.
func FormattedLen64(n int64, cfg *FormatConfig) int64 {
	if n == 0 {
		return 0
	}
	if cfg.notValid() {
		return EncodedLen64(n)
	}
	blockLen := int64(cfg.BlockLen)
	return (n-1)/blockLen*(blockLen*2+int64(len(cfg.Sep))) + ((n-1)%blockLen+1)*2
}

// Format src into dst, with cfg.
// It returns the number of bytes written into dst, exactly FormattedLen(len(src), cfg).
// Format(dst, src, nil) is equivalent to Encode(dst, src) in official package encoding/hex.
func Format(dst, src []byte, cfg *FormatConfig) int {
	if cfg.notValid() {
		upper := false
		if cfg != nil {
			upper = cfg.Upper
		}
		return Encode(dst, src, upper)
	}
	ht := getHexTable(cfg.Upper)
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

// Return formatting of src, with cfg.
// FormatToString(src, nil) is equivalent to EncodeToString(src) in official package encoding/hex.
func FormatToString(src []byte, cfg *FormatConfig) string {
	dst := make([]byte, FormattedLen(len(src), cfg))
	Format(dst, src, cfg)
	return string(dst)
}

// Write formatting of src with cfg to w.
// It returns the bytes formatted from src, and any encountered error.
func FormatTo(w io.Writer, src []byte, cfg *FormatConfig) (n int, err error) {
	dst := make([]byte, FormattedLen(len(src), cfg))
	Format(dst, src, cfg)
	n, err = w.Write(dst)
	if n == len(dst) {
		n = len(src)
	} else {
		n = ParsedLen(n, cfg)
	}
	return n, err
}

// A formatter to write hexadecimal characters with separators to w.
// Formatter should be closed after use.
type Formatter struct {
	w       io.Writer
	cfg     FormatConfig
	bufp    *[]byte // Pointer of the buffer to store formatted characters.
	idx     int     // Index of unused buffer.
	written int     // Index of already written to w.
	sepCd   int     // Countdown for writing a separator, negative if cfg.notValid().
}

// Create a formatter to write hexadecimal characters with separators to w.
// Note that the created formatter will keep a copy of cfg,
// so you can't change its config after create it.
// Formatter should be closed after use.
func NewFormatter(w io.Writer, cfg *FormatConfig) *Formatter {
	if w == nil {
		panic("hex: NewFormatter: w is nil")
	}
	f := &Formatter{w: w}
	if cfg != nil {
		f.cfg = *cfg
	}
	if f.cfg.notValid() {
		f.sepCd = -1
	} else {
		f.sepCd = f.cfg.BlockLen
	}
	return f
}

func (f *Formatter) Write(p []byte) (n int, err error) {
	if f.bufp == nil {
		f.bufp = formatBufferPool.Get().(*[]byte)
		f.idx, f.written = 0, 0
	}
	return f.write(getHexTable(f.cfg.Upper), p)
}

// Flush the buffer.
// It reports no error if formatter is nil.
func (f *Formatter) Flush() error {
	if f == nil || f.w == nil {
		return nil
	}
	return f.flush()
}

// Flush the buffer.
// It reports no error is formatter is nil.
func (f *Formatter) Close() error {
	if f == nil || f.w == nil {
		return nil
	}
	err := f.flush()
	if err != nil {
		return err
	}
	f.w = nil
	f.sepCd = 0
	return nil
}

func (f *Formatter) WriteByte(c byte) error {
	if f.bufp == nil {
		f.bufp = formatBufferPool.Get().(*[]byte)
		f.idx, f.written = 0, 0
	}
	return f.writeByte(getHexTable(f.cfg.Upper), c)
}

func (f *Formatter) ReadFrom(r io.Reader) (n int64, err error) {
	if f.bufp == nil {
		f.bufp = formatBufferPool.Get().(*[]byte)
		f.idx, f.written = 0, 0
	}
	ht := getHexTable(f.cfg.Upper)
	bufp := chunkPool.Get().(*[]byte)
	defer chunkPool.Put(bufp)
	buf := *bufp
	var readLen int
	var readErr, writeErr error
	for {
		readLen, readErr = r.Read(buf)
		if readLen > 0 {
			n += int64(readLen)
			_, writeErr = f.write(ht, buf[:readLen])
		}
		err = readErr
		if err == io.EOF {
			err = nil
		}
		if readErr != nil {
			if err != nil {
				return n, err
			}
			return n, writeErr
		} else if writeErr != nil {
			return n, writeErr
		}
	}
}

// Caller should guarantee that f != nil and f.w != nil.
func (f *Formatter) flush() error {
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

// Caller should guarantee that f != nil and f.w != nil.
func (f *Formatter) flushAndGetBuffer() error {
	err := f.flush()
	if err == nil && f.bufp == nil {
		f.bufp = formatBufferPool.Get().(*[]byte)
	}
	return err
}

// Caller should guarantee that f != nil, f.w != nil and f.bufp != nil.
// ht is getHexTable(f.cfg.Upper).
func (f *Formatter) writeByte(ht string, b byte) error {
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

// Caller should guarantee that f != nil, f.w != nil and f.bufp != nil.
// ht is getHexTable(f.cfg.Upper).
func (f *Formatter) write(ht string, p []byte) (n int, err error) {
	for _, b := range p {
		err = f.writeByte(ht, b)
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
