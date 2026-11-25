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
//
// Write data in csv file on SD card

package sd

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"gpsmonitord/cfg"
	"gpsmonitord/helper"
)

const (
	// DateTime(System);DateTime(GPS);Latitude;Longitude;Altitude;
	// Speed;Course;MagVar;Mode;Quality;
	// Sats(used);Sats(total);Lat(sigma);Lon(Sigma);Alt(Sigma);VDOP;HDOP;PDOD;Modeel;Errors
	DATA_HEADERLINE = "# GPS Monitor - %s\r\n# Version: %s\r\n" +
		"# Mode: [1=NoFix, 2=2D, 3=3D]\r\n" +
		"# Quality: [0=NoFix, 1=NormalFix, 2=DiffFix, 3=PPSFix, 4=RTKFix, 5=RTKFloat, 6=Estimated, 7=Manual, 8=Simul]\r\n" +
		"DateTime (System);DateTime (GPS);Latitude [°];Longitude [°];Altitude [m];" +
		"Speed [m/s];Course [°];MagVar [°];Mode;Quality;" +
		"Sats (used);Sats (total);Latitude (σ) [m];Longitude (σ) [m];Altitude (σ) [m];VDOP;HDOP;PDOD;Model;Errors\r\n"
)

type SdData struct {
	config *cfg.Cfg
	folder string
}

type SdDataEntry struct {
	Longitude    *float64
	Latitude     *float64
	Altitude     *float32
	Speed        *float32
	Course       *float32
	MagVar       *float32
	Mode         *int
	Quality      *int
	NumUsedSats  *int
	NumTotalSats *int
	StdLongitude *float32
	StdLatitude  *float32
	StdAltitude  *float32
	HDop         *float32
	VDop         *float32
	PDop         *float32
	DateTime     *time.Time
	Model        string
	Errors       *uint32
}

func NewSdData(config *cfg.Cfg) (*SdData, error) {
	folder := config.GetStringValue(cfg.SECTION_DATA, cfg.KEY_DATA_FOLDER, cfg.DEF_DATA_FOLDER)
	err := os.MkdirAll(folder, 0o755)
	if err != nil {
		return nil, err
	}
	sd := SdData{
		config: config,
		folder: folder,
	}
	return &sd, nil
}

func (sd *SdData) buildFileName(t time.Time) string {
	return filepath.Join(sd.folder, fmt.Sprintf("%04d", t.Year()), fmt.Sprintf("%02d", t.Month()),
		fmt.Sprintf("%04d-%02d-%02d.csv", t.Year(), t.Month(), t.Day()))
}

func (sd *SdData) writeHeader(file string) error {
	err := os.MkdirAll(filepath.Dir(file), 0o755)
	if err != nil {
		return err
	}
	return os.WriteFile(file, fmt.Appendf(nil, DATA_HEADERLINE,
		sd.config.GetStringValue(cfg.SECTION_GENERAL, cfg.KEY_GENERAL_NAME, cfg.DEF_NAME),
		cfg.GPS_MONITOR_VERSION), 0o644)
}

func (sd *SdData) WriteEntry(data *SdDataEntry) error {
	now := time.Now()
	return sd.writeEntryInternal(sd.buildFileName(now), data, &now)
}

func getFloatAsString[T ~float32 | ~float64](val *T, p int) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%.*f", p, *val)
}

func getIntAsString[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~byte | ~uint16 | ~uint32 | ~uint64](val *T) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%d", *val)
}

func getDateTimeAsString(val *time.Time) string {
	if val == nil {
		return ""
	}
	val2 := val.Local()
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		val2.Year(), val2.Month(), val2.Day(), val2.Hour(), val2.Minute(), val2.Second())
}

func (sd *SdData) writeEntryInternal(file string, data *SdDataEntry, now *time.Time) error {
	ok, err := helper.ExistFile(file)
	if err != nil {
		return err
	}
	if !ok {
		err = sd.writeHeader(file)
		if err != nil {
			return err
		}
	}
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// DateTime(System);DateTime(GPS);Latitude;Longitude;Altitude;
	// Speed;Course;MagVar;Mode;Quality;
	// Sats(used);Sats(total);Lat(sigma);Lon(Sigma);Alt(Sigma);VDOP;HDOP;PDOD;Model;Errors

	line := fmt.Sprintf("%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s\r\n",
		getDateTimeAsString(now), getDateTimeAsString(data.DateTime), getFloatAsString(data.Latitude, 7), getFloatAsString(data.Longitude, 7), getFloatAsString(data.Altitude, 2),
		getFloatAsString(data.Speed, 3), getFloatAsString(data.Course, 2), getFloatAsString(data.MagVar, 2), getIntAsString(data.Mode), getIntAsString(data.Quality),
		getIntAsString(data.NumUsedSats), getIntAsString(data.NumTotalSats), getFloatAsString(data.StdLatitude, 3), getFloatAsString(data.StdLongitude, 3), getFloatAsString(data.StdAltitude, 2),
		getFloatAsString(data.VDop, 2), getFloatAsString(data.HDop, 2), getFloatAsString(data.PDop, 2), data.Model, getIntAsString(data.Errors),
	)
	_, err = f.WriteString(line)
	return err
}

type GPX struct {
	XMLName        xml.Name `xml:"gpx"`
	Version        string   `xml:"version,attr"`
	Creator        string   `xml:"creator,attr"`
	Xmlns          string   `xml:"xmlns,attr"`
	XmlnsXsi       string   `xml:"xmlns:xsi,attr"`
	XmlnsGpxtpx    string   `xml:"xmlns:gpxtpx,attr"`
	SchemaLocation string   `xml:"xsi:schemaLocation,attr"`
	Metadata       Metadata `xml:"metadata"`
	Trk            []Trk    `xml:"trk"`
}

type Metadata struct {
	Name    string `xml:"name"`
	Time    string `xml:"time"`
	Comment string `xml:",comment"`
}

type Trk struct {
	Name   string   `xml:"name"`
	Trkseg []Trkseg `xml:"trkseg"`
}

type Trkseg struct {
	Trkpt []Trkpt `xml:"trkpt"`
}

type Trkpt struct {
	Latitude   float64     `xml:"lat,attr"`
	Longitude  float64     `xml:"lon,attr"`
	Altitude   *float32    `xml:"ele,omitempty"`
	Time       string      `xml:"time,omitempty"`
	TimeAsTime time.Time   `xml:"-"`
	Fix        string      `xml:"fix,omitempty"`
	SatsUsed   *int        `xml:"sat,omitempty"`
	Hdop       *float32    `xml:"hdop,omitempty"`
	Vdop       *float32    `xml:"vdop,omitempty"`
	Pdop       *float32    `xml:"pdop,omitempty"`
	Extensions *Extensions `xml:"extensions,omitempty"`
}

type Extensions struct {
	TrackPointExtension TrackPointExtension `xml:"gpxtpx:TrackPointExtension"`
}

type TrackPointExtension struct {
	Speed  *float32 `xml:"gpxtpx:speed,omitempty"`
	Course *float32 `xml:"gpxtpx:course,omitempty"`
}

func (sd *SdData) GetFilePath(t time.Time) string {
	path := filepath.Join(sd.folder, fmt.Sprintf("%04d", t.Year()))
	path = filepath.Join(path, fmt.Sprintf("%02d", t.Month()))
	path = filepath.Join(path, fmt.Sprintf("%04d-%02d-%02d.csv", t.Year(), t.Month(), t.Day()))
	return path
}

func parseFloat[T ~float32 | ~float64](s string) *T {
	var f T
	n, err := fmt.Sscanf(s, "%f", &f)
	if err == nil && n == 1 {
		return &f
	}
	return nil
}

func parseInt(s string) *int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}

func (sd *SdData) ReadCsv(path string, start time.Time, end time.Time, entries *[]*Trkpt) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 2048), 1024*1024)
	var version uint
	_ = version
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line[0] == '#' {
			var verMaj uint
			var verMin uint
			n, err := fmt.Sscanf(line, "# Version: %d.%d", &verMaj, &verMin)
			if err == nil && n == 2 {
				version = ((verMaj & 0xFF) << 8) | (verMin & 0xFF)
			}
			continue
		}
		if line[0] < '0' || line[0] > '9' {
			continue
		}
		var entry Trkpt
		parts := strings.Split(line, ";")
		dateTime, err := time.ParseInLocation("2006-01-02 15:04:05", parts[0], time.Local)
		if err != nil {
			continue
		}
		if !dateTime.Before(start) && !dateTime.After(end) {
			entry.TimeAsTime = dateTime.UTC()
			entry.Time = entry.TimeAsTime.Format(time.RFC3339)
			n, err := fmt.Sscanf(parts[2], "%f", &entry.Latitude)
			if err != nil || n != 1 {
				continue
			}
			n, err = fmt.Sscanf(parts[3], "%f", &entry.Longitude)
			if err != nil || n != 1 {
				continue
			}
			entry.Altitude = parseFloat[float32](parts[4])
			speed := parseFloat[float32](parts[5])
			course := parseFloat[float32](parts[6])
			if speed != nil || course != nil {
				entry.Extensions = &Extensions{
					TrackPointExtension: TrackPointExtension{
						Speed:  speed,
						Course: course,
					},
				}
			}
			fix, err := strconv.Atoi(parts[8])
			if err == nil {
				switch fix {
				case 1:
					entry.Fix = "none"
				case 2:
					entry.Fix = "2d"
				case 3:
					entry.Fix = "3d"
				}
			}
			entry.SatsUsed = parseInt(parts[10])

			entry.Vdop = parseFloat[float32](parts[15])
			entry.Hdop = parseFloat[float32](parts[16])
			entry.Pdop = parseFloat[float32](parts[17])

			*entries = append(*entries, &entry)
		}
	}

	err = scanner.Err()
	if err != nil {
		return err
	}
	return nil
}

func (sd *SdData) GetData(startTime, endTime time.Time) ([]byte, error) {
	start := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	end := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
	var path string
	day := start
	entries := make([]*Trkpt, 0, 1000)
	for day.Before(end) || day.Equal(end) {
		path = sd.GetFilePath(day)
		sd.ReadCsv(path, startTime, endTime, &entries)
		day = day.AddDate(0, 0, 1)
	}

	track := make([]Trk, 0, 1)
	trackName := "GPS Track"
	if len(entries) > 0 {
		trackName = "Track " + entries[0].TimeAsTime.Local().Format("2006-01-02 15:04")
	}
	track = append(track, Trk{
		Name: trackName,
		Trkseg: []Trkseg{
			{
				Trkpt: make([]Trkpt, 0, len(entries)),
			},
		},
	})

	index := 0
	var lastTime time.Time
	for i, item := range entries {
		if !lastTime.IsZero() && item.TimeAsTime.Sub(lastTime) > time.Duration(5*time.Minute) {
			track[0].Trkseg = append(track[0].Trkseg, Trkseg{
				Trkpt: make([]Trkpt, 0, len(entries)-i),
			})
			index = index + 1
		}
		track[0].Trkseg[index].Trkpt = append(track[0].Trkseg[index].Trkpt, *item)
		lastTime = item.TimeAsTime
	}
	gpxData := GPX{
		Version:        "1.1",
		Creator:        "GpsMonitor " + cfg.GPS_MONITOR_VERSION,
		Xmlns:          "http://www.topografix.com/GPX/1/1",
		XmlnsXsi:       "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsGpxtpx:    "http://www.garmin.com/xmlschemas/TrackPointExtension/v2",
		SchemaLocation: "http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd",
		Metadata: Metadata{
			Name:    "GPS Monitor export",
			Time:    time.Now().UTC().Format(time.RFC3339),
			Comment: "https://github.com/bytemystery-com/gpsmonitor",
		},
		Trk: track,
	}

	output, err := xml.MarshalIndent(gpxData, "", "  ")
	if err != nil {
		return nil, err
	}

	finalXML := xml.Header + string(output)
	return []byte(finalXML), nil
}

type listEntry struct {
	Name string `json:"name"`
}

type listData struct {
	Path  string      `json:"path"`
	Dirs  []listEntry `json:"dirs"`
	Files []listEntry `json:"files"`
}

// returns true if we are in basefolder
func (sd *SdData) checkSdPath(dir string) (bool, error) {
	// Verhindern, dass wir sd.folder verlassen
	rel, err := filepath.Rel(sd.folder, dir)
	if err != nil {
		return false, err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false, fmt.Errorf("invalid path")
	}
	return rel == ".", nil
}

func (sd *SdData) GetSd(file string) (*os.File, os.FileInfo, error) {
	file = filepath.Join(sd.folder, filepath.Clean(file))
	_, err := sd.checkSdPath(file)
	if err != nil {
		return nil, nil, err
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	return f, info, nil
}

func (sd *SdData) ListSd(dir string) ([]byte, error) {
	var ld listData
	ld.Path = dir
	ld.Dirs = make([]listEntry, 0, 15)
	ld.Files = make([]listEntry, 0, 40)
	dir = filepath.Join(sd.folder, filepath.Clean(dir))
	inRoot, err := sd.checkSdPath(dir)
	if err != nil {
		return nil, err
	}
	if !inRoot {
		ld.Dirs = append(ld.Dirs, listEntry{Name: ".."})
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			ld.Dirs = append(ld.Dirs, listEntry{Name: entry.Name()})
		} else {
			ld.Files = append(ld.Files, listEntry{Name: entry.Name()})
		}
	}
	slices.SortFunc(ld.Dirs, func(a, b listEntry) int {
		A := a.Name
		B := b.Name
		an := strings.ToLower(A)
		bn := strings.ToLower(B)
		if an == bn {
			if A < B {
				return -1
			}
			if A > B {
				return 1
			}
			return 0
		}
		if an < bn {
			return -1
		}
		return 1
	})
	slices.SortFunc(ld.Files, func(a, b listEntry) int {
		A := a.Name
		B := b.Name
		an := strings.ToLower(A)
		bn := strings.ToLower(B)
		if an == bn {
			if A < B {
				return -1
			}
			if A > B {
				return 1
			}
			return 0
		}
		if an < bn {
			return -1
		}
		return 1
	})
	data, err := json.Marshal(ld)
	if err != nil {
		return nil, err
	}
	return data, nil
}

const earthRadiusMeters = 6371000.0 // Mittlerer Erdradius in Metern

type Point struct {
	Lat float64
	Lon float64
	Alt float64 // Höhe in Metern
}

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180.0
}

// Distance2D berechnet die horizontale Distanz auf der Erdoberfläche in Metern (Haversine).
func Distance2D(p1, p2 Point) float64 {
	lat1 := toRadians(p1.Lat)
	lat2 := toRadians(p2.Lat)
	dLat := toRadians(p2.Lat - p1.Lat)
	dLon := toRadians(p2.Lon - p1.Lon)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusMeters * c
}

// Distance3D bezieht zusätzlich den Höhenunterschied mit ein (in Metern).
func Distance3D(p1, p2 Point) float64 {
	d2d := Distance2D(p1, p2)
	dAlt := p2.Alt - p1.Alt
	return math.Sqrt(d2d*d2d + dAlt*dAlt)
}
