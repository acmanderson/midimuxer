// +build !test

package main

import (
	"fmt"
	"sort"

	"github.com/acmanderson/midimuxer"
	"github.com/acmanderson/midimuxer/portmidi"
	"github.com/manifoldco/promptui"
)

func getDeviceSelection(choices []midimuxer.Device, label string, selectedGlyph string) (midimuxer.Device, error) {
	devices := make([]midimuxer.Device, 0)
	for _, device := range choices {
		devices = append(devices, device)
	}
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Name() < devices[j].Name()
	})

	deviceTemplates := &promptui.SelectTemplates{
		Active:   fmt.Sprintf("%s {{ .Name | cyan | bold}}", selectedGlyph),
		Inactive: "{{ .Name | cyan }}",
		Selected: fmt.Sprintf("%s {{ .Name }}", selectedGlyph),
	}

	prompt := promptui.Select{
		Label:     label,
		Items:     devices,
		Templates: deviceTemplates,
	}
	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	return devices[i], nil
}

func prompt(router *midimuxer.Router) error {
	selectedInput, err := getDeviceSelection(router.Inputs(), "Inputs", "\U0001F3B9") // musical keyboard emoji
	if err != nil {
		return err
	}
	selectedOutput, err := getDeviceSelection(router.Outputs(), "Outputs", "\U0001F50A") // speaker emoji
	if err != nil {
		return err
	}
	return router.AddRoute(
		selectedInput,
		selectedOutput,
		nil,
		nil,
	)
}

func main() {
	router := midimuxer.NewRouter(&portmidi.Source{})
	router.Start()
	defer router.Stop()

	for {
		if err := prompt(router); err != nil {
			switch err {
			case promptui.ErrInterrupt:
				return
			default:
				panic(err)
			}
		}
	}
}
