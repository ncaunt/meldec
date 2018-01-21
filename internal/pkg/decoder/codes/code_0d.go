package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode0d(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		BoilerFlowTemp   int16 `structs:"temperatures/system/boiler_flow"`
		_                byte
		BoilerReturnTemp int16 `structs:"temperatures/system/boiler_return"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	m = structs.Map(s)
	for _, k := range []string{"temperatures/system/boiler_flow", "temperatures/system/boiler_return"} {
		t, _ := m[k].(int16)
		m[k] = float64(t) / 100.0
	}
	return
}
