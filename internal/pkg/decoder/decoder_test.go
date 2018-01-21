package decoder_test

import (
	"sync"
	"testing"

	"github.com/ncaunt/meldec/internal/pkg/decoder"
	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
)

func processCode(in []byte, expect int) (ret []interface{}, err error) {
	dec := decoder.NewDecoder()

	wg := sync.WaitGroup{}
	wg.Add(expect)

	// read codes from channel
	go func() {
		for code := range dec.Stats() {
			ret = append(ret, code)
			wg.Done()
		}
	}()

	c, err := codes.NewCode(in)
	if err != nil {
		return
	}

	err = dec.Decode(c)
	if err != nil {
		return
	}

	wg.Wait()
	return
}

func TestDecoder(t *testing.T) {
	var codeTests = []struct {
		code   []byte
		values map[string]string
	}{
		{[]byte("fc62027a10011005120b24280800000000000000008b"), map[string]string{"timestamp/year": "16", "timestamp/month": "5", "timestamp/day": "18", "timestamp/hour": "11", "timestamp/minute": "36", "timestamp/second": "40"}},
		{[]byte("fc62027a100b0802f0c4f0c40b09c40b6c0000000046"), map[string]string{"temperatures/indoor/zone1": "20.5", "temperatures/indoor/zone2": "-39.0", "temperatures/outdoor/front": "14"}},
	}
	for _, code := range codeTests {
		result, err := processCode(code.code, len(code.values))
		if err != nil {
			t.Error(err)
		}

		if len(result) != len(code.values) {
			t.Errorf("expected %d values but got %d", len(code.values), len(result))
		}
		for _, v := range result {
			switch v.(type) {
			case error:
				t.Errorf("expected value %s but got %s", v.(error))
			default:
				s, ok := v.(decoder.Stat)
				if !ok {
					t.Errorf("got an incorrect type")
				}
				v, ok := code.values[s.Name]
				if !ok {
					t.Errorf("key %s in result was unexpected", s.Name)
				}
				if s.Value != v {
					t.Errorf("expected value %f but got %f", v, s.Value)
				}
			}
		}
	}
}
