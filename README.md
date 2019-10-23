# go-xrr-gamma

This module provides three packages:

* `gamma` provides a completely hardware-independent interface for querying and programming the CRTC lookup tables in terms of simple, real-number functions.

* `gamma/animate` provides a framework for writing and executing animated, event-responsive gamma transitions.

* `gamma/animate/alert` provides an event-responsive animation that alerts the users attention with varying degrees of gentleness and emphasis.

### What good is this?

With `gamma`, you can dim the screen, change its gamma compensation, change its color temperature, invert its colors, or increase its contrast.

With `gamma/animate`, you can do all of that continuously over time.

### What are CRTCs, and why is this called "gamma" when it's not strictly about gamma correction?

CRTCs are "cathode ray tube controllers," and they're still a part modern day video cards' output stage.
There's a nonlinear relationship between a CRT's electron gun voltage and its display brightness, and this relationship may vary from tube to tube and even from channel to channel within a tube.
The power-law function that approximates this nonlinear relationship is called the monitor's "gamma."
To compensate for gammas that differed between tubes and between channels, CRTCs were equipped with per-channel lookup tables that could be programmed with the inverse of the monitor's gamma function, providing a virtually linear relationship between the colors expressed in software and the colors displayed on the screen.

Today, cathode ray tube displays are increasingly rare, and the LCDs that have displaced them are smart enough to correct for their nonlinearities themselves.
Nevertheless, the original CRTC lookup tables persist, since they provide a convenient, generic, and ubiquitous target for color correction and color temperature adjustments.
Although these are no longer strictly "gamma" functions, the nomenclature has stuck around, since it's short and matches code and documentation that was written back when gamma was the principal concern.
