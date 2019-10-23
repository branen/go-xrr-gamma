// Copyright 2019 Branen Salmon
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"github.com/branen/go-xrr-gamma/gamma"
	"log"
	"os"
)

type Dim struct{}

func init()                  { cmds = append(cmds, Dim{}) }
func (cmd Dim) Name() string { return "dim" }

func (cmd Dim) Help(args []string) {
	fmt.Printf("%s %s\n", os.Args[0], args[0])
	fmt.Println("Dim by 50%.")
	return
}

func (cmd Dim) Main(args []string) {
	var (
		cl     *gamma.Client
		s      *gamma.Session
		err    error
		baseFn gamma.XferFn
	)
	if cl, err = gamma.NewClient(); err != nil {
		log.Fatal(err)
	}
	if s, err = cl.NewSession(); err != nil {
		log.Fatal(err)
	}
	if lut, err := s.GetLookupTable(); err != nil {
		log.Fatal(err)
	} else {
		baseFn = lut.XferFn()
	}
	s.SetGamma(gamma.DimFn(0.5).Mul(baseFn))
	return
}
