package src

import (
	"github.com/ILkUVayne/utlis-go/v2/time"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
)

// SRedisDB 数据库结构
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

// del SRedisDB.data by key
func (db *SRedisDB) dictDel(key *SRobj) int {
	return db.data.dictDelete(key)
}

// del SRedisDB.expire by key
func (db *SRedisDB) expireDel(key *SRobj) int {
	return db.expire.dictDelete(key)
}

// del SRedisDB.data and SRedisDB.expire by key if exist
func (db *SRedisDB) dbDel(key *SRobj) int {
	// 重新创建一个新的key，如果直接用传入的key是expire库的key
	// 删除expire后会被提前释放(s.refCount == 0),导致dictDel报错
	key = createSRobj(SR_STR, key.strVal())
	if !isEmpty(db.expire) {
		db.expireDel(key)
	}
	return db.dictDel(key)
}

// get SRedisDB.data by key
func (db *SRedisDB) dictGet(key *SRobj) *SRobj {
	return db.data.dictGet(key)
}

// get SRedisDB.expire by key
func (db *SRedisDB) expireGet(key *SRobj) *SRobj {
	return db.expire.dictGet(key)
}

// Return the expireTime of the specified key, or -1 if no expire is associated with this key
func (db *SRedisDB) expireTime(key *SRobj) int64 {
	expire := db.expireGet(key)
	if expire == nil {
		return -1
	}
	t, _ := expire.intVal()
	return t
}

// set SRedisDB.data
func (db *SRedisDB) dictSet(key *SRobj, val *SRobj) {
	server.db.data.dictSet(key, val)
}

// set SRedisDB.expire
func (db *SRedisDB) expireSet(key *SRobj, val *SRobj) {
	server.db.expire.dictSet(key, val)
}

// 检查是否过期，如果过期了，就删除
func (db *SRedisDB) expireIfNeeded(key *SRobj) bool {
	e := db.expireGet(key)
	if e == nil {
		return false
	}

	when, _ := e.intVal()
	return db.expireIfNeeded1(when, key)
}

// 检查是否过期，如果过期了，就删除
//
// when 表示key的过期时间戳
func (db *SRedisDB) expireIfNeeded1(when int64, key *SRobj) bool {
	if when > time.GetMsTime() {
		return false
	}
	db.dbDel(key)
	return true
}

// check if it is expired and return SRedisDB.data by key, return nil if it is expired or not exists.
func (db *SRedisDB) lookupKeyWrite(key *SRobj) *SRobj {
	db.expireIfNeeded(key)
	return db.dictGet(key)
}

// check if it is expired and return SRedisDB.data by key, return nil if it is expired or not exists.
func (db *SRedisDB) lookupKeyRead(key *SRobj) *SRobj {
	db.expireIfNeeded(key)
	return db.dictGet(key)
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

// 获取一个随机数据
func (db *SRedisDB) dataRandomKey() *dictEntry {
	return db.data.dictGetRandomKey()
}

// 获取一个有过期时间的随机数据
func (db *SRedisDB) expireRandomKey() *dictEntry {
	return db.expire.dictGetRandomKey()
}

func (db *SRedisDB) dbRandomKey() *SRobj {
	for {
		de := db.dataRandomKey()
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

// return SRedisDB.expire size
func (db *SRedisDB) dbExpireSize() int64 {
	return sLen(db.expire)
}

// return SRedisDB.data size
func (db *SRedisDB) dbDataSize() int64 {
	return sLen(db.data)
}

// 获取一个数据库迭代器（dictIterator）
func (db *SRedisDB) dbDataDi() *dictIterator {
	return server.db.data.dictGetIterator()
}

// 尝试执行一步rehash（如果当前数据库正在rehash）
func tryRehash() {
	server.db.data.dictRehashStep()
}

// 尝试缩容，如果需要的话
func tryResizeHashTables() {
	if server.db.data.htNeedResize() {
		server.db.data.dictResize()
	}
	if server.db.expire.htNeedResize() {
		server.db.expire.dictResize()
	}
}

//-----------------------------------------------------------------------------
// db commands
//-----------------------------------------------------------------------------

func ttlGenericCommand(c *SRedisClient, outputMs bool) {
	key := c.args[1]
	if c.db.lookupKeyRead(key) == nil {
		c.addReplyLongLong(-2)
		return
	}
	expireTime := c.db.expireTime(key)
	if expireTime == -1 {
		c.addReplyLongLong(-1)
		return
	}
	ttl := expireTime - time.GetMsTime()
	if outputMs {
		c.addReplyLongLong(ttl)
		return
	}
	c.addReplyLongLong((ttl + 500) / 1000)
}

// expire key value
func expireCommand(c *SRedisClient) {
	key := c.args[1]
	val := c.args[2]
	if !val.checkType(c, SR_STR) {
		return
	}

	eval, err := val.intVal()
	if err != nil {
		ulog.ErrorP(err)
		c.addReply(shared.syntaxErr)
		return
	}

	if c.db.lookupKeyReadOrReply(c, key, nil) == nil {
		return
	}

	expire := eval
	if eval < MAX_EXPIRE {
		expire = time.GetMsTime() + (eval * 1000)
	}

	expireObj := createFromInt(expire)
	c.db.expireSet(key, expireObj)
	expireObj.decrRefCount()
	c.addReply(shared.ok)
	server.incrDirtyCount(c, 1)
}

// object encoding key
func objectCommand(c *SRedisClient) {
	val := c.args[2]
	if !val.checkType(c, SR_STR) {
		return
	}
	value := c.db.lookupKeyReadOrReply(c, val, nil)
	if value != nil {
		c.addReplyBulk(value.getEncoding())
	}
}

// TYPE key
func typeCommand(c *SRedisClient) {
	val := c.args[1]
	if !val.checkType(c, SR_STR) {
		return
	}
	value := c.db.lookupKeyReadOrReply(c, val, shared.none)
	if value != nil {
		c.addReplyStatus(value.strType())
	}
}

// del key [key ...]
func delCommand(c *SRedisClient) {
	deleted := 0
	for i := 1; i < len(c.args); i++ {
		if c.db.dbDel(c.args[i]) == REDIS_OK {
			deleted++
		}
	}
	c.addReplyLongLong(int64(deleted))
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
		if allKeys || StringMatchLen(pattern, key.strVal(), false) {
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
		if c.db.lookupKeyRead(c.args[i]) != nil {
			count++
		}
	}
	c.addReplyLongLong(int64(count))
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

// FLUSHDB
func flushDbCommand(c *SRedisClient) {
	server.incrDirtyCount(c, server.db.dbDataSize())
	server.db.data.dictEmpty()
	server.db.expire.dictEmpty()
	c.addReply(shared.ok)
}

// dbsize
func dbSizeCommand(c *SRedisClient) {
	c.addReplyLongLong(c.db.dbDataSize())
}
