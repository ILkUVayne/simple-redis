package src

//-----------------------------------------------------------------------------
// Hash type commands
//-----------------------------------------------------------------------------

func hSetCommand(c *SRedisClient) {
	var o *SRobj
	if o = hashTypeLookupWriteOrCreate(c, c.args[1]); o == nil {
		return
	}
	hashTypeTryObjectEncoding(o, c.args[2], c.args[3])
	update := hashTypeSet(o, c.args[2], c.args[3])
	if update == DICT_SET {
		c.addReply(shared.cone)
	} else {
		c.addReply(shared.czero)
	}
	server.dirty++
}

func hGetCommand(c *SRedisClient) {
	var o *SRobj
	o = c.db.lookupKeyReadOrReply(c, c.args[1], shared.nullBulk)
	if o == nil || o.checkType(c, SR_DICT) == false {
		return
	}
	addHashFieldToReply(c, o, c.args[2])
}
