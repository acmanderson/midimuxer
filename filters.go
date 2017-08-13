package midimuxer

type Filter interface {
	Filter(Event) bool
}

type ChannelFilter struct {
	Channel int64
}

func (f ChannelFilter) Filter(event Event) bool {
	return (event.Status >= 0xF0) || ((event.Status%16)+1 == f.Channel)
}

const (
	GreaterThan = iota
	LessThan    = iota
)

type NoteFilter struct {
	Note      int64
	Condition int
}

func (f NoteFilter) Filter(event Event) bool {
	if event.Status >= 0xB0 {
		return true
	}

	switch f.Condition {
	case GreaterThan:
		return event.Data1 > f.Note
	case LessThan:
		return event.Data1 < f.Note
	default:
		return false
	}
}
