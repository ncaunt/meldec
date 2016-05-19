package main

type Result struct {
	LSV string `xml:",chardata"`
}

type Report struct {
	Codes []struct {
		Date string `xml:"DATDATE"`
		Code []byte `xml:"VALUE"`
	} `xml:"UPLOAD>CODE>DATA"`
}

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

type UnitStatus struct {
	RoomTempZone1    float32
	OutsideTemp      int8
	WaterFeedTemp    float32
	WaterReturnTemp  float32
	HotWaterTemp     float32
	BoilerFlowTemp   float32
	BoilerReturnTemp float32
}
