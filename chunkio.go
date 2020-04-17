// Copyright (c) 2019 Jason T. Lenz.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

/*
Package chunkio provides functionality for transparently reading a subset of a
stream containing a user defined ending byte sequence.  When the byte sequence
is reached an EOF is returned.  This sub stream can be accessed or passed to
other routines as standard Reader objects.
*/
package chunkio

import (
	"bytes"
	"errors"
	"io"
)

const (
	minKeyLength = 1
	bufAdd       = 4096 // buffAdd plus key length = buffer size
)

var ErrInvalidKey = errors.New("chunkio: invalid key definition")

// Reader implements chunkio functionality wrapped around an io.Reader object
type Reader struct {
	rd      io.Reader    // Underlying Reader
	key     []byte       // key that delineates end of chunk
	buf     bytes.Buffer // A buffer to provide "read ahead" ability
	bufSize int          // The target buffer size
	err     error        // Current error state of chunkio Reader
	ierr    error        // Current error state of underlying Reader
	scan    int          // Number of bytes in buffer that have already been scanned for key
	found   bool         // True if key exists in buffer. Position is in scan in that case
}

// NewReader creates a new chunk reader.
func NewReader(rd io.Reader) *Reader {
	return &Reader{
		rd:      rd,
		key:     nil,
		buf:     bytes.Buffer{},
		bufSize: 0,
		err:     nil,
		ierr:    nil,
		scan:    0,
		found:   false,
	}
}

// GetKey returns the key for the current active chunkio stream.
func (c *Reader) GetKey() []byte {
	return c.key
}

// GetErr returns the error status for the current active chunkio stream.
func (c *Reader) GetErr() error {
	return c.err
}

// SetKey updates the search key.  The search key can also be cleared by
// providing a nil key.
func (c *Reader) SetKey(key []byte) error {
	if key == nil {
		c.key = key
		return nil
	}
	if len(key) < minKeyLength {
		return ErrInvalidKey
	}
	c.key = key
	c.bufSize = bufAdd + len(c.key)
	if c.buf.Cap() < c.bufSize {
		c.buf.Grow(c.bufSize - c.buf.Cap())
	}
	c.scan = 0
	return nil
}

// Reset puts the chunkio stream back into a readable state.  This can be used
// when the end of a chunk is reached to enable reading the next chunk.
func (c *Reader) Reset() {
	if c.buf.Len() == 0 && c.ierr != nil {
		c.err = io.ErrUnexpectedEOF
	} else {
		c.err = nil
	}
	c.scan = 0
	c.found = false
}

func (c *Reader) readScanned(p []byte) (int, error) {
	var n int

	if c.scan > len(p) {
		n, _ = c.buf.Read(p)
	} else {
		n, _ = c.buf.Read(p[:c.scan])
	}
	c.scan = c.scan - n
	if n > 0 && c.scan >= 0 {
		return n, nil
	} else {
		return 0, io.ErrUnexpectedEOF
	}
}

func (c *Reader) readEOF() (int, error) {
	// Discard key from input stream
	r := make([]byte, len(c.key))
	n, err := c.buf.Read(r)
	if n != len(c.key) || err != nil {
		panic("Error: Unexpected error in chunkio.readEOF()")
	}
	// Set / return EOF
	c.err = io.EOF
	return 0, io.EOF
}

func (c *Reader) bufFill() error {
	for c.buf.Len() < c.bufSize {
		t := make([]byte, c.bufSize-c.buf.Len())
		n, err := c.rd.Read(t)
		c.buf.Write(t[:n])
		if err != nil {
			return err
		}
	}
	return nil
}

// Read implements the standard Reader interface allowing chunkio to be used
// anywhere a standard Reader can be used.  Read puts data into p.  It returns
// the number of bytes read into p.  The bytes are taken from at most one read
// on the underlying Reader, hence n may be less than len(p).  When the key is
// reached (EOF for the stream chunk), the count will be zero and err will be
// io.EOF.  If the key has been set to nil, the Read function performs exactly
// like the underlying stream Read function (no key scanning).
func (c *Reader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if c.err != nil {
		return 0, c.err
	}
	if c.key == nil {
		if c.buf.Len() > 0 {
			return c.buf.Read(p)
		}
		return c.rd.Read(p)
	}
	if c.scan > 0 {
		return c.readScanned(p)
	}
	if c.found {
		return c.readEOF()
	}
	c.ierr = c.bufFill()
	pos := bytes.Index(c.buf.Bytes(), c.key)
	switch pos {
	case -1:
		if c.ierr != nil {
			// Reached input EOF w/o key
			c.scan = c.buf.Len()
			c.err = io.ErrUnexpectedEOF
			return c.readScanned(p)
		}
		c.scan = c.buf.Len() - len(c.key)
		if c.scan <= 0 {
			panic("Error: Unexpected error in chunkio.Read()")
		}
		return c.readScanned(p)
	case 0:
		c.scan = 0
		c.found = true
		return c.readEOF()
	default:
		c.scan = pos
		c.found = true
		return c.readScanned(p)
	}
}
