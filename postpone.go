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
	"fmt"
	"io/ioutil"
	"os"
)

type postpone struct {
	r      io.Reader
	rs     io.ReadSeeker
	getr   func() (io.Reader, error)
	getrs  func() (io.ReadSeeker, error)
	err    error
	loaded bool
	bad    bool
}

// NewFile takes a filepath, and returns an io.ReadSeeker.
// This ReadSeeker will wait to open the file until the
// first call to either Read or Seek.
func NewFile(file string) io.ReadSeeker {
	return NewFunc(func() (io.ReadSeeker, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		return f, nil
	})
}

// NewFilePre takes a filepath, and returns an io.ReadSeeker.
// This ReadSeeker will wait to open the file until the
// first call to either Read or Seek. Upon this first call,
// the entire contents of file, or as much as is available,
// will be read into an internal buffer, and the file
// will be closed.
func NewFilePre(file string) io.ReadSeeker {
	return NewFuncPre(func() (io.Reader, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		return f, nil
	})
}

// NewFunc takes a function which returns an io.ReadSeeker.
// This is so the given resource doesn't have to be
// opened until it is needed. Upon the first loaded
// or Seek call, r is called, the resultant loadedSeeker
// is stored, and r is discarded.
func NewFunc(r func() (io.ReadSeeker, error)) io.ReadSeeker {
	return &postpone{nil, nil, nil, r, nil, false, false}
}

// NewFuncPre is identical to NewFunc except it takes
// a reader rather than a loadedSeeker, and upon the first 
// loaded or Seek call, it not only retreives the reader, 
// it also preloads all of the data from the reader into 
// an internal buffer, and discards the reader.
func NewFuncPre(r func() (io.Reader, error)) io.ReadSeeker {
	return &postpone{nil, nil, r, nil, nil, false, false}
}

// NewReader takes an io.Reader and, upon the first
// call to loaded or Seek, preloads all available data
// into an internal buffer, and discards the reader
func NewReader(r io.Reader) io.ReadSeeker {
	return &postpone{r, nil, nil, nil, nil, false, false}
}

func (p *postpone) Read(buf []byte) (int, error) {
	if !p.loaded {
		p.retreive()
	}
	if p.bad {
		return 0, p.err
	}
	i, err := p.rs.Read(buf)
	fmt.Println(err == io.EOF)
	err = errlist.NewError(err).AddError(p.err).Err()
	fmt.Println(err == io.EOF)
	return i, errlist.NewError(err).AddError(p.err).Err()
}

func (p *postpone) Seek(offset int64, whence int) (int64, error) {
	if !p.loaded {
		p.retreive()
	}
	if p.bad {
		return 0, p.err
	}
	i, err := p.rs.Seek(offset, whence)
	return i, errlist.NewError(err).AddError(p.err).Err()
}

func (p *postpone) retreive() {
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
		}
	}
	p.loaded = true
}
