package midimuxer_test

import (
	. "github.com/acmanderson/midimuxer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rakyll/portmidi"
)

var _ = Describe("Filters", func() {
	var (
		event1 Event = Event{Event: portmidi.Event{Status: 0x80, Data1: 65, Data2: 127}, Device: nil}
		event2 Event = Event{Event: portmidi.Event{Status: 0x81, Data1: 30, Data2: 50}, Device: nil}
		filter Filter
	)

	Describe("ChannelFilter", func() {
		filter = ChannelFilter{Channel: 1}

		Context("With an event matching its channel", func() {
			It("should send the event", func() {
				Expect(filter.Filter(event1)).To(Equal(true))
			})
		})

		Context("With an event not matching its channel", func() {
			It("should not send the event", func() {
				Expect(filter.Filter(event2)).To(Equal(false))
			})
		})
	})

	Describe("NoteFilter", func() {
		Context("With a LessThan Condition", func() {
			filter := NoteFilter{Note: 60, Condition: LessThan}

			It("should not send events whose note is greater than its Note value", func() {
				Expect(filter.Filter(event1)).To(Equal(false))
			})
			It("should send events whose note is less than its Note value", func() {
				Expect(filter.Filter(event2)).To(Equal(true))
			})
		})

		Context("With a GreaterThan Condition", func() {
			filter := NoteFilter{Note: 35, Condition: GreaterThan}

			It("should send events whose note is greater than its Note value", func() {
				Expect(filter.Filter(event1)).To(Equal(true))
			})
			It("should not send events whose note is less than its Note value", func() {
				Expect(filter.Filter(event2)).To(Equal(false))
			})
		})
	})
})
