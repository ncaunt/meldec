package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode15(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		Group21a int8 `structs:"status/group21a"`
		Group21b int8 `structs:"status/group21b"`
		_        [4]byte
		Group21c int8 `structs:"status/group21c"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	return structs.Map(s), err
}
