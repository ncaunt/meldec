package codes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

func groupCodeGeneric(g byte) (f decoderfn) {
	f = func(b []byte) (m map[string]interface{}, err error) {
		var s struct {
			Code1, Code2, Code3, Code4, Code5, Code6, Code7 int16
			Code8                                           int8
		}
		err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, &s)
		if err != nil {
			return
		}
		m = make(map[string]interface{})
		v := reflect.ValueOf(s)
		for i := 0; i < v.NumField(); i++ {
			m[fmt.Sprintf("status/group%d/%s", g, v.Type().Field(i).Name)] = v.Field(i).Interface()
		}
		return
	}
	return
}
