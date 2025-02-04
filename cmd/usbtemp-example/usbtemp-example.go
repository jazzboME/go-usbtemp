package main

import (
	"flag"
	"log"

	"github.com/jazzboME/go-usbtemp"
)
var serial string

func main() {
	flag.StringVar(&serial, "port", "/dev/ttyUSB0", "Port Name where usbtemp device is connected")
	flag.Parse()
	var probe = usbtemp.USBtemp{}
	if err := probe.Open(serial); err != nil {
		log.Fatalf("Open() failed: %v", err)
	}
	
	defer probe.Close()

	rom, err := probe.Rom()
	if err != nil {
		log.Fatalf("ROM probe failed: %v", err)
	}
	temp, err := probe.Temperature(true)
	if err != nil {
		log.Fatalf("Temperature probe failed: %v", err)
	}

	log.Printf("\nName: %s\nSerial: %s\nRom: %s\nTemperature: %.3f\n", 
				probe.Name, probe.SerialNumber, rom, temp)
}
