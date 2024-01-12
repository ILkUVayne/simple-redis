package src

import "simple-redis/utils"

type SRedisDB struct {
	data   *dict
	expire *dict
}

var dbDictType = dictType{
	hashFunc:   SRStrHash,
	keyCompare: SRStrCompare,
}

var keyPtrDictType = dictType{
	hashFunc:   SRStrHash,
	keyCompare: SRStrCompare,
}

func (db *SRedisDB) dictDel(key *SRobj) int {
	return db.data.dictDelete(key)
}

func (db *SRedisDB) expireDel(key *SRobj) int {
	return db.expire.dictDelete(key)
}

func (db *SRedisDB) dbDel(key *SRobj) int {
	if db.expire.dictSize() > 0 {
		db.expireDel(key)
	}
	return db.dictDel(key)
}

func (db *SRedisDB) dictGet(key *SRobj) *SRobj {
	return db.data.dictGet(key)
}

func (db *SRedisDB) expireGet(key *SRobj) *SRobj {
	return db.expire.dictGet(key)
}

func (db *SRedisDB) dictSet(key *SRobj, val *SRobj) {
	server.db.data.dictSet(key, val)
}

func (db *SRedisDB) expireIfNeeded(key *SRobj) bool {
	e := db.expireGet(key)
	if e == nil {
		return false
	}

	intVal, _ := e.intVal()
	if when := intVal; when > utils.GetMsTime() {
		return false
	}
	db.expireDel(key)
	db.dictDel(key)
	return true
}

func (db *SRedisDB) lookupKey(key *SRobj) *SRobj {
	return db.dictGet(key)
}

func (db *SRedisDB) lookupKeyWrite(key *SRobj) *SRobj {
	db.expireIfNeeded(key)
	return db.lookupKey(key)
}

func (db *SRedisDB) lookupKeyRead(key *SRobj) *SRobj {
	db.expireIfNeeded(key)
	return db.lookupKey(key)
}

func (db *SRedisDB) lookupKeyReadOrReply(c *SRedisClient, key *SRobj, reply *SRobj) *SRobj {
	o := db.lookupKeyRead(key)
	if o == nil {
		c.addReply(shared.nullBulk)
	}
	return o
}
