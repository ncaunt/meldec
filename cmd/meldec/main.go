package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/ncaunt/meldec/internal/pkg/decoder"
	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
	"github.com/ncaunt/meldec/internal/pkg/doc"
	"github.com/ncaunt/meldec/internal/pkg/reporter"
	"github.com/ncaunt/meldec/internal/pkg/uploader"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/tarm/serial"
)

func main() {
	var err error
	var serialDevice = flag.String("d", "/dev/ttyS0", "serial device (default /dev/ttyS0)")
	var baud = flag.Int("b", 2400, "baud rate (default 2400)")
	var parity = flag.String("p", "E", "parity (default E: even)")
	var timeout = flag.String("t", "200ms", "communication timeout")
	// XXX: implement
	_ = parity

	flag.Parse()
	to, err := time.ParseDuration(*timeout)
	if err != nil {
		log.Fatal("invalid timeout specified")
	}

	c := &serial.Config{Name: *serialDevice, Baud: *baud, Parity: serial.ParityEven, ReadTimeout: to}
	t, err := NewTTY(c)
	if err != nil {
		log.Fatalf("error with serial configuration: %s", err)
	}

	r, err := reporter.NewMQTTReporter()

	metrics := prometheus.NewRegistry()
	metrics.MustRegister(
		version.NewCollector("meldec"),
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)
	var g run.Group
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(metrics, promhttp.HandlerOpts{}))
	httpBindAddr := "0.0.0.0:9100"
	l, err := net.Listen("tcp", httpBindAddr)
	if err != nil {
		log.Fatal(errors.Wrap(err, "listen metrics address"))
	}

	g.Add(func() error {
		log.Fatal(errors.Wrap(http.Serve(l, mux), "serve metrics"))
		return nil
	}, func(error) {
	})

	ticker := time.NewTicker(60 * time.Second)
	g.Add(func() error {
		u := uploader.NewHTTPUploader("http://leswifidata.meuk.mee.com/upload")
		return loop(ticker.C, t, r, u)
	}, func(error) {
		ticker.Stop()
	})

	g.Run()
}

func loop(ticker <-chan time.Time, t SerialComm, r reporter.Reporter, u uploader.Uploader) (err error) {
	if err = t.Init(); err != nil {
		return
	}

	d := decoder.NewDecoder()
	gcs := []byte{
		0x01,
		0x02,
		0x03,
		0x04,
		0x05,
		0x06,
		0x07,
		0x09,
		0x0b,
		0x0c,
		0x0d,
		0x0e,
		0x10,
		0x11,
		0x13,
		0x14,
		0x15,
		0x16,
		0x17,
		0x18,
		0x19,
		0x1a,
		0x1c,
		0x1d,
		0x1e,
		0x1f,
		0x20,
		0x26,
		0x27,
		0x28,
		0x29,
		0xa1,
		0xa2,
	}

	for range ticker {
		u.Init()

		// base packet
		for _, i := range gcs {
			pkt := []byte{0xfc, 0x42, 0x02, 0x7a, 0x10, i, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

			// calc checksum
			pkt[21] = sum(pkt[1:21])

			// send packet and await response
			c, err := t.Send(pkt)
			if err != nil {
				log.Print(err)
			}

			// HP might not return data for a groupcode so ignore it
			if len(c) == 0 {
				continue
			}
			log.Printf("received\t[% x]\n", c)
			c2, err := codes.NewCode(c)
			if err != nil {
				log.Print(err)
			}
			stats, err := d.Decode(c2)
			if err != nil {
				log.Print(err)
			}

			err = u.AddCode(c2)
			for _, s := range stats {
				log.Printf("stat: %s=%s\n", s.Name, s.Value)
				r.Publish(s)
			}

		}

		r, err := u.Send(handler)
		if err != nil {
			log.Printf("error sending HTTP request: %s\n", err)
		}
		//fmt.Printf("codes in response: %+v\n", r)
		for _, change := range r {
			// send packet and await response
			codeBytes, err := change.ToBytes()
			if err != nil {
				log.Fatal(err)
			}
			c, err := t.Send(codeBytes)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("received\t[% x]\n", c)
		}

	}
	return
}

func handler(r io.Reader) (c []codes.Code, err error) {
	body, _ := ioutil.ReadAll(r)
	fmt.Printf("response: %s\n", string(body))
	doc, err := doc.NewCSVDoc(body)
	for _, rawCode := range doc.Codes {
		code, err_ := codes.NewCodeFromHex(rawCode)
		if err_ != nil {
			err = err_
			return
		}
		c = append(c, code)
	}
	return
}

func sum(data []byte) (s byte) {
	total := byte(0)
	for _, i := range data {
		total += i
	}
	return -total
}

/*
func keepalive(c []byte) (err error) {
	ticker := time.Tick(15 * time.Second)
	for {
		select {
		case <-ticker:
			r, err := send(c)
			if err != nil {
				break
			}
			log.Printf("keepalive resp: [% x]\n", r)
		}
	}
	return
}
*/
