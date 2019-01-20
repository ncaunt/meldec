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

/*type dec struct {
	stats chan decoder.Stat
}

func (d *dec) Decode(c codes.Code) (n int, err error) {
	d.stats <- c
	return
}

func (d *dec) Stats() (c chan decoder.Stat) {
	return
}*/

type NullDecoder struct {
	c chan interface{}
}

func (n *NullDecoder) Decode(c codes.Code) (s []decoder.Stat, err error) { n.c <- struct{}{}; return }
func (n *NullDecoder) Stats() chan interface{}                           { return n.c }

type NullReporter struct{}

func (r *NullReporter) Publish(s decoder.Stat) {}

func TestHttpd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Logf("httptest: %s %s %s\n", req.Method, req.URL, req.Proto)
	}))
	defer ts.Close()

	//d := &dec{stats: make(chan interface{})}
	c := make(chan interface{})
	d := &NullDecoder{c: c}
	r := &NullReporter{}
	go httpd(d, r)
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

	wait := time.Tick(10 * time.Second)

	var expected = 31
	var i = 0
X:
	// 31 codes are present in the test doc
	for {
		select {
		case y := <-c:
			t.Logf("got value: %+v\n", y)
			i++
			if i == expected {
				break X
			}
		case <-wait:
			t.Fatalf("expected %d codes but received %d\n", expected, i)
		}
	}
}
