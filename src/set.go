package src

import (
	"simple-redis/cgo/qsort"
	"simple-redis/utils"
	"sort"
)

//-----------------------------------------------------------------------------
// Set commands API
//-----------------------------------------------------------------------------

type setTypeIterator struct {
	subject  *SRobj
	ii       int
	encoding uint8
	di       *dictIterator
}

// return -1 if next entry is nil
func (si *setTypeIterator) setTypeNext(objEle **SRobj, llEle *int64) int {
	if si.encoding == REDIS_ENCODING_INTSET {
		if !assertIntSet(si.subject).intSetGet(uint32(si.ii), llEle) {
			return -1
		}
		si.ii++
	}
	if si.encoding == REDIS_ENCODING_HT {
		de := si.di.dictNext()
		if de == nil {
			return -1
		}
		*objEle = de.getKey()
	}
	return int(si.encoding)
}

func (si *setTypeIterator) setTypeReleaseIterator() {
	if si.encoding == REDIS_ENCODING_HT {
		si.di.dictReleaseIterator()
	}
	si.di = nil
	si.subject = nil
}

func setTypeInitIterator(subject *SRobj) *setTypeIterator {
	checkSetEncoding(subject)
	si := new(setTypeIterator)
	si.subject = subject
	si.encoding = subject.encoding
	if si.encoding == REDIS_ENCODING_HT {
		si.di = assertDict(subject).dictGetIterator()
		return si
	}
	// REDIS_ENCODING_INTSET
	si.ii = 0
	return si
}

var setDictType = dictType{
	hashFunc:      SRStrHash,
	keyCompare:    SRStrCompare,
	keyDestructor: nil,
	valDestructor: nil,
}

func setTypeCreate(value *SRobj) *SRobj {
	if value.isObjectRepresentableAsInt64(nil) == REDIS_OK {
		return createIntSetObject()
	}
	return createSetObject()
}

func setTypeAdd(subject, value *SRobj) bool {
	checkSetEncoding(subject)
	var intVal int64
	// hashtable
	if subject.encoding == REDIS_ENCODING_HT {
		return assertDict(subject).dictAdd(value, nil)
	}
	// intSet
	if value.isObjectRepresentableAsInt64(&intVal) == REDIS_OK {
		var success bool
		assertIntSet(subject).intSetAdd(intVal, &success)
		return success
	}
	// change to ht
	setTypeConvert(subject, REDIS_ENCODING_HT)
	return assertDict(subject).dictAdd(value, nil)
}

func setTypeConvert(setObj *SRobj, enc uint8) {
	if setObj.Typ != SR_SET || setObj.encoding != REDIS_ENCODING_INTSET {
		utils.ErrorF("setTypeConvert err: setObj.Typ = %d,setObj.encoding = %d", setObj.Typ, setObj.encoding)
	}
	if enc != REDIS_ENCODING_HT {
		panic("Unsupported set conversion")
	}

	d := dictCreate(&setDictType)
	d.dictExpand(int64(assertIntSet(setObj).intSetLen()))
	si := setTypeInitIterator(setObj)
	var intEle int64
	for si.setTypeNext(nil, &intEle) != -1 {
		element := createFromInt(intEle)
		d.dictAdd(element, nil)
	}

	si.setTypeReleaseIterator()

	setObj.encoding = REDIS_ENCODING_HT
	setObj.Val = d
}

func setTypeSize(setObj *SRobj) int64 {
	checkSetEncoding(setObj)
	if setObj.encoding == REDIS_ENCODING_HT {
		return assertDict(setObj).dictSize()
	}
	// REDIS_ENCODING_INTSET
	return int64(assertIntSet(setObj).intSetLen())
}

func setTypeIsMember(setObj, value *SRobj) bool {
	var intVal int64
	checkSetEncoding(setObj)
	if setObj.encoding == REDIS_ENCODING_HT {
		_, e := assertDict(setObj).dictFind(value)
		return e != nil
	}
	if value.isObjectRepresentableAsInt64(&intVal) == REDIS_OK {
		return assertIntSet(setObj).intSetFind(intVal)
	}
	return false
}

// qSortSet use c qsort function
// has err: cgo argument has Go pointer to Go pointer
// need set env: export GODEBUG=cgocheck=0
func qSortSet(set []*SRobj) {
	qsort.Slice(set, func(a, b int) bool {
		return setTypeSize(set[a]) < setTypeSize(set[b])
	})
}

func sortSet(set []*SRobj) {
	sort.Slice(set, func(a, b int) bool {
		return setTypeSize(set[a]) < setTypeSize(set[b])
	})
}

func checkSetEncoding(subject *SRobj) {
	if subject.encoding != REDIS_ENCODING_INTSET && subject.encoding != REDIS_ENCODING_HT {
		panic("Unknown set encoding")
	}
}
