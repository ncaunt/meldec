package main

import (
	"testing"
	"time"

	"github.com/ncaunt/meldec/internal/pkg/decoder"
	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
	"github.com/ncaunt/meldec/internal/pkg/uploader"
)

type FakeTTY struct {
	c chan []byte
}

func (f *FakeTTY) Init() {
	//f.c <- []byte{0xBA}
}
func (f *FakeTTY) Send(d []byte) (c []byte, err error) {
	f.c <- d
	return
}

type NullReporter struct{}

func (r *NullReporter) Publish(s decoder.Stat) {}

type NullUploader struct{}

func (u *NullUploader) Init() {}
func (u *NullUploader) AddCode(c codes.Code) (err error) {
	return
}
func (u *NullUploader) Send(h uploader.UploadResponseHandler) (c []codes.Code, err error) {
	return
}

// tests that the main loop sends (at least) 33 requests for data
func TestSendsRequests(t *testing.T) {
	expected := 33

	ticker := make(chan time.Time)
	codes := make(chan []byte, 0)
	f := &FakeTTY{c: codes}
	r := &NullReporter{}
	u := &NullUploader{}
	go loop(ticker, f, r, u)

	ticker <- time.Now()

	// allow 10 seconds to receive data
	wait := time.Tick(10 * time.Second)

	var i = 0
X:
	for {
		select {
		case c := <-codes:
			t.Logf("% x\n", c)
			i++
			if i == expected {
				break X
			}
		case <-wait:
			t.Fatalf("expected %d codes but received %d\n", expected, i)
		}
	}
}
