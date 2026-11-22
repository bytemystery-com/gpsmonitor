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
	"fmt"
	"os"
	"path/filepath"
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
