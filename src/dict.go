package src

import (
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"hash/fnv"
	"math"
	"math/bits"
	"math/rand"
)

// dict是否能执行缩容
//
// 默认true，当在进行BGREWRITEAOF或者BGSAVE时会被禁用
var dictCanResize = true

type dictIterator struct {
	d                *dict
	table, index     int
	entry, nextEntry *dictEntry
}

// return next dictEntry by dictIterator
func (di *dictIterator) dictNext() *dictEntry {
	for {
		if di.entry != nil {
			di.entry = di.nextEntry
		} else {
			ht := di.d.ht[di.table]
			if di.index == -1 && di.table == 0 {
				di.d.iterators++
			}
			di.index++
			if di.index >= int(ht.size) {
				if di.d.isRehash() && di.table == 0 {
					di.table++
					di.index = 0
					ht = di.d.ht[1]
				} else {
					break
				}
			}
			di.entry = ht.table[di.index]
		}
		if di.entry != nil {
			di.nextEntry = di.entry.next
			return di.entry
		}
	}
	return nil
}

// Release dict Iterator
func (di *dictIterator) dictReleaseIterator() {
	if !(di.index == -1 && di.table == 0) {
		di.d.iterators--
	}
	di.d = nil
	di.entry = nil
	di.nextEntry = nil
}

// dictEntry 结构体
type dictEntry struct {
	key  *SRobj
	val  *SRobj
	next *dictEntry // next dictEntry 用于处理哈希冲突（头插法）
}

// return dictEntry.key
func (de *dictEntry) getKey() *SRobj {
	return de.key
}

// return dictEntry.val
func (de *dictEntry) getVal() *SRobj {
	return de.val
}

// dictType 声明了一些通用的函数
type dictType struct {
	hashFunc      func(key *SRobj) int64       // 哈希函数
	keyCompare    func(key1, key2 *SRobj) bool // 关键字比较函数
	keyDestructor func(key *SRobj)             // 关键字销毁（释放）函数
	valDestructor func(val *SRobj)             // 值销毁（释放）函数
}

type dictScanFunction func(priVData any, de *dictEntry)

// dictht 哈希表结构
type dictht struct {
	table []*dictEntry // 哈希表
	size  int64        // 哈希表容量（因为存在哈希冲突，在还没有触发rehash时，实际存储数据量可能会超过）
	used  int64        // 已存储的数据量
	mask  int64        // 哈希表大小掩码，用于计算关键字索引，等于size-1
}

// dict 字典结构
type dict struct {
	dType     *dictType
	ht        [2]*dictht // 哈希表，ht[1]用于rehash
	rehashIdx int64      // rehash索引，默认-1，表示当前未进行rehash
	// iterators
	iterators int64
}

// ----------------------------- dict type func -------------------------

// SRStrHash 计算哈希值函数
func SRStrHash(key *SRobj) int64 {
	if key.Typ != SR_STR {
		return 0
	}
	hash := fnv.New64()
	_, err := hash.Write([]byte(key.strVal()))
	if err != nil {
		ulog.Error("simple-redis server: hashFunc err: ", err)
	}
	return int64(hash.Sum64())
}

// SRStrCompare 键比较函数
func SRStrCompare(key1, key2 *SRobj) bool {
	if key1.Typ != SR_STR || key2.Typ != SR_STR {
		return false
	}
	return key1.strVal() == key2.strVal()
}

// SRStrDestructor 键销毁函数
func SRStrDestructor(key *SRobj) {
	key.decrRefCount()
}

// ObjectDestructor 值销毁函数
func ObjectDestructor(val *SRobj) {
	if val != nil {
		val.decrRefCount()
	}
}

// -------------------------------- api ----------------------------

// free dict key by dictEntry
func (d *dict) dictFreeKey(de *dictEntry) {
	if d.dType.keyDestructor != nil {
		d.dType.keyDestructor(de.getKey())
	}
}

// free dict val by dictEntry
func (d *dict) dictFreeVal(de *dictEntry) {
	if d.dType.valDestructor != nil {
		d.dType.valDestructor(de.getVal())
	}
}

// free dict key and val by dictEntry
func (d *dict) dictFreeEntry(e *dictEntry) {
	d.dictFreeKey(e)
	d.dictFreeVal(e)
}

// return dict current size
func (d *dict) cap() int64 {
	s := d.ht[0].size
	if d.ht[1] != nil {
		s += d.ht[1].size
	}
	return s
}

// return dict current used
func (d *dict) len() int64 {
	s := d.ht[0].used
	if d.ht[1] != nil {
		s += d.ht[1].used
	}
	return s
}

// check if it is Empty
func (d *dict) isEmpty() bool {
	return sLen(d) == 0
}

// return dict iterators
func (d *dict) dictGetIterator() *dictIterator {
	return &dictIterator{d: d, index: -1}
}

// init or reset dict.ht
func (d *dict) initHt() {
	d.ht[0] = &dictht{
		mask:  DICT_HT_INITIAL_SIZE - 1,
		size:  DICT_HT_INITIAL_SIZE,
		table: make([]*dictEntry, DICT_HT_INITIAL_SIZE),
	}
	d.rehashIdx = -1
	d.iterators = 0
}

func (d *dict) dictScan(v uint64, fn dictScanFunction, priVData any) int64 {
	if d.isEmpty() {
		return 0
	}

	if !d.isRehash() {
		t := d.ht[0]
		m := uint64(t.mask)

		de := t.table[v&m]
		for de != nil {
			fn(priVData, de)
			de = de.next
		}

		v |= ^m

		v = bits.Reverse64(v)
		v++
		v = bits.Reverse64(v)
		return int64(v)
	}

	t0, t1 := d.ht[0], d.ht[1]

	if t0.size > t1.size {
		t0, t1 = t1, t0
	}

	m0, m1 := uint64(t0.mask), uint64(t1.mask)

	de := t0.table[v&m0]
	for de != nil {
		fn(priVData, de)
		de = de.next
	}

	for {
		de = t1.table[v&m1]
		for de != nil {
			fn(priVData, de)
			de = de.next
		}

		v |= ^m1
		v = bits.Reverse64(v)
		v++
		v = bits.Reverse64(v)

		if v&(m0^m1) == 0 {
			break
		}
	}
	return int64(v)
}

// check if the current rehash is in progress
func (d *dict) isRehash() bool {
	return d.rehashIdx != -1
}

// rehash n step
func (d *dict) dictRehash(step int) {
	if !d.isRehash() {
		return
	}
	for ; step > 0; step-- {
		// Check if we already rehashed the whole table
		if d.ht[0].used == 0 {
			d.ht[0] = d.ht[1]
			d.ht[1] = nil
			d.rehashIdx = -1
			return
		}
		// find a not nil elem
		rehashNoNullStep := server.rehashNullStep
		if rehashNoNullStep > (d.ht[0].size - d.rehashIdx) {
			rehashNoNullStep = d.ht[0].size - d.rehashIdx
		}
		for ; rehashNoNullStep > 0; rehashNoNullStep-- {
			if d.ht[0].table[d.rehashIdx] == nil {
				d.rehashIdx++
			}
		}
		entry := d.ht[0].table[d.rehashIdx]
		// cannot find a not nil elem
		if entry == nil {
			return
		}
		for entry != nil {
			nextEntry := entry.next
			h := d.dType.hashFunc(entry.key) & d.ht[1].mask
			entry.next = d.ht[1].table[h]
			d.ht[1].table[h] = entry
			d.ht[0].used--
			d.ht[1].used++
			entry = nextEntry
		}
		d.ht[0].table[d.rehashIdx] = nil
		d.rehashIdx++
	}
}

// rehash DEFAULT_REHASH_STEP step
func (d *dict) dictRehashStep() {
	d.dictRehash(DEFAULT_REHASH_STEP)
}

// return dict expand size
func (d *dict) dictNextPower(size int64) int64 {
	i := DICT_HT_INITIAL_SIZE
	if size > math.MaxInt64 {
		return math.MaxInt64
	}
	for {
		if i >= size {
			return i
		}
		i *= EXPEND_RATIO
	}
}

// 判断是否需要调整dict容量
//
// used使用量占比小于10%时，需要调整dict容量
func (d *dict) htNeedResize() bool {
	size, used := sCap(d), sLen(d)
	return size > DICT_HT_INITIAL_SIZE && (used*100/size < HT_MIN_FILL)
}

// 调整dict容量
func (d *dict) dictResize() int {
	if !dictCanResize || d.isRehash() {
		return DICT_ERR
	}
	minimal := sLen(d)
	if minimal < DICT_HT_INITIAL_SIZE {
		minimal = DICT_HT_INITIAL_SIZE
	}
	return d.dictExpand(minimal)
}

// 扩容
func (d *dict) dictExpand(size int64) int {
	realSize := d.dictNextPower(size)

	if d.isRehash() || d.ht[0].used > size {
		return DICT_ERR
	}

	ht := new(dictht)
	ht.used = 0
	ht.size = realSize
	ht.mask = realSize - 1
	ht.table = make([]*dictEntry, realSize)

	d.ht[1] = ht
	d.rehashIdx = 0
	return DICT_OK
}

// 检查是否需要扩容
func (d *dict) dictExpandIfNeeded() int {
	if d.isRehash() {
		return DICT_OK
	}
	if d.ht[0].used > d.ht[0].size && (dictCanResize || float64(d.ht[0].used)/float64(d.ht[0].size) > float64(server.loadFactor)) {
		return d.dictExpand(d.ht[0].size * EXPEND_RATIO)
	}
	return DICT_OK
}

// return -1 if err or exist
func (d *dict) dictKeyIndex(key *SRobj) int64 {
	if err := d.dictExpandIfNeeded(); err != DICT_OK {
		return -1
	}
	idx, _ := d.dictFind(key)
	return idx
}

// find index and val
func (d *dict) dictFind(key *SRobj) (int64, *dictEntry) {
	if d.ht[0].size == 0 {
		return -1, nil
	}
	if d.isRehash() {
		d.dictRehashStep()
	}
	h := d.dType.hashFunc(key)
	var idx int64
	for table := 0; table <= 1; table++ {
		idx = h & d.ht[table].mask
		he := d.ht[table].table[idx]
		for he != nil {
			if d.dType.keyCompare(key, he.key) {
				return -1, he
			}
			he = he.next
		}
		if !d.isRehash() {
			break
		}
	}
	return idx, nil
}

// add key to dict,return dictEntry
func (d *dict) dictAddRaw(key *SRobj) *dictEntry {
	// maybe after flushdb
	if d.ht[0].size == 0 {
		d.initHt()
	}
	idx := d.dictKeyIndex(key)
	if idx == -1 {
		return nil
	}
	ht := d.ht[0]
	if d.isRehash() {
		ht = d.ht[1]
	}
	entry := dictEntry{key: key, next: ht.table[idx]}
	key.incrRefCount()
	ht.table[idx] = &entry
	ht.used++
	return &entry
}

// dict add by key and val
func (d *dict) dictAdd(key, val *SRobj) bool {
	entry := d.dictAddRaw(key)
	if entry == nil {
		return false
	}
	entry.val = val
	if val != nil {
		val.incrRefCount()
	}
	return true
}

// return DICT_SET if new key, else DICT_REP if replacer
func (d *dict) dictSet(key, val *SRobj) int {
	if d.dictAdd(key, val) {
		return DICT_SET
	}
	_, entry := d.dictFind(key)
	if entry.val != nil {
		entry.val.decrRefCount()
	}
	entry.val = val
	if entry.val != nil {
		entry.val.incrRefCount()
	}
	return DICT_REP
}

// dict get val by key
func (d *dict) dictGet(key *SRobj) *SRobj {
	_, entry := d.dictFind(key)
	if entry == nil {
		return nil
	}
	return entry.val
}

// dict del by key
func (d *dict) dictDelete(key *SRobj) int {
	if d.ht[0].size == 0 {
		return DICT_ERR
	}

	if d.isRehash() {
		d.dictRehashStep()
	}

	h := d.dType.hashFunc(key)
	var idx int64
	for table := 0; table <= 1; table++ {
		idx = h & d.ht[table].mask
		he := d.ht[table].table[idx]
		var preHe *dictEntry
		for he != nil {
			if d.dType.keyCompare(key, he.key) {
				if preHe == nil {
					d.ht[table].table[idx] = he.next
				} else {
					preHe.next = he.next
				}
				d.dictFreeEntry(he)
				d.ht[table].used--
				return DICT_OK
			}
			preHe = he
			he = he.next
		}
		if !d.isRehash() {
			break
		}
	}
	return DICT_ERR
}

// get a non-empty bucket
func (d *dict) dictGetRandomKey1() *dictEntry {
	var he *dictEntry
	var slotIdx int64
	if d.isRehash() {
		for he == nil {
			slotIdx = rand.Int63n(sCap(d))
			if slotIdx >= d.ht[0].size {
				he = d.ht[1].table[slotIdx-d.ht[0].size]
				continue
			}
			he = d.ht[0].table[slotIdx]
		}
		return he
	}
	for he == nil {
		slotIdx = rand.Int63n(d.ht[0].size)
		he = d.ht[0].table[slotIdx]
	}
	return he
}

// get a random key
func (d *dict) dictGetRandomKey() *dictEntry {
	if isEmpty(d) {
		return nil
	}
	if d.isRehash() {
		d.dictRehashStep()
	}
	// find a non-empty bucket
	he := d.dictGetRandomKey1()
	// get a random key from bucket
	listLen, origHe := int64(0), he
	for he != nil {
		he = he.next
		listLen++
	}
	listEle := rand.Int63n(listLen)
	he = origHe
	for ; listEle < 0; listEle-- {
		he = he.next
	}
	return he
}

// 清空哈希表数据
func (d *dict) _dictClear(ht *dictht) int {
	var he *dictEntry

	if ht == nil {
		return DICT_OK
	}
	for i := int64(0); i < ht.size; i++ {
		he = ht.table[i]
		if he == nil {
			continue
		}
		for he != nil {
			nextHe := he.next
			d.dictFreeEntry(he)
			ht.used--
			he = nextHe
		}
	}
	// 重置哈希表
	ht.table = nil
	ht.size = 0
	ht.used = 0
	ht.mask = 0
	return DICT_OK
}

// 清空数据库（字典）中的数据
func (d *dict) dictEmpty() {
	d._dictClear(d.ht[0])
	d._dictClear(d.ht[1])
	d.rehashIdx = -1
	d.iterators = 0
}

// return new dict
func dictCreate(dType *dictType) *dict {
	d := &dict{dType: dType}
	d.initHt()
	return d
}

// 启用数据库缩容
func dictEnableResize() {
	dictCanResize = true
}

// 禁用数据库缩容
func dictDisableResize() {
	dictCanResize = false
}
