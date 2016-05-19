MELCloud Data Decoder (`meldec`)
---------------------

What?
-----
**TL;DR**
See the status of your Mitsubishi Ecodan heat pump with this one weird old tip!

Mitsubishi Ecodan heat pumps can be equipped with a MELCloud wifi data reporter which sends system parameters to Mitsubishi's servers, and an official portal allows review of the data with a degree of control over the unit (operation mode, setting temperatures, etc). While what's offered is good, I wanted to capture the data for myself rather than rely on a third party service. I also suspected that additional information was being transmitted beyond what was made available in the portal, and I'm curious about such things.

This project will sniff packets on your network and attempt to decode the transmissions from the MELCloud unit. Additionally it can be used offline by reading a capture file.

Quick setup guide for the impatient
-----------------------------------
Install Go. Build this code. I'm not holding your hand with this bit so RTFM if you need to: https://golang.org/doc/code.html

Usage of `meldec`:

      -assembly_debug_log
            If true, the github.com/google/gopacket/tcpassembly library will log verbose debugging information (at least one line per packet)
      -assembly_memuse_log
            If true, the github.com/google/gopacket/tcpassembly library will log information regarding its memory use every once in a while.
      -cpuprofile string
            Where to write CPU profile
      -f string
            BPF filter for pcap (default "tcp and dst port 80")
      -i string
            Interface to get packets from (default "eth0")
      -r string
            Filename to read from, overrides -i
      -s int
            SnapLen for pcap packet capture (default 1600)
      -v    Logs every packet in great detail
      -w    Show all raw values

The rest depends on the specifics of your wireless network and the hardware you have available but this is how I do it with help from aircrack-ng.

Have at least 3 terminals ready. You'll need to know the name of your wireless interface (wlan0 in my case, wlan0mon in monitor mode), the key for a WEP or WPA network and the MAC addresses of your access point and the MELCloud device (found on a sticker on its case, or pick it up from your router's ARP table or something).

Execute these commands in order (replacing stuff in angle brackets, natch):
```
airmon-ng start <IFACE>
airodump-ng --channel <CHANNEL> <MONITOR IFACE> (new terminal, keep running)
airtun-ng -p <PSK> -a <AP MAX> -e <SSID> <MONITOR IFACE> (new terminal, keep running)
ifconfig at0 up
aireplay-ng -a <AP MAC> -c <MELCLOUD MAC> -0 1 <MONITOR IFACE>
```

At this point the terminal in which airtun-ng is running should report something like this:
WPA handshake: <MELCLOUD MAC>

You're good to go. Run `meldec` and wait for some data to appear. The MELCloud unit should communicate with the mothership every minute so it won't take long to see something assuming you're close enough to it (physically, not emotionally). If it doesn't look like it's working, read on.

Addendum to setup guide that explains more stuff for people who are interested in knowing what's going down and appreciate longer headings
-----------------------------------------
1. Get [aircrack-ng](http://aircrack-ng.org/). If you're lucky your operating system will have a package you can install. If you're not so lucky you can download the source and build it yourself. If you're not running an operating system that aircrack-ng supports then maybe you should stop bitching and think about running an operating system aircrack-ng supports. I mean, you can even do this shit in a VM using a free VM player and a free operating system. For free.

2. Put your wireless adapter into monitor mode so aircrack-ng has a chance at receiving the packets. This will depend on the adapter, its driver and the OS so check the aircrack-ng docs for supported devices or just grow a pair and try it out.

3. Run airodump-ng to lock the interface to the channel used by your network. If you've got an exotic channel-hopping setup then you'll probably have a bad time.

4. The clever bit: create a tunnel with airtun-ng. Give it your network key and it will provide a new interface with all the traffic decrypted. Well, there is a catch. In order to decrypt communcation with a device it must first observe a handshake where the device attempts to associate with the access point. So do we have to wait for that to happen? Turn it off and back on again? Screw that idea! aireplay-ng can spoof packets that force a device off the network in the hope that it will try and rejoin.

5. Run `meldec` and wait for some data. Errors are quite likely (see the FAQ) but cross your fingers and eventually you'll get a decoded transmission.

Data format
-----------
HTTP POST request with a body containing an XML document. The XML contains a single node named LSV which contains _another_ XML document, base64-encoded. This is where the interesting stuff is found.

In addition to some system information, there are a number of hexadecimal strings which encode thermistor readings and status information. The majority are prefixed by a fixed 5 byte preamble followed by a single byte which corresponds to a "group code", which is duplicated in a sibling node in the XML. The remaining characters are a mixture of 8 and 16 bit integers with a final trailing byte which I suspect is a checksum. This smells to me like data packets that were originally designed to be transmitted over radio frequency to a local receiver, now re-purposed for transmission over the Internet. It might even be the same communication mechanism used by the wireless remote control panel, but I haven't got around to investigating that yet. Watch this space, my RTL-SDR dongle is standing by...

How did I get here?
-------------------
I started by capturing 802.11 packets and decrypting them with aircrack-ng. Then I could load the capture into Wireshark and save the assembled HTTP streams for analysis. It was obvious the XML contained something base64-encoded so I decoded that and found another XML document. Yo dawg...

Did a lot of messing with Python one-liners to make sense of the hex strings I found, but finally made decent progress when I inspected the HTTP responses from the MELCloud server in combination with changing temperatures in the portal. Some diffing and guesswork revealed enough of the structure in the strings to correlate byte sequences with data points.


Notes
-----
Boiler flow and return temps (THWB1 and THWB2) are fixed at 25ºC when no boiler is present in the system (probably configured with DIP switches).

Outside temp has a precision of 1ºC and is represented differently from other thermistor readings, which have a precision of 0.5ºC.

FAQ
---
**How can you have frequently asked questions when this thing has just been published?**
Shut up, I'm trying to be helpful.

**I need help with aircrack-ng! Why don't I get any data?**
RTFM. That's what I did. Trial and error, my friend. The instructions above worked for me, I promise.

**It keeps moaning about invalid HTTP streams or bad XML encoding! What's going on?**
If you're using the aircrack-ng tunnel method then this is just something you'll have to put up with. Since it's eavesdropping on the network, the normal TCP flow control doesn't apply; it's not ACKing packets and if any are missed they will not be retransmitted. Move your wireless receiver to a better location or get freaky by re-routing packets over a more reliable transport.

**Can I run this on a Raspberry Pi?**
Yes, probably. I haven't tested that yet but I see no reason why not.

**Why did you write this in Go? / Why didn't you use [my preferred language]?**
Because I felt like it and the libpcap library (gopacket) works effortlessly. / Feel free to port it, snowflake.

**Where's the Vagrantfile?**
Good question. Haven't got round to making one yet, that's all.

**Where's the Docker container?**
It's over there behind your mum.

Future
------
Some stuff I fancy adding in no particular order.

 - Find out why the outside temperature readings are encoded differently
 - Vagrantfile
 - Test on a Raspberry Pi
 - Output data to statsd or similar
 - Have the MELCloud unit associate with a SoftAP for more reliable capture
 - Debian packaging shiz

Licence
-------
A portion of this code is based on https://godoc.org/github.com/google/gopacket/examples/httpassembly