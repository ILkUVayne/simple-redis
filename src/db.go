package src

import "simple-redis/utils"

type SRedisDB struct {
	data   *dict // data dict
	expire *dict // expire dict
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

// Return the expire time of the specified key, or -1 if no expire is associated with this key
func (db *SRedisDB) expireTime(key *SRobj) int64 {
	expire := db.expireGet(key)
	if expire == nil {
		return -1
	}
	t, _ := expire.intVal()
	return t
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

// return db value,if null while reply error to client
func (db *SRedisDB) lookupKeyReadOrReply(c *SRedisClient, key *SRobj, reply *SRobj) *SRobj {
	o := db.lookupKeyRead(key)
	if o == nil {
		if reply != nil {
			c.addReply(reply)
		} else {
			c.addReply(shared.nullBulk)
		}
	}
	return o
}

func (db *SRedisDB) dbRandomKey() *SRobj {
	for {
		de := db.data.dictGetRandomKey()
		if de == nil {
			return nil
		}
		keyObj := de.getKey()
		if db.expireIfNeeded(keyObj) {
			continue
		}
		return keyObj
	}
}

func (db *SRedisDB) dbDataSize() int64 {
	return db.data.dictSize()
}
