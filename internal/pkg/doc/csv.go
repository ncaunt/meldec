package doc

import "encoding/xml"

type csvWrapper struct {
	XMLName xml.Name `xml:"CSV"`
	CSV     string   `xml:",chardata"`
}
