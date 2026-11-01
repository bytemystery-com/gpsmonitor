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

// GPS NMEA Parser

package gps

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"maps"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// https://www.nmea.org/nmea-0183.html
// https://avionicsduino.com/wp-content/uploads/2023/04/NEO-M9N_Interface_description_UBX-19035940.pdf
// https://www.ardusimple.com/nmea-protocol-and-message-description/

// NEO-M9N
// https://www.u-blox.com/sites/default/files/u-blox-M9-SPG-4.04_InterfaceDescription_UBX-21022436.pdf

const (
	KNOT_IN_MSEC               = (1852.0 / 3600.0)
	KMH_IN_MSEC                = (1.0 / 3.6)
	CLEAMUP_INTERVAL           = 5   // sec
	SATELLITE_TIMEOUT_INTERVAL = 120 // sec
)

type Trace interface {
	Print()
}

type Quality int

const (
	NotFix Quality = iota
	NormalFix
	DiffFix
	PPSFix
	RTKFix
	RTKFloat
	Estimated
	Manual
	Simul
)

func (q Quality) String() string {
	switch q {
	case NotFix:
		return "NotFix"
	case NormalFix:
		return "NormalFix"
	case DiffFix:
		return "DiffFix"
	case PPSFix:
		return "PPSFix"
	case RTKFix:
		return "RTKFix"
	case RTKFloat:
		return "RTKFloat"
	case Estimated:
		return "Estimated"
	case Manual:
		return "Manual"
	case Simul:
		return "Simul"
	default:
		return "???"
	}
}

type OpMode int

const (
	Manually OpMode = iota
	Automatically
)

type NavMode int

const (
	NavNoFix NavMode = iota + 1
	Nav2D
	Nav3D
)

type SignalId int

const (
	SIGNALID_UNKNOWN SignalId = iota
	GPS_L1CA
	GPS_L2CL
	GPS_L2CM
	GPS_L5I
	GPS_L5Q

	SBAS_L1CA

	GALILIEO_E1
	GALILIEO_E5A
	GALILIEO_E5B

	BEIDOU_B1I
	BEIDOU_B2I
	BEIDOU_B2A

	QZSS_L1CA
	QZSS_L1S
	QZSS_L2CM
	QZSS_L2CL
	QZSS_L5I
	QZSS_L5Q

	GLONASS_L1OF
	GLONASS_L2OF

	NAVIC_L5A
)

func (signalId SignalId) String() string {
	switch signalId {
	case SIGNALID_UNKNOWN:
		return "---"
	case GPS_L1CA:
		return "GPS L1 C/A"
	case GPS_L2CL:
		return "GPS L2 CL"
	case GPS_L2CM:
		return "GPS L2 CM"
	case GPS_L5I:
		return "GPS L5 I"
	case GPS_L5Q:
		return "GPS L5 Q"
	case SBAS_L1CA:
		return "SBAS L1 C/A"
	case GALILIEO_E1:
		return "Galileo E1"
	case GALILIEO_E5A:
		return "Galileo E5 a"
	case GALILIEO_E5B:
		return "Galileo E5 b"
	case BEIDOU_B1I:
		return "BeiDou B1 I"
	case BEIDOU_B2I:
		return "BeiDou B2 I"
	case BEIDOU_B2A:
		return "BeiDou B2 A"
	case QZSS_L1CA:
		return "QZSS L1 C/A"
	case QZSS_L1S:
		return "QZSS L1 S"
	case QZSS_L2CM:
		return "QZSS L2 CM"
	case QZSS_L2CL:
		return "QZSS L2 CL"
	case QZSS_L5I:
		return "QZSS L5 I"
	case QZSS_L5Q:
		return "QZSS L5 Q"
	case GLONASS_L1OF:
		return "GLONASS L1 OF"
	case GLONASS_L2OF:
		return "GLONASS L2 OF"
	case NAVIC_L5A:
		return "NaviC L5A"
	default:
		return "???"
	}
}

func (n NavMode) String() string {
	switch n {
	case NavNoFix:
		return "NoFix"
	case Nav2D:
		return "Nav2D"
	case Nav3D:
		return "Nav3D"
	default:
		return "???"
	}
}

type GGA struct {
	TalkerId      string
	Time          time.Time
	Latitude      float64
	Longitude     float64
	Quality       Quality
	NumSats       int
	HDop          float32
	Altitude      float32
	AltSep        float32 // Unterschied von Meereshöe zu Ellipsoid
	DiffAge       float32 // Alter der Korrekturdaten
	DiffStationId int
}

type SatDataDetail struct {
	Id       int
	Elev     int
	Azimuth  int
	Strength float32
}

type GSV struct {
	TalkerId string
	MsgTotal int
	MsgNr    int
	NumSat   int
	Sats     []SatDataDetail
	SignalId int
}

func GetSignalId(signalId int, gpsNetwork GPSNetwork) SignalId {
	switch signalId {
	case 1:
		switch gpsNetwork {
		case NetworkGPS:
			return GPS_L1CA
		case NetworkSBAS:
			return SBAS_L1CA
		case NetworkGalileo:
			return GALILIEO_E5A
		case NetworkBeiDou:
			return BEIDOU_B1I
		case NetworkQZSS:
			return QZSS_L1CA
		case NetworkGLONASS:
			return GLONASS_L1OF
		case NetworkNaviC:
			return NAVIC_L5A
		}
	case 2:
		switch gpsNetwork {
		case NetworkGalileo:
			return GALILIEO_E5B
		}
	case 3:
		switch gpsNetwork {
		case NetworkGLONASS:
			return GLONASS_L2OF
		}
	case 4:
		switch gpsNetwork {
		case NetworkQZSS:
			return QZSS_L1S
		}
	case 5:
		switch gpsNetwork {
		case NetworkGPS:
			return GPS_L2CM
		case NetworkBeiDou:
			return BEIDOU_B2A
		case NetworkQZSS:
			return QZSS_L2CM
		}
	case 6:
		switch gpsNetwork {
		case NetworkGPS:
			return GPS_L2CL
		case NetworkQZSS:
			return QZSS_L2CL
		}
	case 7:
		switch gpsNetwork {
		case NetworkGPS:
			return GPS_L5I
		case NetworkGalileo:
			return GALILIEO_E1
		case NetworkQZSS:
			return QZSS_L5I
		}
	case 8:
		switch gpsNetwork {
		case NetworkGPS:
			return GPS_L5Q
		case NetworkQZSS:
			return QZSS_L5Q
		}
	case 11:
		switch gpsNetwork {
		case NetworkBeiDou:
			return BEIDOU_B2I
		case NetworkQZSS:
			return QZSS_L5Q
		}
	}
	return SIGNALID_UNKNOWN
}

func getGpsNetworkFromSatId(id int) GPSNetwork {
	if id >= 1 && id <= 32 {
		return NetworkGPS
	}
	if id >= 65 && id <= 96 {
		return NetworkGLONASS
	}
	if (id >= 1 && id <= 36) || (id >= 301 && id <= 336) {
		return NetworkGalileo
	}
	if (id >= 1 && id <= 37) || (id >= 401 && id <= 437) {
		return NetworkBeiDou
	}
	if (id >= 33 && id <= 64) || (id >= 152 && id <= 158) {
		return NetworkSBAS
	}
	if id >= 173 && id <= 182 {
		return NetworkIMES
	}
	if id >= 193 && id <= 202 {
		return NetworkQZSS
	}
	if id >= 1 && id <= 14 { // wird nie erreicht
		return NetworkNaviC
	}
	return NetworkUnknown
}

// Seite 17 der Interface spec
func getNetwork(talkerId string, satId, systemId int) GPSNetwork {
	switch talkerId {
	case "GP":
		return NetworkGPS
	case "GL":
		return NetworkGLONASS
	case "GB":
		return NetworkBeiDou
	case "GA":
		return NetworkGalileo
	case "GQ":
		return NetworkQZSS
	case "GI":
		return NetworkNaviC
	default: // GN and unknown - try to identify by satID
		switch systemId {
		case 1:
			if (satId >= 33 && satId <= 64) || (satId >= 152 && satId <= 158) {
				return NetworkSBAS
			}
			return NetworkGPS
		case 2:
			return NetworkGLONASS
		case 3:
			return NetworkGalileo
		case 4:
			return NetworkBeiDou
		case 5:
			return NetworkQZSS
		case 6:
			return NetworkNaviC
		default:
			return getGpsNetworkFromSatId(satId)
		}
	}
}

func (gsv *GSV) getNetwork(sat *SatDataDetail) GPSNetwork {
	return getNetwork(gsv.TalkerId, sat.Id, -1)
}

type RMC struct {
	TalkerId   string
	DateTime   time.Time
	Latitude   float64
	Longitude  float64
	Valid      bool
	Speed      float32
	CourseTrue float32
	MagVar     float32
	Quality    Quality
}

type VTG struct {
	TalkerId   string
	CourseTrue float32
	CourseMag  float32
	Speed      float32
	Quality    Quality
}

type GLL struct {
	TalkerId  string
	Time      time.Time
	Latitude  float64
	Longitude float64
	Valid     bool
	Quality   Quality
}

type SatData struct {
	Id int
}

type GSA struct {
	TalkerId      string
	OperatingMode OpMode
	NavMode       NavMode
	SatDatas      []SatData
	Pdop          float32
	Vdop          float32
	Hdop          float32
	SystemId      int
}

func (gsa *GSA) getNetwork(sat *SatData) GPSNetwork {
	return getNetwork(gsa.TalkerId, sat.Id, gsa.SystemId)
}

type MsgType int

const (
	MsgError MsgType = iota
	MsgWarning
	MsgNotice
	MsgUser = 7
)

type TXT struct {
	TalkerId string
	MsgTotal int
	MsgNr    int
	MsgType  MsgType
	Txt      string
}

type GST struct {
	TalkerId     string
	Time         time.Time
	Rms          float32 // Rms von Satelitten
	StdMajor     float32 // Fehlerellipse - grosse Achse 95% Vertrauenintervall
	StdMinor     float32 // Fehlerellipse - kleine Achse 95% Vertrauenintervall
	Orientation  float32 // Fehlerellipse - Orientierung
	StdLatitude  float32 // 68% Vertrauenintervall
	StdLongitude float32 // 68% Vertrauenintervall
	StdAltitude  float32 // 68% Vertrauenintervall
}

func (g *GGA) Print() {
	log.Printf("GGA")
	log.Printf("  ID            : %s", g.TalkerId)
	log.Printf("  Time          : %v", g.Time)
	log.Printf("  Long          : %.5f", g.Longitude)
	log.Printf("  Lat           : %.5f", g.Latitude)
	log.Printf("  Alt           : %.2f", g.Altitude)
	log.Printf("  Quality       : %d", g.Quality)
	log.Printf("  NumSats       : %d", g.NumSats)
	log.Printf("  HDop          : %.2f", g.HDop)
	log.Printf("  DiffStationId : %d", g.DiffStationId)
	log.Printf("  Maps          : %s", GetGoogleMapsLink(g.Latitude, g.Longitude))
	log.Printf("  Earth         : %s", GetGoogleEarthLink(g.Latitude, g.Longitude))
}

func (r *RMC) Print() {
	log.Printf("RMC")
	log.Printf("  ID          : %s", r.TalkerId)
	log.Printf("  Valid       : %v", r.Valid)
	log.Printf("  Date / Time : %v", r.DateTime)
	log.Printf("  Long        : %.5f", r.Longitude)
	log.Printf("  Lat         : %.5f", r.Latitude)
	log.Printf("  Speed       : %.3f", r.Speed)
	log.Printf("  CourseTru   : %.3f", r.CourseTrue)
	log.Printf("  Quality     : %d", r.Quality)
	log.Printf("  MagVar      : %.3f", r.MagVar)
}

func (l *GLL) Print() {
	log.Printf("GLL")
	log.Printf("  ID          : %s", l.TalkerId)
	log.Printf("  Valid       : %v", l.Valid)
	log.Printf("  Time        : %v", l.Time)
	log.Printf("  Long        : %.5f", l.Longitude)
	log.Printf("  Lat         : %.5f", l.Latitude)
	log.Printf("  Quality     : %d", l.Quality)
}

func (v *VTG) Print() {
	log.Printf("VTG")
	log.Printf("  ID          : %s", v.TalkerId)
	log.Printf("  CourseTru   : %.3f", v.CourseTrue)
	log.Printf("  CourseMag   : %.3f", v.CourseTrue)
	log.Printf("  Speed       : %.3f", v.Speed)
	log.Printf("  Quality     : %d", v.Quality)
}

func (t *TXT) Print() {
	log.Printf("TXT")
	log.Printf("  ID          : %s", t.TalkerId)
	log.Printf("  TotalMsgs   : %d", t.MsgTotal)
	log.Printf("  MsgNr       : %d", t.MsgNr)
	log.Printf("  Type        : %d", t.MsgType)
	log.Printf("  Text        : %s", t.Txt)
}

func (g *GSA) Print() {
	s := ""
	for _, item := range g.SatDatas {
		if s != "" {
			s += ", "
		}
		s += strconv.Itoa(item.Id)
	}
	log.Printf("GSA")
	log.Printf("  ID          : %s", g.TalkerId)
	log.Printf("  OpMode      : %d", g.OperatingMode)
	log.Printf("  NavMode     : %v", g.NavMode)
	log.Printf("  SystemId    : %d", g.SystemId)
	log.Printf("  HDOP        : %.2f", g.Hdop)
	log.Printf("  VDOP        : %.2f", g.Vdop)
	log.Printf("  PDOP        : %.2f", g.Pdop)
	log.Printf("  Satellites  : %s", s)
}

func (g *GSV) Print() {
	log.Printf("GSV")
	log.Printf("  ID          : %s", g.TalkerId)
	log.Printf("  TotalMsgs   : %d", g.MsgTotal)
	log.Printf("  MsgNr       : %d", g.MsgNr)
	log.Printf("  NumSat      : %d", g.NumSat)
	for _, item := range g.Sats {
		log.Printf("    ID       : %d", item.Id)
		log.Printf("    Elev     : %d", item.Elev)
		log.Printf("    Azim     : %d", item.Azimuth)
		log.Printf("    Strength : %.3f", item.Strength)
	}

	log.Printf("  SignalID    : %d", g.SignalId)
}

func (g *GST) Print() {
	log.Printf("GST")
	log.Printf("  ID           : %s", g.TalkerId)
	log.Printf("  Time         : %v", g.Time)
	log.Printf("  Rms          : %.3f", g.Rms)
	log.Printf("  StdMajor     : %.3f", g.StdMajor)
	log.Printf("  StdMinor     : %.3f", g.StdMinor)
	log.Printf("  Orientation  : %.3f", g.Orientation)
	log.Printf("  StdLatitude  : %.3f", g.StdLatitude)
	log.Printf("  StdLongitude : %.3f", g.StdLongitude)
	log.Printf("  StdAltitude  : %.3f", g.StdAltitude)
}

type GPSNetwork int

const (
	NetworkUnknown GPSNetwork = iota
	NetworkGPS
	NetworkGLONASS
	NetworkGalileo
	NetworkBeiDou
	NetworkQZSS
	NetworkSBAS
	NetworkIMES
	NetworkNaviC
)

func (n GPSNetwork) String() string {
	switch n {
	case NetworkUnknown:
		return "Unknown"
	case NetworkGPS:
		return "GPS"
	case NetworkGLONASS:
		return "GLONASS"
	case NetworkGalileo:
		return "Galileo"
	case NetworkBeiDou:
		return "BeiDou"
	case NetworkQZSS:
		return "QZSS"
	case NetworkSBAS:
		return "SBAS"
	case NetworkIMES:
		return "IMES"
	case NetworkNaviC:
		return "NaviC"
	default:
		return "???"
	}
}

type satKey struct {
	network GPSNetwork
	id      int
}

type SatInfo struct {
	Network        GPSNetwork
	Id             int
	Elev           int
	Azimuth        int
	Strength       map[SignalId]float32 // SignalID as Key
	Used           bool
	LastUpdate     time.Time
	LastUsedUpdate time.Time
}

func (s *SatInfo) Print() {
	log.Printf("Sat:")
	log.Printf("  Id        : %d", s.Id)
	log.Printf("  Network   : %s", s.Network.String())
	log.Printf("  Elevation : %d", s.Elev)
	log.Printf("  Azimuth   : %d", s.Azimuth)
	log.Printf("  Used      : %v", s.Used)
	log.Printf("  Strength  :")
	for key, val := range s.Strength {
		log.Printf("    %-15s  : %.2f", key.String(), val)
	}
}

type GpsData struct {
	Time          atomic.Value  // time.Time
	Latitude      atomic.Uint64 // float64
	Longitude     atomic.Uint64 // float64
	Altitude      atomic.Uint32 // float32
	Speed         atomic.Uint32 // float32
	CourseTrue    atomic.Uint32 // float32
	MagVar        atomic.Uint32 // float32
	Quality       atomic.Uint32 // Quality
	NavMode       atomic.Uint32 // NavMode
	NumSats       atomic.Uint32 // int
	HDop          atomic.Uint32 // float32
	VDop          atomic.Uint32 // float32
	PDop          atomic.Uint32 // float32
	AltSep        atomic.Uint32 // float32 // Unterschied von Meereshöe zu Ellipsoid
	Rms           atomic.Uint32 // float32 // Rms von Satelitten
	StdMajor      atomic.Uint32 // float32 // Fehlerellipse - grosse Achse
	StdMinor      atomic.Uint32 // float32 // Fehlerellipse - kleine Achse
	Orientation   atomic.Uint32 // float32 // Fehlerellipse - Orientierung
	StdLatitude   atomic.Uint32 // float32
	StdLongitude  atomic.Uint32 // float32
	StdAltitude   atomic.Uint32 // float32
	DiffAge       atomic.Uint32 // float32 // Alter der Korrekturdaten
	DiffStationId atomic.Uint32 // int
	Messages      atomic.Value  // []string
}

func (gps *GpsData) Print() {
	log.Printf("GPSDATA:")
	t, _ := gps.Time.Load().(time.Time)
	log.Printf("  DateTime      : %v", t)
	log.Printf("  Latitude      : %.7f °", math.Float64frombits(gps.Latitude.Load()))
	log.Printf("  Longitude     : %.7f °", math.Float64frombits(gps.Longitude.Load()))
	log.Printf("  Altitude      : %.1f m", math.Float32frombits(gps.Altitude.Load()))
	log.Printf("  Speed         : %.1f m/s", math.Float32frombits(gps.Speed.Load()))
	log.Printf("  CourseTrue    : %.2f °", math.Float32frombits(gps.CourseTrue.Load()))
	log.Printf("  MagVar        : %.2f °", math.Float32frombits(gps.MagVar.Load()))
	log.Printf("  Quality       : %s", Quality(gps.Quality.Load()).String())
	log.Printf("  NavMode       : %s", NavMode(gps.NavMode.Load()).String())
	log.Printf("  NumSats       : %d", gps.NumSats.Load())
	log.Printf("  HDop          : %.1f", math.Float32frombits(gps.HDop.Load()))
	log.Printf("  VDop          : %.1f", math.Float32frombits(gps.VDop.Load()))
	log.Printf("  PDop          : %.1f", math.Float32frombits(gps.PDop.Load()))
	log.Printf("  AltSep        : %.1f m", math.Float32frombits(gps.AltSep.Load()))
	log.Printf("  Rms           : %.1f m", math.Float32frombits(gps.Rms.Load()))
	log.Printf("  StdMajor      : %.1f m", math.Float32frombits(gps.StdMajor.Load()))
	log.Printf("  StdMinor      : %.1f m", math.Float32frombits(gps.StdMinor.Load()))
	log.Printf("  Orientation   : %.1f °", math.Float32frombits(gps.Orientation.Load()))
	log.Printf("  StdLatitude   : %.1f m", math.Float32frombits(gps.StdLatitude.Load()))
	log.Printf("  StdLongitude  : %.1f m", math.Float32frombits(gps.StdLongitude.Load()))
	log.Printf("  StdAltitude   : %.1f m", math.Float32frombits(gps.StdAltitude.Load()))
	log.Printf("  DiffAge       : %.1f s", math.Float32frombits(gps.DiffAge.Load()))
	log.Printf("  DiffStationId : %d", gps.DiffStationId.Load())
	log.Printf("  Messages:")
	m, _ := gps.Messages.Load().([]string)
	for _, item := range m {
		log.Printf("    - %s", item)
	}
}

func (gps *GpsData) init() {
	gps.Time.Store(time.Time{})
	gps.Latitude.Store(math.Float64bits(math.NaN()))
	gps.Longitude.Store(math.Float64bits(math.NaN()))
	gps.Altitude.Store(math.Float32bits(float32(math.NaN())))
	gps.Speed.Store(math.Float32bits(float32(math.NaN())))
	gps.CourseTrue.Store(math.Float32bits(float32(math.NaN())))
	gps.MagVar.Store(math.Float32bits(float32(math.NaN())))
	gps.Quality.Store(^uint32(0))
	gps.NavMode.Store(^uint32(0))
	gps.NumSats.Store(^uint32(0))
	gps.HDop.Store(math.Float32bits(float32(math.NaN())))
	gps.VDop.Store(math.Float32bits(float32(math.NaN())))
	gps.PDop.Store(math.Float32bits(float32(math.NaN())))
	gps.AltSep.Store(math.Float32bits(float32(math.NaN())))
	gps.Rms.Store(math.Float32bits(float32(math.NaN())))
	gps.StdMajor.Store(math.Float32bits(float32(math.NaN())))
	gps.StdMinor.Store(math.Float32bits(float32(math.NaN())))
	gps.Orientation.Store(math.Float32bits(float32(math.NaN())))
	gps.StdLatitude.Store(math.Float32bits(float32(math.NaN())))
	gps.StdLongitude.Store(math.Float32bits(float32(math.NaN())))
	gps.StdAltitude.Store(math.Float32bits(float32(math.NaN())))
	gps.DiffAge.Store(math.Float32bits(float32(math.NaN())))
	gps.DiffStationId.Store(^uint32(0))
}

type Gps struct {
	Messages []any
	reader   io.Reader
	writer   io.Writer
	notify   chan<- any

	satMap      map[satKey]SatInfo
	satInfoLock sync.RWMutex

	GpsSumData GpsData
	Trace      bool
}

func NewGps(reader io.Reader, writer io.Writer) (*Gps, error) {
	gps := &Gps{
		Messages: make([]any, 0, 100),
		reader:   reader,
		writer:   writer,
		satMap:   make(map[satKey]SatInfo, 20),
	}
	gps.GpsSumData.init()
	return gps, nil
}

func (gps *Gps) parseFloat32(s string) float32 {
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return float32(f)
	} else {
		return float32(math.NaN())
	}
}

func (gps *Gps) parseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return n
}

func (gps *Gps) SendNotify(val any) {
	if gps.notify != nil {
		select {
		case gps.notify <- val:
		default:
			log.Println("Notification channel blocked, skipping frame.")
		}
	}
}

func (gps *Gps) parseLatitude(s string) float64 {
	if len(s) < 3 {
		return math.NaN()
	}
	dd, err := strconv.Atoi(s[:2])
	if err != nil {
		return math.NaN()
	}
	mm, err := strconv.ParseFloat(s[2:], 64)
	if err != nil {
		return math.NaN()
	}
	return float64(dd) + mm/60.0
}

func (gps *Gps) parseLongitude(s string) float64 {
	if len(s) < 4 {
		return math.NaN()
	}
	dd, err := strconv.Atoi(s[:3])
	if err != nil {
		return math.NaN()
	}
	mm, err := strconv.ParseFloat(s[3:], 64)
	if err != nil {
		return math.NaN()
	}
	return float64(dd) + mm/60.0
}

func (gps *Gps) verifyNmeaChecksum(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	length := len(s)
	if length < 4 {
		return nil, fmt.Errorf("empty or invalid string")
	}
	if s[0] != '$' || s[length-3] != '*' {
		return nil, fmt.Errorf("incomplete")
	}
	line, crc, ok := strings.Cut(s[:length], "*")
	if !ok {
		return nil, fmt.Errorf("* not found")
	}
	sArray := strings.Split(line, ",")
	var cSum byte = 0
	for i := 1; i < len(line); i++ {
		cSum ^= line[i]
	}
	c, err := strconv.ParseUint(crc, 16, 8)
	if err != nil {
		return nil, err
	}
	if byte(c) != byte(cSum) {
		return nil, fmt.Errorf("checksum mismatch %v / %v", c, cSum)
	}
	return sArray, nil
}

func (gps *Gps) StoreFloat64InAtomic(v float64, a *atomic.Uint64) {
	if math.IsNaN(v) {
		return
	}
	a.Store(math.Float64bits(v))
}

func (gps *Gps) StoreFloat32InAtomic(v float32, a *atomic.Uint32) {
	if math.IsNaN(float64(v)) {
		return
	}
	a.Store(math.Float32bits(v))
}

func (gps *Gps) StoreIntInAtomic(v int, a *atomic.Uint32) {
	if v < 0 {
		return
	}
	a.Store(uint32(v))
}

func (gps *Gps) StoreTimeInAtomic(v time.Time, a *atomic.Value) {
	if v.IsZero() {
		return
	}
	a.Store(v)
}

func (gps *Gps) LoadTimeFromAtomic(a *atomic.Value) time.Time {
	var td time.Time
	v := a.Load()
	if v != nil {
		t, ok := v.(time.Time)
		if ok {
			td = t
		}
	}
	return td
}

func (gps *Gps) UpdateTimeInAtomic(t time.Time, a *atomic.Value) {
	told := gps.LoadTimeFromAtomic(a)
	t = time.Date(told.Year(), told.Month(), told.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	gps.StoreTimeInAtomic(t, a)
}

func (gps *Gps) LoadStringSliceFromAtomic(a *atomic.Value) []string {
	var s []string
	v := a.Load()
	if v != nil {
		t, ok := v.([]string)
		if ok {
			s = t
		}
	}
	return s
}

func (gps *Gps) AddStringToSliceInAtomic(msg string, maxNum int, a *atomic.Value) {
	if msg != "" {
		sl := gps.LoadStringSliceFromAtomic(a)
		l := min(maxNum, len(sl)+1)
		nsl := make([]string, l)
		if len(sl) > 0 {
			if len(sl) >= maxNum {
				copy(nsl, sl[1:])
			} else {
				copy(nsl, sl)
			}
		}
		nsl[len(nsl)-1] = msg
		a.Store(nsl)
	}
}

func (gps *Gps) parseGGA(sArray []string) (*GGA, error) {
	l := len(sArray)
	if l != 15 {
		return nil, fmt.Errorf("unexpected number of fields")
	}

	g := GGA{
		TalkerId: sArray[0][1:3],
	}

	t, err := time.Parse("150405.00", sArray[1])
	if err == nil {
		g.Time = t
	}

	g.Latitude = gps.parseLatitude(sArray[2])
	if sArray[3] == "S" {
		g.Latitude = -g.Latitude
	}

	g.Longitude = gps.parseLongitude(sArray[4])
	if sArray[5] == "W" {
		g.Longitude = -g.Longitude
	}

	g.Quality = Quality(gps.parseInt(sArray[6]))

	g.NumSats = gps.parseInt(sArray[7])

	g.HDop = gps.parseFloat32(sArray[8])

	g.Altitude = gps.parseFloat32(sArray[9])

	g.AltSep = gps.parseFloat32(sArray[11])

	g.DiffAge = gps.parseFloat32(sArray[13])

	g.DiffStationId = gps.parseInt(sArray[14])

	return &g, nil
}

func (gps *Gps) getQualityFromLetter(s string) Quality {
	switch s {
	case "A":
		return NormalFix
	case "N":
		return NotFix
	case "E":
		return Estimated
	case "D":
		return DiffFix
	case "F":
		return RTKFloat
	case "R":
		return RTKFix
	}
	return -1
}

func (gps *Gps) parseRMC(sArray []string) (*RMC, error) {
	l := len(sArray)
	if l != 14 {
		return nil, fmt.Errorf("unexpected number of fields")
	}

	r := RMC{
		TalkerId: sArray[0][1:3],
	}

	t, err := time.Parse("150405.00", sArray[1])
	if err == nil {
		r.DateTime = t
	}

	if sArray[2] == "A" {
		r.Valid = true
	}

	r.Latitude = gps.parseLatitude(sArray[3])
	if sArray[4] == "S" {
		r.Latitude = -r.Latitude
	}

	r.Longitude = gps.parseLongitude(sArray[5])
	if sArray[6] == "W" {
		r.Longitude = -r.Longitude
	}

	r.Speed = gps.parseFloat32(sArray[7]) * KNOT_IN_MSEC

	r.CourseTrue = gps.parseFloat32(sArray[8])

	d, err := time.Parse("020106", sArray[9])
	if err == nil {
		r.DateTime = time.Date(
			d.Year(), d.Month(), d.Day(), r.DateTime.Hour(), r.DateTime.Minute(), r.DateTime.Second(), r.DateTime.Nanosecond(), time.UTC)
	}
	r.MagVar = gps.parseFloat32(sArray[10])
	if sArray[11] == "W" {
		r.MagVar = -r.MagVar
	}

	r.Quality = gps.getQualityFromLetter(sArray[12])

	return &r, nil
}

func (gps *Gps) parseGLL(sArray []string) (*GLL, error) {
	l := len(sArray)
	if l != 9 {
		return nil, fmt.Errorf("unexpected number of fields")
	}
	g := GLL{
		TalkerId: sArray[0][1:3],
	}

	g.Latitude = gps.parseLatitude(sArray[1])
	if sArray[2] == "S" {
		g.Latitude = -g.Latitude
	}

	g.Longitude = gps.parseLongitude(sArray[3])
	if sArray[4] == "W" {
		g.Longitude = -g.Longitude
	}

	t, err := time.Parse("150405.00", sArray[5])
	if err == nil {
		g.Time = t
	}

	if sArray[6] == "A" {
		g.Valid = true
	}

	g.Quality = gps.getQualityFromLetter(sArray[12])

	return &g, nil
}

func (gps *Gps) parseGSA(sArray []string) (*GSA, error) {
	l := len(sArray)
	if l != 19 {
		return nil, fmt.Errorf("unexpected number of fields")
	}
	g := GSA{
		TalkerId: sArray[0][1:3],
		SatDatas: make([]SatData, 0, 12),
	}

	if sArray[1] == "A" {
		g.OperatingMode = Automatically
	}

	g.NavMode = NavMode(gps.parseInt(sArray[2]))

	for i := range 12 {
		if sArray[3+i] != "" {
			n, err := strconv.Atoi(sArray[3+i])
			if err == nil {
				g.SatDatas = append(g.SatDatas, SatData{Id: n})
			}
		}
	}
	g.Pdop = gps.parseFloat32(sArray[15])

	g.Hdop = gps.parseFloat32(sArray[16])

	g.Vdop = gps.parseFloat32(sArray[17])

	g.SystemId = gps.parseInt(sArray[18])

	return &g, nil
}

func (gps *Gps) parseVTG(sArray []string) (*VTG, error) {
	l := len(sArray)
	if l != 10 {
		return nil, fmt.Errorf("unexpected number of fields")
	}
	v := VTG{
		TalkerId: sArray[0][1:3],
	}

	v.CourseTrue = gps.parseFloat32(sArray[1])

	v.CourseMag = gps.parseFloat32(sArray[3])

	v.Speed = gps.parseFloat32(sArray[5]) * KNOT_IN_MSEC

	v.Speed = gps.parseFloat32(sArray[7]) * KMH_IN_MSEC

	v.Quality = gps.getQualityFromLetter(sArray[9])

	return &v, nil
}

func (gps *Gps) parseGSV(sArray []string) (*GSV, error) {
	l := len(sArray)
	if l < 5 {
		return nil, fmt.Errorf("unexpected number of fields")
	}
	g := GSV{
		TalkerId: sArray[0][1:3],
		Sats:     make([]SatDataDetail, 0, 4),
	}

	g.MsgTotal = gps.parseInt(sArray[1])

	g.MsgNr = gps.parseInt(sArray[2])

	g.NumSat = gps.parseInt(sArray[3])

	n := (l - 5) / 4
	for i := range n {
		d := SatDataDetail{
			Id:       gps.parseInt(sArray[4+i*4]),
			Elev:     gps.parseInt(sArray[5+i*4]),
			Azimuth:  gps.parseInt(sArray[6+i*4]),
			Strength: gps.parseFloat32(sArray[7+i*4]),
		}
		g.Sats = append(g.Sats, d)
	}
	g.SignalId = gps.parseInt(sArray[4+n*4])

	return &g, nil
}

func (gps *Gps) parseTXT(sArray []string) (*TXT, error) {
	l := len(sArray)
	if l != 5 {
		return nil, fmt.Errorf("unexpected number of fields")
	}
	t := TXT{
		TalkerId: sArray[0][1:3],
	}

	t.MsgTotal = gps.parseInt(sArray[1])

	t.MsgNr = gps.parseInt(sArray[2])

	t.MsgType = MsgType(gps.parseInt(sArray[3]))

	t.Txt = sArray[4]

	return &t, nil
}

func (gps *Gps) parseGST(sArray []string) (*GST, error) {
	l := len(sArray)
	if l != 9 {
		return nil, fmt.Errorf("unexpected number of fields")
	}
	g := GST{
		TalkerId: sArray[0][1:3],
	}

	t, err := time.Parse("150405.00", sArray[1])
	if err == nil {
		g.Time = t
	}

	g.Rms = gps.parseFloat32(sArray[2])

	g.StdMajor = gps.parseFloat32(sArray[3])

	g.StdMinor = gps.parseFloat32(sArray[4])

	g.Orientation = gps.parseFloat32(sArray[5])

	g.StdLatitude = gps.parseFloat32(sArray[6])

	g.StdLongitude = gps.parseFloat32(sArray[7])

	g.StdAltitude = gps.parseFloat32(sArray[8])

	return &g, nil
}

func (gps *Gps) appendMessage(msg any) {
	if len(gps.Messages) < cap(gps.Messages) {
		gps.Messages = append(gps.Messages, msg)
	} else {
		copy(gps.Messages, gps.Messages[1:])
		gps.Messages[len(gps.Messages)-1] = msg
	}
}

func (gps *Gps) parseNmea(line string) {
	if gps.Trace {
		log.Printf("Line: %s", line)
	}
	sArray, err := gps.verifyNmeaChecksum(line)
	if err != nil {
		return
	}
	// log.Printf("Token: %s", sArray[0])
	switch sArray[0][3:] {
	case "GGA":
		gga, err := gps.parseGGA(sArray)
		if err == nil {
			gps.appendMessage(gga)

			gps.UpdateTimeInAtomic(gga.Time, &gps.GpsSumData.Time)
			gps.StoreFloat64InAtomic(gga.Latitude, &gps.GpsSumData.Latitude)
			gps.StoreFloat64InAtomic(gga.Longitude, &gps.GpsSumData.Longitude)
			gps.StoreIntInAtomic(int(gga.Quality), &gps.GpsSumData.Quality)
			gps.StoreIntInAtomic(gga.NumSats, &gps.GpsSumData.NumSats)
			gps.StoreFloat32InAtomic(gga.HDop, &gps.GpsSumData.HDop)
			gps.StoreFloat32InAtomic(gga.Altitude, &gps.GpsSumData.Altitude)
			gps.StoreFloat32InAtomic(gga.AltSep, &gps.GpsSumData.AltSep)
			gps.StoreFloat32InAtomic(gga.DiffAge, &gps.GpsSumData.DiffAge)
			gps.StoreIntInAtomic(gga.DiffStationId, &gps.GpsSumData.DiffStationId)

			gps.SendNotify(gga)
		}

	case "RMC":
		rmc, err := gps.parseRMC(sArray)
		if err == nil {
			gps.appendMessage(rmc)

			if rmc.Valid {
				gps.StoreTimeInAtomic(rmc.DateTime, &gps.GpsSumData.Time)
			}
			gps.StoreFloat64InAtomic(rmc.Latitude, &gps.GpsSumData.Latitude)
			gps.StoreFloat64InAtomic(rmc.Longitude, &gps.GpsSumData.Longitude)
			gps.StoreFloat32InAtomic(rmc.Speed, &gps.GpsSumData.Speed)
			gps.StoreFloat32InAtomic(rmc.CourseTrue, &gps.GpsSumData.CourseTrue)
			gps.StoreFloat32InAtomic(rmc.MagVar, &gps.GpsSumData.MagVar)
			gps.StoreIntInAtomic(int(rmc.Quality), &gps.GpsSumData.Quality)

			gps.SendNotify(rmc)
		}

	case "GLL":
		gll, err := gps.parseGLL(sArray)
		if err == nil {
			gps.appendMessage(gll)

			if gll.Valid {
				gps.UpdateTimeInAtomic(gll.Time, &gps.GpsSumData.Time)
				gps.StoreFloat64InAtomic(gll.Latitude, &gps.GpsSumData.Latitude)
				gps.StoreFloat64InAtomic(gll.Longitude, &gps.GpsSumData.Longitude)
				gps.StoreIntInAtomic(int(gll.Quality), &gps.GpsSumData.Quality)
			}

			gps.SendNotify(gll)
		}

	case "GSA":
		gsa, err := gps.parseGSA(sArray)
		if err == nil {
			gps.appendMessage(gsa)

			gps.StoreIntInAtomic(int(gsa.NavMode), &gps.GpsSumData.NavMode)
			gps.StoreFloat32InAtomic(gsa.Pdop, &gps.GpsSumData.PDop)
			gps.StoreFloat32InAtomic(gsa.Vdop, &gps.GpsSumData.VDop)
			gps.StoreFloat32InAtomic(gsa.Hdop, &gps.GpsSumData.HDop)

			gps.satInfoLock.Lock()
			for _, item := range gsa.SatDatas {
				key := satKey{
					id:      item.Id,
					network: gsa.getNetwork(&item),
				}
				info, ok := gps.satMap[key]
				if !ok {
					info = SatInfo{
						Id:       item.Id,
						Network:  key.network,
						Strength: make(map[SignalId]float32, 5),
						Azimuth:  -1,
						Elev:     -1,
					}
				}
				info.Used = true
				now := time.Now()
				info.LastUsedUpdate = now
				info.LastUpdate = now

				gps.satMap[key] = info
			}
			gps.satInfoLock.Unlock()

			gps.SendNotify(gsa)
		}

	case "VTG":
		vtg, err := gps.parseVTG(sArray)
		if err == nil {
			gps.appendMessage(vtg)

			gps.StoreFloat32InAtomic(vtg.CourseTrue, &gps.GpsSumData.CourseTrue)
			gps.StoreFloat32InAtomic(vtg.Speed, &gps.GpsSumData.Speed)
			gps.StoreIntInAtomic(int(vtg.Quality), &gps.GpsSumData.Quality)

			gps.SendNotify(vtg)
		}

	case "GSV":
		gsv, err := gps.parseGSV(sArray)
		if err == nil {
			gps.appendMessage(gsv)
			gps.satInfoLock.Lock()
			for _, item := range gsv.Sats {
				key := satKey{
					id:      item.Id,
					network: gsv.getNetwork(&item),
				}
				info, ok := gps.satMap[key]
				if !ok {
					info = SatInfo{
						Id:       item.Id,
						Network:  key.network,
						Strength: make(map[SignalId]float32, 5),
					}
				}
				info.Strength[GetSignalId(gsv.SignalId, key.network)] = item.Strength
				info.Elev = item.Elev
				info.Azimuth = item.Azimuth
				info.LastUpdate = time.Now()
				gps.satMap[key] = info
			}
			gps.satInfoLock.Unlock()
			gps.SendNotify(gsv)
		}

	case "TXT":
		txt, err := gps.parseTXT(sArray)
		if err == nil {
			gps.appendMessage(txt)
			gps.AddStringToSliceInAtomic(txt.Txt, 20, &gps.GpsSumData.Messages)
			gps.SendNotify(txt)
		}

	case "GST":
		gst, err := gps.parseGST(sArray)
		if err == nil {
			gps.appendMessage(gst)

			gps.UpdateTimeInAtomic(gst.Time, &gps.GpsSumData.Time)
			gps.StoreFloat32InAtomic(gst.Rms, &gps.GpsSumData.Rms)
			gps.StoreFloat32InAtomic(gst.StdMajor, &gps.GpsSumData.StdMajor)
			gps.StoreFloat32InAtomic(gst.StdMinor, &gps.GpsSumData.StdMinor)
			gps.StoreFloat32InAtomic(gst.Orientation, &gps.GpsSumData.Orientation)
			gps.StoreFloat32InAtomic(gst.StdLatitude, &gps.GpsSumData.StdLatitude)
			gps.StoreFloat32InAtomic(gst.StdLongitude, &gps.GpsSumData.StdLongitude)
			gps.StoreFloat32InAtomic(gst.StdAltitude, &gps.GpsSumData.StdAltitude)

			gps.SendNotify(gst)
		}

	default:
		log.Printf("Unknownn token: %s", sArray[0])
	}
}

func (gps *Gps) ubxChecksum(data []byte) (byte, byte) {
	var ckA, ckB byte
	for _, b := range data {
		ckA += b
		ckB += ckA
	}
	return ckA, ckB
}

func (gps *Gps) SendUBXConfig(layer byte, key uint32, value uint32) error {
	// 1. Bestimmen, wie groß der Wert ist, basierend auf der u-blox Key-ID.
	// Die oberen Bits der Key-ID verraten die Datengröße (0x01 = 1-bit, 0x02 = 1-byte, 0x03 = 2-byte, 0x05 = 4-byte)

	// Korrektur: Mit 0x0F maskieren, um das komplette obere Hex-Nibble zu bekommen
	sizeGroup := (key >> 28) & 0x0F
	var valueSize int

	switch sizeGroup {
	case 0x01, 0x02: // 1-Bit oder 1-Byte
		valueSize = 1
	case 0x03: // 2-Byte
		valueSize = 2
	case 0x04: // 4-Byte (HIER landet jetzt deine Baudrate 0x40520001!)
		valueSize = 4
	case 0x05: // 8-Byte
		valueSize = 8
	default:
		valueSize = 1
	}

	// 2. Payload vorbereiten (8 Bytes Header + dynamische Wertgröße)
	payloadSize := 8 + valueSize
	payload := make([]byte, payloadSize)

	payload[0] = 0x00            // Version
	payload[1] = layer           // 0x01 = RAM, 0x07 = Flash
	payload[2] = 0x00            // Reserviert
	payload[3] = 0x00            // Reserviert
	payload[4] = byte(key)       // Key Byte 0 (Little Endian)
	payload[5] = byte(key >> 8)  // Key Byte 1
	payload[6] = byte(key >> 16) // Key Byte 2
	payload[7] = byte(key >> 24) // Key Byte 3

	// Wert in Little Endian in die Payload schreiben
	for i := 0; i < valueSize; i++ {
		payload[8+i] = byte(value >> (8 * i))
	}

	// 3. Header vorbereiten (6 Bytes)
	header := []byte{
		0xB5, 0x62, // Sync Chars
		0x06, 0x8A, // Class (0x06) & ID (0x8A) für UBX-CFG-VALSET
		byte(len(payload)), byte(len(payload) >> 8), // Dynamische Länge der Payload
	}

	// 4. Komplettes Paket zusammensetzen
	packet := append(header, payload...)

	// Prüfsumme berechnen (ohne die ersten zwei Sync-Chars)
	cka, ckb := gps.ubxChecksum(packet[2:])
	packet = append(packet, cka, ckb)

	// 5. Paket abschicken
	_, err := gps.writer.Write(packet)
	return err
}

func (gps *Gps) doSatMapCleanup() {
	for {
		time.Sleep(CLEAMUP_INTERVAL * time.Second)
		gps.satInfoLock.Lock()
		now := time.Now()
		for key, item := range gps.satMap {
			if now.Sub(item.LastUpdate) > SATELLITE_TIMEOUT_INTERVAL*time.Second {
				delete(gps.satMap, key)
			} else if item.Used && now.Sub(item.LastUsedUpdate) > CLEAMUP_INTERVAL*time.Second {
				item.Used = false
				gps.satMap[key] = item
			}
		}
		gps.satInfoLock.Unlock()
	}
}

func (gps *Gps) Start(notify chan<- any, isUbx bool) {
	gps.notify = notify
	reader := bufio.NewReader(gps.reader)
	if isUbx {
		// gps.SendUBXConfig(0x01, 0x40520001, 115200)
		gps.SendUBXConfig(0x01, 0x20930001, 41) // NMEA 4.11
		gps.SendUBXConfig(0x01, 0x20930007, 1)  // Extended einschalten
		gps.SendUBXConfig(0x01, 0x209100d4, 1)  // GST einschalten
		gps.SendUBXConfig(0x01, 0x20910087, 1)  // TXT einschalten
		// gps.SendUBXConfig(0x01, 0x209100b2, 1)  // GNS einschalten ???
		gps.SendUBXConfig(0x01, 0x209100b7, 1) // VTG einschalten
	}

	go gps.doSatMapCleanup()

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return
		}
		switch b {
		case '$':
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = "$" + line
			gps.parseNmea(line)

		// UBX stuff
		case 0xB5:
			// UBX: zweites Synchronisationsbyte prüfen
			b2, err := reader.ReadByte()
			if err != nil {
				return
			}

			if b2 == 0x62 {
				// TODO: UBX-Nachricht anhand der Länge lesen
				// und Prüfsumme prüfen
				classId, err := reader.ReadByte()
				if err != nil {
					return
				}
				messageId, err := reader.ReadByte()
				if err != nil {
					return
				}
				bufLength := make([]byte, 2)
				_, err = io.ReadFull(reader, bufLength)
				if err != nil {
					return
				}
				l := binary.LittleEndian.Uint16(bufLength)
				bufPayload := make([]byte, l)
				_, err = io.ReadFull(reader, bufPayload)
				if err != nil {
					return
				}
				log.Printf("Class=%02X Message=%02X Length=%d Payload=% X", classId, messageId, l, bufPayload)
				// log.Printf("Payload %v", bufPayload)
				bufCrc := make([]byte, 2)
				_, err = io.ReadFull(reader, bufCrc)
				if err != nil {
					return
				}
				data := make([]byte, 0, 4+len(bufPayload))
				data = append(data, classId, messageId)
				data = append(data, bufLength...)
				data = append(data, bufPayload...)
				ckA, ckB := gps.ubxChecksum(data)
				if bufCrc[0] != ckA || bufCrc[1] != ckB {
					log.Printf("CRC Fehler")
				}
			}
		}
	}
}

func (gps *Gps) GetSattelites() []SatInfo {
	gps.satInfoLock.RLock()
	defer gps.satInfoLock.RUnlock()

	result := make([]SatInfo, 0, len(gps.satMap))
	for _, item := range gps.satMap {
		item.Strength = maps.Clone(item.Strength)
		result = append(result, item)
	}
	return result
}

// PrintGoogleMapsLink gibt den klickbaren URL-Link in der Konsole aus
func GetGoogleMapsLink(latitude, longitude float64) string {
	return fmt.Sprintf("https://www.google.com/maps?q=%f,%f", latitude, longitude)
}

func GetGoogleEarthLink(latitude, longitude float64) string {
	return fmt.Sprintf("https://earth.google.com/web/search/%.8f,%.8f", latitude, longitude)
}
