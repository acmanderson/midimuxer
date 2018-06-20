package portmidi

import (
	"github.com/acmanderson/midimuxer"
	"github.com/rakyll/portmidi"
)

type (
	device struct {
		id   portmidi.DeviceID
		info *portmidi.DeviceInfo

		inputStream    *portmidi.Stream
		outputStream   *portmidi.Stream
		incomingEvents chan *midimuxer.Event
		outgoingEvents chan *midimuxer.Event
	}

	Source struct {
		devices []*device
	}
)

func (d *device) IsInput() bool {
	return d.info.IsInputAvailable
}

func (d *device) IsOutput() bool {
	return d.info.IsOutputAvailable
}

func (d *device) Incoming() (<-chan *midimuxer.Event, error) {
	if d.inputStream == nil {
		stream, err := portmidi.NewInputStream(d.id, 1024)
		if err != nil {
			return nil, err
		}
		d.inputStream = stream
		d.incomingEvents = make(chan *midimuxer.Event)
		go func() {
			for event := range d.inputStream.Listen() {
				d.incomingEvents <- &midimuxer.Event{
					Status: event.Status,
					Data1:  event.Data1,
					Data2:  event.Data2,
				}
			}
		}()
	}
	return d.incomingEvents, nil
}

func (d *device) Outgoing() (chan<- *midimuxer.Event, error) {
	if d.outputStream == nil {
		stream, err := portmidi.NewOutputStream(d.id, 1024, 0)
		if err != nil {
			return nil, err
		}
		d.outputStream = stream
		d.outgoingEvents = make(chan *midimuxer.Event)
		go func() {
			for event := range d.outgoingEvents {
				d.outputStream.WriteShort(event.Status, event.Data1, event.Data2)
			}
		}()
	}
	return d.outgoingEvents, nil
}

func (d *device) Name() string {
	return d.info.Name
}

func (s *Source) loadDevices() {
	for i := 0; i < portmidi.CountDevices(); i++ {
		deviceID := portmidi.DeviceID(i)
		s.devices = append(s.devices, &device{
			id:   deviceID,
			info: portmidi.Info(portmidi.DeviceID(deviceID)),
		})
	}
}

func (s *Source) Start() error {
	if err := portmidi.Initialize(); err != nil {
		return err
	}

	s.loadDevices()

	return nil
}

func (s *Source) Stop() error {
	for _, device := range s.devices {
		if device.inputStream != nil {
			if err := device.inputStream.Close(); err != nil {
				return err
			}
			close(device.incomingEvents)
		}
		if device.outputStream != nil {
			if err := device.outputStream.Close(); err != nil {
				return err
			}
			close(device.outgoingEvents)
		}
	}

	if err := portmidi.Terminate(); err != nil {
		return err
	}

	return nil
}

func (s *Source) Devices() []midimuxer.Device {
	var devices []midimuxer.Device
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	return devices
}
