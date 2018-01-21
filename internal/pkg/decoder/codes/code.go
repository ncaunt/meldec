package codes

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type Code struct {
	Preamble  [5]byte
	GroupCode byte
	Data      [15]byte
	Checksum  byte
}

type decoderfn func([]byte) (map[string]interface{}, error)

var groupCodes = map[byte]decoderfn{
	0x01: groupCode01,
	0x03: groupCode03,
	0x04: groupCode04,
	0x07: groupCode07,
	0x09: groupCode09,
	0x0b: groupCode0b,
	0x0c: groupCode0c,
	0x0d: groupCode0d,
	0x13: groupCode13,
	0x14: groupCode14,
	0x15: groupCode15,
	0x26: groupCode26,
}

var codeLen = 22 // length in bytes, when hex decoded

func NewCode(c []byte) (nc Code, err error) {
	decoded := make([]byte, hex.DecodedLen(len(c)))
	n, err := hex.Decode(decoded, c)
	if err != nil {
		return
	}

	if n != codeLen {
		err = fmt.Errorf("expected %d bytes but got %d", codeLen, len(c))
		return
	}

	if !verify_checksum(decoded, decoded[codeLen-1]) {
		err = fmt.Errorf("checksum invalid")
		return
	}

	buf := bytes.NewBuffer(decoded)
	err = binary.Read(buf, binary.BigEndian, &nc)
	return
}

func (c Code) Decode() (m map[string]interface{}, err error) {
	dec, ok := groupCodes[c.GroupCode]
	if !ok {
		//err = fmt.Errorf("no decoder for group 0x%x", c.GroupCode)
		//return

		// XXX: use default decoder
		dec = groupCodeGeneric(c.GroupCode)
	}

	m, err = dec(c.Data[:])
	return
}
