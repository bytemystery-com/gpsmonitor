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
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
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

func getFloat32ForJsonFromAtomic(val *atomic.Uint32) *float32 {
	v := math.Float32frombits(val.Load())
	if math.IsNaN(float64(v)) {
		return nil
	}
	return &v
}

func getFloat64ForJsonFromAtomic(val *atomic.Uint64) *float64 {
	v := math.Float64frombits(val.Load())
	if math.IsNaN(v) {
		return nil
	}
	return &v
}

func getIntForJsonFromAtomic(val *atomic.Uint32) *int {
	v := val.Load()
	if v == ^uint32(0) {
		return nil
	}
	vi := int(v)
	return &vi
}

func getTimeForJsonFromAtomic(val *atomic.Value) *time.Time {
	v, ok := val.Load().(time.Time)
	if !ok || v.IsZero() {
		return nil
	}
	return &v
}

func (w WebCallbacksImpl) GetData() *webserver.WebDataAll {
	if Data.Gps.GpsSumData.IsReady.Load() {
		used := 0
		sats := Data.Gps.GetSattelites(&used)
		total := len(sats)
		data := webserver.WebDataAll{
			Gps: webserver.WebDataGps{
				Longitude:    getFloat64ForJsonFromAtomic(&Data.Gps.GpsSumData.Longitude),
				Latitude:     getFloat64ForJsonFromAtomic(&Data.Gps.GpsSumData.Latitude),
				Altitude:     getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.Altitude),
				Speed:        getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.Speed),
				Course:       getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.CourseTrue),
				MagVar:       getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.MagVar),
				Mode:         getIntForJsonFromAtomic(&Data.Gps.GpsSumData.NavMode),
				Quality:      getIntForJsonFromAtomic(&Data.Gps.GpsSumData.Quality),
				NumUsedSats:  &used,
				NumTotalSats: &total,
				StdLongitude: getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.StdLongitude),
				StdLatitude:  getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.StdLatitude),
				StdAltitude:  getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.StdAltitude),
				HDop:         getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.HDop),
				VDop:         getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.VDop),
				PDop:         getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.PDop),
				DateTime:     getTimeForJsonFromAtomic(&Data.Gps.GpsSumData.Time),
			},
			Sats: sats,
		}
		return &data
	}
	return nil
}

func main() {
	var err error
	var debug bool
	str := os.Getenv("GPSMONITORD_DEBUG")
	if strings.EqualFold(str, "true") || strings.EqualFold(str, "on") || strings.EqualFold(str, "yes") || strings.EqualFold(str, "1") {
		debug = true
	}

	d := flag.Bool("d", false, "debug")
	flag.Parse()
	if *d {
		debug = true
	}
	log.Printf("Debug: %v, Go version: %v", debug, runtime.Version()[2:])
	// journalctl -u weatherd.service

	if !debug {
		log.SetOutput(io.Discard)
	} else {
		// curl http://localhost:6060/debug/pprof/
		// go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
		go func() {
			http.ListenAndServe("0.0.0.0:6060", nil)
		}()
	}

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
		DataBits: Data.Config.GetIntValue(cfg.SECTION_HARDWARE, cfg.KEY_HARDWARE_DATA_BITS, cfg.DEF_UART_DATA_BITS),
		Parity:   serial.Parity(Data.Config.GetIntValue(cfg.SECTION_HARDWARE, cfg.KEY_HARDWARE_DATA_PARITY, cfg.DEF_UART_PARITY)),
		StopBits: serial.StopBits(Data.Config.GetIntValue(cfg.SECTION_HARDWARE, cfg.KEY_HARDWARE_DATA_STOP_BITS, cfg.DEF_UART_STOP_BITS)),
	}

	baudRate := Data.Config.GetIntValue(cfg.SECTION_HARDWARE, cfg.KEY_HARDWARE_BAUD_RATE, cfg.DEF_UART_BAUD_RATE)
	dev := Data.Config.GetStringValue(cfg.SECTION_HARDWARE, cfg.KEY_HARDWARE_UART_DEV, cfg.DEF_UART_DEV)
	if baudRate == 0 {
		// Auto detect
		baudRates := []int{38400, 460800, 115200, 9600, 4800, 57600, 230400}
		var port serial.Port
		for _, br := range baudRates {
			mode.BaudRate = br
			if port == nil {
				port, err = serial.Open(dev, mode)
				if err != nil {
					log.Printf("open port failed %v", err)
					continue
				}
				port.SetReadTimeout(1200 * time.Millisecond)
			} else {
				port.SetMode(mode)
			}
			_ = port.ResetInputBuffer()
			err = gps.TestGPSRead(port)
			if err != nil {
				log.Printf("Auto baudrate test for %d - %v", br, err)
				continue
			}
			baudRate = br
			log.Printf("Auto baud rate found - %d", br)
			break
		}
		if port != nil {
			port.Close()
			time.Sleep(250 * time.Millisecond)
		}
	}

	if baudRate == 0 {
		log.Printf("No valid baud rate found")
		return
	}
	mode.BaudRate = baudRate
	port, err := serial.Open(dev, mode)
	if err != nil {
		log.Printf("open port failed %v", err)
		return
	}
	g, err := gps.NewGps(port, port, func() {
		port.ResetOutputBuffer()
		port.ResetInputBuffer()
	})
	// g.Trace = true
	if err != nil {
		port.Close()
		return
	}
	Data.Gps = g
	go g.Start(nil)

	var callBacks WebCallbacksImpl
	Data.WebServer = webserver.NewWebServer(callBacks, Data.Config)
	webError := make(chan error, 2)
	Data.WebServer.StartWebServer(webError)

	/*
		for {
			time.Sleep(250 * time.Millisecond)
			if Data.Gps.GpsSumData.IsReady.Load() {
				g.GpsSumData.Print()
				// sats := g.GetSattelites()
				// log.Printf("Sats: %d", len(sats))
					for _, item := range sats {
						item.Print()
					}
			}

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
		}
	*/

	x := <-webError
	fmt.Printf("WebError %v", x)

	fmt.Println("Exit")
}
