package src

import (
	"hash/fnv"
	"math"
	"simple-redis/utils"
)

// LOAD_FACTOR 负载因子
// BG_PERSISTENCE_LOAD_FACTOR bgsave或者bgrewriteaof 的负载因子
const (
	DICK_OK                    = 0
	DICK_ERR                   = 1
	DEFAULT_REHASH_STEP        = 1
	DICT_HT_INITIAL_SIZE int64 = 4
	EXPEND_RATIO         int64 = 2
	LOAD_FACTOR                = 1
	//BG_PERSISTENCE_LOAD_FACTOR       = 5
)

type dictEntry struct {
	key  *SRobj
	val  *SRobj
	next *dictEntry
}

type dictType struct {
	hashFunc   func(key *SRobj) int64
	keyCompare func(key1, key2 *SRobj) bool
}

type dictht struct {
	table []*dictEntry
	size  int64
	used  int64
	mask  int64
}

type dict struct {
	dType     *dictType
	ht        [2]*dictht
	rehashIdx int64
	// iterators
}

// ----------------------------- hash func -------------------------

func SRStrHash(key *SRobj) int64 {
	if key.Typ != SR_STR {
		return 0
	}
	hash := fnv.New64()
	_, err := hash.Write([]byte(key.strVal()))
	if err != nil {
		utils.Error("simple-redis server: hashFunc err: ", err)
	}
	return int64(hash.Sum64())
}

func SRStrCompare(key1, key2 *SRobj) bool {
	if key1.Typ != SR_STR || key2.Typ != SR_STR {
		return false
	}
	return key1.strVal() == key2.strVal()
}

// -------------------------------- api ----------------------------

func dictCreate(dType *dictType) *dict {
	d := new(dict)
	d.dType = dType
	ht := new(dictht)
	ht.mask = DICT_HT_INITIAL_SIZE - 1
	ht.size = DICT_HT_INITIAL_SIZE
	ht.used = 0
	ht.table = make([]*dictEntry, DICT_HT_INITIAL_SIZE)
	d.ht[0] = ht
	d.rehashIdx = -1
	return d
}

func freeDictEntry(e *dictEntry) {
	e.key.decrRefCount()
	e.val.decrRefCount()
}

func (d *dict) isRehash() bool {
	return d.rehashIdx != -1
}

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

func (d *dict) dictRehashStep() {
	d.dictRehash(DEFAULT_REHASH_STEP)
}

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

func (d *dict) dictExpand(size int64) int {
	realSize := d.dictNextPower(size)

	if d.isRehash() || d.ht[0].used > size {
		return DICK_ERR
	}

	ht := new(dictht)
	ht.used = 0
	ht.size = realSize
	ht.mask = realSize - 1
	ht.table = make([]*dictEntry, realSize)

	d.ht[1] = ht
	d.rehashIdx = 0
	return DICK_OK
}

func (d *dict) dictExpandIfNeeded() int {
	if d.isRehash() {
		return DICK_OK
	}
	if d.ht[0].used > d.ht[0].size && (float64(d.ht[0].used)/float64(d.ht[0].size) > float64(server.loadFactor)) {
		return d.dictExpand(d.ht[0].size * EXPEND_RATIO)
	}
	return DICK_OK
}

// return -1 if err or exist
func (d *dict) dictKeyIndex(key *SRobj) int64 {
	if err := d.dictExpandIfNeeded(); err != DICK_OK {
		return -1
	}
	idx, _ := d.dictFind(key)
	return idx
}

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

func (d *dict) dictAddRaw(key *SRobj) *dictEntry {
	var idx int64
	var entry dictEntry
	var ht *dictht
	if idx = d.dictKeyIndex(key); idx == -1 {
		return nil
	}
	ht = d.ht[0]
	if d.isRehash() {
		ht = d.ht[1]
	}
	entry.key = key
	entry.next = ht.table[idx]
	key.incrRefCount()
	ht.table[idx] = &entry
	ht.used++
	return &entry
}

func (d *dict) dictAdd(key, val *SRobj) int {
	entry := d.dictAddRaw(key)
	if entry == nil {
		return DICK_ERR
	}
	entry.val = val
	val.incrRefCount()
	return DICK_OK
}

func (d *dict) dictSet(key, val *SRobj) {
	if d.dictAdd(key, val) == DICK_OK {
		return
	}
	_, entry := d.dictFind(key)
	entry.val.decrRefCount()
	entry.val = val
	entry.val.incrRefCount()
}

func (d *dict) dictGet(key *SRobj) *SRobj {
	_, entry := d.dictFind(key)
	if entry == nil {
		return nil
	}
	return entry.val
}

func (d *dict) dictDelete(key *SRobj) int {
	if d.ht[0].size == 0 {
		return DICK_ERR
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
				freeDictEntry(he)
				d.ht[table].used--
				return DICK_OK
			}
			preHe = he
			he = he.next
		}
		if !d.isRehash() {
			break
		}
	}
	return DICK_ERR
}
