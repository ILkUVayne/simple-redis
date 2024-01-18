package src

//-----------------------------------------------------------------------------
// String commands
//-----------------------------------------------------------------------------

func getCommand(c *SRedisClient) {
	key := c.args[1]
	val := server.db.lookupKeyRead(key)
	if val == nil {
		c.addReply(shared.nullBulk)
		return
	}
	if val.Typ != SR_STR {
		c.addReply(shared.typeErr)
		return
	}
	c.addReplyBulk(val)
}

func setCommand(c *SRedisClient) {
	key := c.args[1]
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReply(shared.typeErr)
		return
	}
	val.tryObjectEncoding()
	server.db.dictSet(key, val)
	server.db.expireDel(key)
	c.addReply(shared.ok)
	server.dirty++
}
