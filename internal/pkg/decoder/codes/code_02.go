package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode03(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		_      [7]byte
		Status int16 `structs:"status/group3"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	return structs.Map(s), err
}
