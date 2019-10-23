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

package alert_test

import (
	"github.com/branen/go-xrr-gamma/gamma"
	"github.com/branen/go-xrr-gamma/gamma/animate"
	"github.com/branen/go-xrr-gamma/gamma/animate/alert"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Example() {
	var (
		cl  *gamma.Client
		err error

		sigChan   chan os.Signal
		errChan   <-chan error
		eventChan animate.EventChan
	)

	// Connect to XRandR.
	if cl, err = gamma.NewClient(); err != nil {
		log.Fatal(err)
	}
	defer cl.Close()

	// Start the animation goroutine.
	// We don't use cancelFunc, since alert.Xft provides an Exit event.
	errChan, eventChan, _ = animate.Animate(cl, alert.Xft())

	// Wait and handle signals until the animation goroutine exits.
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGHUP)
	for {
		select {
		// Exit when the animation goroutine exits.
		case err, ok := <-errChan:
			if ok {
				if err != nil {
					log.Fatal(err)
				}
			}
			return
		case c := <-sigChan:
			switch c {
			// Exit the animation on SIGINT
			case syscall.SIGINT:
				eventChan <- alert.Exit
			// Strobe the animation on SIGHUP
			case syscall.SIGHUP:
				eventChan <- alert.Strobe
			}
		}
	}
}
