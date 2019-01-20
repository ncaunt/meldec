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
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"

	"github.com/ncaunt/meldec/internal/pkg/decoder"
	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
	"github.com/ncaunt/meldec/internal/pkg/doc"
	"github.com/ncaunt/meldec/internal/pkg/reporter"
)

var iface = flag.String("i", "", "Interface to get packets from")
var fname = flag.String("r", "", "Filename to read from, overrides -i")
var snaplen = flag.Int("s", 1600, "SnapLen for pcap packet capture")
var filter = flag.String("f", "tcp and dst port 80", "BPF filter for pcap")
var verbose = flag.Bool("v", false, "verbose; show decoded values")

var d *decoder.StatDecoder

// HTTP stream handler
func (h *httpStream) run() {
	defer tcpreader.DiscardBytesToEOF(&h.r)
	buf := bufio.NewReader(&h.r)
	for {
		fmt.Println("begin http.ReadRequest()")
		req, err := http.ReadRequest(buf)
		fmt.Println("end http.ReadRequest()")
		fmt.Println(err)
		if err == io.EOF {
			// We must read until we see an EOF... very important!
			fmt.Printf("EOF when reading HTTP request\n")
			return
		} else if err != nil {
			log.Println("Error reading stream", h.net, h.transport, ":", err)
			tcpreader.DiscardBytesToEOF(&h.r)
			return
		} else {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Println("ioutil.ReadAll", err)
				return
			}
			req.Body.Close()
			doc, err := doc.NewDoc(body)
			if err != nil {
				return
			}

			for _, c := range doc.Codes {
				c2, err := codes.NewCode(c)
				if err != nil {
					continue
				}
				d.Decode(c2)
			}
		}
	}
}

func main() {
	flag.Parse()
	if *fname == "" && *iface == "" {
		log.Fatal("one of -i or -r must be specified")
	}

	d = decoder.NewDecoder()
	r, err := reporter.NewMQTTReporter()
	if err != nil {
		log.Fatal("failed to initialise MQTT client")
	}

	go func(d *decoder.StatDecoder) {
		c := d.Stats()
		for s := range c {
			if *verbose {
				fmt.Printf("stat: %s\n", s)
			}
			r.Publish(s)
		}
	}(d)

	packets(d)
}

func packets(d *decoder.StatDecoder) {
	defer func() {
		fmt.Println("packets() returned")
		util.Run()
		fmt.Println("util.Run() finished")
	}()
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

	err = handle.SetBPFFilter(*filter)
	if err != nil {
		log.Fatal(err)
	}

	// Set up assembly
	streamFactory := &httpStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	// Read in packets, pass to assembler.
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	ticker := time.Tick(5 * time.Second)
	for {
		select {
		case packet, ok := <-packets:
			if *verbose {
				log.Println(packet)
			}
			// A nil packet indicates the end of a pcap file.
			if !ok || packet == nil {
				fmt.Println("nil packet")
				return
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
			pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
		}
	}
}
