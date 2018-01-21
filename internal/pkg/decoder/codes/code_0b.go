package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode0b(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		RoomTempZone1 int16 `structs:"temperatures/indoor/zone1"`
		_             [2]byte
		RoomTempZone2 int16 `structs:"temperatures/indoor/zone2"`
		_             [4]byte
		OutdoorTemp   int8 `structs:"temperatures/outdoor/front"`
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
	m["temperatures/outdoor/front"] = m["temperatures/outdoor/front"].(int8)/2 - 40
	return
}
