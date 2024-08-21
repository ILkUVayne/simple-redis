package src

//-----------------------------------------------------------------------------
// Hash type commands
//-----------------------------------------------------------------------------

func genericHGetAllCommand(c *SRedisClient, flags int) {
	o := c.db.lookupKeyReadOrReply(c, c.args[1], shared.emptyMultiBulk)
	if o == nil || !o.checkType(c, SR_DICT) {
		return
	}
	multiplier := int64(0)
	if flags&OBJ_HASH_KEY == OBJ_HASH_KEY {
		multiplier++
	}
	if flags&OBJ_HASH_VALUE == OBJ_HASH_VALUE {
		multiplier++
	}
	length := hashTypeLength(o) * multiplier
	c.addReplyMultiBulkLen(length, false)

	hi := assertDict(o).dictGetIterator()
	for de := hi.dictNext(); de != nil; de = hi.dictNext() {
		if flags&OBJ_HASH_KEY == OBJ_HASH_KEY {
			c.addReplyBulk(de.getKey())
		}
		if flags&OBJ_HASH_VALUE == OBJ_HASH_VALUE {
			c.addReplyBulk(de.getVal())
		}
	}
}

// hset key field value
func hSetCommand(c *SRedisClient) {
	o := hashTypeLookupWriteOrCreate(c, c.args[1])
	if o == nil {
		return
	}
	hashTypeTryObjectEncoding(o, c.args[2], c.args[3])
	update := hashTypeSet(o, c.args[2], c.args[3])
	server.incrDirtyCount(c, 1)
	if update == DICT_SET {
		c.addReply(shared.cone)
		return
	}
	c.addReply(shared.czero)
}

// hget key field
func hGetCommand(c *SRedisClient) {
	o := c.db.lookupKeyReadOrReply(c, c.args[1], nil)
	if o != nil && o.checkType(c, SR_DICT) {
		addHashFieldToReply(c, o, c.args[2])
	}
}

// HDEL key field [field ...]
func hDelCommand(c *SRedisClient) {
	deleted := int64(0)
	o := c.db.lookupKeyReadOrReply(c, c.args[1], nil)
	if o == nil || !o.checkType(c, SR_DICT) {
		return
	}
	for i := 2; i < len(c.args); i++ {
		if !hashTypeDel(o, c.args[i]) {
			continue
		}
		deleted++
		if hashTypeLength(o) == 0 {
			c.db.dbDel(c.args[1])
			break
		}
	}
	c.addReplyLongLong(deleted)
}

// hexists key field
func hExistsCommand(c *SRedisClient) {
	o := c.db.lookupKeyReadOrReply(c, c.args[1], shared.czero)
	if o == nil || !o.checkType(c, SR_DICT) {
		return
	}
	if hashTypeExists(o, c.args[2]) {
		c.addReply(shared.cone)
		return
	}
	c.addReply(shared.czero)
}

// hlen key
func hLenCommand(c *SRedisClient) {
	o := c.db.lookupKeyReadOrReply(c, c.args[1], shared.czero)
	if o != nil && o.checkType(c, SR_DICT) {
		c.addReplyLongLong(hashTypeLength(o))
	}
}

// HKEYS key
func hKeysCommand(c *SRedisClient) {
	genericHGetAllCommand(c, OBJ_HASH_KEY)
}

// HVALS key
func hValsCommand(c *SRedisClient) {
	genericHGetAllCommand(c, OBJ_HASH_VALUE)
}

// HGETALL key
func hGetAllCommand(c *SRedisClient) {
	genericHGetAllCommand(c, OBJ_HASH_KEY|OBJ_HASH_VALUE)
}
