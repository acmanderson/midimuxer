package midimuxer_test

import (
	. "github.com/acmanderson/midimuxer"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
)

type mockDevice struct {
	isInput, isOutput  bool
	incoming, outgoing chan *Event
}

func (d *mockDevice) IsInput() bool {
	return d.isInput
}

func (d *mockDevice) IsOutput() bool {
	return d.isOutput
}

func (d *mockDevice) Incoming() (<-chan *Event, error) {
	return d.incoming, nil
}

func (d *mockDevice) Outgoing() (chan<- *Event, error) {
	return d.outgoing, nil
}

func (d *mockDevice) Name() string {
	return ""
}

type mockSource struct {
	devices []*mockDevice
}

func (s *mockSource) Start() error {
	for _, device := range s.devices {
		device.incoming = make(chan *Event)
		device.outgoing = make(chan *Event)
	}
	return nil
}

func (s *mockSource) Stop() error {
	for _, device := range s.devices {
		close(device.incoming)
		close(device.outgoing)
	}
	return nil
}

func (s *mockSource) Devices() []Device {
	var devices []Device
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	return devices
}

var _ = Describe("Router", func() {
	var (
		router *Router

		input   = &mockDevice{isInput: true}
		output1 = &mockDevice{isOutput: true}
		output2 = &mockDevice{isOutput: true}
	)

	BeforeEach(func() {
		router = NewRouter(&mockSource{
			[]*mockDevice{
				input,
				output1,
				output2,
			},
		})
		router.Start()
	})

	AfterEach(func() {
		router.Stop()
	})

	Describe("Routing", func() {
		Context("Without filters/transformers", func() {
			It("should route events from input to output", func() {
				router.AddRoute(input, output1)
				router.AddRoute(input, output2)
				event := &Event{Status: 5}
				input.incoming <- event

				Expect(<-output1.outgoing).To(Equal(event))
				Expect(<-output2.outgoing).To(Equal(event))
			})
		})

		Context("With filters/transformers", func() {
			It("should route/transform events", func() {
				router.AddRoute(
					input,
					output1,
					WithFilter(func(event Event) bool {
						return event.Status > 3
					}),
					WithFilter(func(event Event) bool {
						return event.Status%2 == 1
					}),
					WithTransformer(func(event Event) Event {
						event.Status = event.Status + 1
						return event
					}),
					WithTransformer(func(event Event) Event {
						event.Status = event.Status / 2
						return event
					}),
				)

				event := &Event{Status: 2}
				input.incoming <- event
				Eventually(output1.outgoing).ShouldNot(Receive())
				event.Status = 5
				input.incoming <- event
				Expect(*(<-output1.outgoing)).To(Equal(Event{Status: 3}))
			})
		})
	})
})
