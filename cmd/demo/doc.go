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
Command demo demonstrates some of the capabilities of the go-xrr-gamma module.

Write-only

Reset the lookup tables to their default.  (Same as "demo power 1".)
    $ demo reset

Apply a power law function with exponent POWER and coefficient 1.
    $ demo power POWER

Make all three color channels channels bilevel.
    $ demo bilevel

Read and Write-back

Dim the existing lookup tables by 50%.
    $ demo dim

Animation

Make the screen pulse.
    $ demo pulse

Demo an "alert" effect with smooth transitions and event-driven accents.
(Send SIGUSR1 to the process to "strobe" the screen, SIGUSR2 to "warble" the screen, or SIGINT to exit.)
    $ demo alert
*/
package main
