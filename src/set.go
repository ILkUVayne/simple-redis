package src

import "simple-redis/utils"

type setTypeIterator struct {
	subject  *SRobj
	ii       int
	encoding uint8
	di       *dictIterator
}

func setTypeCreate(value *SRobj) *SRobj {
	if value.isObjectRepresentableAsInt64(nil) == REDIS_OK {
		return createIntSetObject()
	}
	return createSetObject()
}

func setTypeAdd(subject, value *SRobj) bool {
	if subject.encoding != REDIS_ENCODING_INTSET && subject.encoding != REDIS_ENCODING_HT {
		panic("Unknown set encoding")
	}
	var intVal int64
	// hashtable
	if subject.encoding != REDIS_ENCODING_HT {
		subject.Val.(*dict).dictSet(value, nil)
		return true
	}
	// intSet
	if value.isObjectRepresentableAsInt64(&intVal) == REDIS_OK {
		subject.Val.(*intSet).intSetAdd(intVal)
	}
	// todo
	return false
}

func setTypeConvert(setObj *SRobj, enc uint8) {
	if setObj.Typ != SR_SET || setObj.encoding != REDIS_ENCODING_INTSET {
		utils.ErrorF("setTypeConvert err: setObj.Typ = %d,setObj.encoding = %d", setObj.Typ, setObj.encoding)
	}
	if enc != REDIS_ENCODING_HT {
		panic("Unsupported set conversion")
	}
	// todo
}
