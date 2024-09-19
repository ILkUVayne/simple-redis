package src

import (
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"simple-redis/cgo/qsort"
	"sort"
)

//-----------------------------------------------------------------------------
// Set commands API
//-----------------------------------------------------------------------------

type setTypeIterator struct {
	subject  *SRobj
	ii       int64
	encoding uint8
	di       *dictIterator
}

// return -1 if next entry is nil
func (si *setTypeIterator) setTypeNext(objEle **SRobj, llEle *int64) int {
	if si.encoding == REDIS_ENCODING_INTSET {
		if !assertIntSet(si.subject).intSetGet(si.ii, llEle) {
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

// release set iterator
func (si *setTypeIterator) setTypeReleaseIterator() {
	if si.encoding == REDIS_ENCODING_HT {
		si.di.dictReleaseIterator()
	}
	si.di = nil
	si.subject = nil
}

// return a setTypeIterator
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

// create set obj.
//
// return intset obj if value is int.
// return dict obj if value is not int.
func setTypeCreate(value *SRobj) *SRobj {
	if value.isObjectRepresentableAsInt64(nil) == nil {
		return createIntSetObject()
	}
	return createSetObject()
}

// 往集合中添加数据，当底层结构是intset但需添加的数据不是数字时，会先转换为哈希表并存储
func setTypeAdd(subject, value *SRobj) bool {
	checkSetEncoding(subject)
	var intVal int64
	// hashtable
	if subject.encoding == REDIS_ENCODING_HT {
		return assertDict(subject).dictAdd(value, nil)
	}
	// intSet
	if value.isObjectRepresentableAsInt64(&intVal) == nil {
		var success bool
		assertIntSet(subject).intSetAdd(intVal, &success)
		return success
	}
	// change to ht
	setTypeConvert(subject, REDIS_ENCODING_HT)
	return assertDict(subject).dictAdd(value, nil)
}

// 转换intset为哈希表并且数据迁移
func setTypeConvert(setObj *SRobj, enc uint8) {
	if setObj.Typ != SR_SET || setObj.encoding != REDIS_ENCODING_INTSET {
		ulog.ErrorF("setTypeConvert err: setObj.Typ = %d,setObj.encoding = %d", setObj.Typ, setObj.encoding)
	}
	if enc != REDIS_ENCODING_HT {
		panic("Unsupported set conversion")
	}

	d := dictCreate(&setDictType)
	d.dictExpand(sLen(assertIntSet(setObj)))
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

// return set length
func setTypeSize(setObj *SRobj) int64 {
	checkSetEncoding(setObj)
	if setObj.encoding == REDIS_ENCODING_HT {
		return sLen(assertDict(setObj))
	}
	// REDIS_ENCODING_INTSET
	return sLen(assertIntSet(setObj))
}

// 验证value是否是集合的成员
func setTypeIsMember(setObj, value *SRobj) bool {
	var intVal int64
	checkSetEncoding(setObj)
	if setObj.encoding == REDIS_ENCODING_HT {
		_, e := assertDict(setObj).dictFind(value)
		return e != nil
	}
	if value.isObjectRepresentableAsInt64(&intVal) == nil {
		return assertIntSet(setObj).intSetFind(intVal)
	}
	return false
}

// 获取一个随机的集合（set）元素
func setTypeRandomElement(setObj *SRobj) (encoding uint8, objEle *SRobj, intEle int64) {
	checkSetEncoding(setObj)
	encoding = setObj.encoding
	if encoding == REDIS_ENCODING_INTSET {
		return encoding, nil, assertIntSet(setObj).intSetRandom()
	}
	// hash table
	objEle = assertDict(setObj).dictGetRandomKey().getKey()
	return encoding, objEle, -123456789
}

// 删除集合元素，value 为需要被删除的元素
func setTypeRemove(setObj, value *SRobj) bool {
	checkSetEncoding(setObj)
	if setObj.encoding == REDIS_ENCODING_HT {
		d := assertDict(setObj)
		if d.dictDelete(value) == REDIS_OK {
			if d.htNeedResize() {
				d.dictResize()
			}
			return true
		}
		return false
	}
	// intSet
	var intVal int64
	if value.isObjectRepresentableAsInt64(&intVal) == nil {
		return assertIntSet(setObj).intSetRemove(intVal)
	}
	return false
}

// qSortSet use c qsort function
//
// has error: cgo argument has Go pointer to Go pointer
//
// need set env: export GODEBUG=cgocheck=0
func qSortSet(sets []*SRobj) {
	qsort.Slice(sets, func(a, b int) bool {
		return setTypeSize(sets[a]) < setTypeSize(sets[b])
	})
}

// 集合切片容量从小到大排序
func sortSet(sets []*SRobj) {
	sort.Slice(sets, func(a, b int) bool {
		return setTypeSize(sets[a]) < setTypeSize(sets[b])
	})
}

func checkSetEncoding(subject *SRobj) {
	if subject.encoding != REDIS_ENCODING_INTSET && subject.encoding != REDIS_ENCODING_HT {
		panic("Unknown set encoding")
	}
}
