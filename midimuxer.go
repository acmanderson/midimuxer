package midimuxer

import (
	"github.com/rakyll/portmidi"
)

type DeviceID portmidi.DeviceID

type Device struct {
	DeviceID   DeviceID
	DeviceInfo *portmidi.DeviceInfo
	Stream     *portmidi.Stream
}

type Event struct {
	portmidi.Event
	Device *Device
}

type Route struct {
	device       *Device
	filters      []Filter
	transformers []Transformer
}

type Router struct {
	inputs, outputs map[DeviceID]*Device

	events chan Event
	routes map[DeviceID][]*Route
}

func NewRouter() *Router {
	return &Router{
		events: make(chan Event),
		routes: make(map[DeviceID][]*Route),
	}
}

func (m *Router) loadDevices() {
	inputs, outputs := make(map[DeviceID]*Device), make(map[DeviceID]*Device)
	for i := 0; i < portmidi.CountDevices(); i++ {
		deviceID := DeviceID(i)
		deviceInfo := portmidi.Info(portmidi.DeviceID(deviceID))
		if deviceInfo.IsInputAvailable {
			if _, ok := inputs[deviceID]; !ok {
				inputs[deviceID] = &Device{deviceID, deviceInfo, nil}
			}
		}
		if deviceInfo.IsOutputAvailable {
			if _, ok := outputs[deviceID]; !ok {
				outputs[deviceID] = &Device{deviceID, deviceInfo, nil}
			}
		}
	}
	m.inputs = inputs
	m.outputs = outputs
}

func (m *Router) routeEvents() {
	for event := range m.events {
		for _, route := range m.routes[event.Device.DeviceID] {
			sendEvent := true
			for _, filter := range route.filters {
				if !filter.Filter(event) {
					sendEvent = false
				}
			}
			if sendEvent {
				for _, transformer := range route.transformers {
					event = transformer.Transform(event)
				}
				route.device.Stream.WriteShort(event.Status, event.Data1, event.Data2)
			}
		}
	}
}

func (m *Router) Start() {
	portmidi.Initialize()
	m.loadDevices()
	go m.routeEvents()
}

func (m *Router) Stop() {
	portmidi.Terminate()
	close(m.events)
	for _, input := range m.inputs {
		if input.Stream != nil {
			input.Stream.Close()
		}
	}
	for _, output := range m.outputs {
		if output.Stream != nil {
			output.Stream.Close()
		}
	}
}

func (m *Router) Inputs() map[DeviceID]*Device {
	return m.inputs
}

func (m *Router) Outputs() map[DeviceID]*Device {
	return m.outputs
}

func (m *Router) AddRoute(input *Device, output *Device, filters []Filter, transformers []Transformer) error {
	m.routes[input.DeviceID] = append(m.routes[input.DeviceID], &Route{output, filters, transformers})

	if input.Stream == nil {
		stream, err := portmidi.NewInputStream(portmidi.DeviceID(input.DeviceID), 1024)
		if err != nil {
			return err
		}
		input.Stream = stream
	}
	if output.Stream == nil {
		stream, err := portmidi.NewOutputStream(portmidi.DeviceID(output.DeviceID), 1024, 0)
		if err != nil {
			return err
		}
		output.Stream = stream
	}

	go func(device *Device, events <-chan portmidi.Event) {
		for event := range events {
			m.events <- Event{event, input}
		}
	}(input, input.Stream.Listen())

	return nil
}
