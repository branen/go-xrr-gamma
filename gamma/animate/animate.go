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

// Package animate provides type XferFnAtTime and function Animate, which
// comprise a simple framework for building and running animated and
// event-responsive gamma transitions.
package animate

import (
	"fmt"
	"github.com/branen/go-xrr-gamma/gamma"
	"time"
)

// ForeignCrtcUpdate is returned when another process updates the CRTC lookup
// tables while an animation is running and the ExitOnForeignUpdate(false)
// Option wasn't passed to Animate.
var ForeignCrtcUpdate error = fmt.Errorf(
	"A foreign process updated the CRTC lookup tables.")

// An XferFnAtTime returns an XferFn fn for use at the specified time t in an
// animation.  Additionally, it is provided baseFn, a XferFn reflecting the
// original state of the CRTC lookup tables (as returned by
// gamma.(Session).GetLookupTable), and event, an interface{} that may be
// passed to the animation loop through the EventChan returned by the Animate
// function.
//
// An XferFnAtTime returns fn, a gamma.XferFn reflecting the desired state for
// time t, sleepFor, a time.Duration that may indicate to the animation loop
// that fn is not expected to change for some period of time, and exit, a bool
// that will cause the animation loop to exit when true.
//
// For a smooth animation, sleepFor should be zero any time fn is changing;
// this will result in the animation loop using its default update rate.  Even
// when sleepFor is non-zero, the animation loop will always be woken
// immediately any time an event is sent on its EventChan.
type XferFnAtTime func(
	t time.Duration, baseFn gamma.XferFn, event interface{},
) (
	fn gamma.XferFn, sleepFor time.Duration, exit bool)

// CancelFunc may be called to cancel a running animation.
type CancelFunc func()

// EventChan may be used to send events to a running animation.
type EventChan chan<- interface{}

type options struct {
	cl     *gamma.Client
	xft    XferFnAtTime
	err    chan error
	cancel chan struct{}
	event  chan interface{}

	startClockBeforeSetup bool
	initialClock          time.Duration
	updateInterval        time.Duration
	exitOnForeignUpdate   bool
	restoreOnExit         bool
}

type Option func(o *options)

// StartClockBeforeSetup, when true, starts the animation clock before the
// animation routine calls gamma.Client.NewSession (which is slow to return).
// This could be useful when restarting an animation.  By default, the clock
// is started after NewSession returns.
func StartClockBeforeSetup(b bool) Option {
	return func(o *options) {
		o.startClockBeforeSetup = b
	}
}

// IntialClock sets the initial animation clock to time t.  This could be
// useful when restarting an animation.  By default, the animation clock starts
// at 0.
func InitialClock(t time.Duration) Option {
	return func(o *options) {
		o.initialClock = t
	}
}

// UpdateInterval sets the minimum interval i at which the CRTCs will
// be reprogrammed.  By default, the CRTCs are updated at most once
// every 33.333ms.  (This is an alternative to UpdatesPerSecond.)
func UpdateInterval(i time.Duration) Option {
	return func(o *options) {
		o.updateInterval = i
	}
}

// UpdatesPerSecond sets the maximum number of times per second that the CRTCs
// will be reprogrammed.  By default, the CRTCs are updated at most 30 times
// per second.  (This is an alternative to UpdateInterval.)
func UpdatesPerSecond(u float64) Option {
	return func(o *options) {
		o.updateInterval = time.Second / time.Duration(u)
	}
}

// ExitOnForeignUpdate, if true, causes the animation to return
// ForeignCrtcUpdate and exit if another process updates the CRTC lookup
// while the animation is running.  This is the default.  If false, the
// animation updates baseFn (see XferFnAtTime) and continues running.
func ExitOnForeignUpdate(b bool) Option {
	return func(o *options) {
		o.exitOnForeignUpdate = b
	}
}

// RestoreOnExit, if true, causes the the baseFn (see XferFnAtTime) to be
// applied to the CRTCs when the animation exits.  This the default.  If false,
// the CRTCs are left with the last state set by the animation loop before
// exit.
func RestoreOnExit(b bool) Option {
	return func(o *options) {
		o.restoreOnExit = b
	}
}

// Animate starts a goroutine that uses XfterFnAtTime xft to update gamma.Client
// cl's CRTC lookup tables.  It returns (<-chan error) e, to which exactly one
// error (or nil) will be written when the animation exits; EventChan ev,
// through which events may be sent to xft; and CancelFunc c, which may be used
// to cancel a running animation.
func Animate(
	cl *gamma.Client, xft XferFnAtTime, opts ...Option,
) (
	e <-chan error, ev EventChan, c CancelFunc,
) {
	err := make(chan error)
	cancel := make(chan struct{})
	o := options{
		cl:     cl,
		xft:    xft,
		err:    err,
		cancel: cancel,
		event:  make(chan interface{}),

		startClockBeforeSetup: false,
		initialClock:          0,
		updateInterval:        time.Second / 30,
		exitOnForeignUpdate:   true,
		restoreOnExit:         true,
	}
	for _, fn := range opts {
		fn(&o)
	}
	e = (<-chan error)(err)
	c = func() CancelFunc {
		var called bool
		return func() {
			if !called {
				close(cancel)
				called = true
			}
			return
		}
	}()
	ev = EventChan(o.event)
	go animate(o)
	return
}

func animate(o options) {
	var (
		s          *gamma.Session
		exit       bool
		err        error
		anchor     time.Time
		thisUpdate time.Time
		lastUpdate time.Time
		extraTime  time.Duration
		sleepFor   time.Duration
		oldLut     gamma.LookupTable
		newLut     gamma.LookupTable
		baseFn     gamma.XferFn
		curFn      gamma.XferFn
		timer      *time.Timer = time.NewTimer(time.Second)
		event      interface{}
	)

	if !timer.Stop() {
		<-timer.C
	}
	if o.startClockBeforeSetup {
		anchor = time.Now().Add(-o.initialClock)
		s, err = o.cl.NewSession()
	} else {
		s, err = o.cl.NewSession()
		anchor = time.Now().Add(-o.initialClock)
	}
	if err != nil {
		goto bail
	}
	defer s.Close()

loop:
	for {
		if exit {
			break loop
		}
		if newLut, err = s.GetLookupTable(); err != nil {
			break loop
		}
		if oldLut.IsZero() {
			baseFn = newLut.XferFn()
		} else {
			if !newLut.Equals(oldLut) {
				if o.exitOnForeignUpdate {
					err = ForeignCrtcUpdate
					o.restoreOnExit = false
					break loop
				} else {
					baseFn = newLut.XferFn()
				}
			}
		}
		curFn, sleepFor, exit = o.xft(
			time.Now().Sub(anchor), baseFn, event)
		s.SetGamma(curFn)
		if oldLut, err = s.GetLookupTable(); err != nil {
			break loop
		}
		thisUpdate = time.Now()
		extraTime = o.updateInterval - thisUpdate.Sub(lastUpdate)
		lastUpdate = thisUpdate

		if sleepFor < extraTime {
			sleepFor = extraTime
		}
		if sleepFor < 0 {
			sleepFor = 0
		}
		timer.Reset(sleepFor)

		event = nil
		select {
		case <-o.cancel:
			break loop
		case event = <-o.event:
			if !timer.Stop() {
				<-timer.C
			}
		case <-timer.C:
		}
	}

	if o.restoreOnExit {
		s.SetGamma(baseFn)
	}
bail:
	if err != nil {
		o.err <- err
	}
	close(o.err)
}
