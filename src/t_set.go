package src

func setTypeCreate(value *SRobj) *SRobj {
	if value.isObjectRepresentableAsInt64(nil) == REDIS_OK {
		return createIntSetObject()
	}
	return createSetObject()
}

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
}
