# midimuxer
Route MIDI events between your MIDI devices!

## Requirements
* [`PortMidi`](http://portmedia.sourceforge.net/portmidi/)

## Usage
A simple command-line interface to the `midimuxer` library is included in the `cli` package of this repo. To build and install it to your `GOPATH`, run `go install github.com/acmanderson/midimuxer/cli`. Plug in some USB MIDI devices and run it, following the prompts to route MIDI messages between them. For example, route the output from one MIDI keyboard to multiple different MIDI inputs.
