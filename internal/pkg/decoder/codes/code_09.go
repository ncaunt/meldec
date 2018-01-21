package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode09(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		SetRoomTempZone1 int16 `structs:"temperatures/indoor/set_zone1"`
		SetRoomTempZone2 int16 `structs:"temperatures/indoor/set_zone2"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	if err != nil {
		return
	}
	m = structs.Map(s)
	for _, k := range []string{"temperatures/indoor/set_zone1", "temperatures/indoor/set_zone2"} {
		t, _ := m[k].(int16)
		m[k] = float64(t) / 100.0
	}
	return
}
