package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ncaunt/meldec/internal/pkg/decoder"
	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
	"github.com/ncaunt/meldec/internal/pkg/doc"
	"github.com/ncaunt/meldec/internal/pkg/reporter"
)

var port = flag.Int("p", 8080, "TCP port on which to listen")
var verbose = flag.Bool("v", false, "verbose; show decoded values")

func main() {
	flag.Parse()
	d := decoder.NewDecoder()
	r, err := reporter.NewMQTTReporter()
	if err != nil {
		fmt.Errorf("failed to initialise MQTT client")
		return
	}

	go func(d decoder.Decoder) {
		c := d.Stats()
		for s := range c {
			if *verbose {
				fmt.Printf("stat: %s\n", s)
			}
			r.Publish(s)
		}
	}(d)

	// XXX: for testing
	/*	go func() {
			ticker := time.Tick(time.Second)
			for range ticker {
				c, _ := codes.NewCode([]byte("fc62027a10011005120b24280800000000000000008b"))
				err := d.Decode(c)
				_ = err
			}
		}()
	*/
	httpd(d)
}

func httpd(d decoder.Decoder) {
	log.Printf("started httpd\n")
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s %s\n", req.Method, req.URL, req.Proto)
		for name, vals := range req.Header {
			for _, val := range vals {
				fmt.Printf("%s: %s\n", name, val)
			}
		}

		body := req.Body
		var b1, b2 bytes.Buffer
		wr := io.MultiWriter(&b1, &b2)
		io.Copy(wr, body)

		func() {
			log.Printf("req.RequestURI: %s\n", req.RequestURI)
			proxyReq, err := http.NewRequest(req.Method, req.RequestURI, &b2)
			if err != nil {
				panic(err)
				log.Printf("proxy request creation failed: %s\n", err)
			}

			/*
				if req.ContentLength < 1024 {
					fmt.Printf("small request. body: %s\n", body2)
				}
			*/

			client := &http.Client{}
			resp, err := client.Do(proxyReq)
			if err != nil {
				panic(err)
				log.Printf("proxy request failed: %s\n", err)
				return
			}
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("response body: %s\n", string(body))
			w.WriteHeader(resp.StatusCode)
			w.Write(body)
			resp.Body.Close()
		}()

		doc, err := doc.NewDoc(b1.Bytes())
		if err != nil {
			log.Print(err)
			return
		}
		log.Printf("1\n")
		for _, c := range doc.Codes {
			c2, err := codes.NewCode(c)
			if err != nil {
				log.Print(err)
				continue
			}
			log.Printf("c2: %+v\n", c2)
			err = d.Decode(c2)
			if err != nil {
				log.Print(err)
			}
			log.Printf("3\n")
		}
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
