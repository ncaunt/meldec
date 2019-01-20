package decoder_test

import (
	"fmt"
	"testing"

	"github.com/ncaunt/meldec/internal/pkg/decoder"
	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
)

func processCode(in []byte, expect int) (ret []interface{}, err error) {
	fmt.Printf("input % x, expecting %d values\n", in, expect)
	dec := decoder.NewDecoder()

	c, err := codes.NewCodeFromHex(in)
	if err != nil {
		return
	}

	s, err := dec.Decode(c)
	if err != nil {
		return
	}

	for _, code := range s {
		ret = append(ret, code)
	}
	return
}

func TestDecoder(t *testing.T) {
	var codeTests = []struct {
		code     []byte
		expected map[string]string
	}{
		{[]byte("fc62027a10011005120b24280800000000000000008b"), map[string]string{"timestamp/year": "16", "timestamp/month": "5", "timestamp/day": "18", "timestamp/hour": "11", "timestamp/minute": "36", "timestamp/second": "40"}},
		{[]byte("fc62027a100b0802f0c4f0c40b09c40b6c0000000046"), map[string]string{"temperatures/indoor/zone1": "20.5", "status/group11/code2": "-39.0", "temperatures/indoor/zone2": "-39.0", "status/group11/code4": "2825", "status/group11/code5": "-15349", "temperatures/outdoor/front": "14"}},
	}

	for _, code := range codeTests {
		result, err := processCode(code.code, len(code.expected))
		if err != nil {
			t.Error(err)
		}

		if len(result) != len(code.expected) {
			t.Errorf("expected %d values but got %d", len(code.expected), len(result))
		}

		for _, v := range result {
			s, ok := v.(decoder.Stat)
			if !ok {
				t.Errorf("got an incorrect type")
			}
			v, ok := code.expected[s.Name]
			if !ok {
				t.Errorf("key %s in result was unexpected", s.Name)
			}
			if s.Value != v {
				t.Errorf("expected value %s but got %s", v, s.Value)
			}
		}
	}
}
