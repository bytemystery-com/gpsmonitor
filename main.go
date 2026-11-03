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
	"math"
	"os"
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
		sats := Data.Gps.GetSattelites()
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
				NumUsedSats:  getIntForJsonFromAtomic(&Data.Gps.GpsSumData.NumSats),
				NumTotalSats: &total,
				StdLongitude: getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.StdLongitude),
				StdLatitude:  getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.StdLatitude),
				StdAltitude:  getFloat32ForJsonFromAtomic(&Data.Gps.GpsSumData.StdAltitude),
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
	g, err := gps.NewGps(port, port, func() {
		port.ResetOutputBuffer()
		port.ResetInputBuffer()
	})
	if err != nil {
		port.Close()
		return
	}
	Data.Gps = g
	go g.Start(nil, true)

	var callBacks WebCallbacksImpl
	Data.WebServer = webserver.NewWebServer(callBacks, Data.Config)
	webError := make(chan error, 2)
	Data.WebServer.StartWebServer(webError)

	for {
		time.Sleep(250 * time.Millisecond)
		if Data.Gps.GpsSumData.IsReady.Load() {
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
}
