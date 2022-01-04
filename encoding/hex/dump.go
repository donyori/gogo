// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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
	"github.com/donyori/gogo/inout"
)

// DumpConfig is a configuration for hexadecimal dumping.
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
}

// dumpCfgLineNotValid returns true if dumping with cfg won't separate lines,
// or add prefixes or suffixes.
func (cfg *DumpConfig) dumpCfgLineNotValid() bool {
	return cfg == nil || cfg.BlockLen <= 0 || cfg.BlocksPerLine <= 0
}

// DumpToString returns a string including a hexadecimal dump of src with cfg.
func DumpToString(src []byte, cfg *DumpConfig) string {
	var builder strings.Builder
	_, err := DumpTo(&builder, src, cfg)
	if err != nil { // It should not happen.
		panic(errors.AutoWrap(err))
	}
	return builder.String()
}

// DumpTo dumps src with cfg to w.
//
// It returns the number of bytes dumped from src,
// and any write error encountered.
func DumpTo(w io.Writer, src []byte, cfg *DumpConfig) (n int, err error) {
	d := NewDumper(w, cfg)
	defer func() {
		closeErr := d.Close()
		// If err != nil, closeErr must be err, so closeErr is ignored in this case.
		if err == nil {
			err = closeErr
		}
		if err != nil {
			err = errors.AutoWrapSkip(err, 1) // skip = 1 to skip the inner function
		}
	}()
	return d.Write(src)
}

// Dumper is a device to dump input data, with specified configuration,
// to the destination writer.
//
// It combines io.Writer, io.ByteWriter, io.ReaderFrom.
// All methods dump input data to the destination writer.
//
// It contains the method Close and the method Flush.
// The client should flush the dumper after use,
// and close it when it is no longer needed.
type Dumper interface {
	io.Writer
	io.ByteWriter
	io.ReaderFrom
	io.Closer
	inout.Flusher

	// DumpDst returns the destination writer of this dumper.
	// It returns nil if the dumper is closed successfully.
	DumpDst() io.Writer

	// DumpCfg returns a copy of the configuration for hexadecimal dumping
	// used by this dumper.
	DumpCfg() *DumpConfig
}

// dumper is an implementation of interface Dumper.
type dumper struct {
	w            io.Writer
	cfg          DumpConfig
	bytesPerLine int // If !dumpCfgLineNotValid(cfg), = BlockLen * BlocksPerLine. Otherwise, = 0.
	err          error
	line         []byte // Buffer to store current line, only valid when cfg.SuffixFn != nil && !dumpCfgLineNotValid(cfg).
	buf          []byte // Buffer to write blocks in one line, exclude the line separator, prefix and suffix.
	idx          int    // Index of unused buffer.
	written      int    // Index of already written to w.
	sepCd        int    // Countdown for writing a separator, negative if formatCfgNotValid(cfg).
	lineCd       int    // Countdown for writing a separator, negative if dumpCfgLineNotValid(cfg).
	used         bool   // Indicate whether anything has been asked to be written to the dumper or not.
}

// NewDumper creates a dumper to dump hexadecimal characters to w.
// The format is specified by cfg.
//
// It panics if w is nil.
//
// Note that the created dumper will keep a copy of cfg,
// so you can't change its config after create it.
func NewDumper(w io.Writer, cfg *DumpConfig) Dumper {
	if w == nil {
		panic(errors.AutoMsg("w is nil"))
	}
	d := &dumper{w: w}
	if cfg != nil {
		d.cfg = *cfg
	}
	if d.cfg.formatCfgNotValid() {
		d.sepCd = -1
	} else {
		d.sepCd = d.cfg.BlockLen
	}
	if d.cfg.dumpCfgLineNotValid() {
		d.bytesPerLine = 0
		d.buf = make([]byte, 1024)
		d.lineCd = -1
	} else {
		d.bytesPerLine = d.cfg.BlockLen * d.cfg.BlocksPerLine
		d.buf = make([]byte, (d.cfg.BlockLen*2+len(d.cfg.Sep))*d.cfg.BlocksPerLine-len(d.cfg.Sep))
		d.lineCd = d.bytesPerLine
		if d.cfg.SuffixFn != nil {
			d.line = make([]byte, d.lineCd)
		}
	}
	return d
}

// Write dumps p to the destination writer.
//
// It conforms to interface io.Writer.
func (d *dumper) Write(p []byte) (n int, err error) {
	if d.err != nil {
		return 0, d.err
	}
	ht := lowercaseHexTable
	if d.cfg.Upper {
		ht = uppercaseHexTable
	}
	n, err = d.write(ht, p)
	d.err = errors.AutoWrap(err)
	return n, d.err
}

// WriteByte dumps c to the destination writer.
//
// It conforms to interface io.ByteWriter.
func (d *dumper) WriteByte(c byte) error {
	if d.err != nil {
		return d.err
	}
	ht := lowercaseHexTable
	if d.cfg.Upper {
		ht = uppercaseHexTable
	}
	d.err = errors.AutoWrap(d.writeByte(ht, c))
	return d.err
}

// ReadFrom dumps data read from r to the destination writer.
//
// It conforms to interface io.ReaderFrom.
func (d *dumper) ReadFrom(r io.Reader) (n int64, err error) {
	if d.err != nil {
		return 0, d.err
	}
	ht := lowercaseHexTable
	if d.cfg.Upper {
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

// Close outputs all buffered content to the destination writer
// and then closes the dumper.
//
// It returns any write error encountered.
//
// All operations, except the method Close, on a closed dumper
// will report the error ErrWriterClosed.
// Calling the method Close on a closed dumper will perform nothing
// and report no error.
func (d *dumper) Close() error {
	if errors.Is(d.err, inout.ErrWriterClosed) {
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
	if !d.cfg.dumpCfgLineNotValid() {
		var appendSuffix bool
		if d.used {
			appendSuffix = d.lineCd < d.bytesPerLine
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
	d.err = errors.AutoWrap(inout.ErrWriterClosed)
	d.buf = nil
	d.idx = 0
	d.written = 0
	d.sepCd = 0
	d.lineCd = 0
	d.used = false
	return nil
}

// Flush outputs all buffered content to the destination writer.
//
// It returns any write error encountered.
func (d *dumper) Flush() error {
	if d.err != nil {
		return d.err
	}
	d.err = errors.AutoWrap(d.flush())
	return d.err
}

// DumpDst returns the destination writer of this dumper.
// It returns nil if the dumper is closed successfully.
func (d *dumper) DumpDst() io.Writer {
	return d.w
}

// DumpCfg returns a copy of the configuration for hexadecimal dumping
// used by this dumper.
func (d *dumper) DumpCfg() *DumpConfig {
	cfg := new(DumpConfig)
	*cfg = d.cfg
	return cfg
}

// flush writes all data in the buffer to d.w,
// and maintains indexes d.idx and d.written.
//
// It returns any write error encountered.
//
// Caller should guarantee that d != nil and d.w != nil.
func (d *dumper) flush() error {
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

// writeByte dumps b to the destination writer.
//
// Caller should guarantee that d != nil and d.w != nil.
func (d *dumper) writeByte(ht string, b byte) error {
	d.used = true
	if d.lineCd == d.bytesPerLine && d.cfg.PrefixFn != nil {
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
		d.lineCd = d.bytesPerLine
	}
	return nil
}

// write dumps p to the destination writer.
//
// Caller should guarantee that d != nil and d.w != nil.
func (d *dumper) write(ht string, p []byte) (n int, err error) {
	for _, b := range p {
		err = d.writeByte(ht, b)
		if err != nil {
			return
		}
		n++
	}
	return
}
