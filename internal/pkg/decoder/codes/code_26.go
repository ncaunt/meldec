package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode26(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		_               [7]byte
		SetHotWaterTemp int16 `structs:"temperatures/system/set_hot_water"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	if err != nil {
		return
	}
	m = structs.Map(s)
	for _, k := range []string{"temperatures/system/set_hot_water"} {
		t, _ := m[k].(int16)
		m[k] = float64(t) / 100.0
	}
	return
}
