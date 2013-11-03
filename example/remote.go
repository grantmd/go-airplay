package main

import (
	"bufio"
	"fmt"
	"github.com/grantmd/go-airplay"
	"os"
)

var (
	deviceList []airplay.AirplayDevice
)

func main() {
	// Discover some devices
	fmt.Println("Waiting for remotes...")

	deviceChan := make(chan []airplay.AirplayDevice)
	go airplay.Discover(deviceChan)

	//var device airplay.Remote
	for {
		deviceList = <-deviceChan

		for i := range deviceList {
			// Connect to the first one that has properties that make sense
			if deviceList[i].IP == nil || deviceList[i].Type != "remote" {
				continue
			}

			fmt.Println(deviceList[i].String())

			fmt.Print("Enter the pin: ")
			bio := bufio.NewReader(os.Stdin)
			line, _, err := bio.ReadLine()
			if err != nil {
				panic(err)
			}
			//fmt.Println()

			_, err = airplay.Pair(deviceList[i], string(line))
			if err != nil {
				panic(err)
			}

			// We connected
			fmt.Println("Connected")
			break
		}
	}
}
