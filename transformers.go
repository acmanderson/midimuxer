package midimuxer

type Transformer func(Event) Event

func AftertouchToPitchBend(event Event) Event {
	if !((event.Status >= 0xA0 && event.Status <= 0xAF) || (event.Status >= 0xD0 && event.Status <= 0xDF)) {
		return event
	}

	event.Status = 0xE0 + event.Status%16
	event.Data2 = event.Data1
	event.Data1 = 0

	return event
}
