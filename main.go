package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/rakyll/portmidi"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

type Device struct {
	deviceID   portmidi.DeviceID
	deviceInfo *portmidi.DeviceInfo
	stream     *portmidi.Stream
}

type Event struct {
	portmidi.Event
	deviceID portmidi.DeviceID
}

type Router struct {
	events chan Event
	routes map[portmidi.DeviceID][]*Device
}

func (r *Router) AddRoute(input *Device, output *Device) {
	r.routes[input.deviceID] = append(r.routes[input.deviceID], output)
	go func(deviceID portmidi.DeviceID, events <-chan portmidi.Event) {
		for event := range events {
			r.events <- Event{event, deviceID}
		}
	}(input.deviceID, input.stream.Listen())
}

func (r *Router) RouteEvents() {
	for event := range r.events {
		for _, device := range r.routes[event.deviceID] {
			device.stream.WriteShort(event.Status, event.Data1, event.Data2)
		}
	}
}

func parseInput(input string, devices map[portmidi.DeviceID]*Device) (*Device, error) {
	index, err := strconv.ParseInt(input, 10, 0)
	if err != nil {
		return nil, err
	}

	if device, ok := devices[portmidi.DeviceID(index)]; ok {
		return device, nil
	}

	fmt.Printf("%+v %d\n", devices, index)
	return nil, errors.New("Selection is invalid.")
}

func getDeviceSelection(reader *bufio.Reader, choices map[portmidi.DeviceID]*Device, deviceType string) (*Device, error) {
	fmt.Printf("Select an %s =>\n", deviceType)
	for i, device := range choices {
		fmt.Printf("\t%d) %v\n", i, device.deviceInfo.Name)
	}
	fmt.Printf("> ")
	selection, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	device, err := parseInput(strings.Trim(selection, "\n"), choices)
	if err != nil {
		return nil, err
	}

	return device, nil
}

func inputLoop(router *Router, inputs, outputs map[portmidi.DeviceID]*Device, streams chan *portmidi.Stream, done chan bool) {
	for i := 0; i < portmidi.CountDevices(); i++ {
		deviceID := portmidi.DeviceID(i)
		deviceInfo := portmidi.Info(deviceID)
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

	reader := bufio.NewReader(os.Stdin)
	input, err := getDeviceSelection(reader, inputs, "input")
	if err != nil {
		panic(err)
	}
	output, err := getDeviceSelection(reader, outputs, "output")
	if err != nil {
		panic(err)
	}

	if input.stream == nil {
		input.stream, err = portmidi.NewInputStream(input.deviceID, 1024)
		if err != nil {
			panic(err)
		}
		streams <- input.stream
	}

	if output.stream == nil {
		output.stream, err = portmidi.NewOutputStream(output.deviceID, 1024, 0)
		if err != nil {
			panic(err)
		}
		streams <- output.stream
	}

	router.AddRoute(input, output)

	done <- true
}

func main() {
	portmidi.Initialize()
	defer portmidi.Terminate()

	eventChan := make(chan Event)
	defer close(eventChan)
	router := &Router{eventChan, make(map[portmidi.DeviceID][]*Device)}
	go router.RouteEvents()

	inputs, outputs := make(map[portmidi.DeviceID]*Device), make(map[portmidi.DeviceID]*Device)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	streams := []*portmidi.Stream{}
	streamChan := make(chan *portmidi.Stream)
	done := make(chan bool)
	go inputLoop(router, inputs, outputs, streamChan, done)

LOOP:
	for {
		select {
		case <-sigint:
			break LOOP
		case stream := <-streamChan:
			streams = append(streams, stream)
		case <-done:
			go inputLoop(router, inputs, outputs, streamChan, done)
		}
	}

	for _, stream := range streams {
		stream.Close()
	}
}
