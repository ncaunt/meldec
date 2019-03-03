package codes

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type Code struct {
	Start     byte
	MsgType   byte
	Header    [2]byte
	DataLen   byte
	GroupCode byte
	Data      [15]byte
	Checksum  byte
}

type decoderfn func([]byte) (map[string]interface{}, error)

var groupCodes = map[byte]decoderfn{
	0x01: groupCode01,
	0x02: groupCode02,
	0x03: groupCodeGenericBytes(3),
	0x04: groupCodeGenericBytes(4),
	0x05: groupCodeGenericBytes(5),
	0x06: groupCodeGenericBytes(6),
	0x07: groupCodeGenericBytes(7),
	0x09: groupCode09,
	0x0b: groupCode0b,
	0x0c: groupCode0c,
	0x0d: groupCode0d,
	0x13: groupCode13,
	0x14: groupCode14,
	0x15: groupCode15,
	0x26: groupCode26,
	0x28: groupCodeGenericBytes(0x28),
}

var codeLen = 22 // length in bytes, when hex decoded

func NewCode(c []byte) (nc Code, err error) {
	cl := len(c)
	if cl != codeLen {
		err = fmt.Errorf("code length incorrect: %d", cl)
		return
	}
	if !verify_checksum(c) {
		err = fmt.Errorf("checksum invalid")
		return
	}

	buf := bytes.NewBuffer(c)
	err = binary.Read(buf, binary.BigEndian, &nc)
	return
}

func NewCodeFromHex(c []byte) (nc Code, err error) {
	decoded := make([]byte, hex.DecodedLen(len(c)))
	n, err := hex.Decode(decoded, c)
	if err != nil {
		return
	}

	if n != codeLen {
		err = fmt.Errorf("expected %d bytes but got %d", codeLen, len(c))
		return
	}

	nc, err = NewCode(decoded)
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

func (c Code) ToHex() (b []byte, err error) {
	bytes_, err := c.ToBytes()
	if err != nil {
		return
	}
	b = make([]byte, hex.EncodedLen(len(bytes_)))
	hex.Encode(b, bytes_)
	return
}

func (c Code) ToBytes() (b []byte, err error) {
	var buf bytes.Buffer
	// XXX: endianness for non x86 builds might need to change
	err = binary.Write(&buf, binary.LittleEndian, c)
	b = buf.Bytes()
	return
}
