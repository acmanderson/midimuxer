package midimuxer

type Filter func(Event) bool

func ChannelFilter(channel int64) Filter {
	return func(event Event) bool {
		return (event.Status >= 0xF0) || ((event.Status%16)+1 == channel)
	}
}

const (
	GreaterThan = iota
	LessThan    = iota
)

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
