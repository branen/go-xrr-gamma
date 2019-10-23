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

package animate_test

import (
	"github.com/branen/go-xrr-gamma/gamma"
	"github.com/branen/go-xrr-gamma/gamma/animate"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// type animate.XferFnAtTime
func blink(
	t time.Duration, baseFn gamma.XferFn, event interface{},
) (
	fn gamma.XferFn, sleepFor time.Duration, exit bool,
) {
	// Exit after five seconds
	if t > 5*time.Second {
		exit = true
	}

	// Exit if we receive any event
	if event != nil {
		exit = true
	}

	// Dim the screen by 50% for one second, every other second
	if position := t % (2 * time.Second); position < time.Second {
		fn = baseFn.Mul(gamma.DimFn(0.5))
	} else {
		fn = baseFn
	}

	// fn won't change until the start of the next second, so don't
	// update until then.
	sleepFor = time.Second - (t % time.Second)
	return
}

func Example() {
	var (
		cl  *gamma.Client
		err error

		sigChan    chan os.Signal
		errChan    <-chan error
		eventChan  animate.EventChan
		cancelFunc animate.CancelFunc
	)

	// Connect to XRandR.
	if cl, err = gamma.NewClient(); err != nil {
		log.Fatal(err)
	}
	defer cl.Close()

	// Start the animation goroutine.
	errChan, eventChan, cancelFunc = animate.Animate(cl, blink)

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
			// Exit the animation via cancelFunc on SIGINT
			case syscall.SIGINT:
				cancelFunc()
			// Send the animation an event on SIGHUP
			case syscall.SIGHUP:
				eventChan <- struct{}{}
			}
		}
	}

}
