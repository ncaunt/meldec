package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode04(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		Status byte `structs:"status/group4"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	return structs.Map(s), err
}
