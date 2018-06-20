package midimuxer

type (
	Event struct {
		Status int64
		Data1  int64
		Data2  int64
	}

	Device interface {
		IsInput() bool
		IsOutput() bool
		Incoming() (<-chan *Event, error)
		Outgoing() (chan<- *Event, error)
		Name() string
	}

	Source interface {
		Start() error
		Stop() error
		Devices() []Device
	}

	route struct {
		device       Device
		filters      []Filter
		transformers []Transformer
	}

	Router struct {
		routes  map[Device][]*route
		sources []Source
	}
)

func NewRouter(sources ...Source) *Router {
	return &Router{
		routes:  make(map[Device][]*route),
		sources: sources,
	}
}

func (m *Router) Start() error {
	for _, source := range m.sources {
		if err := source.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Router) Stop() error {
	for _, source := range m.sources {
		if err := source.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Router) Inputs() []Device {
	var inputs []Device
	for _, source := range m.sources {
		for _, device := range source.Devices() {
			if device.IsInput() {
				inputs = append(inputs, device)
			}
		}
	}
	return inputs
}

func (m *Router) Outputs() []Device {
	var outputs []Device
	for _, source := range m.sources {
		for _, device := range source.Devices() {
			if device.IsOutput() {
				outputs = append(outputs, device)
			}
		}
	}
	return outputs
}

func (m *Router) AddRoute(input Device, output Device, filters []Filter, transformers []Transformer) error {
	_, ok := m.routes[input]
	m.routes[input] = append(m.routes[input], &route{output, filters, transformers})

	if !ok {
		incoming, err := input.Incoming()
		if err != nil {
			return err
		}
		go func(device Device, events <-chan *Event) {
			for event := range events {
				for _, route := range m.routes[device] {
					sendEvent := true
					for _, filter := range route.filters {
						if !filter.Filter(*event) {
							sendEvent = false
						}
					}
					if sendEvent {
						for _, transformer := range route.transformers {
							*event = transformer.Transform(*event)
						}
						outgoing, _ := route.device.Outgoing()
						outgoing <- &Event{event.Status, event.Data1, event.Data2}
					}
				}
			}
		}(input, incoming)
	}

	return nil
}
