package src

import (
	"testing"
)

func TestSetTypeCreate(t *testing.T) {
	intVal := createSRobj(SR_STR, "21312")
	res := setTypeCreate(intVal)
	if res.encoding != REDIS_ENCODING_INTSET {
		t.Error("setTypeCreate err: res.encoding = ", res.encoding)
	}
	strVal := createSRobj(SR_STR, "asd312")
	res = setTypeCreate(strVal)
	if res.encoding != REDIS_ENCODING_HT {
		t.Error("setTypeCreate err: res.encoding = ", res.encoding)
	}
}

func TestSetTypeAdd(t *testing.T) {
	intVal := createSRobj(SR_STR, "21312")
	set := setTypeCreate(intVal)
	intVal.tryObjectEncoding()
	setTypeAdd(set, intVal)
	if set.encoding != REDIS_ENCODING_INTSET {
		t.Error("setTypeAdd err: set.encoding = ", set.encoding)
	}
	strVal := createSRobj(SR_STR, "asd312")
	setTypeAdd(set, strVal)
	if set.encoding != REDIS_ENCODING_HT {
		t.Error("setTypeAdd err: res.encoding = ", set.encoding)
	}
	if sLen(assertDict(set)) != 2 {
		t.Error("setTypeAdd err: sLen(assertDict(set)) = ", sLen(assertDict(set)))
	}
}
