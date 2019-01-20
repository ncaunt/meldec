package codes_test

import (
	"testing"

	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
)

func TestChecksum(t *testing.T) {
	var checksumTests = []struct {
		code        []byte
		expectValid bool
	}{
		{[]byte("fc62027a10011005120b24280800000000000000008b"), true},
		{[]byte("fc62027a10011005120b2428080000000000000000ff"), false},
	}

	for _, test := range checksumTests {
		_, err := codes.NewCodeFromHex(test.code)
		t.Logf("%+v\n", err)
		if err == nil {
			if test.expectValid == false {
				t.Errorf("expected invalid checksum but it was valid")
			}
		} else if err.Error() == "checksum invalid" {
			if test.expectValid == true {
				t.Errorf("expected valid checksum but it was invalid")
			}
		} else {
			t.Error(err)
		}
	}
}
