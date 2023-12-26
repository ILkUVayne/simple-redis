package src

import (
	"testing"
)

func TestZslCreateNode(t *testing.T) {
	str := "qqqq"
	score := 1.2
	zn := zslCreateNode(7, 1.2, createSRobj(SR_STR, "qqqq"))
	if zn.obj.strVal() != str {
		t.Error("zslCreateNode err: zn.obj = ", zn.obj.strVal())
	}
	if zn.score != score {
		t.Error("zslCreateNode err: score = ", score)
	}
	if len(zn.level) != 7 {
		t.Error("zslCreateNode err: len(zn.level) = ", len(zn.level))
	}
}

func TestZslCreate(t *testing.T) {
	zsl := zslCreate()
	if len(zsl.header.level) != ZSKIPLIST_MAXLEVEL {
		t.Error("zslCreate err: len(zsl.header.level) = ", len(zsl.header.level))
	}
	if zsl.length != 0 {
		t.Error("zslCreate err: zsl.length = ", zsl.length)
	}
	if zsl.level != 1 {
		t.Error("zslCreate err: zsl.level = ", zsl.level)
	}
}

func TestZslInsert(t *testing.T) {
	zsl := zslCreate()
	o1 := createSRobj(SR_STR, "qqqq")
	zsl.insert(5.1, o1)
	o2 := createSRobj(SR_STR, "qqq")
	zsl.insert(4.6, o2)
	o3 := createSRobj(SR_STR, "qqq11")
	zsl.insert(4.8, o3)
	o4 := createSRobj(SR_STR, "qqq112")
	zsl.insert(4.7, o4)
	o5 := createSRobj(SR_STR, "qqq1122")
	zsl.insert(5.0, o5)
	if zsl.length != 5 {
		t.Error("zslInsert err: zsl.length = ", zsl.length)
	}
}

func TestZslDelete(t *testing.T) {
	zsl := zslCreate()
	o1 := createSRobj(SR_STR, "qqqq")
	zsl.insert(5.1, o1)
	o2 := createSRobj(SR_STR, "qqq")
	zsl.insert(4.6, o2)
	o3 := createSRobj(SR_STR, "qqq11")
	zsl.insert(4.8, o3)
	o4 := createSRobj(SR_STR, "qqq112")
	zsl.insert(4.7, o4)
	o5 := createSRobj(SR_STR, "qqq1122")
	zsl.insert(5.0, o5)
	res := zsl.delete(5.2, o5)
	if zsl.length != 5 {
		t.Error("delete err: zsl.length = ", zsl.length)
	}
	if res {
		t.Error("delete err: res = ", res)
	}
	res = zsl.delete(4.8, o3)
	if zsl.length != 4 {
		t.Error("delete err: zsl.length = ", zsl.length)
	}
	if !res {
		t.Error("delete err: res = ", res)
	}
}
