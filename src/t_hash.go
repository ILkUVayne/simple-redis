package src

//-----------------------------------------------------------------------------
// Hash type commands
//-----------------------------------------------------------------------------

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
	if o == nil || !o.checkType(c, SR_DICT) {
		return
	}
	addHashFieldToReply(c, o, c.args[2])
}

// HDEL key field [field ...]
func hDelCommand(c *SRedisClient) {
	deleted := 0
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
	c.addReplyLongLong(int64(deleted))
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
