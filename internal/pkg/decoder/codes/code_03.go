package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode03(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		Code1  int8 `structs:"status/group3/Code1"`
		Code2  int8 `structs:"status/group3/Code2"`
		_      [5]byte
		Status int16 `structs:"status/group3"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	return structs.Map(s), err
}
