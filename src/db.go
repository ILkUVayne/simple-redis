package src

import "simple-redis/utils"

type SRedisDB struct {
	data   *dict // data dict
	expire *dict // expire dict
}

// Db->dict
var dbDictType = dictType{
	hashFunc:      SRStrHash,
	keyCompare:    SRStrCompare,
	keyDestructor: SRStrDestructor,
	valDestructor: ObjectDestructor,
}

// Db->expires
var keyPtrDictType = dictType{
	hashFunc:      SRStrHash,
	keyCompare:    SRStrCompare,
	keyDestructor: SRStrDestructor,
	valDestructor: ObjectDestructor,
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

//-----------------------------------------------------------------------------
// db commands
//-----------------------------------------------------------------------------

func ttlGenericCommand(c *SRedisClient, outputMs bool) {
	key := c.args[1]
	c.db.expireIfNeeded(key)
	if c.db.lookupKey(key) == nil {
		c.addReplyLongLong(-2)
		return
	}
	expireTime := c.db.expireTime(key)
	if expireTime == -1 {
		c.addReplyLongLong(-1)
		return
	}
	ttl := expireTime - utils.GetMsTime()
	if outputMs {
		c.addReplyLongLong(int(ttl))
		return
	}
	c.addReplyLongLong(int((ttl + 500) / 1000))
}

// expire key value
func expireCommand(c *SRedisClient) {
	key := c.args[1]
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReply(shared.typeErr)
		return
	}

	eval, res := val.intVal()
	if res == REDIS_ERR {
		c.addReply(shared.syntaxErr)
		return
	}

	if c.db.lookupKeyReadOrReply(c, key, nil) == nil {
		return
	}

	expire := eval
	if eval < MAX_EXPIRE {
		expire = utils.GetMsTime() + (eval * 1000)
	}

	expireObj := createFromInt(expire)
	c.db.expire.dictSet(key, expireObj)
	expireObj.decrRefCount()
	c.addReply(shared.ok)
	server.incrDirtyCount(c, 1)
}

// object encoding key
func objectCommand(c *SRedisClient) {
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReply(shared.typeErr)
		return
	}
	value := c.db.lookupKeyReadOrReply(c, val, nil)
	if value == nil {
		return
	}
	c.addReplyBulk(value.getEncoding())
}

// del key [key ...]
func delCommand(c *SRedisClient) {
	deleted := 0
	for i := 1; i < len(c.args); i++ {
		if c.db.dbDel(c.args[i]) == REDIS_OK {
			deleted++
		}
	}
	c.addReplyLongLong(deleted)
}

// keys pattern
func keysCommand(c *SRedisClient) {
	pattern := c.args[1].strVal()
	numKeys := 0
	allKeys := false
	if pattern[0] == '*' && len(pattern) == 1 {
		allKeys = true
	}
	replyLen := c.addDeferredMultiBulkLength()
	di := c.db.data.dictGetIterator()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		key := de.getKey()
		if allKeys || utils.StringMatch(pattern, key.strVal(), false) {
			if !c.db.expireIfNeeded(key) {
				c.addReplyBulk(key)
				numKeys++
			}
		}
	}
	di.dictReleaseIterator()
	c.setDeferredMultiBulkLength(replyLen, numKeys)
}

// EXISTS key [key ...]
func existsCommand(c *SRedisClient) {
	count := 0
	for i := 1; i < len(c.args); i++ {
		c.db.expireIfNeeded(c.args[i])
		if c.db.lookupKey(c.args[i]) != nil {
			count++
		}
	}
	c.addReplyLongLong(count)
}

// TTL key, return s
func ttlCommand(c *SRedisClient) {
	ttlGenericCommand(c, false)
}

// PTTL key, return ms
func pTtlCommand(c *SRedisClient) {
	ttlGenericCommand(c, true)
}

// PERSIST key
func persistCommand(c *SRedisClient) {
	key := c.args[1]
	c.db.expireIfNeeded(key)
	if c.db.expireGet(key) == nil {
		c.addReply(shared.czero)
		return
	}
	if c.db.expireDel(key) == REDIS_OK {
		c.addReply(shared.cone)
		server.incrDirtyCount(c, 1)
		return
	}
	c.addReply(shared.czero)
}

// RANDOMKEY
func randomKeyCommand(c *SRedisClient) {
	key := c.db.dbRandomKey()
	if key == nil {
		c.addReply(shared.nullBulk)
		return
	}
	c.addReplyBulk(key)
}

func flushDbCommand(c *SRedisClient) {
	server.incrDirtyCount(c, server.db.dbDataSize())
	server.db.data.dictEmpty()
	server.db.expire.dictEmpty()
	c.addReply(shared.ok)
}
