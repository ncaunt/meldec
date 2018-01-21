MELCloud Data Decoder (`meldec`)
---------------------

What?
-----

**TL;DR**
See the status of your Mitsubishi Ecodan heat pump.

Mitsubishi Ecodan heat pumps can be equipped with a MELCloud wifi data reporter which sends system parameters to Mitsubishi's servers. An official portal allows review of the data with a degree of control over the unit (operation mode, setting temperatures, etc). While what's offered is good, I wanted to capture the data for myself rather than rely on a third party service. I also suspected that additional information was being transmitted beyond what was made available in the portal, and I'm curious about such things.

This project will attempt to decode the transmissions from the MELCloud unit and offers a number of options for obtaining the data. The binary `meldec` starts an HTTP proxy which intercepts communications on their way to the command and control server. The `meldec-pcap` binary can sniff packets on the local network or be used offline by reading a capture file.

Building
--------
Tested with Go up to version 1.9.2.

Build the HTTP proxy binary:
```
go build ./cmd/meldec
```

Network packet capture version:
```
go build ./cmd/meldec-pcap
```


HTTP proxy mode
---------------
This is currently my preferred way of obtaining the heat pump's data but it requires a way to coerce the MELCloud unit to make HTTP requests to the proxy. If you have control of your network's DNS server then it's a simple task to add an entry to direct Mitsubishi's hostname to the address of the machine where you're running `meldec`.
Another method with which I've had success is to spoof the ARP address of the network's gateway so MELCloud connects to the proxy in preference. `arpspoof` from the [dsniff toolkit](https://www.monkey.org/~dugsong/dsniff/) is able to accomplish this task. One caveat with this method is that DNS requests can also be sent to the proxy server so these will need to be handled in order to avoid data loss.

Network packet capture mode
---------------------------

Have at least 3 terminals ready (`screen` or `tmux` is useful here). You'll need to know the name of your wireless interface (wlan0 in my case, wlan0mon in monitor mode), the key for a WEP or WPA network and the MAC addresses of your access point and the MELCloud device (found on a sticker on its case, or pick it up from your router's ARP table or something).

Execute these commands in order (replacing parameters in angle brackets):
```
airmon-ng start <IFACE>
airodump-ng --channel <CHANNEL> <MONITOR IFACE> (new terminal, keep running)
airtun-ng -p <PSK> -a <AP MAX> -e <SSID> <MONITOR IFACE> (new terminal, keep running)
ifconfig at0 up
aireplay-ng -a <AP MAC> -c <MELCLOUD MAC> -0 1 <MONITOR IFACE>
```
At this point the terminal in which airtun-ng is running should report something like this:
```
WPA handshake: aa:bb:cc:dd:ee:ff
```
You're good to go. Run `meldec-pcap` and wait for some data to appear. The MELCloud unit should communicate with the mothership every minute so it won't take long to see something assuming you're close enough to it (physically, not emotionally). If it doesn't look like it's working, read on.

Addendum to setup guide
-----------------------

1. Get [aircrack-ng](http://aircrack-ng.org/). If you're lucky your operating system will have a package you can install. If you're not so lucky you can download the source and build it yourself.
2. Put your wireless adapter into monitor mode so `aircrack-ng` has a chance at receiving the packets. This will depend on the adapter, its driver and the OS so check the `aircrack-ng` docs for supported devices.
3. Run `airodump-ng` to lock the interface to the channel used by your network. If you've got an exotic channel-hopping setup then you'll probably have a bad time.
4. The clever bit: create a tunnel with `airtun-ng`. Give it your network key and it will provide a new interface with all the traffic decrypted. Well, there is a catch. In order to decrypt communication with a device it must first observe a handshake where the device attempts to associate with the access point. `aireplay-ng` can spoof packets that force a device off the network in the hope that it will try and rejoin.
5. Run `meldec-pcap` and wait for some data. Errors are quite likely (see the FAQ) but hopefully you'll get a decoded transmission.

Data format
-----------
HTTP POST request with a body containing an XML document. The XML contains a single node named LSV which contains _another_ XML document, base64-encoded. This is where the interesting stuff is found.

In addition to some system information, there are a number of hexadecimal strings which encode thermistor readings and status information. The majority are prefixed by a fixed 5 byte preamble followed by a single byte which corresponds to a "group code", which is duplicated in a sibling node in the XML. The remaining characters are a mixture of 8 and 16 bit integers with a final trailing byte which <s>I suspect</s> is a checksum. This smells to me like data packets that were originally designed to be transmitted over radio frequency to a local receiver, now re-purposed for transmission over the Internet. It might even be the same communication mechanism used by the wireless remote control panel, but I haven't got around to investigating that yet.

How did I get here?
-------------------
I started by capturing 802.11 packets and decrypting them with `aircrack-ng` before loading the capture into Wireshark and saving the assembled HTTP streams for analysis. It was obvious the XML contained something base64-encoded so I decoded that and found another XML document. Yo dawg, we heard you like XML...

Did a lot of messing with Python one-liners to make sense of the hex strings I found, but finally made decent progress when I inspected the HTTP responses from the MELCloud server in combination with changing temperatures in the portal. Some diffing and guesswork revealed enough of the structure in the strings to correlate byte sequences with data points.

Notes
-----
Boiler flow and return temps (THWB1 and THWB2) are fixed at 25ºC when no boiler is present in the system (probably configured with DIP switches).

Outside temp has a precision of 1ºC and is represented differently from other thermistor readings, which have a precision of 0.5ºC.

FAQ
---

**I need help with aircrack-ng! Why don't I get any data?**
Since it doesn't require associating with a base station this method is somewhat hit and miss. Missing packets won't be retransmitted and duplicate packets might appear unexpectedly. You ideally will need to be in range of both the MELCloud unit and the access point. Your operating system, wireless adapter and drivers will affect your chances of success.

**It keeps moaning about invalid HTTP streams or bad XML encoding! What's going on?**
If you're using the `aircrack-ng` tunnel method then this is just something you'll have to put up with. Since it's eavesdropping on the network, the normal TCP flow control doesn't apply; it's not ACKing packets and if any are missed they will not be retransmitted. Move your wireless receiver to a better location or try using the HTTP proxy method.

**Can I run this on a Raspberry Pi?**
Yes! I have been running this on an original Pi model B for months on end without major issues. See the `build-rpi.sh` script in the repository for clues on cross-compiling since building directly on the Pi can take a long time. A very long time.

**Why did you write this in Go? / Why didn't you use [my preferred language]?**
Because I felt like it and the libpcap library (gopacket) works effortlessly. I would be interesting to see it ported to another language. I might try with Rust one of these days.

Future
------
Some stuff I fancy adding in no particular order.

 - Find out why the outside temperature readings are encoded differently.
 - Decode more of the data.
 - Control the device in a similar manner to the MELCloud app.

Licence
-------
A portion of this code is based on https://godoc.org/github.com/google/gopacket/examples/httpassembly

