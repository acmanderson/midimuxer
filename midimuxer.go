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

func (r *Router) Start() error {
	for _, source := range r.sources {
		if err := source.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) Stop() error {
	for _, source := range r.sources {
		if err := source.Stop(); err != nil {
			return err
		}
	}
	return nil
}

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

func WithFilter(filter Filter) func(*route) {
	return func(r *route) {
		r.filters = append(r.filters, filter)
	}
}

func WithTransformer(transformer Transformer) func(*route) {
	return func(r *route) {
		r.transformers = append(r.transformers, transformer)
	}
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
