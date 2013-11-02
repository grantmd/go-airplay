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

	for {
		deviceList = <-deviceChan

		for i := range deviceList {
			// Connect to the first one that has properties that make sense
			if deviceList[i].IP == nil || deviceList[i].Type != "remote" {
				continue
			}

			fmt.Println(deviceList[i].String())

			// We connected
			fmt.Println("Connected")
			break
		}
	}
}
