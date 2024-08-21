package src

import "testing"

func TestStrEncoding(t *testing.T) {
	strObj := createSRobj(SR_STR, "hello")
	if strObj.strEncoding() != encodingMaps[REDIS_ENCODING_RAW] {
		t.Error("strObj.strEncoding() = ", strObj.strEncoding())
	}
	strObj.encoding = 100
	if strObj.strEncoding() != UNKNOWN {
		t.Error("strObj.strEncoding() = ", strObj.strEncoding())
	}
}

func TestStrType(t *testing.T) {
	strObj := createSRobj(SR_STR, "hello")
	if strObj.strType() != TypeMaps[SR_STR] {
		t.Error("strObj.strType() = ", strObj.strType())
	}
	strObj.Typ = 100
	if strObj.strType() != UNKNOWN {
		t.Error("strObj.strType() = ", strObj.strType())
	}
}

func TestIntVal(t *testing.T) {
	o := createSRobj(SR_STR, "15")
	n, err := o.intVal()
	if err != nil {
		t.Error(err)
	}
	if n != 15 {
		t.Error("intVal err: n = ", n)
	}
}

func TestFloatVal(t *testing.T) {
	o := createSRobj(SR_STR, "15.5")
	n, err := o.floatVal()
	if err != nil {
		t.Error(err)
	}
	if n != 15.5 {
		t.Error("intVal err: n = ", n)
	}
}

func TestCreateZsetSRobj(t *testing.T) {
	zs := createZsetSRobj()
	if zs == nil {
		t.Error("createZsetSRobj err: zs == nil")
	}
	if zs.Typ != SR_ZSET {
		t.Error("createZsetSRobj err: zs.Typ != SR_ZSET")
	}
	if assertZSet(zs).zsl.length != 0 {
		t.Error("createZsetSRobj err: assertZSet(zs).zsl.length != 0")
	}
}

func TestCreateIntSetObject(t *testing.T) {
	is := createIntSetObject()
	if is == nil {
		t.Error("createIntSetObject err: is == nil")
	}
	if is.Typ != SR_SET {
		t.Error("createIntSetObject err: is.Typ != SR_SET")
	}
	if sLen(assertIntSet(is)) != 0 {
		t.Error("createIntSetObject err: assertIntSet(is).intSetLen() != 0")
	}
}

func TestCreateSetObject(t *testing.T) {
	set := createSetObject()
	if set == nil {
		t.Error("createSetObject err: is == nil")
	}
	if set.Typ != SR_SET {
		t.Error("createSetObject err: set.Typ != SR_SET")
	}
	if sLen(assertDict(set)) != 0 {
		t.Error("createSetObject err: sLen(assertDict(set)) != 0")
	}
}

func TestCreateListObject(t *testing.T) {
	l := createListObject()
	if l == nil {
		t.Error("createListObject err: is == nil")
	}
	if l.Typ != SR_LIST {
		t.Error("createListObject err: l.Typ != SR_LIST")
	}
	if sLen(assertList(l)) != 0 {
		t.Error("createListObject err: assertList(l).len() != 0")
	}
}

func TestCreateHashObject(t *testing.T) {
	h := createHashObject()
	if h == nil {
		t.Error("createHashObject err: is == nil")
	}
	if h.Typ != SR_DICT {
		t.Error("createHashObject err: h.Typ != SR_DICT")
	}
	if sLen(assertDict(h)) != 0 {
		t.Error("createHashObject err: sLen(assertDict(h)) != 0")
	}
}
