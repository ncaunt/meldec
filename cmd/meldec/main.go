package main

import (
	"flag"
	"io"
	"log"
	"sync"
	"time"

	"github.com/ncaunt/meldec/internal/pkg/decoder"
	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
	"github.com/tarm/serial"
)

var port *serial.Port
var m sync.Mutex

func main() {

	/*
		init serial port + hp comms
		loop
			request data
			decode
			mqtt
		end
	*/

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
	port, err = serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	/*
		// XXX: apparently the melcloud unit sends these codes but I have not found it necessary
		cmds := [][]byte{
			{0x02, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x02},
			{0x02, 0xff, 0xff, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00},
			{0xfc, 0x5a, 0x02, 0x7a, 0x02, 0xca, 0x01, 0x5d},
			{0xfc, 0x5b, 0x02, 0x7a, 0x01, 0xc9, 0x5f},
			{0xfc, 0x41, 0x02, 0x7a, 0x10, 0x34, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0xfd},
		}

		for _, x := range cmds {
			_, err = send(x)
			if err != nil {
				log.Fatal(err)
			}
		}

		var w sync.WaitGroup
		defer w.Wait()
		go keepalive([]byte{0xfc, 0x41, 0x02, 0x7a, 0x10, 0x34, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xfe})
		w.Add(1)
	*/

	d := decoder.NewDecoder()
	go func(d decoder.Decoder) {
		c := d.Stats()
		for s := range c {
			log.Printf("stat: %s\n", s)
		}
	}(d)

	pkt := []byte{0xfc, 0x42, 0x02, 0x7a, 0x10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	for i := byte(0x01); i <= 0x2a; i++ {
		pkt[5] = i
		pkt[21] = sum(pkt[1:21])
		c, err := send(pkt)
		if err != nil {
			log.Fatal(err)
		}

		// HP might not return data for a groupcode so ignore it
		if len(c) == 0 {
			continue
		}
		log.Printf("received [% x]\n", c)
		c2, err := codes.NewCode(c)
		if err != nil {
			log.Fatal(err)
		}
		err = d.Decode(c2)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func sum(data []byte) (s byte) {
	total := byte(0)
	for _, i := range data {
		total += i
	}
	return -total
}

func send(data []byte) (c []byte, err error) {
	log.Printf("sending [% x]\n", data)
	m.Lock()
	defer func() {
		m.Unlock()
	}()
	n, err := port.Write(data)
	if err != nil {
		return
	}
	_ = n

	c = make([]byte, 0, 22)
	buf := make([]byte, 0, 22)
	//reader := bufio.NewReader(port)
	//reply, err := reader.ReadBytes('\x00')
	//nb := int(0)
	for {
		//log.Printf("reading up to %d bytes\n", cap(buf))
		n, err = port.Read(buf[:cap(buf)])
		buf = buf[:n]
		//log.Printf("got %d bytes\n", n)
		//log.Printf("len %d\n", len(buf))
		//nb += n
		//log.Printf("bytes %d\n", nb)
		//log.Printf("received: [% x]\n", buf)
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				err = nil
				break
			}
			return
		}
		c = append(c, buf[:n]...)
	}
	//log.Printf("received: [% x]\n", c)
	return
}

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
