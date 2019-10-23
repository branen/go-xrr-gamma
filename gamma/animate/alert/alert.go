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

// Package alert provides Xft, an event-responsive animate.XferFnAtTime
// that turns the screen a soft red; two emphasis events (one gentle,
// one bold); and an exit event that causes the animation to fade out
// smoothly.
package alert

import (
	"github.com/branen/go-xrr-gamma/gamma"
	"github.com/branen/go-xrr-gamma/gamma/animate"
	"math"
	"time"
)

// Cmd specifies a command for a running Alert animation.
type Cmd int

const (
	noCmd Cmd = iota
	// Gently modulates a running animation to remind the user it's there.
	Warble
	// Adds a moment of emphasis to a running animation.
	Strobe
	// Exits the animation with a smooth fade-out.
	Exit
)

type effect struct {
	start time.Duration
	apply func(since time.Duration, in float64) (out float64, done bool)
}

func warble(since time.Duration, in float64) (out float64, done bool) {
	const period = 125 * time.Millisecond
	const cycles = 5
	const weight = 1.0 / 12.0
	const duration = period * cycles
	if since > duration {
		out = in
		done = true
	} else {
		_, pos := math.Modf(float64(since) / float64(period))
		pow := math.Cos(2*math.Pi*pos)/2 + 0.5
		out = 1 - ((1 - in) * ((1 - weight) + weight*pow))
	}
	return
}

func strobe(since time.Duration, in float64) (out float64, done bool) {
	const duration = 1250 * time.Millisecond
	if since > duration {
		out = in
		done = true
	} else {
		pos := float64(since) / float64(duration)
		out = 1 - ((1 - in) * (0.5 + 0.5*pos))
	}
	return
}

// Xft returns an animate.XferFnAtTime instance that accepts events of type Cmd
// through animate.Animate's EventChan.
func Xft() animate.XferFnAtTime {
	type stageT int
	const (
		enter stageT = iota
		static
		exit
	)
	const enterExitDuration = 250 * time.Millisecond
	var (
		stage      stageT
		stageStart time.Duration
		sinceStage time.Duration
		strength   float64
	)
	var effects []effect = make([]effect, 0, 16)

	var (
		cmd            Cmd
		effectStrength float64
		rCmp, oCmp     float64
	)

	return func(
		t time.Duration, baseFn gamma.XferFn, event interface{},
	) (
		fn gamma.XferFn, sleepFor time.Duration, exitFlag bool,
	) {
		if event == nil {
			cmd = noCmd
		} else {
			cmd = event.(Cmd)
		}

		setStage := func(s stageT) {
			stage = s
			stageStart = t
			sinceStage = 0
		}

		sinceStage = t - stageStart
		switch cmd {
		case Warble:
			effects = append(effects, effect{t, warble})
		case Strobe:
			effects = append(effects, effect{t, strobe})
		case Exit:
			switch stage {
			case static:
				setStage(exit)
			case enter:
				stage = exit
				sinceStage = enterExitDuration - sinceStage
				stageStart = t - sinceStage
			}
		}
		cmd = noCmd
		switch stage {
		case static:
			strength = 1
			sleepFor = 2 * time.Second
		case enter:
			strength = float64(sinceStage) / float64(
				enterExitDuration)
			sleepFor = 0
			if strength >= 1 {
				strength = 1
				setStage(static)
			}
		case exit:
			strength = 1 - float64(sinceStage)/float64(
				enterExitDuration)
			if strength < 0 {
				strength = 0
				exitFlag = true
			}
			sleepFor = 0
		}

		effectStrength = 0
		for idx := 0; idx < len(effects); {
			var done bool
			effect := effects[idx]
			effectStrength, done = effect.apply(t-effect.start,
				effectStrength)
			if done {
				if idx < len(effects)-1 {
					effects[idx] = effects[len(effects)-1]
				}
				effects = effects[0 : len(effects)-1]
			} else {
				sleepFor = 0
				idx++
			}
		}

		rCmp = 0.2 + effectStrength*0.6
		oCmp = 0 + effectStrength*0.6

		fn = func(ch gamma.Channel, in float64) (out float64) {
			base := baseFn(ch, in)
			var fx float64
			switch ch {
			case gamma.Red:
				fx = base*(1-rCmp) + rCmp
			case gamma.Green, gamma.Blue:
				fx = base * (1 - oCmp)
			}
			out = strength*fx + (1-strength)*base
			return
		}
		return
	}
}
