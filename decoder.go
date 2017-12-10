package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"strconv"
)

func decode(body []byte) (stat UnitStatus, err error) {
	// pick the LSV node from the XML document
	xmldat := Result{}
	if err = xml.Unmarshal(body, &xmldat); err != nil {
		return
	}

	// decode the chardata that was present in the LSV node
	decoded, _ := base64.StdEncoding.DecodeString(xmldat.LSV)

	// unmarshal the XML from the decoded LSV node
	report := Report{}
	if err = xml.Unmarshal(decoded, &report); err != nil {
		return
	}

	// iterate each code in the HTTP request
	for _, cRaw := range report.Codes {
		var (
			c  Code
			c2 CodeAlt
			c3 CodeAlt2
			c4 CodeAlt3
			c5 CodeAlt4
		)
		cDec := make([]byte, 22)
		hex.Decode(cDec, cRaw.Code)
		checksum(cDec)
		err := binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c)
		binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c2)
		binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c3)
		binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c4)
		binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c5)
		if err != nil {
			logf("error parsing %s", cRaw.Code)
			continue
		}

		logf("%s %s\n", cRaw.Date, cRaw.Code)
		logf("Code: %+v\n", c)
		logf("CodeAlt: %+v\n", c2)
		logf("CodeAlt2: %+v\n", c3)
		logf("CodeAlt3: %+v\n", c4)

		switch c.GroupCode {
		case 1:
			var c struct {
				Preamble  [5]byte
				GroupCode byte
				Year      uint8
				Month     uint8
				Day       uint8
				Hour      uint8
				Minute    uint8
				Seconds   uint8
				NotSure   uint8
				_         [8]byte
				Checksum  byte
			}
			binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c)
			logf("%04d-%02d-%02d %02d:%02d:%02d\n", c.Year, c.Month, c.Day, c.Hour, c.Minute, c.Seconds)
		case 3:
			stat.Status.Group3 = c4.Val5
			publish("status/group3", strconv.Itoa(int(stat.Status.Group3)))
		case 4:
			stat.Status.Group4 = c4.Val1
			publish("status/group4", strconv.Itoa(int(stat.Status.Group4)))
			logf("Operation mode? %d\n", c4.Val1)
		case 7:
			stat.Status.Group7a = c3.Val2
			stat.Status.Group7b = c3.Val3
			publish("status/group7a", strconv.Itoa(int(stat.Status.Group7a)))
			publish("status/group7b", strconv.Itoa(int(stat.Status.Group7b)))
		case 9:
			logf("Set room temp zone 1: %d, zone 2: %d, unknown1: %d, unknown2: %d, legionella temp?: %d\n", c2.Val1, c2.Val2, c2.Val3, c2.Val4, c2.Val5)
			stat.SetRoomTempZone1 = float32(c2.Val1) / 100
			stat.SetRoomTempZone2 = float32(c2.Val2) / 100
		case 11:
			var c CodeAlt
			err := binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c)
			if err != nil {
				logf("error parsing %s", cRaw.Code)
				continue
			}
			temp := int8(c.Val6/2) - 40
			stat.RoomTempZone1 = float32(c.Val1) / 100
			stat.RoomTempZone2 = float32(c.Val3) / 100
			stat.OutsideTemp = temp
			logf("Room temp zone 1: %d, zone 2: %d, outside: %d\n", c.Val1, c.Val2, stat.OutsideTemp)

			publish("temperatures/indoor/zone1", strconv.Itoa(int(stat.RoomTempZone1)))
			publish("temperatures/indoor/zone2", strconv.Itoa(int(stat.RoomTempZone2)))
			publish("temperatures/outdoor/front", strconv.Itoa(int(stat.OutsideTemp)))
		case 12:
			logf("Water feed: %d, return: %d, hot water tank: %d\n", c.Val1, c.Val3, c.Val5)
			stat.WaterFeedTemp = float32(c.Val1) / 100
			stat.WaterReturnTemp = float32(c.Val3) / 100
			stat.HotWaterTemp = float32(c.Val5) / 100
			publish("temperatures/indoor/heating_feed", strconv.FormatFloat(float64(stat.WaterFeedTemp), 'e', -1, 32))
			publish("temperatures/indoor/heating_return", strconv.FormatFloat(float64(stat.WaterReturnTemp), 'e', -1, 32))
			publish("temperatures/indoor/hot_water", strconv.FormatFloat(float64(stat.HotWaterTemp), 'e', -1, 32))
		case 13:
			logf("THWB1: %d, THWB2: %d\n", c.Val1, c.Val3)
			stat.BoilerFlowTemp = float32(c.Val1) / 100
			stat.BoilerReturnTemp = float32(c.Val3) / 100
		case 14:
			//logf("Outside temp: %d, something: %d\n", c.Val2, c.Val3)
			//stat.OutsideTemp = int8(c.Val2)
		case 19:
			stat.Status.Group19 = c4.Val1
			publish("status/group19", strconv.Itoa(int(stat.Status.Group19)))
		case 20:
			stat.Status.Group20 = c.Val8
			publish("status/group20", strconv.Itoa(int(stat.Status.Group20)))
		case 21:
			stat.Status.Group21a = c5.Val1
			stat.Status.Group21b = c5.Val2
			stat.Status.Group21c = c5.Val7
			publish("status/group21a", strconv.Itoa(int(stat.Status.Group21a)))
			publish("status/group21b", strconv.Itoa(int(stat.Status.Group21b)))
			publish("status/group21c", strconv.Itoa(int(stat.Status.Group21c)))
		case 38:
			stat.SetHotWaterTemp = float32(c4.Val5)
			// c4.Val6 == set room temp zone 1 (duplicate of groupcode 9)
		case 40:
			// c4.Val6 a flag (?)
		case 41:
			// c4.Val3 == set room temp zone 1 (duplicate of groupcode 9)
		}
	}

	return
}
