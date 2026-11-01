// Copyright (c) 2026 Reiner Pröls
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// SPDX-License-Identifier: MIT
//
// Author: Reiner Pröls

// GPS parser - Testprogram

package main

import (
	"log"
	"os"
	"time"

	"gpsmonitord/cfg"
	"gpsmonitord/data"
	"gpsmonitord/gps"
	"gpsmonitord/webserver"

	"go.bug.st/serial"
)

var Data data.Data

type WebCallbacksImpl struct{}

var _ webserver.WebCallBacks = (*WebCallbacksImpl)(nil)

func (w WebCallbacksImpl) GetData() *webserver.WebDataAll {
	return nil
}

func main() {
	var err error
	cfgFile := os.Getenv("GPSMONITORD_INI_FILE")
	if cfgFile == "" {
		cfgFile = cfg.DEF_CFG_FILE
	}

	Data.Config, err = cfg.NewCfg(cfgFile, func() {
	})
	if err != nil {
		return
	}

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
	go g.Start(nil, true)

	var callBacks WebCallbacksImpl
	Data.WebServer = webserver.NewWebServer(callBacks, Data.Config)
	webError := make(chan error, 2)
	Data.WebServer.StartWebServer(webError)

	for {
		time.Sleep(250 * time.Millisecond)
		g.GpsSumData.Print()
		// sats := g.GetSattelites()
		// log.Printf("Sats: %d", len(sats))
		/*
			for _, item := range sats {
				item.Print()
			}*/
	}
	/*
		notify := make(chan any, 20)
		go g.Sart(notify)
		for val := range notify {
			// fmt.Printf("Received: %T\n", val)
			switch data := val.(type) {
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
			case *gps.TXT:
				data.Print()
			case *gps.GST:
				data.Print()
			}
		}
	*/
}
