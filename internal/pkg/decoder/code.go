package decoder

/*
type Code struct {
	Preamble  [5]byte
	GroupCode byte
	Val1      uint16
	Val2      byte
	Val3      uint16
	Val4      byte
	Val5      uint16
	Val6      byte

	Val7     uint16
	Val8     byte
	Val9     uint16
	Val10    byte
	Checksum byte
}

type Code struct {
	Preamble  [5]byte
	GroupCode byte
	Data      [15]byte
	Checksum  byte
}

func NewCode(c []byte) (nc *Code, err error) {
	decoded := make([]byte, hex.DecodedLen(len(c)))
	_, err = hex.Decode(decoded, c)
	if err != nil {
		return
	}

	if !verify_checksum(decoded, decoded[21]) {
		err = errors.New("checksum invalid")
		return
	}

	buf := bytes.NewBuffer(decoded)
	nc = &Code{}
	err = binary.Read(buf, binary.BigEndian, nc)
	if err != nil {
		return
	}

	return
}
*/

type CodeAlt struct {
	Val1 int16
	Val2 int16
	Val3 int16
	Val4 int16
	Val5 int16
	Val6 uint8
	_    [4]byte
}

type CodeAlt2 struct {
	Preamble  [5]byte
	GroupCode byte
	Val1      int16
	Val2      int16
	Val3      int16
	Val4      int16
	Val5      int16
	Val6      int16
	Val7      int16
	Val8      byte
	Checksum  byte
}

type CodeAlt3 struct {
	Preamble  [5]byte
	GroupCode byte
	Val1      byte
	Val2      int16
	Val3      int16
	Val4      int16
	Val5      int16
	Val6      int16
	Val7      int16
	Val8      int16
	Checksum  byte
}

type CodeAlt4 struct {
	Preamble  [5]byte
	GroupCode byte
	Val1, Val2, Val3, Val4, Val5, Val6, Val7, Val8, Val9, Val10,
	Val11, Val12, Val13, Val14, Val15, Checksum byte
}

type UnitStatus struct {
	RoomTempZone1    float32
	RoomTempZone2    float32
	SetRoomTempZone1 float32
	SetRoomTempZone2 float32
	OutsideTemp      int8
	WaterFeedTemp    float32
	WaterReturnTemp  float32
	HotWaterTemp     float32
	SetHotWaterTemp  float32
	BoilerFlowTemp   float32
	BoilerReturnTemp float32
	Status           struct {
		Group3                       int16
		Group4                       byte
		Group7a, Group7b             int16
		Group19, Group20             byte
		Group21a, Group21b, Group21c byte
	}
}
