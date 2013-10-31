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

	deviceList = <-deviceChan

	fmt.Println(deviceList[0].String())

	// Connect to the first one
	// TODO: Validate the TXT record properties first?
	_, err := airplay.Dial(deviceList[0].IP, deviceList[0].Port, "")
	if err != nil {
		panic(err)
	}

	// We connected, now announce something
}
