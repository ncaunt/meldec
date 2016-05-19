package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
)

func decode(body []byte) (*UnitStatus, error) {
	// pick the LSV node from the XML document
	xmldat := Result{}
	if err := xml.Unmarshal(body, &xmldat); err != nil {
		return nil, err
	}

	// decode the chardata that was present in the LSV node
	decoded, _ := base64.StdEncoding.DecodeString(xmldat.LSV)

	// unmarshal the XML from the decoded LSV node
	report := Report{}
	if err := xml.Unmarshal(decoded, &report); err != nil {
		return nil, err
	}

	stat := UnitStatus{}

	// iterate each code in the HTTP request
	for _, cRaw := range report.Codes {
		var c Code
		cDec := make([]byte, 22)
		hex.Decode(cDec, cRaw.Code)
		err := binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c)
		if err != nil {
			logf("error parsing %s", cRaw.Code)
			continue
		}

		logf("%s %s\n", cRaw.Date, cRaw.Code)
		if c.GroupCode == 11 {
			temp := int8((c.Val7 - 2896) / 2)
			logf("Room temp zone 1: %d, outside temp: %d\n", c.Val1, temp)
			stat.RoomTempZone1 = float32(c.Val1) / 100
			stat.OutsideTemp = temp
		} else if c.GroupCode == 12 {
			logf("Water feed: %d, return: %d, hot water tank: %d\n", c.Val1, c.Val3, c.Val5)
			stat.WaterFeedTemp = float32(c.Val1) / 100
			stat.WaterReturnTemp = float32(c.Val3) / 100
			stat.HotWaterTemp = float32(c.Val5) / 100
		} else if c.GroupCode == 13 {
			logf("THWB1: %d, THWB2: %d\n", c.Val1, c.Val3)
			stat.BoilerFlowTemp = float32(c.Val1) / 100
			stat.BoilerReturnTemp = float32(c.Val3) / 100
		}
	}

	return &stat, nil
}
