package src

//-----------------------------------------------------------------------------
// Set commands
//-----------------------------------------------------------------------------

func sAddCommand(c *SRedisClient) {
	key := c.args[1]
	set := server.db.lookupKeyWrite(key)
	if set != nil && set.Typ != SR_SET {
		c.addReply(shared.wrongTypeErr)
		return
	}
	if set == nil {
		set = setTypeCreate(c.args[2])
		server.db.dictSet(key, set)
	}
	// add...
	added := 0
	for j := 2; j < len(c.args); j++ {
		c.args[j].tryObjectEncoding()
		if setTypeAdd(set, c.args[j]) {
			added++
		}
	}
	c.addReplyLongLong(added)
}
