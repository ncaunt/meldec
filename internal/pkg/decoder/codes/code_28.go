package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode28(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		Code1, Code2, Code3, Code4, Code5, Code6, Code7, Code8, Code9, Code10, Code11, Code12, Code13, Code14, Code15 int8
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	return structs.Map(s), err
}
