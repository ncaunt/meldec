package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode0c(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		WaterFeedTemp   int16 `structs:"temperatures/system/heating_feed"`
		_               byte
		WaterReturnTemp int16 `structs:"temperatures/system/heating_return"`
		_               byte
		HotWaterTemp    int16 `structs:"temperatures/system/hot_water"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	m = structs.Map(s)
	for _, k := range []string{"temperatures/system/heating_feed", "temperatures/system/heating_return", "temperatures/system/hot_water"} {
		t, _ := m[k].(int16)
		m[k] = float64(t) / 100.0
	}
	return
}
