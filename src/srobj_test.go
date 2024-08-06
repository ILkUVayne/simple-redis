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
