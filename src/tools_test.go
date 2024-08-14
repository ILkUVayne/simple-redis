package src

import (
	"testing"
)

func TestStringMatchLen(t *testing.T) {
	pattern1 := "*"
	s1 := "asda"
	res := StringMatchLen(pattern1, s1, len(pattern1), len(s1), false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
	pattern1 = "*d*"
	s1 = "asda"
	res = StringMatchLen(pattern1, s1, len(pattern1), len(s1), false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
	pattern1 = "a*a"
	s1 = "asda"
	res = StringMatchLen(pattern1, s1, len(pattern1), len(s1), false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
	pattern1 = "as?a"
	s1 = "asda"
	res = StringMatchLen(pattern1, s1, len(pattern1), len(s1), false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
	pattern1 = "as[de]a"
	s1 = "asda"
	res = StringMatchLen(pattern1, s1, len(pattern1), len(s1), false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
	s1 = "asea"
	res = StringMatchLen(pattern1, s1, len(pattern1), len(s1), false)
	if !res {
		t.Error("StringMatchLen err: res = ", res)
	}
}
