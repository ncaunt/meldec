package main

import (
	"io"
	"log"
	"sync"

	"github.com/tarm/serial"
)

type SerialComm interface {
	Init()
	Send([]byte) ([]byte, error)
}

type TTY struct {
	port *serial.Port
	m    sync.Mutex
}

func NewTTY(conf *serial.Config) (tty *TTY, err error) {
	p, err := serial.OpenPort(conf)
	if err != nil {
		return
	}
	tty = &TTY{port: p}
	return
}

func (s *TTY) Init() {
	// initialisation codes
	// these must be sent first (at least after each HP power cycle) or the HP will not respond
	cmds := [][]byte{
		{0x02, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x02},
		{0x02, 0xff, 0xff, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00},
		{0xfc, 0x5a, 0x02, 0x7a, 0x02, 0xca, 0x01, 0x5d},
		{0xfc, 0x5b, 0x02, 0x7a, 0x01, 0xc9, 0x5f},
		{0xfc, 0x41, 0x02, 0x7a, 0x10, 0x34, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0xfd},
	}

	// the following two codes do the same thing (set zone1 to 18°C)
	// melcloud.com responds with both when changing the temperature via the website/app
	// only one needs to be sent in order to change the temperature
	//"FC41027A103280000102000002138807080DAC00001E",
	//"FC41027A1035020000070800000000000000000000F2",

	for _, x := range cmds {
		/*
			decoded := make([]byte, hex.DecodedLen(len(x)))
			_, err = hex.Decode(decoded, []byte(x))
			if err != nil {
				log.Fatal(err)
			}
			decoded[21] = sum(decoded[1:21])
		*/
		decoded := x
		_, err := s.Send(decoded)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s *TTY) Send(data []byte) (c []byte, err error) {
	log.Printf("sending\t[% x]\n", data)
	s.m.Lock()
	defer func() {
		s.m.Unlock()
	}()
	n, err := s.port.Write(data)
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
		n, err = s.port.Read(buf[:cap(buf)])
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
