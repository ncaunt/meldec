# MELCloud Data Decoder (`meldec`)

[![Gitter](https://badges.gitter.im/ncaunt/meldec.svg)](https://gitter.im/ncaunt/meldec?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Introduction

`meldec` is designed to decode data read from a Mitsubishi Ecodan heat pump flow controller unit. It can act as an almost-complete open source replacement for the MELCloud wifi data reporter. Direct communication with a heat pump is possible over a serial connection and the software can also be used as an HTTP proxy and supports packet capture for passive network sniffing.

An `MQTT` server is required in order to publish decoded data.

## Building
Tested with golang up to version 1.11.5.
Build the serial comms binary:
```shell
go mod init github.com/ncaunt/meldec
go mod tidy
go build ./cmd/meldec
```

Other binaries can be found in the `cmd` directory and are built in a similar way.

## Options
```shell
$ ./meldec --help
Usage of ./meldec:
  -b int
      baud rate (default 2400) (default 2400)
  -d string
      serial device (default "/dev/ttyS0")
  -mqttdebug
      Turn on MQTT debugging messages
  -mqttserver string
      MQTT server (default "tcp://127.0.0.1:1883")
  -p string
      parity (default E: even) (default "E")
  -t string
      communication timeout (default "200ms")
```

## Data acquisition methods

### Direct serial communication
This method is very reliable and I have been using this for several months with no issues to report. The serial port on the FTC5 flow controller operates at 5 volts so it is important that the connected device supports this voltage either natively or with a level shifter. This has been tested with a USB FTDI cable ([Adafruit FTDI Friend](https://www.adafruit.com/product/284) configured for 5 V) and a standard 3.3 V UART with a level shifter. The connection for most, perhaps all, units is at 2400 baud with even parity. Although this seems slow it is possible to scan all data within approximately 20 seconds.

`meldec` will construct an HTTP request which is compatible with what is expected by Mitsubishi's servers. It will send to the heat pump any commands that are received in the response, thus allowing control of the unit in the same manner as with a MELCloud unit connected.

### HTTP proxy
`cmd/meldec-http` is a man-in-the-middle HTTP proxy. It requires a way to coerce the MELCloud unit to make HTTP requests to the proxy. If you have control of your network's DNS server then it's a simple task to add an entry to direct Mitsubishi's hostname to the address of the machine where you're running `meldec-http`. Another method with which I've had success is to spoof the ARP address of the network's gateway so MELCloud connects to the proxy in preference. `arpspoof` from the [dsniff toolkit](https://www.monkey.org/~dugsong/dsniff/) is able to accomplish this task. One caveat with this method is that DNS requests can also be sent to the proxy server so these will need to be handled in order to avoid data loss.

### Network packet capture
`cmd/meldec-pcap` passively captures data on a network (presumably wireless if used with the MELCloud device) but there is no support for proxying the data as the original packets will still reach Mitsubishi's service. This method is somewhat unreliable as packets might be missed or retransmitted leading either to no data being received or duplicates appearing. WPA encryption can also cause issues but can be worked around using `aircrack-ng`.

Have at least 3 terminals ready (`screen` or `tmux` is useful here). You'll need to know the name of your wireless interface (wlan0 in my case, wlan0mon in monitor mode), the key for a WEP or WPA network and the MAC addresses of your access point and the MELCloud device (found on a sticker on its case, or pick it up from your router's ARP table).

Execute these commands in order (replacing variables as necessary):
```shell
airmon-ng start $IFACE                                # start interface in monitor mode
airodump-ng --channel $CHANNEL $MONITOR_IFACE         # capture packets on monitor interface (optional; new terminal, keep running
airtun-ng -p $PSK -a $AP_MAC -e $SSID $MONITOR_IFACE  # set up decrypted tunnel )new terminal, keep running)
ifconfig at0 up                                       # bring up tunnel interface
aireplay-ng -a $AP_MAC -c $MELCLOUD_MAC -0 1 $MONITOR_IFACE   # send spoofed de-authenticate packets to MELCloud device
```
At this point the terminal in which airtun-ng is running should report something like this:
```shell
WPA handshake: aa:bb:cc:dd:ee:ff
```
You're good to go. Run `meldec-pcap` and wait for some data to appear. The MELCloud unit should report data every minute so it won't take long to see something assuming the device is in range. See below for further tips.

1. Get [aircrack-ng](http://aircrack-ng.org/). If you're lucky your operating system will have a package you can install. If you're not so lucky you can download the source and build it yourself.
2. Put your wireless adapter into monitor mode so `aircrack-ng` has a chance at receiving the packets. This will depend on the adapter, its driver and the OS so check the `aircrack-ng` docs for supported devices.
3. Run `airodump-ng` to lock the interface to the channel used by your network. If you've got an exotic channel-hopping setup then you'll probably have a bad time.
4. The clever bit: create a tunnel with `airtun-ng`. Give it your network key and it will provide a new interface with all the traffic decrypted. Well, there is a catch. In order to decrypt communication with a device it must first observe a handshake where the device attempts to associate with the access point. `aireplay-ng` can spoof packets that force a device off the network in the hope that it will try and rejoin.
5. Run `meldec-pcap` and wait for some data. Errors are quite likely (see the FAQ) but hopefully you'll get a decoded transmission.

## Data format
The MELCloud device sends an HTTP POST request with a body containing an XML document (notably, the connection does not use TLS). The XML contains a single node named LSV which contains _another_ XML document, base64-encoded. This is where the interesting stuff is found.

In addition to some system information, there are a number of hexadecimal strings which encode thermistor readings and status information. Each string contains some header bytes and a single byte "group code", which is duplicated in a sibling node in the XML. The remaining characters are a mixture of 8 and 16 bit integers with a final trailing checksum byte.

## How did I get here?
I started by capturing 802.11 packets and decrypting them with `aircrack-ng` before loading the capture into Wireshark and saving the assembled HTTP streams for analysis. It was obvious the XML contained something base64-encoded so I decoded that and found another XML document. Yo dawg, we heard you like XML...

After a lot of messing with Python one-liners to make sense of the hex strings, I finally made decent progress when I inspected the HTTP responses from the MELCloud server in combination with changing temperatures in the portal. Some diffing and guesswork revealed enough of the structure in the strings to correlate byte sequences with data points.

## Notes
Boiler flow and return temps (THWB1 and THWB2) are fixed at 25ºC when no boiler is present in the system (probably configured with DIP switches).

Outside temp has a precision of 1ºC and is represented differently from other thermistor readings, which have a precision of 0.5ºC.

## FAQ

**I need help with aircrack-ng! Why don't I get any data?**
Since it doesn't require associating with a base station this method is somewhat hit and miss. Missing packets won't be retransmitted and duplicate packets might appear unexpectedly. You ideally will need to be in range of both the MELCloud unit and the access point. Your operating system, wireless adapter and drivers will affect your chances of success.

**It keeps moaning about invalid HTTP streams or bad XML encoding! What's going on?**
If you're using the `aircrack-ng` tunnel method then this is just something you'll have to put up with. Since it's eavesdropping on the network, the normal TCP flow control doesn't apply; it's not ACKing packets and if any are missed they will not be retransmitted. Move your wireless receiver to a better location or try using the HTTP proxy method.

**Can I run this on a Raspberry Pi?**
Yes! I have been running this on an original Pi model B for months on end without major issues. See the `build-rpi.sh` script in the repository for clues on cross-compiling since building directly on the Pi can take a long time. A very long time.

**Why did you write this in Go? / Why didn't you use [my preferred language]?**
Because I felt like it and the libpcap library (gopacket) works effortlessly. I would be interesting to see it ported to another language. I might try with Rust one of these days.

## Future
Some stuff I fancy adding in no particular order.

 - Modularise the codebase.
 - Expand on Prometheus metrics.
 - Decode more of the data.

## Licence
A portion of this code is based on https://godoc.org/github.com/google/gopacket/examples/httpassembly (noted in the file header).
