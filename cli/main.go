package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/acmanderson/midimuxer"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

func parseInput(input string, devices map[midimuxer.DeviceID]*midimuxer.Device) (*midimuxer.Device, error) {
	index, err := strconv.ParseInt(input, 10, 0)
	if err != nil {
		return nil, err
	}

	if device, ok := devices[midimuxer.DeviceID(index)]; ok {
		return device, nil
	}

	fmt.Printf("%+v %d\n", devices, index)
	return nil, errors.New("Selection is invalid.")
}

func getDeviceSelection(reader *bufio.Reader, choices map[midimuxer.DeviceID]*midimuxer.Device, deviceType string) (*midimuxer.Device, error) {
	fmt.Printf("Select an %s =>\n", deviceType)
	for i, device := range choices {
		fmt.Printf("\t%d) %v\n", i, device.DeviceInfo.Name)
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

func inputLoop(router *midimuxer.Router, done chan bool) {
	inputs, outputs := router.Inputs(), router.Outputs()
	reader := bufio.NewReader(os.Stdin)
	input, err := getDeviceSelection(reader, inputs, "input")
	if err != nil {
		panic(err)
	}
	output, err := getDeviceSelection(reader, outputs, "output")
	if err != nil {
		panic(err)
	}

	router.AddRoute(input, output)

	done <- true
}

func main() {
	router := midimuxer.NewRouter()
	router.Start()
	defer router.Stop()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	done := make(chan bool)
	go inputLoop(router, done)

LOOP:
	for {
		select {
		case <-sigint:
			break LOOP
		case <-done:
			go inputLoop(router, done)
		}
	}
}
