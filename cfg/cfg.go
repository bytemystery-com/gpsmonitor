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

package cfg

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"gpsmonitord/helper"

	"gopkg.in/ini.v1"
)

const (
	GPS_MONITOR_VERSION = "1.1"

	SECTION_GENERAL             = "general"
	KEY_GENERAL_NAME            = "Name"
	KEY_GENERAL_USE_PASSWORD    = "UsePassword"
	KEY_GENERAL_PASSWORD        = "Password"
	KEY_GENERAL_UPDATE_INTERVAL = "UpdateInterval"

	SECTION_WEBSERVER               = "webserver"
	KEY_WEBSERVER_HTTP_PORT         = "HTTPPort"
	KEY_WEBSERVER_HTTP_PORT_NOROOT  = "HTTPPortNonRoot"
	KEY_WEBSERVER_HTTPS_PORT        = "HTTPSPort"
	KEY_WEBSERVER_HTTPS_PORT_NOROOT = "HTTPSPortNonRoot"
	KEY_WEBSERVER_HTTP_ENABLED      = "HttpEnabled"
	KEY_WEBSERVER_HTTPS_ENABLED     = "HttpsEnabled"
	KEY_WEBSERVER_KEY_FILE          = "SSLKeyFile"
	KEY_WEBSERVER_CERT_FILE         = "SSLCertFile"

	SECTION_SYSTEM               = "system"
	KEY_SYSTEM_RESTART_POFF_TIME = "PoweroffTime"

	SECTION_HARDWARE            = "hardware"
	KEY_HARDWARE_UART_DEV       = "UartDev"
	KEY_HARDWARE_BAUD_RATE      = "BaudRate"
	KEY_HARDWARE_DATA_BITS      = "DataBits"
	KEY_HARDWARE_DATA_PARITY    = "Partity"
	KEY_HARDWARE_DATA_STOP_BITS = "StopBits"
	KEY_HARDWARE_IS_UBLOX       = "IsUblox"

	SECTION_RUN    = "run"
	KEY_RUN_REBOOT = "Reboots"

	DEF_CFG_FILE                    = "/etc/gpsmonitord/gpsmonitord.cfg"
	DEF_NAME                        = "GPS Monitor"
	DEF_USE_PASSWORD                = false
	DEF_UPDATE_INTERVAL             = 1000 // msec
	DEF_WEBSERVER_HTTP_PORT         = 80
	DEF_WEBSERVER_HTTP_PORT_NOROOT  = 8080
	DEF_WEBSERVER_HTTPS_PORT        = 443
	DEF_WEBSERVER_HTTPS_PORT_NOROOT = 4430
	DEF_WEBSERVER_ENABLED_HTTP      = true
	DEF_WEBSERVER_ENABLED_HTTPS     = true
	DEF_WEBSERVER_KEY_FILE          = "/etc/gpsmonitord/certs/gpsmonitord.key"
	DEF_WEBSERVER_CERT_FILE         = "/etc/gpsmonitord/certs/gpsmonitord.crt"
	DEF_UART_DEV                    = "/dev/ttyAMA0"
	DEF_UART_BAUD_RATE              = 38400
	DEF_UART_DATA_BITS              = 8
	DEF_UART_PARITY                 = 0
	DEF_UART_STOP_BITS              = 0
	DEF_IS_UBLOX                    = true
	DEF_RESTART_POFF_TIME           = 60 // sec
)

type Cfg struct {
	ini      *ini.File
	doReload func()
	cfgFile  string
	lock     sync.RWMutex
}

func NewCfg(file string, doReload func()) (*Cfg, error) {
	err := os.MkdirAll(filepath.Dir(file), 0o755)
	if err != nil {
		log.Printf("Mkdir for '%s' failed with %v", filepath.Dir(file), err)
		return nil, err
	}

	ok, err := helper.ExistFile(file)
	if err != nil {
		log.Printf("Exist file for '%s' failed with %v", file, err)
		return nil, err
	}
	c := Cfg{
		cfgFile: file,
		ini:     ini.Empty(),
	}
	c.SetDeaultCfg()
	if !ok {
		log.Printf("CFG: New created in %s", file)
		err := c.Save()
		if err != nil {
			log.Printf("Save cfg failed %v", err)
		}
	} else {
		log.Printf("CFG: Loaded from %s", file)
		err := c.ini.Append(file)
		if err != nil {
			log.Printf("Append cfg failed %v", err)
			return nil, err
		}
		err = c.Save()
		if err != nil {
			log.Printf("Save2 cfg failed %v", err)
		}
	}
	return &c, nil
}

func (c *Cfg) Save() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	err := c.ini.SaveTo(c.cfgFile)
	if err != nil {
		return err
	}
	if c.doReload != nil {
		c.doReload()
	}
	return nil
}

// Getter
func (c *Cfg) GetIntValue(section, key string, def int) int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.ini.Section(section).Key(key).MustInt(def)
}

func (c *Cfg) GetInt64Value(section, key string, def int64) int64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.ini.Section(section).Key(key).MustInt64(def)
}

func (c *Cfg) GetStringValue(section, key string, def string) string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.ini.Section(section).Key(key).MustString(def)
}

func (c *Cfg) GetFloatValue(section, key string, def float32) float32 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return float32(c.ini.Section(section).Key(key).MustFloat64(float64(def)))
}

func (c *Cfg) GetBoolValue(section, key string, def bool) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.ini.Section(section).Key(key).MustBool(def)
}

// Setter
func (c *Cfg) SetStringValue(section, key, value, comment string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ini.Section(section).Key(key).SetValue(value)
	if comment != "" {
		c.ini.Section(section).Key(key).Comment = comment
	}
}

func (c *Cfg) SetIntValue(section, key string, value int, comment string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ini.Section(section).Key(key).SetValue(strconv.Itoa(value))
	if comment != "" {
		c.ini.Section(section).Key(key).Comment = comment
	}
}

func (c *Cfg) SetInt64Value(section, key string, value int64, comment string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ini.Section(section).Key(key).SetValue(strconv.FormatInt(value, 10))
	if comment != "" {
		c.ini.Section(section).Key(key).Comment = comment
	}
}

func (c *Cfg) SetFloatValue(section, key string, value float32, comment string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ini.Section(section).Key(key).SetValue(strconv.FormatFloat(float64(value), 'f', -1, 32))
	if comment != "" {
		c.ini.Section(section).Key(key).Comment = comment
	}
}

func (c *Cfg) SetBoolValue(section, key string, value bool, comment string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ini.Section(section).Key(key).SetValue(strconv.FormatBool(value))
	if comment != "" {
		c.ini.Section(section).Key(key).Comment = comment
	}
}

func (c *Cfg) SetDeaultCfg() {
	c.ini = ini.Empty()
	c.SetStringValue(SECTION_GENERAL, KEY_GENERAL_NAME, DEF_NAME, "Name of station (title in browser)")
	c.SetBoolValue(SECTION_GENERAL, KEY_GENERAL_USE_PASSWORD, DEF_USE_PASSWORD, "Use password")
	c.SetStringValue(SECTION_GENERAL, KEY_GENERAL_PASSWORD, "", "Password hash")
	c.SetIntValue(SECTION_GENERAL, KEY_GENERAL_UPDATE_INTERVAL, DEF_UPDATE_INTERVAL, "Main page update interval [msec]")

	c.SetIntValue(SECTION_WEBSERVER, KEY_WEBSERVER_HTTP_PORT, DEF_WEBSERVER_HTTP_PORT, "Port for webserver if running as root")
	c.SetIntValue(SECTION_WEBSERVER, KEY_WEBSERVER_HTTP_PORT_NOROOT, DEF_WEBSERVER_HTTP_PORT_NOROOT, "Port for webserver if not running as root")
	c.SetIntValue(SECTION_WEBSERVER, KEY_WEBSERVER_HTTPS_PORT, DEF_WEBSERVER_HTTPS_PORT, "Port for webserver for using SSL if running as root")
	c.SetIntValue(SECTION_WEBSERVER, KEY_WEBSERVER_HTTPS_PORT_NOROOT, DEF_WEBSERVER_HTTPS_PORT_NOROOT, "Port for webserver for using SSL and if not running as root")
	c.SetBoolValue(SECTION_WEBSERVER, KEY_WEBSERVER_HTTP_ENABLED, DEF_WEBSERVER_ENABLED_HTTP, "Enable HTTP")
	c.SetBoolValue(SECTION_WEBSERVER, KEY_WEBSERVER_HTTPS_ENABLED, DEF_WEBSERVER_ENABLED_HTTPS, "Enable HTTPS")
	c.SetStringValue(SECTION_WEBSERVER, KEY_WEBSERVER_KEY_FILE, DEF_WEBSERVER_KEY_FILE, "The SSL key file")
	c.SetStringValue(SECTION_WEBSERVER, KEY_WEBSERVER_CERT_FILE, DEF_WEBSERVER_CERT_FILE, "The SSL cert file")

	c.SetIntValue(SECTION_SYSTEM, KEY_SYSTEM_RESTART_POFF_TIME, DEF_RESTART_POFF_TIME, "Restart after restart command [sec]")

	c.SetStringValue(SECTION_HARDWARE, KEY_HARDWARE_UART_DEV, DEF_UART_DEV, "UART device")
	c.SetIntValue(SECTION_HARDWARE, KEY_HARDWARE_BAUD_RATE, DEF_UART_BAUD_RATE, "Baudrate (9600, 38400, 115200)")
	c.SetIntValue(SECTION_HARDWARE, KEY_HARDWARE_DATA_BITS, DEF_UART_DATA_BITS, "Data bits")
	c.SetIntValue(SECTION_HARDWARE, KEY_HARDWARE_DATA_PARITY, DEF_UART_PARITY, "Parity: 0=none, 1=Odd, 2=Even, 3=Mark, 4=Space")
	c.SetIntValue(SECTION_HARDWARE, KEY_HARDWARE_DATA_STOP_BITS, DEF_UART_STOP_BITS, "SopBits: 0=1, 1=1.5, 2=2")

	c.SetIntValue(SECTION_RUN, KEY_RUN_REBOOT, 0, "Number of reboots")
}

func (c *Cfg) GetJson() []byte {
	type general struct {
		Name        string `json:"name"`
		UsePassword bool   `json:"use_pass"`
	}
	g := general{
		Name:        c.GetStringValue(SECTION_GENERAL, KEY_GENERAL_NAME, DEF_NAME),
		UsePassword: c.GetBoolValue(SECTION_GENERAL, KEY_GENERAL_USE_PASSWORD, DEF_USE_PASSWORD),
	}
	data := map[string]any{
		"general": g,
	}

	b, _ := json.Marshal(data)
	return b
}

func (c *Cfg) SetCfg(r *http.Request) error {
	var allErrors error = nil

	val := r.FormValue("name")
	if val != "" {
		c.SetStringValue(SECTION_GENERAL, KEY_GENERAL_NAME, val, "")
	}
	val = r.FormValue("use_pass")
	if val != "" {
		h, err := strconv.Atoi(val)
		if err == nil {
			c.SetBoolValue(SECTION_GENERAL, KEY_GENERAL_USE_PASSWORD, h == 1, "")
		} else {
			allErrors = errors.Join(allErrors, err)
		}
	}
	val = r.FormValue("pass")
	if val != "" {
		if !helper.CheckPassword(val, c.GetStringValue(SECTION_GENERAL, KEY_GENERAL_PASSWORD, "")) {
			val, err := helper.HashPassword(val)
			if err == nil {
				c.SetStringValue(SECTION_GENERAL, KEY_GENERAL_PASSWORD, val, "")
			} else {
				allErrors = errors.Join(allErrors, err)
			}
		}
	}
	err := c.Save()
	if err != nil {
		allErrors = errors.Join(allErrors, err)
	}
	return allErrors
}
