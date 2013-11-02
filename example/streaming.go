package main

import (
	"fmt"
	"github.com/grantmd/go-airplay"
)

var (
	deviceList []airplay.AirplayDevice
)

func main() {
	// Discover some devices
	fmt.Println("Looking for devices...")

	deviceChan := make(chan []airplay.AirplayDevice)
	go airplay.Discover(deviceChan)

	var device airplay.Airplay
	for device.IsConnected() == false {
		deviceList = <-deviceChan

		for i := range deviceList {
			// Connect to the first one that has properties that make sense
			if deviceList[i].IP == nil || deviceList[i].Type != "airplay" {
				continue
			}

			fmt.Println(deviceList[i].String())
			// TODO: Validate the TXT record properties first?
			var err error
			device, err = airplay.Dial(deviceList[i].IP, deviceList[i].Port, "")
			if err != nil {
				panic(err)
			}

			// We connected, now announce something
			fmt.Println("Connected")
			break
		}
	}
}
