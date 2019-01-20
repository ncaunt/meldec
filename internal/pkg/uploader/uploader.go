package uploader

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
	"github.com/ncaunt/meldec/internal/pkg/doc"
)

type UploadResponseHandler func(io.Reader) ([]codes.Code, error)

type Uploader interface {
	Init()
	AddCode(codes.Code) error
	Send(UploadResponseHandler) ([]codes.Code, error)
}

type HTTPUploader struct {
	url   string
	codes []codes.Code
}

func (u *HTTPUploader) Init() {
	u.codes = u.codes[:0]
}

func (u *HTTPUploader) AddCode(c codes.Code) (err error) {
	u.codes = append(u.codes, c)
	return
}

func (u *HTTPUploader) Send(h UploadResponseHandler) (c []codes.Code, err error) {
	b, err := doc.FromCodes(u.codes)
	if err != nil {
		return
	}

	/*
		Connection: keep-alive
		Host: leswifidata.meuk.mee.com
		User-Agent: MAC557 IF
		Content-Type: text/plain;chrset=UTF-8
		Pragma: no-cache
		Content-Length: 9322
	*/
	req, err := http.NewRequest("POST", u.url, bytes.NewReader(b))
	if err != nil {
		return
	}
	req.Header.Add("Connection", "close")
	req.Header.Add("User-Agent", "MAC557 IF")
	req.Header.Add("Content-Type", "text/plain;chrset=UTF-8")
	req.Header.Add("Pragma", "no-cache")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("proxy request failed: %s\n", err)
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("upload response was %d", resp.StatusCode)
		return
	}

	c, err = h(resp.Body)
	return
}

func NewHTTPUploader(url string) (u *HTTPUploader) {
	u = &HTTPUploader{url: url}
	return
}
