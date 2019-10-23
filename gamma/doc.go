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

/*
Package gamma provides a completely hardware-independent interface for querying and programming the CRTC lookup tables in terms of simple, real-number functions.
This can be used to dim the screen, change its gamma compensation, change its color temperature, invert its colors, or increase its contrast, among other things.

(With the package gamma/animate, part of the same module, you can do all of that continuously over time.)

This package depends on the XRandR extension to X11 and requires its headers to build.

Nomenclature

CRTCs are "cathode ray tube controllers."
Despite the name, they're still a part of modern-day display controllers.

There's a nonlinear relationship between a CRT's electron gun voltage and its display brightness, and this relationship may vary from tube to tube and even from channel to channel within a tube.
The power-law function that approximates this nonlinear relationship is called the CRT's "gamma."
To compensate for gammas that differed between tubes and between channels, CRTCs were equipped with per-channel lookup tables that could be programmed with the inverse of the monitor's gamma function, providing a virtually linear relationship between the colors expressed in software and the colors displayed on the screen.

Today, CRT displays are increasingly rare, and the LCDs that have displaced them are smart enough to correct for their nonlinearities themselves.
Nevertheless, the original CRTC lookup tables persist, since they provide a convenient, generic, and ubiquitous target for color correction and color temperature adjustments.
Although these are no longer strictly "gamma" functions, the nomenclature has stuck around, since it's short and matches code and documentation that was written back when gamma was the principal concern.
*/
package gamma
