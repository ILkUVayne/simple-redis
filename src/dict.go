package src

import (
	"hash/fnv"
	"simple-redis/utils"
)

const (
	DICK_OK                    = 0
	DICK_ERR                   = 1
	DEFAULT_REHASH_STEP        = 1
	DICT_HT_INITIAL_SIZE       = 4
	EXPEND_RATIO               = 2
	LOAD_FACTOR                = 1
	BG_PERSISTENCE_LOAD_FACTOR = 5
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

func (d *dict) isRehash() bool {
	return d.rehashIdx != -1
}

func (d *dict) dictRehash(step int) {
	// TODO
}

func (d *dict) dictRehashStep() {
	d.dictRehash(DEFAULT_REHASH_STEP)
}

func (d *dict) dictExpand(size int64) int {
	// TODO
	return DICK_OK
}

func (d *dict) dictExpandIfNeeded() int {
	if d.isRehash() {
		return DICK_OK
	}
	if d.ht[0].used > d.ht[0].size && (d.ht[0].used/d.ht[0].size > server.loadFactor) {
		return d.dictExpand(d.ht[0].size * EXPEND_RATIO)
	}
	return DICK_OK
}

// return -1 if err or exist
func (d *dict) dictKeyIndex(key *SRobj) int64 {
	if err := d.dictExpandIfNeeded(); err != DICK_OK {
		return -1
	}
	h := d.dType.hashFunc(key)
	var idx int64
	for table := 0; table < 1; table++ {
		idx = h & d.ht[table].mask
		he := d.ht[table].table[idx]
		for he != nil {
			if d.dType.keyCompare(key, he.key) {
				return -1
			}
			he = he.next
		}
		if d.isRehash() {
			break
		}
	}
	return idx
}

func (d *dict) dictAddRaw(key *SRobj) *dictEntry {
	if d.isRehash() {
		d.dictRehashStep()
	}

	var idx int64
	var entry *dictEntry
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
	ht.table[idx] = entry
	ht.used++
	return entry
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
