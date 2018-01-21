package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode07(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		_       [2]byte
		Group7a int16 `structs:"status/group7a"`
		Group7b int16 `structs:"status/group7b"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	return structs.Map(s), err
}
