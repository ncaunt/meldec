package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode14(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		_       [11]byte
		Group20 int8 `structs:"status/group20"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	return structs.Map(s), err
}
