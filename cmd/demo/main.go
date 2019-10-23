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
	"os"
)

type Command interface {
	Name() string
	Main(args []string)
	Help(args []string)
}

var cmds []Command = make([]Command, 0)

type Help struct{}

func init()                       { cmds = append(cmds, Help{}) }
func (_ Help) Name() string       { return "help" }
func (_ Help) Help(args []string) { return }
func (_ Help) Main(args []string) {
	if len(args) > 1 {
		for _, cmd := range cmds {
			if args[1] == cmd.Name() {
				cmd.Help(args[1:len(args)])
				return
			}
		}
	}
	for _, cmd := range cmds {
		if cmd.Name() != "help" {
			fmt.Printf("%s %s\n", os.Args[0], cmd.Name())
		} else {
			fmt.Printf("%s %s ...\n", os.Args[0], cmd.Name())
		}
	}
	return
}

func main() {
	if len(os.Args) < 2 {
		Help{}.Main(nil)
		os.Exit(1)
	}
	for _, cmd := range cmds {
		if os.Args[1] == cmd.Name() {
			cmd.Main(os.Args[1:len(os.Args)])
			os.Exit(0)
		}
	}
	Help{}.Main(nil)
	os.Exit(1)
}
