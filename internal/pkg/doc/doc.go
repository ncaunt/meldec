package doc

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"log"
	"time"

	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
)

// outer XML document
type xmlWrapper struct {
	XMLName xml.Name `xml:"LSV"`
	LSV     string   `xml:",chardata"`
}

// inner XML document
type xmlCodes struct {
	XMLName xml.Name `xml:"COMMAND"`
	CmdType string   `xml:"CMDTYPE"`
	Upload  upload   `xml:"UPLOAD"`
}

type upload struct {
	Certification certification `xml:"CERTIFICATION"`
	CommInfo      comminfo      `xml:"COMMINFO"`
	UnitInfo      unitinfo      `xml:"UNITINFO"`
	ProfileCode   profilecode   `xml:"PROFILECODE"`
	Info          info          `xml:"INFO"`
	Codes         []code        `xml:"CODE"`
}

type certification struct {
	Mac    string `xml:"MAC"`
	Ip     string `xml:"IP"`
	Serial string `xml:"SERIAL"`
}

type comminfo struct {
	Reason string `xml:"REASON"`
	Retry  string `xml:"RETRY"`
}

type unitinfo struct {
	Type    string `xml:"TYPE"`
	Version string `xml:"VERSION"`
}

type profilecode struct {
	GroupCode string           `xml:"GROUPCODE"`
	Data      profilecode_data `xml:"DATA"`
}

type profilecode_data struct {
	Value string `xml:"VALUE"`
}

type info struct {
	Rssi         int          `xml:"RSSI"`
	ItVer        string       `xml:"IT_VER"`
	ItStatus     string       `xml:"IT_STATUS"`
	Version      info_version `xml:"VERSION"`
	Position     string       `xml:"POSITION"`
	PCycle       string       `xml:"PCYCLE"`
	RecordNumMax string       `xml:"RECORDNUMMAX"`
	Timeout      string       `xml:"TIMEOUT"`
}

type info_version struct {
	App1App2Ver string `xml:"APP1_APP2_VER"`
	WlanMacVer  string `xml:"WLAN_MAC_VER"`
	WebVer      string `xml:"WEB_VER"`
}

type code struct {
	GroupCode string    `xml:"GROUPCODE"`
	RecordNum string    `xml:"RECORDNUM"`
	Data      code_data `xml:"DATA"`
}

type code_data struct {
	DatDate string `xml:"DATDATE"`
	Value   []byte `xml:"VALUE"`
}

type Doc struct {
	Codes [][]byte
}

func NewDoc(b []byte) (d Doc, err error) {
	w := xmlWrapper{}
	err = xml.Unmarshal(b, &w)
	if err != nil {
		return
	}

	codes, err := unpack(w.LSV)
	if err != nil {
		return
	}

	d = Doc{
		Codes: codes,
	}
	return
}

func NewCSVDoc(b []byte) (d Doc, err error) {
	w := csvWrapper{}
	err = xml.Unmarshal(b, &w)
	if err != nil {
		return
	}

	codes, err := unpack(w.CSV)
	if err != nil {
		return
	}

	d = Doc{
		Codes: codes,
	}
	return
}

func FromCodes(codes []codes.Code) (b []byte, err error) {
	var _codes []code
	for _, c := range codes {
		t := time.Now()
		h, err := c.ToHex()
		if err != nil {
			return b, err
		}
		h = bytes.ToUpper(h)
		x := code{
			GroupCode: fmt.Sprintf("%02x", c.GroupCode),
			RecordNum: "01",
			Data: code_data{
				DatDate: t.Format("2006/01/02 15:04:05"),
				Value:   h,
			},
		}
		_codes = append(_codes, x)
	}

	y := xmlCodes{
		CmdType: "request",
		Upload: upload{
			Certification: certification{
				Mac:    "00:1d:c9:92:ba:44",
				Ip:     "192.168.1.115",
				Serial: "1508275096557e0000",
			},
			CommInfo: comminfo{
				"NORMAL",
				"0000",
			},
			UnitInfo: unitinfo{
				"ATW",
				"01.00",
			},
			ProfileCode: profilecode{
				"C9",
				profilecode_data{
					"fc7b027a10c903000100140200000000000000000016",
				},
			},
			Info: info{
				Rssi:     -88,
				ItVer:    "03.00",
				ItStatus: "NORMAL",
				Version: info_version{
					"04.00",
					"3.5.9",
					"02.00",
				},
				Position:     "unregistered",
				PCycle:       "01",
				RecordNumMax: "01",
				Timeout:      "01",
			},
			Codes: _codes,
		},
	}
	xm, err := xml.Marshal(y)
	if err != nil {
		return
	}
	log.Printf("%s\n", xm)

	lsv := base64.StdEncoding.EncodeToString(xm)

	x := xmlWrapper{LSV: lsv}
	xm, err = xml.Marshal(x)
	if err != nil {
		return
	}
	xm = []byte(fmt.Sprintf("%s%s", xml.Header, xm))
	log.Printf("xml:\n%s\n", xm)

	return xm, nil
}

func unpack(w string) (cl [][]byte, err error) {
	l, err := base64.StdEncoding.DecodeString(w)
	if err != nil {
		return nil, err
	}

	m := xmlCodes{}
	err = xml.Unmarshal(l, &m)
	if err != nil {
		return nil, err
	}

	for _, c := range m.Upload.Codes {
		if c.Data.Value != nil {
			cl = append(cl, c.Data.Value)
		}
	}
	return
}
