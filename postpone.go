// Copyright 2012 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The postpone package provides an io.ReadSeeker wrapper, and various functions
// which handle readers with different postponements such as open on read and
// preload to RAM
package postpone

import (
	"bytes"
	"github.com/joshlf13/errlist"
	"io"
	"io/ioutil"
	"os"
)

// Postpone fulfills the io.ReadSeeker interface.
type Postpone struct {
	r      io.Reader
	rs     io.ReadSeeker
	getr   func() (io.Reader, error)
	getrs  func() (io.ReadSeeker, error)
	err    error
	loaded bool
	c      bool
	bad    bool
}

// NewFile takes a filepath, and returns an io.ReadSeeker.
// This ReadSeeker will wait to open the file until the
// first call to either Read or Seek.
func NewFile(file string, close bool) *Postpone {
	return NewFunc(func() (io.ReadSeeker, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		return f, nil
	}, false)
}

// NewFilePre takes a filepath, and returns an io.ReadSeeker.
// This ReadSeeker will wait to open the file until the
// first call to either Read or Seek. Upon this first call,
// the entire contents of file, or as much as is available,
// will be read into an internal buffer, and the file
// will be closed.
func NewFilePre(file string) *Postpone {
	return NewFuncPre(func() (io.Reader, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		return f, nil
	}, true)
}

// NewFunc takes a function which returns an io.ReadSeeker.
// This is so the given resource doesn't have to be
// opened until it is needed. Upon the first Read
// or Seek call, r is called, the resultant ReadSeeker
// is stored, and r is discarded.
//
// If r returns an io.Closer, c optionally tells
// the reader to close the io.Closer once it's been
// read from.
func NewFunc(r func() (io.ReadSeeker, error), c bool) *Postpone {
	return &Postpone{nil, nil, nil, r, nil, false, c, false}
}

// NewFuncPre is identical to NewFunc except it takes
// a reader rather than a ReadSeeker, and upon the first 
// Read or Seek call, it not only retreives the reader, 
// it also preloads all of the data from the reader into 
// an internal buffer, and discards the reader.
//
// If r returns an io.Closer, c optionally tells
// the reader to close the io.Closer once it's been
// read from.
func NewFuncPre(r func() (io.Reader, error), c bool) *Postpone {
	return &Postpone{nil, nil, r, nil, nil, false, c, false}
}

// NewReader takes an io.Reader and, upon the first
// call to Read or Seek, preloads all available data
// into an internal buffer, and discards the reader
//
// If r is an io.Closer, c optionally tells
// the reader to close r once it's been read from.
func NewReader(r io.Reader, c bool) *Postpone {
	return &Postpone{r, nil, nil, nil, nil, false, c, false}
}

// Load performs the same operation which would
// normally be performed during the first call
// to Read or Seek
func (p *Postpone) Load() {
	p.retreive()
}

// Loaded returns whether or not Load, Read,
// or Seek has been called yet.
func (p *Postpone) Loaded() bool {
	return p.loaded
}

func (p *Postpone) Read(buf []byte) (int, error) {
	if !p.loaded {
		p.retreive()
	}
	if p.bad {
		return 0, p.err
	}
	i, err := p.rs.Read(buf)
	return i, errlist.NewError(err).AddError(p.err).Err()
}

func (p *Postpone) Seek(offset int64, whence int) (int64, error) {
	if !p.loaded {
		p.retreive()
	}
	if p.bad {
		return 0, p.err
	}
	i, err := p.rs.Seek(offset, whence)
	return i, errlist.NewError(err).AddError(p.err).Err()
}

func (p *Postpone) retreive() {
	if p.getr != nil {
		var r io.Reader
		r, p.err = p.getr()
		p.getr = nil
		if r == nil || p.err != nil {
			p.bad = true
		} else {
			buf, err := ioutil.ReadAll(r)
			p.err = err
			p.rs = bytes.NewReader(buf)
		}
		c, ok := r.(io.Closer)
		if ok {
			c.Close()
		}
	} else if p.getrs != nil {
		p.rs, p.err = p.getrs()
		p.getrs = nil
		if p.rs == nil {
			p.bad = true
		}
	} else {
		var buf []byte
		if p.r == nil {
			p.bad = true
		} else {
			buf, p.err = ioutil.ReadAll(p.r)
			p.rs = bytes.NewReader(buf)
			c, ok := p.r.(io.Closer)
			if ok {
				c.Close()
			}
		}
	}
	p.loaded = true
}
