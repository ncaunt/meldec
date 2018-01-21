package doc

import (
	"encoding/base64"
	"encoding/xml"
)

// outer XML document
type xmlWrapper struct {
	LSV string `xml:",chardata"`
}

// inner XML document
type xmlCodes struct {
	Codes []struct {
		//		Date string `xml:"DATDATE"`
		Code []byte `xml:"VALUE"`
	} `xml:"UPLOAD>CODE>DATA"`
}

type Doc struct {
	Codes [][]byte
}

func NewDoc(b []byte) (d Doc, err error) {
	codes, err := unpack(b)
	if err != nil {
		return
	}

	d = Doc{
		Codes: codes,
	}
	return
}

func unpack(body []byte) (cl [][]byte, err error) {
	w := xmlWrapper{}
	err = xml.Unmarshal(body, &w)
	if err != nil {
		return nil, err
	}

	l, err := base64.StdEncoding.DecodeString(w.LSV)
	if err != nil {
		return nil, err
	}

	m := xmlCodes{}
	err = xml.Unmarshal(l, &m)
	if err != nil {
		return nil, err
	}

	for _, c := range m.Codes {
		cl = append(cl, c.Code)
	}
	return
}
