/*
This file is based on code
Copyright (c) 2012 Google, Inc. All rights reserved.
Copyright (c) 2009-2011 Andreas Krennmair. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Andreas Krennmair, Google, nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var port = flag.Int("p", 0, "TCP port on which to listen")
var iface = flag.String("i", "", "Interface to get packets from")
var fname = flag.String("r", "", "Filename to read from, overrides -i")
var snaplen = flag.Int("s", 1600, "SnapLen for pcap packet capture")
var filter = flag.String("f", "tcp and dst port 80", "BPF filter for pcap")
var logAllPackets = flag.Bool("v", false, "Logs every packet in great detail")
var raw = flag.Bool("w", false, "Show all raw values")

func logf(format string, v ...interface{}) {
	if *raw {
		fmt.Printf(fmt.Sprintf(format, v...))
	}
}

// HTTP stream handler
func (h *httpStream) run() {
	buf := bufio.NewReader(&h.r)
	for {
		req, err := http.ReadRequest(buf)
		if err == io.EOF {
			// We must read until we see an EOF... very important!
			fmt.Printf("EOF when reading HTTP request\n")
			return
		} else if err != nil {
			log.Println("Error reading stream", h.net, h.transport, ":", err)
		} else {
			body, _ := ioutil.ReadAll(req.Body)
			req.Body.Close()

			stat, err := decode(body)
			if err != nil {
				log.Println("error with stream", err)
				continue
			}

			fmt.Printf("%+v\n", stat)
		}
	}
}

func main() {
	flag.Parse()
	if *port == 0 && *fname == "" && *iface == "" {
		log.Fatal("one of -p, -i or -f must be specified")
	}

	init_mqtt()

	if *port > 0 {
		httpd()
	} else {
		packets()
	}
}

func httpd() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("%s %s %s\n", req.Method, req.URL, req.Proto)
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
			proxyReq, err := http.NewRequest(req.Method, req.RequestURI, &b2)

			/*
				if req.ContentLength < 1024 {
					fmt.Printf("small request. body: %s\n", body2)
				}
			*/
			client := &http.Client{}
			resp, err := client.Do(proxyReq)
			if err != nil {
				log.Printf("proxy request failed: %s\n", err)
				return
			}
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("response body: %s\n", string(body))
			w.WriteHeader(resp.StatusCode)
			w.Write(body)
		}()
		stat, err := decode(b1.Bytes())
		if err != nil {
			log.Println("error with stream", err)
		}

		fmt.Printf("%+v\n", stat)
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func packets() {
	defer util.Run()()
	var handle *pcap.Handle
	var err error

	// Set up pcap packet capture
	if *fname != "" {
		log.Printf("Reading from pcap dump %q", *fname)
		handle, err = pcap.OpenOffline(*fname)
	} else {
		log.Printf("Starting capture on interface %q", *iface)
		handle, err = pcap.OpenLive(*iface, int32(*snaplen), true, pcap.BlockForever)
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := handle.SetBPFFilter(*filter); err != nil {
		log.Fatal(err)
	}

	// Set up assembly
	streamFactory := &httpStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	log.Println("reading in packets")
	// Read in packets, pass to assembler.
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	ticker := time.Tick(time.Minute)
	for {
		select {
		case packet := <-packets:
			// A nil packet indicates the end of a pcap file.
			if packet == nil {
				return
			}
			if *logAllPackets {
				log.Println(packet)
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				log.Println("Unusable packet")
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)

		case <-ticker:
			// Every minute, flush connections that haven't seen activity in the past 2 minutes.
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}
}
