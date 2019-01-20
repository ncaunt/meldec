package codes

func verify_checksum(c []byte) bool {
	cl := len(c)
	sum := calc_checksum(c[1 : cl-1])
	return sum == c[cl-1]
}

// the final byte in the code is the checksum itself and not included in the calculation
func calc_checksum(c []byte) (sum byte) {
	for _, i := range c {
		sum -= i
	}
	return
}
