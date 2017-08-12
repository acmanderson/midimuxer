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
}

func parseInput(input string, devices []*Device) (*Device, error) {
	index, err := strconv.ParseInt(input, 10, 0)
	if err != nil {
		return nil, err
	}

	if device := devices[index]; device != nil {
		return device, nil
	}

	return nil, errors.New("Selection is invalid.")
}

func getDeviceSelection(reader *bufio.Reader, choices []*Device, deviceType string) (*Device, error) {
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

func main() {
	portmidi.Initialize()
	defer portmidi.Terminate()

	inputs, outputs := []*Device{}, []*Device{}

	for i := 0; i < portmidi.CountDevices(); i++ {
		deviceInfo := portmidi.Info(portmidi.DeviceID(i))
		if deviceInfo.IsInputAvailable {
			inputs = append(inputs, &Device{portmidi.DeviceID(i), deviceInfo})
		}
		if deviceInfo.IsOutputAvailable {
			outputs = append(outputs, &Device{portmidi.DeviceID(i), deviceInfo})
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

	in, err := portmidi.NewInputStream(input.deviceID, 1024)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	out, err := portmidi.NewOutputStream(output.deviceID, 1024, 0)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	midiChannel := in.Listen()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

LOOP:
	for {
		select {
		case midiEvent := <-midiChannel:
			out.WriteShort(midiEvent.Status, midiEvent.Data1, midiEvent.Data2)
		case <-sigint:
			break LOOP
		}
	}
}
