package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode0b(b []byte) (m map[string]interface{}, err error) {
	//fc 62 02 7a 10 0b 08 98 f0 c4 f0 c4 0b 09 c4 0b 6e 00 00 00 00 ae
	//fc 62 02 7a 10 0b 08 ca f0 c4 f0 c4 0b 09 c4 0b 8a 00 00 00 00 60
	// -99, really 27
	// 8c == -98 == 30
	var s struct {
		RoomTempZone1 int16 `structs:"temperatures/indoor/zone1"`
		Code2         int16 `structs:"status/group11/code2"`
		RoomTempZone2 int16 `structs:"temperatures/indoor/zone2"`
		Code4         int16 `structs:"status/group11/code4"`
		Code5         int16 `structs:"status/group11/code5"`
		OutdoorTemp   uint8 `structs:"temperatures/outdoor/front"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	if err != nil {
		return
	}
	m = structs.Map(s)
	for _, k := range []string{"temperatures/indoor/zone1", "temperatures/indoor/zone2"} {
		t, _ := m[k].(int16)
		m[k] = float64(t) / 100.0
	}
	// outdoor temperature has a different representation
	t := int8(m["temperatures/outdoor/front"].(uint8)/2 - 40)
	s2 := struct {
		RoomTempZone1 float64 `structs:"temperatures/indoor/zone1"`
		Code2         float64 `structs:"status/group11/code2"`
		RoomTempZone2 float64 `structs:"temperatures/indoor/zone2"`
		Code4         int16   `structs:"status/group11/code4"`
		Code5         int16   `structs:"status/group11/code5"`
		OutdoorTemp   int8    `structs:"temperatures/outdoor/front"`
	}{
		float64(s.RoomTempZone1) / 100.0,
		float64(s.Code2) / 100.0,
		float64(s.RoomTempZone2) / 100.0,
		m["status/group11/code4"].(int16),
		m["status/group11/code5"].(int16),
		t,
	}
	return structs.Map(s2), err
}
