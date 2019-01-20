package uploader

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
)

func TestHTTPUploader(t *testing.T) {
	u := NewHTTPUploader("foobar")
	c, err := codes.NewCodeFromHex([]byte("fc62027a102900000007d007d000000000000000003b"))
	if err != nil {
		t.Error(err)
	}
	u.AddCode(c)
	u.Send(func(h io.Reader) (c []codes.Code, err error) { return })
}

type HandlingUploader struct{}

func (u *HandlingUploader) Init()                            {}
func (u *HandlingUploader) AddCode(c codes.Code) (err error) { return }
func (u *HandlingUploader) Send(h UploadResponseHandler) (err error) {
	h(bytes.NewReader([]byte{}))
	return
}

func TestHandlerCalled(t *testing.T) {
	x := make(chan int, 1)
	u := HandlingUploader{}
	u.Send(func(r io.Reader) (c []codes.Code, err error) {
		x <- 42
		return
	})
	wait := time.Tick(1 * time.Second)
Y:
	for {
		select {
		case f := <-x:
			if f == 42 {
				break Y
			}
			t.Fatalf("unexpected value received: %d\n", f)
		case <-wait:
			t.Fatalf("expected handler to be called\n")
		}
	}
}
