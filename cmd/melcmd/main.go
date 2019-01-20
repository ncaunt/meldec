package main

import (
	"encoding/hex"
	"flag"
	"io"
	"log"
	"strings"
	"sync"
	"time"

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
	var cmd = flag.String("c", "", "command to send")
	var append_checksum = flag.Bool("a", false, "append checksum to end of data")
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

	// initialisation codes
	// these must be sent first (after each power cycle?) or the heat pump will not respond
	cmds := [][]byte{
	//		{0x02, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x02},
	//		{0x02, 0xff, 0xff, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00},
	//		{0xfc, 0x5a, 0x02, 0x7a, 0x02, 0xca, 0x01, 0x5d},
	//		{0xfc, 0x5b, 0x02, 0x7a, 0x01, 0xc9, 0x5f},
	//		{0xfc, 0x41, 0x02, 0x7a, 0x10, 0x34, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0xfd},
	}

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
		res, err := send(decoded)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("received\t[% x]\n", res)
	}

	in := strings.Split(*cmd, ",")
	for _, x := range in {
		l := hex.DecodedLen(len(x))
		if *append_checksum {
			l++
		}
		decoded := make([]byte, l)
		_, err = hex.Decode(decoded, []byte(x))
		if err != nil {
			log.Fatal(err)
		}
		if *append_checksum {
			var sum byte
			for _, i := range decoded[1 : l-1] {
				sum -= i
			}
			decoded[l-1] = sum
		}

		res, err := send(decoded)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("received\t[% x]\n", res)
	}
}

func send(data []byte) (c []byte, err error) {
	log.Printf("sending\t[% x]\n", data)
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
