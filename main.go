package main

import (
	"log"

	"sertest/gps"

	"go.bug.st/serial"
)

func main() {
	ports, _ := serial.GetPortsList()
	for _, item := range ports {
		log.Printf("Port: %v", item)
	}
	/*
		ports2, _ := enumerator.GetDetailedPortsList()
		for _, item := range ports2 {
			log.Printf("Port: %v", item.Manufacturer)
		}*/
	mode := &serial.Mode{
		BaudRate: 38400,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open("/dev/ttyAMA0", mode)
	if err != nil {
		log.Printf("open port failed %v", err)
	}
	g, err := gps.NewGps(port, port)
	if err != nil {
		port.Close()
		return
	}
	notify := make(chan any, 20)
	go g.Sart(notify)
	for val := range notify {
		// fmt.Printf("Received: %T\n", val)
		switch data := val.(type) {
		/*
			case *gps.GGA:
				data.Print()
			case *gps.RMC:
				data.Print()
			case *gps.GLL:
				data.Print()
			case *gps.GSA:
				data.Print()
			case *gps.VTG:
				data.Print()
			case *gps.GSV:
				data.Print()
		*/
		case *gps.TXT:
			data.Print()
		}
	}
}
