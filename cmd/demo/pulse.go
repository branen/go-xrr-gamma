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
	"github.com/branen/go-xrr-gamma/gamma/animate"
	"log"
	"math"
	"os"
	"os/signal"
	"time"
)

type Pulse struct{}

func init()                    { cmds = append(cmds, Pulse{}) }
func (cmd Pulse) Name() string { return "pulse" }

func (cmd Pulse) Help(args []string) {
	fmt.Printf("%s %s\n", os.Args[0], args[0])
	fmt.Println("Make the screen pulse.")
	return
}

func (cmd Pulse) Main(args []string) {
	var (
		cl         *gamma.Client
		errChan    <-chan error
		cancelFunc animate.CancelFunc
		sigChan    chan os.Signal = make(chan os.Signal)
		err        error
	)
	if cl, err = gamma.NewClient(); err != nil {
		log.Fatal(err)
	}
	signal.Notify(sigChan, os.Interrupt)
	errChan, _, cancelFunc = animate.Animate(cl, pulse)
	for {
		select {
		case err, ok := <-errChan:
			if ok {
				if err != nil {
					log.Fatal(err)
				}
			}
			return
		case _, _ = <-sigChan:
			cancelFunc()
		}
	}
}

func pulse(t time.Duration, baseFn gamma.XferFn, event interface{}) (fn gamma.XferFn, sleepFor time.Duration, exit bool) {
	absStage, position := math.Modf(float64(t) / float64(time.Second) * 2)
	stage := math.Mod(absStage, 4)
	from := func(start, end float64) float64 {
		return start*(1-position) + end*position
	}
	var exp float64
	switch stage {
	case 0:
		exp = from(1, 0.25)
	case 1:
		exp = from(0.25, 1)
	case 2:
		exp = from(1, 4)
	case 3:
		exp = from(4, 1)
	}
	return gamma.PowerFn(exp), 0, t >= 12*time.Second
}
