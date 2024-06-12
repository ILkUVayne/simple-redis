package src

import (
	"fmt"
	"testing"
)

func TestSRStrHash(t *testing.T) {
	str := createSRobj(SR_STR, "name")
	ints := createSRobj(SR_DICT, 1)
	h1 := SRStrHash(ints)
	if h1 != 0 {
		t.Error("SRStrHash err: h == ", h1)
	}
	if SRStrHash(str) != SRStrHash(str) {
		t.Error("SRStrHash err")
	}
}

func TestSRStrCompare(t *testing.T) {
	str1 := createSRobj(SR_STR, "name")
	str2 := createSRobj(SR_STR, "name1")
	res := SRStrCompare(str1, str2)
	if res {
		t.Error("SRStrCompare err: res == ", res)
	}
}

func TestDictCreate(t *testing.T) {
	d := dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare})
	if d.rehashIdx != -1 {
		t.Error("dict init rehashidx err: rehashidx == ", d.rehashIdx)
	}
	if d.ht[0].size != DICT_HT_INITIAL_SIZE {
		t.Error("dict init ht size err: size == ", d.ht[0].size)
	}
}

func TestFreeDictEntry(t *testing.T) {
	key := createSRobj(SR_STR, "name")
	val := createSRobj(SR_STR, "ly")
	e := new(dictEntry)
	e.key = key
	e.key.incrRefCount()
	e.val = val
	e.val.incrRefCount()
	if e.key.refCount != 2 || e.val.refCount != 2 {
		t.Errorf("incrRefCount err: key = %d, val = %d", e.key.refCount, e.val.refCount)
	}
	freeDictEntry(e)
	if e.key.refCount != 1 || e.val.refCount != 1 {
		t.Errorf("incrRefCount err: key = %d, val = %d", e.key.refCount, e.val.refCount)
	}
}

func TestIsRehash(t *testing.T) {
	d := dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare})
	if d.isRehash() {
		t.Error("isRehash err: d.isRehash == ", d.isRehash())
	}
	d.rehashIdx = 0
	if !d.isRehash() {
		t.Error("isRehash err: d.isRehash == ", d.isRehash())
	}
}

func TestDictRehash(t *testing.T) {
	server.rehashNullStep = 10
	server.loadFactor = LOAD_FACTOR
	d := dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare})
	d.dictSet(createSRobj(SR_STR, "name"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name1"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name2"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name3"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name4"), createSRobj(SR_STR, "ly"))
	if d.isRehash() {
		t.Error("isRehash err: d.isRehash == ", d.isRehash())
	}
	d.dictExpandIfNeeded()
	d.dictRehash(1)
	d.dictRehash(1)
	d.dictRehash(1)
	d.dictRehash(1)
	d.dictRehash(1)
	if d.isRehash() {
		t.Error("isRehash err: d.isRehash == ", d.isRehash())
	}
}

func TestDictNextPower(t *testing.T) {
	d := dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare})
	s1 := d.dictNextPower(5)
	if s1 != 8 {
		t.Error("dictNextPower err: size == ", s1)
	}
}

func TestDictResize(t *testing.T) {
	d := dictCreate(&dbDictType)
	d.dictSet(createSRobj(SR_STR, "name"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name1"), createSRobj(SR_STR, "ly"))
	d.dictExpand(5)
	d.dictGet(createSRobj(SR_STR, "name"))
	d.dictGet(createSRobj(SR_STR, "name"))
	if d.dictResize() != DICT_ERR {
		t.Error("dictResize err: isRehash now")
	}
	d.dictGet(createSRobj(SR_STR, "name"))
	if d.dictResize() == DICT_ERR {
		t.Error("dictResize err: result is false")
	}
	if d.ht[1].size != 4 {
		t.Error("dictExpand err: size == ", d.ht[1].size)
	}
}

func TestDictExpand(t *testing.T) {
	server.rehashNullStep = 10
	server.loadFactor = LOAD_FACTOR
	d := dictCreate(&dbDictType)
	d.dictSet(createSRobj(SR_STR, "name"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name1"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name2"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name3"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name4"), createSRobj(SR_STR, "ly"))
	d.dictExpand(5)
	if d.ht[1].size != 8 {
		t.Error("dictExpand err: size == ", d.ht[1].size)
	}
}

func TestDictKeyIndex(t *testing.T) {
	d := dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare})
	key := createSRobj(SR_STR, "name")
	val := createSRobj(SR_STR, "ly")
	d.dictSet(key, val)
	idx := d.dictKeyIndex(key)
	if idx != -1 {
		t.Errorf("dictKeyIndex err: idx = %d", idx)
	}
	idx = d.dictKeyIndex(createSRobj(SR_STR, "name1"))
	if idx == -1 {
		t.Errorf("dictKeyIndex err: idx = %d", idx)
	}
}

func TestDictFind(t *testing.T) {
	d := dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare})
	key := createSRobj(SR_STR, "name")
	val := createSRobj(SR_STR, "ly")
	d.dictSet(key, val)
	idx, e := d.dictFind(key)
	if idx != -1 {
		t.Errorf("dictFind err: idx = %d", idx)
	}
	if val.strVal() != e.val.strVal() {
		t.Errorf("dictFind err: val = %s", d.ht[0].table[idx].val.strVal())
	}
}

func TestDictSet(t *testing.T) {
	server.rehashNullStep = 10
	server.loadFactor = LOAD_FACTOR
	d := dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare})
	d.dictSet(createSRobj(SR_STR, "name"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name1"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name2"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name3"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name4"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name5"), createSRobj(SR_STR, "ly"))
	if !d.isRehash() {
		t.Error("dictSet err")
	}
	d.dictGet(createSRobj(SR_STR, "name"))
	d.dictGet(createSRobj(SR_STR, "name"))
	d.dictGet(createSRobj(SR_STR, "name"))
	d.dictGet(createSRobj(SR_STR, "name"))
	if d.isRehash() {
		t.Error("dictSet err")
	}
}

func TestDictGet(t *testing.T) {
	d := dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare})
	key := createSRobj(SR_STR, "name")
	val := createSRobj(SR_STR, "ly")
	d.dictSet(key, val)
	v := d.dictGet(key)
	if v.strVal() != val.strVal() {
		t.Error("dictGet err")
	}
	v = d.dictGet(createSRobj(SR_STR, "name1"))
	if v != nil {
		t.Error("dictGet err")
	}
}

func TestDictDelete(t *testing.T) {
	d := dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare})
	key := createSRobj(SR_STR, "name")
	val := createSRobj(SR_STR, "ly")
	d.dictSet(key, val)
	d.dictSet(createSRobj(SR_STR, "name1"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name2"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name3"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name4"), createSRobj(SR_STR, "ly"))
	d.dictSet(createSRobj(SR_STR, "name5"), createSRobj(SR_STR, "ly"))

	if d.dictGet(key).strVal() != val.strVal() {
		t.Error("dictGet err")
	}
	d.dictDelete(key)
	if d.dictGet(key) != nil {
		t.Error("dictDelete err")
	}
}

func TestDictGetRandomKey(t *testing.T) {
	d := dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare})
	d.dictSet(createSRobj(SR_STR, "name1"), createSRobj(SR_STR, "ly1"))
	d.dictSet(createSRobj(SR_STR, "name2"), createSRobj(SR_STR, "ly2"))
	d.dictSet(createSRobj(SR_STR, "name3"), createSRobj(SR_STR, "ly3"))
	d.dictSet(createSRobj(SR_STR, "name4"), createSRobj(SR_STR, "ly4"))
	d.dictSet(createSRobj(SR_STR, "name5"), createSRobj(SR_STR, "ly5"))
	d.dictSet(createSRobj(SR_STR, "name6"), createSRobj(SR_STR, "ly6"))
	d.dictSet(createSRobj(SR_STR, "name7"), createSRobj(SR_STR, "ly7"))
	e := d.dictGetRandomKey()
	if e == nil {
		t.Error("dictGetRandomKey err")
	}
	fmt.Println(e.val.strVal())
}
