package codes

func verify_checksum(c []byte, sum byte) bool {
	sum = calc_checksum(c)
	return sum == c[codeLen-1]
}

// the first byte does not seem to be involved in the checksum
// the final byte in the code is the checksum itself and not included in the calculation
func calc_checksum(c []byte) (sum byte) {
	for _, i := range c[1 : codeLen-1] {
		sum -= i
	}

	return
}
