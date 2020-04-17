// Copyright (c) 2019 Jason T. Lenz.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package chunkio_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"git.lenzplace.org/lenzj/chunkio"
	"strings"
	"testing"
)

func ExampleUppercase() {
	example := []byte("the quick {U}brown fox jumps{L} over the lazy dog")
	cio := chunkio.NewReader(bytes.NewReader(example))
	cio.SetKey([]byte("{U}"))
	s1, _ := ioutil.ReadAll(cio)
	cio.Reset()
	cio.SetKey([]byte("{L}"))
	s2, _ := ioutil.ReadAll(cio)
	cio.Reset()
	s3, _ := ioutil.ReadAll(cio)
	fmt.Print(string(s1) + strings.ToUpper(string(s2)) + string(s3))
	// Output: the quick BROWN FOX JUMPS over the lazy dog
}

func TestShortNewReader(t *testing.T) {
	c := chunkio.NewReader(bytes.NewReader([]byte("")))
	if c.GetKey() != nil {
		t.Errorf("GetKey. Expected \"%v\", got \"%v\"", nil, c.GetKey())
	}
	if c.GetErr() != nil {
		t.Errorf("GetErr. Expected \"%v\", got \"%v\"", nil, c.GetErr())
	}
	resid, err := ioutil.ReadAll(c)
	if bytes.Compare(resid, []byte("")) != 0 {
		t.Errorf("Read raw Bufio. Expected \"%v\", got \"%v\"", []byte(""), resid)
	}
	if err != nil {
		t.Errorf("Read raw Bufio. Expected error code \"%v\", got \"%v\"", nil, err)
	}
}

func TestShortSetKey(t *testing.T) {
	c := chunkio.NewReader(bytes.NewReader([]byte("")))
	c.SetKey([]byte("123"))
	if bytes.Compare(c.GetKey(), []byte("123")) != 0 {
		t.Errorf("SetKey. Expected key to be \"%v\", got \"%v\"", []byte("123"), c.GetKey())
	}
}

func TestShortRead(t *testing.T) {
	cases := []struct {
		desc string
		in   []byte
		key1 []byte
		out1 []byte
		err1 error
		key2 []byte
		out2 []byte
		err2 error
	}{
		{
			desc: "No key detected",
			key1: []byte("123456"),
			in:   []byte("---\nauthor : Jason\n---\nqwerty"),
			out1: []byte("---\nauthor : Jason\n---\nqwerty"),
			err1: io.ErrUnexpectedEOF,
			key2: []byte("123456"),
			out2: []byte(""),
			err2: io.ErrUnexpectedEOF,
		},
		{
			desc: "Simple key detected at start",
			in:   []byte("---\nauthor : Jason\n---\nqwerty"),
			key1: []byte("---\n"),
			out1: []byte(""),
			err1: nil,
			key2: []byte("---\n"),
			out2: []byte("author : Jason\n"),
			err2: nil,
		},
		{
			desc: "Simple key detected mid stream",
			in:   []byte("ytrewq\n---\nauthor : Jason"),
			key1: []byte("---\n"),
			out1: []byte("ytrewq\n"),
			err1: nil,
			key2: []byte("---\n"),
			out2: []byte("author : Jason"),
			err2: io.ErrUnexpectedEOF,
		},
		{
			desc: "Simple key detected mid stream then set key to nil",
			in:   []byte("ytrewq\n---\nauthor : Jason"),
			key1: []byte("---\n"),
			out1: []byte("ytrewq\n"),
			err1: nil,
			key2: nil,
			out2: []byte("author : Jason"),
			err2: nil,
		},
		{
			desc: "Empty input stream",
			in:   []byte(""),
			key1: []byte("---\n"),
			out1: []byte(""),
			err1: io.ErrUnexpectedEOF,
			key2: nil,
			out2: []byte(""),
			err2: io.ErrUnexpectedEOF,
		},
	}
	for _, c := range cases {
		rd := chunkio.NewReader(bytes.NewReader(c.in))
		rd.SetKey(c.key1)
		out1, err1 := ioutil.ReadAll(rd)
		if c.err1 != err1 {
			t.Errorf("Case %q. Expected stream read error=\"%v\", got \"%v\"", c.desc, c.err1, err1)
		}
		if bytes.Compare(c.out1, out1) != 0 {
			t.Errorf("Case %q. Expected stream read=%q, got %q", c.desc, c.out1, out1)
		}
		rd.Reset()
		rd.SetKey(c.key2)
		out2, err2 := ioutil.ReadAll(rd)
		if c.err2 != err2 {
			t.Errorf("Case %q. Expected 2nd stream read error=\"%v\", got \"%v\"", c.desc, c.err2, err2)
		}
		if bytes.Compare(c.out2, out2) != 0 {
			t.Errorf("Case %q. Expected 2nd stream read=%q, got %q", c.desc, c.out2, out2)
		}
	}
}

// Test each input length from zero up to a large number.
func TestLongReadSizes(t *testing.T) {
	for i := 0; i < 20000; i++ {
		rd := chunkio.NewReader(bytes.NewReader(append(bytes.Repeat([]byte("X"), i), []byte(";;;")...)))
		rd.SetKey([]byte(";;;"))
		out, err := ioutil.ReadAll(rd)
		if len(out) != i {
			t.Errorf("Failed.  Read byte sequence size %d instead of %d", len(out), i)
		}
		if err != nil {
			t.Errorf("Failed.  Attempted to read byte sequence size %q and got error %v", i, err)
		}
	}
}

// Test a large number of cases randomizing the following:
// * key (length and contents)
// * input data before key (length and contents)
// * additional "garbage" data after key (length and contents)
func TestLongReadRand(t *testing.T) {
	const (
		numcycles = 8000
		maxinput  = 4096 * 10
		maxread   = maxinput / 10
		maxkey    = maxinput / 12
	)

	var (
		key     []byte // Key sequence to use
		input   []byte // Source sequence
		output  []byte // Destination sequence
		garbage []byte // Garbage sequence at end
		readbuf []byte // Buffer to store each Read call in
		num     int    // Number of bytes from Read
		err     error  // Error return code from Read
		cio     *chunkio.Reader
	)

	for i := 0; i < numcycles; i++ {
		output = nil
		key = make([]byte, rand.Intn(maxkey-1)+1)
		rand.Read(key)

		input = make([]byte, rand.Intn(maxinput-1)+1)
		rand.Read(input)

		// Make sure key doesn't exist in input stream
		for {
			p := bytes.Index(input, key)
			if p == -1 {
				break
			}
			input[p] = input[p] + 1
		}

		garbage = make([]byte, rand.Intn(maxinput-1)+1)
		rand.Read(garbage)

		cio = chunkio.NewReader(bytes.NewReader(append(append(input, key...), garbage...)))
		cio.SetKey(key)
		for {
			readbuf = make([]byte, rand.Intn(maxread-1)+1)
			num, err = cio.Read(readbuf)
			if err == nil {
				output = append(output, readbuf[:num]...)
			} else {
				break
			}
		}
		if bytes.Compare(input, output) != 0 {
			t.Errorf("Cycle %d failed. Input and output differ!", i)
		}
		if len(output) != len(input) {
			t.Errorf("Cycle %d failed. Read size should be %d but was %d", i, len(input), len(output))
		}
	}
}
