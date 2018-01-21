package codes

import (
	"bytes"
	"encoding/binary"

	"github.com/fatih/structs"
)

func groupCode13(b []byte) (m map[string]interface{}, err error) {
	var s struct {
		Group19 int8 `structs:"status/group19"`
	}
	err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
	return structs.Map(s), err
}
