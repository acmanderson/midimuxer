// Package midimuxer provides a library for managing the routing of MIDI events between your devices from multiple sources.
// Currently the only supported source is PortMIDI (see the portmidi package). The examples/cli package provides a simple CLI example which uses PortMIDI.
package midimuxer

type (
	// Event represents a MIDI event.
	Event struct {
		// Status represents the status byte of the MIDI event which determines the function of the event.
		Status int64
		// Data1 represents the first data byte of the MIDI event whose meaning is determined by the Status byte.
		Data1 int64
		// Data2 represents the first data byte of the MIDI event whose meaning is determined by the Status byte.
		Data2 int64
	}

	// Device represents a MIDI input or output
	Device interface {
		// IsInput indicates if this device can generate MIDI events.
		IsInput() bool
		// IsOutput indicates if this device can receive MIDI events.
		IsOutput() bool
		// Incoming returns a channel for receiving the device's incoming MIDI events, or an error if it couldn't initialize properly.
		Incoming() (<-chan *Event, error)
		// Incoming returns a channel for sending MIDI events to the device, or an error if it couldn't initialize properly.
		Outgoing() (chan<- *Event, error)
		// Name returns the name of the device.
		Name() string
	}

	// Source represents a source for MIDI devices and their events.
	// See the portmidi package for an implementation using PortMIDI.
	Source interface {
		// Start performs any setup needed for the source and its devices.
		Start() error
		// Stop performs any shutdown steps needed for the source and its devices.
		Stop() error
		// Devices returns a slice of all MIDI devices this source is responsible for.
		Devices() []Device
	}

	route struct {
		device       Device
		filters      []Filter
		transformers []Transformer
	}

	// Router is responsible for managing MIDI sources and routing events between devices.
	Router struct {
		routes  map[Device][]*route
		sources []Source
	}
)

// NewRouter returns a new Router initialized with the specified sources.
func NewRouter(sources ...Source) *Router {
	return &Router{
		routes:  make(map[Device][]*route),
		sources: sources,
	}
}

// Start calls Start() for each source this router is managing.
func (r *Router) Start() error {
	for _, source := range r.sources {
		if err := source.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop calls Stop() for each source this router is managing.
func (r *Router) Stop() error {
	for _, source := range r.sources {
		if err := source.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// Inputs returns all Devices this router manages that can generate MIDI events.
func (r *Router) Inputs() []Device {
	var inputs []Device
	for _, source := range r.sources {
		for _, device := range source.Devices() {
			if device.IsInput() {
				inputs = append(inputs, device)
			}
		}
	}
	return inputs
}

// Outputs returns all Devices this router manages that can receive MIDI events.
func (r *Router) Outputs() []Device {
	var outputs []Device
	for _, source := range r.sources {
		for _, device := range source.Devices() {
			if device.IsOutput() {
				outputs = append(outputs, device)
			}
		}
	}
	return outputs
}

func (r *Router) routeEvents(input Device, events <-chan *Event) {
	for event := range events {
		for _, route := range r.routes[input] {
			sendEvent := true
			for _, filter := range route.filters {
				if !filter(*event) {
					sendEvent = false
					break
				}
			}
			if sendEvent {
				for _, transformer := range route.transformers {
					*event = transformer(*event)
				}
				outgoing, _ := route.device.Outgoing()
				outgoing <- &Event{event.Status, event.Data1, event.Data2}
			}
		}
	}
}

// WithFilter is used in conjunction with Router.AddRoute() to set a Filter when adding a new route.
func WithFilter(filter Filter) func(*route) {
	return func(r *route) {
		r.filters = append(r.filters, filter)
	}
}

// WithTransformer is used in conjunction with Router.AddRoute() to set a Transformer when adding a new route.
func WithTransformer(transformer Transformer) func(*route) {
	return func(r *route) {
		r.transformers = append(r.transformers, transformer)
	}
}

// AddRoute begins routing events from the given input to the given output.
// Route configuration functions can be passed in to set Filters and Transformers for this route (see WithFilter and WithTransformer).
func (r *Router) AddRoute(input Device, output Device, routeConfigs ...func(*route)) error {
	route := &route{device: output}
	for _, config := range routeConfigs {
		config(route)
	}

	_, ok := r.routes[input]
	r.routes[input] = append(r.routes[input], route)

	if !ok {
		incoming, err := input.Incoming()
		if err != nil {
			return err
		}
		go r.routeEvents(input, incoming)
	}

	return nil
}
