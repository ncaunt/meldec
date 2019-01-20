package codes_test

import (
	"fmt"
	"testing"

	"github.com/ncaunt/meldec/internal/pkg/decoder/codes"
)

func assertCode(c []byte, expected map[string]interface{}) (err error) {
	code, err := codes.NewCodeFromHex(c)
	if err != nil {
		return
	}

	result, err := code.Decode()
	if err != nil {
		return
	}

	if len(result) != len(expected) {
		return fmt.Errorf("expected %d code but got %d\n", len(expected), len(result))
	}

	for k, v := range result {
		ev, ok := expected[k]
		if !ok {
			return fmt.Errorf("key %s in result was unexpected", k)
		}
		if v != ev {
			return fmt.Errorf("wanted %s, got %s", ev, v)
		}
	}
	return
}

func TestInvalidLength(t *testing.T) {
	var codeTests = []struct {
		code        []byte
		expectValid bool
	}{
		{[]byte("fc62027a10011005120b24280800000000000000008b"), true},
		{[]byte("deadbeef"), false},
		{[]byte("fc62027a10011005120b24280800000000000000008babc"), false},
	}

	for _, test := range codeTests {
		_, err := codes.NewCodeFromHex(test.code)
		t.Logf("%+v\n", err)
		if err == nil {
			if test.expectValid == false {
				t.Errorf("expected code to be invalid but it was valid")
			}
		} else {
			if test.expectValid == true {
				t.Errorf("expected code to be valid but got error: %s", err)
			}
		}
	}
}

func TestToHex(t *testing.T) {
	src := []byte("fc62027a10011005120b24280800000000000000008b")
	c, err := codes.NewCodeFromHex(src)
	if err != nil {
		t.Error(err)
	}
	got, err := c.ToHex()
	if err != nil {
		t.Error(err)
	}
	t.Logf("original is [%s] and received [%s]\n", src, got)
	if string(got) != string(src) {
		t.Errorf("expected [%s] but got [%s]\n", src, got)
	}
}

func TestGroupCode01(t *testing.T) {
	code := []byte("fc62027a10011005120b24280800000000000000008b")
	expected := map[string]interface{}{"timestamp/year": int8(16), "timestamp/month": int8(5), "timestamp/day": int8(18), "timestamp/hour": int8(11), "timestamp/minute": int8(36), "timestamp/second": int8(40)}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode03(t *testing.T) {
	code := []byte("fc62027a10030080000000000000020000000000008d")
	expected := map[string]interface{}{
		"status/group3/Code1":  uint8(0),
		"status/group3/Code2":  uint8(0x80),
		"status/group3/Code3":  uint8(0),
		"status/group3/Code4":  uint8(0),
		"status/group3/Code5":  uint8(0),
		"status/group3/Code6":  uint8(0),
		"status/group3/Code7":  uint8(0),
		"status/group3/Code8":  uint8(0),
		"status/group3/Code9":  uint8(2),
		"status/group3/Code10": uint8(0),
		"status/group3/Code11": uint8(0),
		"status/group3/Code12": uint8(0),
		"status/group3/Code13": uint8(0),
		"status/group3/Code14": uint8(0),
		"status/group3/Code15": uint8(0),
	}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode04(t *testing.T) {
	code := []byte("fc62027a10041c0000000000000000000000000000f2")
	expected := map[string]interface{}{"status/group4": uint8(28)}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode07(t *testing.T) {
	code := []byte("fc62027a100700000001000501010101000000000001")
	expected := map[string]interface{}{
		"status/group7/Code1":  uint8(0),
		"status/group7/Code2":  uint8(0),
		"status/group7/Code3":  uint8(0),
		"status/group7/Code4":  uint8(1),
		"status/group7/Code5":  uint8(0),
		"status/group7/Code6":  uint8(5),
		"status/group7/Code7":  uint8(1),
		"status/group7/Code8":  uint8(1),
		"status/group7/Code9":  uint8(1),
		"status/group7/Code10": uint8(1),
		"status/group7/Code11": uint8(0),
		"status/group7/Code12": uint8(0),
		"status/group7/Code13": uint8(0),
		"status/group7/Code14": uint8(0),
		"status/group7/Code15": uint8(0),
	}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode09(t *testing.T) {
	code := []byte("fc62027a100907d007d00bb80dac1770388c64000030")
	expected := map[string]interface{}{"temperatures/indoor/set_zone1": 20.0, "temperatures/indoor/set_zone2": 20.0}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode0b(t *testing.T) {
	code := []byte("fc62027a100b0802f0c4f0c40b09c40b6c0000000046")
	expected := map[string]interface{}{"temperatures/indoor/zone1": 20.5, "status/group11/code2": -39.0, "temperatures/indoor/zone2": -39.0, "status/group11/code4": int16(2825), "status/group11/code5": int16(-15349), "temperatures/outdoor/front": int8(14)}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode0c(t *testing.T) {
	code := []byte("fc62027a100c0c808f0af085173ec300000000000054")
	expected := map[string]interface{}{"temperatures/system/heating_feed": 32.0, "temperatures/system/heating_return": 28.0, "temperatures/system/hot_water": 59.5}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode0d(t *testing.T) {
	code := []byte("fc62027a100d09c40b09c40b09c40b09c40b000000a5")
	expected := map[string]interface{}{"temperatures/system/boiler_flow": 25.0, "temperatures/system/boiler_return": 25.0}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode13(t *testing.T) {
	code := []byte("fc62027a1013010011001f00000000000000000000ce")
	expected := map[string]interface{}{"status/group19": int8(1)}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode14(t *testing.T) {
	code := []byte("fc62027a1014000000000000000000000011000000ed")
	expected := map[string]interface{}{"status/group20": int8(17)}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode15(t *testing.T) {
	code := []byte("fc62027a10150100ff010000010000000400000000f7")
	expected := map[string]interface{}{"status/group21a": int8(1), "status/group21b": int8(0), "status/group21c": int8(1)}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}

func TestGroupCode26(t *testing.T) {
	code := []byte("fc62027a102600000102000002138807d00dac0000bc")
	expected := map[string]interface{}{"temperatures/system/set_hot_water": 50.0}

	if err := assertCode(code, expected); err != nil {
		t.Error(err)
	}
}
