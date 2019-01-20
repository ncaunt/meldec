package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode02(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		Code1 int8 `structs:"status/group2/Code1"`
		Code2 int8 `structs:"status/group2/Code2"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	return structs.Map(s), err
}
