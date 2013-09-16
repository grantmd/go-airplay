package main

import (
	"fmt"
)

func main() {
	fmt.Println("Looking for devides...")

	deviceChan := make(chan []AirplayDevice)
	go Discover(deviceChan)

	deviceList = <-deviceChan

	fmt.Printf("%+v\n", deviceList)
}
