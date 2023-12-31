package src

import "simple-redis/utils"

type SRedisDB struct {
	data   *dict
	expire *dict
}

func (db *SRedisDB) dictDel(key *SRobj) {
	db.data.dictDelete(key)
}

func (db *SRedisDB) expireDel(key *SRobj) {
	db.expire.dictDelete(key)
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

func (db *SRedisDB) expireIfNeeded(key *SRobj) {
	e := db.expireGet(key)
	if e == nil {
		return
	}

	if when := e.intVal(); when > utils.GetMsTime() {
		return
	}
	db.expireDel(key)
	db.dictDel(key)
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
