package decoder

import (
	"fmt"
	"strconv"

	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
)

type Stat struct {
	Name, Value string
}

type Decoder interface {
	Decode(codes.Code) error
	Stats() chan Stat
}

type StatDecoder struct {
	stats chan Stat
}

func NewDecoder() (dec *StatDecoder) {
	dec = &StatDecoder{stats: make(chan Stat, 0)}
	return
}

func (dec *StatDecoder) publish(value Stat) {
	dec.stats <- value
}

func (dec *StatDecoder) Stats() (c chan Stat) {
	return dec.stats
}

type groupDecoder struct {
	decode func([]byte) (map[string]interface{}, error)
}

func (dec *StatDecoder) Decode(code codes.Code) (err error) {
	m, err := code.Decode()
	if err != nil {
		return
	}
	for k, v := range m {
		var val string
		switch v.(type) {
		case byte:
			val = strconv.Itoa(int(v.(byte)))
		case int:
			val = strconv.Itoa(v.(int))
		case int8:
			val = strconv.Itoa(int(v.(int8)))
		case int16:
			val = strconv.FormatInt(int64(v.(int16)), 10)
		case float64:
			val = strconv.FormatFloat(v.(float64), 'f', 1, 64)
		default:
			return fmt.Errorf("unknown type for key %s: %T", k, v)
		}
		s := Stat{k, val}
		dec.publish(s)
	}

	/*
		err = binary.Read(bytes.NewBuffer(base.Data[:]), binary.BigEndian, &d)
		if err != nil {
			return
		}

		fmt.Printf("%#v\n", d)
		t := reflect.TypeOf(d)
		v := reflect.ValueOf(d)
		for i := 0; i < t.NumField(); i++ {
			fk := t.Field(i)
			if fk.Anonymous || fk.Name == "_" {
				continue
			}
			fv := v.Field(i)
			tag := fk.Tag.Get("stat")

			switch fv.Kind() {
			case reflect.Uint:
				dec.publish(Stat{tag, float32(fv.Uint())})
			case reflect.Int:
				dec.publish(Stat{tag, float32(fv.Int())})
			default:
				dec.publish(Stat{fv.Kind().String(), 0})
			}
		}
	*/
	return
}

/*
func (dec *Decoder) Decode2(body []byte) (stat UnitStatus, err error) {
	// pick the LSV node from the XML document
	xmldat := Result{}
	if err = xml.Unmarshal(body, &xmldat); err != nil {
		return
	}

	// decode the chardata that was present in the LSV node
	decoded, _ := base64.StdEncoding.DecodeString(xmldat.LSV)

	// unmarshal the XML from the decoded LSV node
	report := Report{}
	if err = xml.Unmarshal(decoded, &report); err != nil {
		return
	}

	// iterate each code in the HTTP request
	for _, cRaw := range report.Codes {
		var (
			c  Code
			c2 CodeAlt
			c3 CodeAlt2
			c4 CodeAlt3
			c5 CodeAlt4
		)
		cDec := make([]byte, 22)
		hex.Decode(cDec, cRaw.Code)
		checksum(cDec)
		err := binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c)
		binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c2)
		binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c3)
		binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c4)
		binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c5)
		if err != nil {
			dec.logf("error parsing %s", cRaw.Code)
			continue
		}

		dec.logf("%s %s\n", cRaw.Date, cRaw.Code)
		dec.logf("Code: %+v\n", c)
		dec.logf("CodeAlt: %+v\n", c2)
		dec.logf("CodeAlt2: %+v\n", c3)
		dec.logf("CodeAlt3: %+v\n", c4)

		switch c.GroupCode {
		case 1:
			var c struct {
				Preamble  [5]byte
				GroupCode byte
				Year      uint8
				Month     uint8
				Day       uint8
				Hour      uint8
				Minute    uint8
				Seconds   uint8
				NotSure   uint8
				_         [8]byte
				Checksum  byte
			}
			binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c)
			dec.logf("%04d-%02d-%02d %02d:%02d:%02d\n", c.Year, c.Month, c.Day, c.Hour, c.Minute, c.Seconds)
		case 3:
			stat.Status.Group3 = c4.Val5
			dec.Publish("status/group3", strconv.Itoa(int(stat.Status.Group3)))
		case 4:
			stat.Status.Group4 = c4.Val1
			dec.Publish("status/group4", strconv.Itoa(int(stat.Status.Group4)))
			dec.logf("Operation mode? %d\n", c4.Val1)
		case 7:
			stat.Status.Group7a = c3.Val2
			stat.Status.Group7b = c3.Val3
			dec.Publish("status/group7a", strconv.Itoa(int(stat.Status.Group7a)))
			dec.Publish("status/group7b", strconv.Itoa(int(stat.Status.Group7b)))
		case 9:
			dec.logf("Set room temp zone 1: %d, zone 2: %d, unknown1: %d, unknown2: %d, legionella temp?: %d\n", c2.Val1, c2.Val2, c2.Val3, c2.Val4, c2.Val5)
			stat.SetRoomTempZone1 = float32(c2.Val1) / 100
			stat.SetRoomTempZone2 = float32(c2.Val2) / 100
		case 11: //0x0b
			var c CodeAlt
			err := binary.Read(bytes.NewBuffer(cDec), binary.BigEndian, &c)
			if err != nil {
				dec.logf("error parsing %s", cRaw.Code)
				continue
			}
			temp := int8(c.Val6/2) - 40
			stat.RoomTempZone1 = float32(c.Val1) / 100
			stat.RoomTempZone2 = float32(c.Val3) / 100
			stat.OutsideTemp = temp
			dec.logf("Room temp zone 1: %d, zone 2: %d, outside: %d\n", c.Val1, c.Val2, stat.OutsideTemp)

			dec.Publish("temperatures/indoor/zone1", strconv.Itoa(int(stat.RoomTempZone1)))
			dec.Publish("temperatures/indoor/zone2", strconv.Itoa(int(stat.RoomTempZone2)))
			dec.Publish("temperatures/outdoor/front", strconv.Itoa(int(stat.OutsideTemp)))
		case 12: //0x0c
			dec.logf("Water feed: %d, return: %d, hot water tank: %d\n", c.Val1, c.Val3, c.Val5)
			stat.WaterFeedTemp = float32(c.Val1) / 100
			stat.WaterReturnTemp = float32(c.Val3) / 100
			stat.HotWaterTemp = float32(c.Val5) / 100
			dec.Publish("temperatures/indoor/heating_feed", strconv.FormatFloat(float64(stat.WaterFeedTemp), 'e', -1, 32))
			dec.Publish("temperatures/indoor/heating_return", strconv.FormatFloat(float64(stat.WaterReturnTemp), 'e', -1, 32))
			dec.Publish("temperatures/indoor/hot_water", strconv.FormatFloat(float64(stat.HotWaterTemp), 'e', -1, 32))
		case 13: //0x0d
			dec.logf("THWB1: %d, THWB2: %d\n", c.Val1, c.Val3)
			stat.BoilerFlowTemp = float32(c.Val1) / 100
			stat.BoilerReturnTemp = float32(c.Val3) / 100
		case 14: //0x0e
			//dec.logf("Outside temp: %d, something: %d\n", c.Val2, c.Val3)
			//stat.OutsideTemp = int8(c.Val2)
		case 19: //0x13
			stat.Status.Group19 = c4.Val1
			dec.Publish("status/group19", strconv.Itoa(int(stat.Status.Group19)))
		case 20: //0x14
			stat.Status.Group20 = c.Val8
			dec.Publish("status/group20", strconv.Itoa(int(stat.Status.Group20)))
		case 21: //0x15
			stat.Status.Group21a = c5.Val1
			stat.Status.Group21b = c5.Val2
			stat.Status.Group21c = c5.Val7
			dec.Publish("status/group21a", strconv.Itoa(int(stat.Status.Group21a)))
			dec.Publish("status/group21b", strconv.Itoa(int(stat.Status.Group21b)))
			dec.Publish("status/group21c", strconv.Itoa(int(stat.Status.Group21c)))
		case 38: //0x26
			stat.SetHotWaterTemp = float32(c4.Val5)
			// c4.Val6 == set room temp zone 1 (duplicate of groupcode 9)
		case 40:
			// c4.Val6 a flag (?)
		case 41:
			// c4.Val3 == set room temp zone 1 (duplicate of groupcode 9)
		}
	}

	return
}
*/
