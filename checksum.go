package main

import (
	"fmt"
)

const magic byte = 19

type ChecksumCode struct {
	Preamble  [5]byte
	GroupCode byte
	Data      [15]byte
	Checksum  byte
}

func checksum(c []byte) {
	g := c[5]
	cs := g - magic
	for _, i := range c[6:21] {
		cs += i
	}
	cs = ^cs

	if cs == c[21] {
		fmt.Printf("checksums match\n")
	} else {
		fmt.Printf("checksums differ by %d\n", c[21]-cs)
	}

}
