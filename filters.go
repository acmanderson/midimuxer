package midimuxer

// Filter takes in an event and returns a bool indicating if the event should be routed to its intended output.
type Filter func(Event) bool

// ChannelFilter only routes events being sent on the given MIDI channel (1-16).
func ChannelFilter(channel int64) Filter {
	return func(event Event) bool {
		return (event.Status >= 0xF0) || ((event.Status%16)+1 == channel)
	}
}

const (
	// GreaterThan is used in conjunction with NoteFilter to route notes higher than a given note.
	GreaterThan = iota
	// LessThan is used in conjunction with NoteFilter to route notes less than a given note.
	LessThan    = iota
)

// NoteFilter only routes events whose note is either greater or less than the given note depending on the given condition (see GreaterThan and LessThan constants).
func NoteFilter(note int64, condition int) Filter {
	return func(event Event) bool {
		if event.Status >= 0xB0 {
			return true
		}

		switch condition {
		case GreaterThan:
			return event.Data1 > note
		case LessThan:
			return event.Data1 < note
		default:
			return false
		}
	}
}
