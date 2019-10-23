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

// Package gamma provides an interface for querying and programming the
// CRTC lookup tables in terms of simple, hardware independent functions.
package gamma

/*
#cgo LDFLAGS: -lX11 -lXrandr
#include <X11/Xlib.h>
#include <X11/extensions/Xrandr.h>

Window GetDefaultRootWindow(Display *dpy) {
	int screen = DefaultScreen(dpy);
	return RootWindow(dpy, screen);
}
*/
import "C"
import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"unsafe"
)

// Channel specifies a primary additive color channel.
type Channel int

const (
	Red Channel = iota
	Green
	Blue
	_channel_cardinality_
)

/*
XferFn specifies a function that maps all values in [0.0, 0.1] to some values
in [0.0, 0.1] for each Channel Red, Green, and Blue, but which needn't
necessarily be continuous or unique.

(In other words, f ∈ F(ℝ, ℝ), and 0 ≤ f(x) ≤ 1 for all x where 0 ≤ x ≤ 1.)
*/
type XferFn func(ch Channel, in float64) (out float64)

// PowerFn returns the XferFn f(ch, in) = math.Pow(in, exp).  In the context of
// traditional CRT gamma correction, exp is the "gamma correction value."
func PowerFn(exp float64) XferFn {
	return func(ch Channel, in float64) (out float64) {
		return math.Pow(in, math.Max(exp, 0))
	}
}

// DimFn returns the XferFn f(ch, in) = coef * in.
func DimFn(coef float64) XferFn {
	coef = math.Max(math.Min(coef, 1), 0)
	return func(ch Channel, in float64) (out float64) {
		return in * coef
	}
}

// Chain combines two XferFns a and b such that a.Chain(b)(x) = b(a(x)).
func (a XferFn) Chain(b XferFn) XferFn {
	return func(ch Channel, in float64) (out float64) {
		return b(ch, a(ch, in))
	}
}

// Mul combines two XferFns a and b such that a.Mul(b)(x) = a(x) * b(x).
func (a XferFn) Mul(b XferFn) XferFn {
	return func(ch Channel, in float64) (out float64) {
		return a(ch, in) * b(ch, in)
	}
}

type crtcGamma struct {
	crtc  C.RRCrtc
	size  C.int
	gamma *C.XRRCrtcGamma
}

type gammaVector *[65536]C.ushort

/*
Client represents a thread-safe, persistent connection to the XRandR extension.
For most applications, one client may be cached for the lifetime of a process.

Client instances must be created by NewClient--its zero value is not valid for
use.
*/
type Client struct {
	dpy   *C.Display
	root  C.Window
	mutex sync.Mutex
	open  bool
}

func NewClient() (cl *Client, err error) {
	cl = new(Client)
	cl.open = true
	if cl.dpy = C.XOpenDisplay(nil); cl.dpy == nil {
		cl = nil
		err = fmt.Errorf("Could not open X display.")
		return
	}
	runtime.SetFinalizer(cl, func(cl *Client) {
		cl.Close()
	})
	cl.root = C.GetDefaultRootWindow(cl.dpy)
	return
}

// Close "closes" a Client, releasing its underlying resources.  Once a Client
// has been closed, it may not be used again.
//
// Calling Close more than once is a no-op.
func (cl *Client) Close() {
	if cl == nil || !cl.open {
		return
	}
	cl.mutex.Lock()
	defer cl.mutex.Unlock()
	C.XCloseDisplay(cl.dpy)
	cl.open = false
}

func (cl *Client) Closed() bool {
	if cl == nil {
		return true
	}
	cl.mutex.Lock()
	defer cl.mutex.Unlock()
	return !cl.open
}

func (cl *Client) check() {
	if cl.dpy == nil {
		panic("Client instances must be created with NewClient.")
	}
	if !cl.open {
		panic("Client has already been closed.")
	}
}

/*
Session represents a "transaction" with the XRandR extension.

Specifically, a Session corresponds to an underlying XRRScreenResources
instance, which may become stale when displays are hotplugged.  Accordingly, a
Session should live no longer than is required to perform one or more closely
consecutive calls to SetGamma.  If the spacing between calls is long enough to
complete a call to NewSession (i.e. hundreds of milliseconds), then separate
sessions should be used.

Session instances must be created by NewSession--its zero value is not valid
for use.
*/
type Session struct {
	cl    *Client
	res   *C.XRRScreenResources
	crtcs []crtcGamma
	open  bool
}

func (cl *Client) NewSession() (s *Session, err error) {
	cl.check()
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	s = new(Session)
	runtime.SetFinalizer(s, func(s *Session) {
		s.Close()
	})
	s.cl = cl
	s.open = true

	s.res = C.XRRGetScreenResourcesCurrent(s.cl.dpy, s.cl.root)
	if s.res == nil {
		err = fmt.Errorf("Error getting XRRScreenResources.")
		return
	}
	s.crtcs = make([]crtcGamma, s.res.ncrtc, s.res.ncrtc)
	for idx := int(0); C.int(idx) < s.res.ncrtc; idx++ {
		var crtc C.RRCrtc = (*[2 << 32]C.RRCrtc)(unsafe.Pointer(s.res.crtcs))[idx]
		var size C.int = C.XRRGetCrtcGammaSize(s.cl.dpy, crtc)
		if size == 0 {
			err = fmt.Errorf("Error getting CrtcGammaSize.")
			return
		}
		if ptr := C.XRRAllocGamma(size); ptr != nil {
			s.crtcs[idx] = crtcGamma{
				crtc:  crtc,
				size:  size,
				gamma: ptr,
			}
		} else {
			err = fmt.Errorf("Error allocating XRRCrtcGamma.")
			return
		}
	}
	return
}

// Close "closes" a Session, releasing its underlying resources.  Once a Session
// has been closed, it may not be used again.
//
// Calling Close more than once is a no-op.
func (s *Session) Close() {
	if s == nil || !s.open {
		return
	}
	s.cl.check()
	s.cl.mutex.Lock()
	defer s.cl.mutex.Unlock()
	if s.res != nil {
		C.XRRFreeScreenResources(s.res)
	}
	if s.crtcs != nil {
		for _, crtc := range s.crtcs {
			if crtc.gamma != nil {
				C.XRRFreeGamma(crtc.gamma)
			}
		}
	}
	s.open = false
}

func (s *Session) Closed() bool {
	if s == nil {
		return true
	}
	s.cl.check()
	s.cl.mutex.Lock()
	defer s.cl.mutex.Unlock()
	return !s.open
}

func (s *Session) check() {
	if s.cl == nil {
		panic("Session instances must be created with NewSession.")
	}
	if !s.open {
		panic("Session has already been closed.")
	}
}

func forGammaChannels(
	gamma *C.XRRCrtcGamma, fn func(ch Channel, gv gammaVector),
) {
	fn(Red, (gammaVector)(unsafe.Pointer(gamma.red)))
	fn(Green, (gammaVector)(unsafe.Pointer(gamma.green)))
	fn(Blue, (gammaVector)(unsafe.Pointer(gamma.blue)))
}

// SetGamma programs the CRTCs gamma lookup tables using an XferFn.
func (s *Session) SetGamma(fn XferFn) {
	s.cl.check()
	s.cl.mutex.Lock()
	defer s.cl.mutex.Unlock()
	for _, crtcGamma := range s.crtcs {
		forGammaChannels(crtcGamma.gamma, func(ch Channel, gv gammaVector) {
			for idx := C.int(0); idx < crtcGamma.size; idx++ {
				base := float64(idx) / float64(crtcGamma.size)
				gv[idx] = C.ushort(fn(ch, base) * 65535.0)
			}
		})
		C.XRRSetCrtcGamma(s.cl.dpy, crtcGamma.crtc, crtcGamma.gamma)
	}
}

/*
GetLookupTable saves the current gamma lookup tables.

NOTE: The non-primary CRTCs don't always read back correctly on some systems,
so for the time being, GetLookupTable ignores all but the primary CRTC.  This
is subject to change in a future minor release.
*/
func (s *Session) GetLookupTable() (LookupTable, error) {
	s.cl.check()
	s.cl.mutex.Lock()
	defer s.cl.mutex.Unlock()
	var t [_channel_cardinality_][][]C.ushort
	/*
		BUG: The non-primary CRTCs don't always read back correctly.  I
		haven't found any documentation of this behavior, and I haven't
		tried to chase it through the video stack.  Ignoring all but
		the primary CRTC should be sufficient for now.

		(To undo this, "crtcs = len(s.crtcs)" instead of "crtcs = 1".)
	*/
	var crtcs int = 1

	for ch := 0; ch < len(t); ch++ {
		t[ch] = make([][]C.ushort, crtcs, crtcs)
	}
	for crtcIdx, crtcGamma := range s.crtcs[0:crtcs] {
		var gamma *C.XRRCrtcGamma
		if gamma = C.XRRGetCrtcGamma(s.cl.dpy, crtcGamma.crtc); gamma == nil {
			return LookupTable{}, fmt.Errorf("Error getting CrtcGamma.")
		}
		forGammaChannels(gamma, func(ch Channel, gv gammaVector) {
			t[int(ch)][crtcIdx] = make([]C.ushort, crtcGamma.size, crtcGamma.size)
			for idx := C.int(0); idx < crtcGamma.size; idx++ {
				t[int(ch)][crtcIdx][idx] = gv[idx]
			}
		})
	}
	return LookupTable{t}, nil
}

// LookupTable represents the state of the CRTC lookup tables at some point in
// time.  Once created, a LookupTable instance does not refer to the underlying
// resources from which it was derived, so its lifespan may exceed that of the
// session from which it was created.
type LookupTable struct {
	// [channel][crtc][idx]
	t [_channel_cardinality_][][]C.ushort
}

// Equals compares two LookupTable instances and returns true if their values
// and topology are the same.  This can be used to detect gamma updates by other
// processes (e.g. redshift).
func (lt LookupTable) Equals(o LookupTable) bool {
	a := lt.t
	b := o.t
	for ch := 0; ch < len(a); ch++ {
		a1 := a[ch]
		b1 := b[ch]
		if len(a1) != len(b1) {
			return false
		}
		for crtc := 0; crtc < len(a1); crtc++ {
			a2 := a1[crtc]
			b2 := b1[crtc]
			if len(a2) != len(b2) {
				return false
			}
			for idx := 0; idx < len(a2); idx++ {
				if a2[idx] != b2[idx] {
					return false
				}
			}
		}
	}
	return true
}

// IsZero returns true if a LookupTable is the zero value.
func (lt LookupTable) IsZero() bool {
	if lt.t[0] == nil {
		return true
	}
	return false
}

// XferFn constructs an XferFn instance from a LookupTable using linear
// interpolation.
func (lt LookupTable) XferFn() XferFn {
	return func(ch Channel, in float64) (out float64) {
		var t [][]C.ushort = lt.t[ch]
		var acc float64
		var crtcs float64 = float64(len(t))
		for crtc := 0; crtc < len(t); crtc++ {
			lut := t[crtc]
			var base, frac float64 = math.Modf(in * float64(len(lut)))
			// We evaluate base here instead of frac so that we
			// don't have to worry about a bounds violation if
			// frac == epsilon.
			if int(base) < len(t)-1 {
				acc += float64(lut[int(base)])*(1.0-frac) +
					float64(lut[int(base)+1])*frac
			} else {
				acc += float64(lut[int(base)])
			}
		}
		return acc / crtcs / 65535.0
	}
}
