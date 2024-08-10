package src

import "testing"

func TestNextLineIdx(t *testing.T) {
	s := "hello  world"
	i := 5
	idx := nextLineIdx(s, i)
	if idx != 7 {
		t.Error("nextLineIdx err: idx = ", idx)
	}
}

func TestSplitArgs(t *testing.T) {
	cmd := ` "get" name`
	ss := splitArgs(cmd)
	if ss[0] != "get" && ss[1] != "name" {
		t.Error("splitArgs err")
	}
	cmd = ` "get" "na\"me"`
	ss = splitArgs(cmd)
	if ss[0] != "get" && ss[1] != "na\"me" {
		t.Error("splitArgs err")
	}
	cmd = ` "get" 'na\'me'`
	ss = splitArgs(cmd)
	if ss[0] != "get" && ss[1] != "na'me" {
		t.Error("splitArgs err")
	}
	cmd = ` "get" 'na me'`
	ss = splitArgs(cmd)
	if ss[0] != "get" && ss[1] != "na me" {
		t.Error("splitArgs err")
	}
	cmd = ` "get" "na\nme"`
	ss = splitArgs(cmd)
	if ss[0] != "get" && ss[1] != "na\nme" {
		t.Error("splitArgs err")
	}
}
