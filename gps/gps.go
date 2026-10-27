package gps

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

// https://www.nmea.org/nmea-0183.html
// https://avionicsduino.com/wp-content/uploads/2023/04/NEO-M9N_Interface_description_UBX-19035940.pdf
// https://www.ardusimple.com/nmea-protocol-and-message-description/

// NEO-M9N
const (
	KNOT_IN_MSEC = (1852.0 / 3600.0)
	KMH_IN_MSEC  = (1.0 / 3.6)
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
	TalkerId      string
	Time          time.Time
	Latitude      float64
	Longitude     float64
	Quality       Quality
	NumSats       int
	HDop          float32
	Altitude      float32
	AltSep        float32
	DiffAge       float32
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

func (gps *Gps) parseFloat32(s string) float32 {
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return float32(f)
	} else {
		return float32(math.NaN())
	}
}

func (gps *Gps) parseFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return f
	} else {
		return math.NaN()
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
				g.SatDatas = append(g.SatDatas, SatData{Number: n})
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

	case "VTG":
		vtg, err := gps.parseVTG(sArray)
		if err == nil {
			gps.Messages = append(gps.Messages, vtg)
			gps.SendNotify(vtg)
		}

	case "GSV":
		gsv, err := gps.parseGSV(sArray)
		if err == nil {
			gps.Messages = append(gps.Messages, gsv)
			gps.SendNotify(gsv)
		}

	case "TXT":
		txt, err := gps.parseTXT(sArray)
		if err == nil {
			gps.Messages = append(gps.Messages, txt)
			gps.SendNotify(txt)
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
