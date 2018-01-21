package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode01(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		Year   int8 `structs:"timestamp/year"`
		Month  int8 `structs:"timestamp/month"`
		Day    int8 `structs:"timestamp/day"`
		Hour   int8 `structs:"timestamp/hour"`
		Minute int8 `structs:"timestamp/minute"`
		Second int8 `structs:"timestamp/second"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	return structs.Map(s), err
}
