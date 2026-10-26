package gps

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

// https://www.nmea.org/nmea-0183.html
// https://avionicsduino.com/wp-content/uploads/2023/04/NEO-M9N_Interface_description_UBX-19035940.pdf
// https://www.ardusimple.com/nmea-protocol-and-message-description/

const (
	KNOT_IN_MSEC = (1852.0 / 3600.0)
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

type GGA struct {
	TalkerId    string
	Time        time.Time
	Latitude    float64
	Longitude   float64
	Quality     Quality
	NumSats     int
	HDop        float32
	Altitude    float32
	AltSep      float32
	DiffAge     float32
	DiffStation int
}

type RMC struct {
	TalkerId  string
	DateTime  time.Time
	Latitude  float64
	Longitude float64
	Valid     bool
	Speed     float32
	Course    float32
	MagVar    float32
	Quality   Quality
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
	Number int
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

func (g *GGA) Print() {
	log.Printf("GGA")
	log.Printf("  ID      : %s", g.TalkerId)
	log.Printf("  Time    : %v", g.Time)
	log.Printf("  Long    : %.5f", g.Longitude)
	log.Printf("  Lat     : %.5f", g.Latitude)
	log.Printf("  Alt     : %.2f", g.Altitude)
	log.Printf("  Quality : %d", g.Quality)
	log.Printf("  NumSats : %d", g.NumSats)
	log.Printf("  HDop    : %.2f", g.HDop)
}

func (r *RMC) Print() {
	log.Printf("RMC")
	log.Printf("  ID          : %s", r.TalkerId)
	log.Printf("  Valid       : %v", r.Valid)
	log.Printf("  Date / Time : %v", r.DateTime)
	log.Printf("  Long        : %.5f", r.Longitude)
	log.Printf("  Lat         : %.5f", r.Latitude)
	log.Printf("  Speed       : %.3f", r.Speed)
	log.Printf("  Course      : %.3f", r.Course)
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

func (g *GSA) Print() {
	s := ""
	for _, item := range g.SatDatas {
		if s != "" {
			s += ", "
		}
		s += strconv.Itoa(item.Number)
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

type Gps struct {
	Messages []any
	reader   io.Reader
	writer   io.Writer
	notify   chan<- any
}

func NewGps(reader io.Reader, writer io.Writer) (*Gps, error) {
	return &Gps{
		Messages: make([]any, 0, 100),
		reader:   reader,
		writer:   writer,
	}, nil
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

func (gps *Gps) parseLatitude(s string) (float64, error) {
	dd, err := strconv.Atoi(s[:2])
	if err != nil {
		return 0, err
	}
	mm, err := strconv.ParseFloat(s[2:], 64)
	if err != nil {
		return 0, err
	}
	return float64(dd) + mm/60.0, nil
}

func (gps *Gps) parseLongitude(s string) (float64, error) {
	dd, err := strconv.Atoi(s[:3])
	if err != nil {
		return 0, err
	}
	mm, err := strconv.ParseFloat(s[3:], 64)
	if err != nil {
		return 0, err
	}
	return float64(dd) + mm/60.0, nil
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
	f, err := gps.parseLatitude(sArray[2])
	if err == nil {
		g.Latitude = f
		if sArray[3] == "S" {
			g.Latitude = -g.Latitude
		}
	}
	f, err = gps.parseLongitude(sArray[4])
	if err == nil {
		g.Longitude = f
		if sArray[5] == "W" {
			g.Longitude = -g.Longitude
		}
	}
	q, err := strconv.Atoi(sArray[6])
	if err == nil {
		g.Quality = Quality(q)
	}
	q, err = strconv.Atoi(sArray[7])
	if err == nil {
		g.NumSats = q
	}
	v, err := strconv.ParseFloat(sArray[8], 64)
	if err == nil {
		g.HDop = float32(v)
	}
	v, err = strconv.ParseFloat(sArray[9], 64)
	if err == nil {
		g.Altitude = float32(v)
	}
	v, err = strconv.ParseFloat(sArray[11], 64)
	if err == nil {
		g.AltSep = float32(v)
	}
	v, err = strconv.ParseFloat(sArray[13], 64)
	if err == nil {
		g.DiffAge = float32(v)
	}
	q, err = strconv.Atoi(sArray[14])
	if err == nil {
		g.DiffStation = q
	}
	return &g, nil
}

func (gps *Gps) getQualityFromLetter(s string) (Quality, error) {
	switch s {
	case "A":
		return NormalFix, nil
	case "N":
		return NotFix, nil
	case "E":
		return Estimated, nil
	case "D":
		return DiffFix, nil
	case "F":
		return RTKFloat, nil
	case "R":
		return RTKFix, nil
	}
	return 0, fmt.Errorf("invalid charakter %s for quality", s)
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
	f, err := gps.parseLatitude(sArray[3])
	if err == nil {
		r.Latitude = f
		if sArray[4] == "S" {
			r.Latitude = -r.Latitude
		}
	}
	f, err = gps.parseLongitude(sArray[5])
	if err == nil {
		r.Longitude = f
		if sArray[6] == "W" {
			r.Longitude = -r.Longitude
		}
	}
	f, err = strconv.ParseFloat(sArray[7], 64)
	if err == nil {
		r.Speed = float32(f * KNOT_IN_MSEC)
	}
	f, err = strconv.ParseFloat(sArray[8], 64)
	if err == nil {
		r.Course = float32(f)
	}
	d, err := time.Parse("020106", sArray[9])
	if err == nil {
		r.DateTime = time.Date(
			d.Year(), d.Month(), d.Day(), r.DateTime.Hour(), r.DateTime.Minute(), r.DateTime.Second(), r.DateTime.Nanosecond(), time.UTC)
	}
	f, err = strconv.ParseFloat(sArray[10], 64)
	if err == nil {
		r.MagVar = float32(f)
		if sArray[11] == "W" {
			r.MagVar = -r.MagVar
		}
	}
	q, err := gps.getQualityFromLetter(sArray[12])
	if err == nil {
		r.Quality = q
	}
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
	lo, err := gps.parseLatitude(sArray[1])
	if err == nil {
		g.Latitude = lo
		if sArray[2] == "S" {
			g.Latitude = -g.Latitude
		}
	}
	la, err := gps.parseLongitude(sArray[3])
	if err == nil {
		g.Longitude = la
		if sArray[4] == "W" {
			g.Longitude = -g.Longitude
		}
	}

	t, err := time.Parse("150405.00", sArray[5])
	if err == nil {
		g.Time = t
	}

	if sArray[6] == "A" {
		g.Valid = true
	}

	q, err := gps.getQualityFromLetter(sArray[12])
	if err == nil {
		g.Quality = q
	}

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

	n, err := strconv.Atoi(sArray[2])
	if err == nil {
		g.NavMode = NavMode(n)
	}
	for i := range 12 {
		if sArray[3+i] != "" {
			n, err = strconv.Atoi(sArray[3+i])
			if err == nil {
				g.SatDatas = append(g.SatDatas, SatData{Number: n})
			}
		}
	}
	f, err := strconv.ParseFloat(sArray[15], 64)
	if err == nil {
		g.Pdop = float32(f)
	}

	f, err = strconv.ParseFloat(sArray[16], 64)
	if err == nil {
		g.Hdop = float32(f)
	}

	f, err = strconv.ParseFloat(sArray[17], 64)
	if err == nil {
		g.Vdop = float32(f)
	}
	n, err = strconv.Atoi(sArray[18])
	if err == nil {
		g.SystemId = n
	}

	return &g, nil
}

func (gps *Gps) parseNmea(line string) {
	// log.Printf("Line: %s", line)
	sArray, err := gps.verifyNmeaChecksum(line)
	if err != nil {
		return
	}
	// log.Printf("Token: %s", sArray[0])
	switch sArray[0][3:] {
	case "GGA":
		gga, err := gps.parseGGA(sArray)
		if err == nil {
			gps.Messages = append(gps.Messages, gga)
			gps.SendNotify(gga)
		}

	case "RMC":
		rmc, err := gps.parseRMC(sArray)
		if err == nil {
			gps.Messages = append(gps.Messages, rmc)
			gps.SendNotify(rmc)
		}

	case "GLL":
		gll, err := gps.parseGLL(sArray)
		if err == nil {
			gps.Messages = append(gps.Messages, gll)
			gps.SendNotify(gll)
		}

	case "GSA":
		gsa, err := gps.parseGSA(sArray)
		if err == nil {
			gps.Messages = append(gps.Messages, gsa)
			gps.SendNotify(gsa)
		}

	default:
		//	log.Printf("Unknownn token: %s", sArray[0])
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

func (gps *Gps) Sart(notify chan<- any) {
	gps.notify = notify
	reader := bufio.NewReader(gps.reader)
	/*
		data := []byte{0xB5, 0x62, 0x06, 0x8B, 0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x93, 0x20, 0x4E, 0xE3}
		ckA, ckB := gps.ubxChecksum(data)
		packet := append(data, ckA, ckB)
		gps.writer.Write(packet)
	*/
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
