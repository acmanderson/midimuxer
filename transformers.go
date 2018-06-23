package midimuxer

// Transformer takes in an event and returns a new or modified event to be forwarded in its place.
type Transformer func(Event) Event

// AftertouchToPitchBend is a Transformer that turns MIDI aftertouch events into pitch bend events.
func AftertouchToPitchBend(event Event) Event {
	if !((event.Status >= 0xA0 && event.Status <= 0xAF) || (event.Status >= 0xD0 && event.Status <= 0xDF)) {
		return event
	}

	event.Status = 0xE0 + event.Status%16
	event.Data2 = event.Data1
	event.Data1 = 0

	return event
}
