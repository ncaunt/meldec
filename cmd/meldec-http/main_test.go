package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ncaunt/meldec/internal/pkg/decoder"
	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
)

type dec struct {
	stats chan interface{}
}

func (d *dec) Decode(c codes.Code) (err error) {
	d.stats <- c
	return
}

func (d *dec) Stats() (c chan decoder.Stat) {
	return
}

func TestHttpd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Logf("httptest: %s %s %s\n", req.Method, req.URL, req.Proto)
	}))
	defer ts.Close()

	d := &dec{stats: make(chan interface{})}
	go httpd(d)
	time.Sleep(100 * time.Millisecond)

	f, err := os.Open("../../examples/encoded_request_xml")
	if err != nil {
		t.Fatal(err)
	}
	s, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}

	hdrs := fmt.Sprintf(`POST %s/ HTTP/1.1
Host: localhost
Connection: keep-alive
User-Agent: MAC557 IF
Content-Type: text/plain;chrset=UTF-8
Pragma: no-cache
Content-Length: %d
Cache-Control: no-cache

`, ts.URL, s.Size())
	conn.Write([]byte(hdrs))
	io.Copy(conn, f)

	// 31 codes are present in the test doc
	for i := 0; i <= 31; i++ {
		select {
		case y := <-d.stats:
			t.Logf("got value: %+v\n", y)
		}
	}
}
