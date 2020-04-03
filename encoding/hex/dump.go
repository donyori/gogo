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

import (
	"io"
	"strings"

	"github.com/donyori/gogo/errors"
	myio "github.com/donyori/gogo/io"
)

// Configuration for hexadecimal dumping.
type DumpConfig struct {
	FormatConfig
	LineSep       string // Line separator.
	BlocksPerLine int    // The number of blocks per line, non-positive values for all blocks in one line.

	// Function to generate a prefix of one line.
	// Only valid when BlockLen > 0 and BlocksPerLine > 0.
	PrefixFn func() []byte
	// Function to generate a suffix of one line.
	// The input is the line before hexadecimal encoding.
	// Only valid when BlockLen > 0 and BlocksPerLine > 0.
	SuffixFn func(line []byte) []byte

	bytesPerLine int // If !dumpCfgLineNotValid(dc), = BlockLen * BlocksPerLine. Otherwise, = 0.
}

// Return true if dumping with cfg won't separate lines, or add prefixes or suffixes.
func dumpCfgLineNotValid(cfg *DumpConfig) bool {
	return cfg == nil || cfg.BlockLen <= 0 || cfg.BlocksPerLine <= 0
}

// Return a string containing a hexadecimal dump of src with cfg.
func Dump(src []byte, cfg *DumpConfig) string {
	var builder strings.Builder
	_, err := DumpTo(&builder, src, cfg)
	if err != nil { // It should not happen.
		panic(errors.AutoWrap(err))
	}
	return builder.String()
}

// Dump src with cfg to w.
// It returns the bytes dumped from src, and any encountered error.
func DumpTo(w io.Writer, src []byte, cfg *DumpConfig) (n int, err error) {
	d := NewDumper(w, cfg)
	defer func() {
		closeErr := d.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			err = errors.AutoWrapSkip(err, 1) // skip = 1 to skip the inner function
		}
	}()
	return d.Write(src)
}

// A dumper to dump hexadecimal characters to w.
// Dumper should be closed after use.
type Dumper struct {
	w       io.Writer
	cfg     DumpConfig
	err     error
	line    []byte // Buffer to store current line, only valid when cfg.SuffixFn != nil && !dumpCfgLineNotValid(cfg).
	buf     []byte // Buffer to write blocks in one line, exclude the line separator, prefix and suffix.
	idx     int    // Index of unused buffer.
	written int    // Index of already written to w.
	sepCd   int    // Countdown for writing a separator, negative if formatCfgNotValid(cfg).
	lineCd  int    // Countdown for writing a separator, negative if dumpCfgLineNotValid(cfg).
	used    bool   // Indicate whether anything has been asked to be written to the dumper or not.
}

// Create a dumper to dump hexadecimal characters to w.
// Note that the created dumper will keep a copy of cfg,
// so you can't change its config after create it.
// Dumper should be closed after use.
func NewDumper(w io.Writer, cfg *DumpConfig) *Dumper {
	if w == nil {
		panic(errors.AutoMsg("w is nil"))
	}
	d := &Dumper{w: w}
	if cfg != nil {
		d.cfg = *cfg
	}
	if formatCfgNotValid(&d.cfg.FormatConfig) {
		d.sepCd = -1
	} else {
		d.sepCd = d.cfg.BlockLen
	}
	if dumpCfgLineNotValid(&d.cfg) {
		d.cfg.bytesPerLine = 0
		d.buf = make([]byte, 1024)
		d.lineCd = -1
	} else {
		d.cfg.bytesPerLine = d.cfg.BlockLen * d.cfg.BlocksPerLine
		d.buf = make([]byte, (d.cfg.BlockLen*2+len(d.cfg.Sep))*d.cfg.BlocksPerLine-len(d.cfg.Sep))
		d.lineCd = d.cfg.bytesPerLine
		if d.cfg.SuffixFn != nil {
			d.line = make([]byte, d.lineCd)
		}
	}
	return d
}

func (d *Dumper) Write(p []byte) (n int, err error) {
	if d.err != nil {
		return 0, d.err
	}
	n, err = d.write(getHexTable(d.cfg.Upper), p)
	d.err = errors.AutoWrap(err)
	return n, d.err
}

// Flush the buffer.
// It reports no error if dumper is nil.
func (d *Dumper) Flush() error {
	if d == nil {
		return nil
	}
	if d.err != nil {
		return d.err
	}
	d.err = errors.AutoWrap(d.flush())
	return d.err
}

// Flush the buffer.
// It reports no error if dumper is nil.
func (d *Dumper) Close() error {
	if d == nil || errors.Is(d.err, myio.ErrWriterClosed) {
		return nil
	}
	if d.err != nil {
		return d.err
	}
	err := d.flush()
	if err != nil {
		d.err = errors.AutoWrap(err)
		return d.err
	}
	if !dumpCfgLineNotValid(&d.cfg) {
		appendSuffix := false
		if d.used {
			appendSuffix = d.lineCd < d.cfg.bytesPerLine
		} else {
			appendSuffix = true
			if d.cfg.PrefixFn != nil {
				prefix := d.cfg.PrefixFn()
				if len(prefix) > 0 {
					_, err = d.w.Write(prefix)
					if err != nil {
						d.err = errors.AutoWrap(err)
						return d.err
					}
				}
			}
		}
		if appendSuffix {
			if d.cfg.SuffixFn != nil {
				suffix := d.cfg.SuffixFn(d.line[:len(d.line)-d.lineCd])
				if len(suffix) > 0 {
					_, err = d.w.Write(suffix)
					if err != nil {
						d.err = errors.AutoWrap(err)
						return d.err
					}
				}
			}
			_, err = d.w.Write([]byte(d.cfg.LineSep))
			if err != nil {
				d.err = errors.AutoWrap(err)
				return d.err
			}
		}
	}
	d.w = nil
	d.line = nil
	d.err = errors.AutoWrap(myio.ErrWriterClosed)
	d.buf = nil
	d.idx = 0
	d.written = 0
	d.sepCd = 0
	d.lineCd = 0
	d.used = false
	return nil
}

func (d *Dumper) WriteByte(c byte) error {
	if d.err != nil {
		return d.err
	}
	d.err = errors.AutoWrap(d.writeByte(getHexTable(d.cfg.Upper), c))
	return d.err
}

func (d *Dumper) ReadFrom(r io.Reader) (n int64, err error) {
	if d.err != nil {
		return 0, d.err
	}
	ht := getHexTable(d.cfg.Upper)
	bufp := chunkPool.Get().(*[]byte)
	defer chunkPool.Put(bufp)
	buf := *bufp
	var readLen int
	var readErr, writeErr error
	for {
		readLen, readErr = r.Read(buf)
		if readLen > 0 {
			n += int64(readLen)
			_, writeErr = d.write(ht, buf[:readLen])
		}
		err = readErr
		if errors.Is(err, io.EOF) {
			err = nil
		}
		if readErr != nil {
			if err != nil {
				return n, errors.AutoWrap(err) // don't record a read error to d.err
			}
			d.err = errors.AutoWrap(writeErr)
			return n, d.err
		} else if writeErr != nil {
			d.err = errors.AutoWrap(writeErr)
			return n, d.err
		}
	}
}

// Caller should guarantee that d != nil and d.w != nil.
func (d *Dumper) flush() error {
	if d.idx == d.written {
		d.idx, d.written = 0, 0
		return nil
	}
	n, err := d.w.Write(d.buf[d.written:d.idx])
	d.written += n
	if err == nil {
		d.idx, d.written = 0, 0
	}
	return err
}

// Caller should guarantee that d != nil and d.w != nil.
// ht is getHexTable(d.cfg.Upper).
func (d *Dumper) writeByte(ht string, b byte) error {
	d.used = true
	if d.lineCd == d.cfg.bytesPerLine && d.cfg.PrefixFn != nil {
		prefix := d.cfg.PrefixFn()
		if len(prefix) > 0 {
			_, err := d.w.Write(prefix)
			if err != nil {
				return err
			}
		}
	}
	if d.sepCd == 0 {
		if len(d.buf)-d.idx < len(d.cfg.Sep) {
			err := d.flush()
			if err != nil {
				return err
			}
		}
		d.idx += copy(d.buf[d.idx:], d.cfg.Sep)
		d.sepCd = d.cfg.BlockLen
	}

	if len(d.buf)-d.idx < 2 {
		err := d.flush()
		if err != nil {
			return err
		}
	}
	if d.line != nil {
		d.line[len(d.line)-d.lineCd] = b
	}
	d.buf[d.idx] = ht[b>>4]
	d.buf[d.idx+1] = ht[b&0x0f]
	d.idx += 2

	if d.sepCd > 0 {
		d.sepCd--
	}
	if d.lineCd > 0 {
		d.lineCd--
	}
	if d.lineCd == 0 {
		err := d.flush()
		if err != nil {
			return err
		}
		if d.cfg.SuffixFn != nil {
			suffix := d.cfg.SuffixFn(d.line)
			if len(suffix) > 0 {
				_, err = d.w.Write(suffix)
				if err != nil {
					return err
				}
			}
		}
		_, err = d.w.Write([]byte(d.cfg.LineSep))
		if err != nil {
			return err
		}
		if d.sepCd >= 0 {
			d.sepCd = d.cfg.BlockLen
		}
		d.lineCd = d.cfg.bytesPerLine
	}
	return nil
}

// Caller should guarantee that d != nil and d.w != nil.
// ht is getHexTable(d.cfg.Upper).
func (d *Dumper) write(ht string, p []byte) (n int, err error) {
	for _, b := range p {
		err = d.writeByte(ht, b)
		if err != nil {
			return
		}
		n++
	}
	return
}
