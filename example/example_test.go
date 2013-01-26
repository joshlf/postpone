// Copyright 2012 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/joshlf13/postpone"
	"io"
	"os"
)

func main() {
	fmt.Println("files/NewFile:")
	rdr := postpone.NewFile("files/NewFile")
	io.Copy(os.Stdout, rdr)
	fmt.Println()

	fmt.Println("files/NewFilePre:")
	rdr = postpone.NewFilePre("files/NewFilePre")
	io.Copy(os.Stdout, rdr)
	fmt.Println()

	fmt.Println("files/NewReader:")
	f, err := os.Open("files/NewReader")
	if err == nil {
		rdr = postpone.NewReader(f, true)
		io.Copy(os.Stdout, rdr)
	}
	fmt.Println()

	fmt.Println("files/nothing:")
	bad_rdr := postpone.NewFile("files/nothing")
	io.Copy(os.Stdout, bad_rdr)
}
