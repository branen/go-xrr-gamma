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

package gamma

import (
	"fmt"
)

func ExampleXferFn() {
	var invert, red, dim XferFn
	invert = func(ch Channel, in float64) (out float64) {
		return 1 - in
	}
	red = func(ch Channel, in float64) (out float64) {
		if ch != Red {
			return 0
		}
		return in
	}
	dim = func(ch Channel, in float64) (out float64) {
		return in / 2
	}
	fmt.Printf("%01.1f\n", invert(Red, 0.8))
	fmt.Printf("%01.1f\n", red(Green, 0.8))
	fmt.Printf("%01.1f\n", dim(Blue, 0.8))
	// Output:
	// 0.2
	// 0.0
	// 0.4
}
