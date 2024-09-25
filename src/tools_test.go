package src

import (
	"testing"
)

func TestStringMatchLen(t *testing.T) {
	pattern1 := "*"
	s1 := "asda"
	res := StringMatchLen(pattern1, s1, false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}

	pattern1 = "***"
	s1 = "asda"
	res = StringMatchLen(pattern1, s1, false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}

	pattern1 = "*d*"
	s1 = "asda"
	res = StringMatchLen(pattern1, s1, false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
	pattern1 = "a*a"
	s1 = "asda"
	res = StringMatchLen(pattern1, s1, false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
	pattern1 = "as?a"
	s1 = "asda"
	res = StringMatchLen(pattern1, s1, false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
	pattern1 = "as[de]a"
	s1 = "asda"
	res = StringMatchLen(pattern1, s1, false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
	s1 = "asea"
	res = StringMatchLen(pattern1, s1, false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
}

func TestFormatFloat(t *testing.T) {
	var f float64
	f = 12
	str := formatFloat(f, 10)
	if str != "12" {
		t.Error("formatFloat err: str = ", str)
	}
	f = 12.1
	str = formatFloat(f, 10)
	if str != "12.1" {
		t.Error("formatFloat err: str = ", str)
	}
	f = 12.12345678919
	str = formatFloat(f, 10)
	if str != "12.1234567892" {
		t.Error("formatFloat err: str = ", str)
	}
}
