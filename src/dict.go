package src

import (
	"hash/fnv"
	"simple-redis/utils"
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

func dictCreate(dType *dictType) *dict {
	d := new(dict)
	d.dType = dType
	d.ht[0] = new(dictht)
	d.rehashIdx = -1
	return d
}
