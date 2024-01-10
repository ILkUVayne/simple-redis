package src

import "testing"

func TestIntSetNew(t *testing.T) {
	is := intSetNew()
	if is.intSetLen() != 0 {
		t.Error("intSetNew err: is.length = ", is.length)
	}
	if len(is.contents) != DEFAULT_INTSET_BUF {
		t.Error("intSetNew err: len(is.contents) = ", len(is.contents))
	}
}

func TestIntSetAdd(t *testing.T) {
	is := intSetNew()
	var success bool
	is.intSetAdd(10, &success)
	if is.intSetLen() != 1 {
		t.Error("intSetAdd err: is.length = ", is.length)
	}
	is.intSetAdd(10, &success)
	if is.intSetLen() != 1 {
		t.Error("intSetAdd err: is.length = ", is.length)
	}
	is.intSetAdd(7, &success)
	if is.contents[0] != 7 {
		t.Error("intSetAdd err: is.contents[0] = ", is.contents[0])
	}
	is.intSetAdd(8, &success)
	if is.contents[1] != 8 {
		t.Error("intSetAdd err: is.contents[0] = ", is.contents[0])
	}
	is.intSetAdd(5, &success)
	is.intSetAdd(9, &success)
	is.intSetAdd(1, &success)
	if is.intSetLen() != 6 {
		t.Error("intSetAdd err: is.length = ", is.length)
	}
}

func TestIntSetRemove(t *testing.T) {
	var success bool
	is := intSetNew()
	is.intSetRemove(10)
	is.intSetAdd(2, &success)
	is.intSetAdd(28, &success)
	is.intSetAdd(5, &success)
	is.intSetAdd(9, &success)
	is.intSetAdd(10, &success)
	is.intSetAdd(13, &success)
	is.intSetRemove(10)
	if is.intSetLen() != 5 {
		t.Error("intSetRemove err: is.length = ", is.length)
	}
}
