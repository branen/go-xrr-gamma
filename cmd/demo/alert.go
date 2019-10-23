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
	"github.com/branen/go-xrr-gamma/gamma/animate/alert"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Alert struct{}

func init()                    { cmds = append(cmds, Alert{}) }
func (cmd Alert) Name() string { return "alert" }

func (cmd Alert) Help(args []string) {
	fmt.Printf("%s %s\n", os.Args[0], args[0])
	fmt.Println("Demo an \"alert\" effect with smooth transitions and event-driven accents.")
	return
}

func (cmd Alert) Main(args []string) {
	fmt.Printf("Send SIGUSR1 to pid %d to \"strobe\" the screen, SIGUSR2 to \"warble\" the screen, or SIGINT to exit.\n", os.Getpid())
	var (
		cl         *gamma.Client
		errChan    <-chan error
		cancelFunc animate.CancelFunc
		eventChan  animate.EventChan
		sigChan    chan os.Signal = make(chan os.Signal)
		err        error
		exiting    bool
	)
	if cl, err = gamma.NewClient(); err != nil {
		log.Fatal(err)
	}
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2)
	errChan, eventChan, cancelFunc = animate.Animate(cl, alert.Xft())
	for {
		select {
		case err, ok := <-errChan:
			if ok {
				if err != nil {
					log.Fatal(err)
				}
			}
			return
		case c := <-sigChan:
			switch c {
			case syscall.SIGINT:
				eventChan <- alert.Exit
				if exiting {
					cancelFunc()
				}
				exiting = true
			case syscall.SIGUSR1:
				eventChan <- alert.Strobe
			case syscall.SIGUSR2:
				eventChan <- alert.Warble
			}
		}
	}
}
