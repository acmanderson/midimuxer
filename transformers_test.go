package midimuxer_test

import (
	. "github.com/acmanderson/midimuxer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transformers", func() {
	var (
		transformer Transformer
	)

	Describe("AfterTouchToPitchBendTransformer", func() {
		transformer = AftertouchToPitchBend

		Context("With a non-aftertouch event", func() {
			It("should not transform the event", func() {
				event := Event{Status: 0x80, Data1: 65, Data2: 127}
				Expect(transformer(event)).To(Equal(event))
			})
		})

		Context("With an aftertouch event", func() {
			It("should transform the event into a pitch bend event", func() {
				event := Event{Status: 0xD5, Data1: 100}
				transformedEvent := Event{Status: 0xE5, Data1: 0, Data2: 100}
				Expect(transformer(event)).To(Equal(transformedEvent))
			})
		})
	})
})
