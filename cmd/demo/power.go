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

type Power struct{}

func init()                    { cmds = append(cmds, Power{}) }
func (cmd Power) Name() string { return "power" }

func (cmd Power) Help(args []string) {
	fmt.Printf("%s %s EXPONENT\n", os.Args[0], args[0])
	fmt.Println("Apply a power law function with a coefficient of 1.")
	return
}

func (cmd Power) Main(args []string) {
	var (
		cl  *gamma.Client
		s   *gamma.Session
		err error
		pow float64
	)
	if len(args) < 2 {
		cmd.Help(args)
		return
	}
	{
		n, err := fmt.Sscanf(args[1], "%f", &pow)
		if err != nil {
			log.Fatal(err)
		}
		if n != 1 {
			log.Fatal("Error parsing arguments.")
		}
	}
	if cl, err = gamma.NewClient(); err != nil {
		log.Fatal(err)
	}
	if s, err = cl.NewSession(); err != nil {
		log.Fatal(err)
	}
	s.SetGamma(gamma.PowerFn(pow))
	return
}
