package src

//-----------------------------------------------------------------------------
// List commands
//-----------------------------------------------------------------------------

func pushGenericCommand(c *SRedisClient, where int) {
	var pushed int64
	lObj := c.db.lookupKeyWrite(c.args[1])

	if lObj != nil && lObj.Typ != SR_LIST {
		c.addReply(shared.wrongTypeErr)
		return
	}

	for i := 2; i < len(c.args); i++ {
		c.args[i].tryObjectEncoding()
		if lObj == nil {
			lObj = createListObject()
			c.db.dictSet(c.args[1], lObj)
		}
		listTypePush(lObj, c.args[i], where)
		pushed++
	}
	c.addReplyLongLong(lObj.Val.(*list).len())
	server.incrDirtyCount(c, pushed)
}

func lPushCommand(c *SRedisClient) {
	pushGenericCommand(c, REDIS_HEAD)
}

func rPushCommand(c *SRedisClient) {
	pushGenericCommand(c, REDIS_TAIL)
}

func popGenericCommand(c *SRedisClient, where int) {
	lObj := c.db.lookupKeyReadOrReply(c, c.args[1], shared.nullBulk)
	if lObj == nil || lObj.checkType(c, SR_LIST) == false {
		return
	}

	value := listTypePop(lObj, where)
	if value == nil {
		c.addReply(shared.nullBulk)
		return
	}
	c.addReplyBulk(value)
	value.decrRefCount()
	if lObj.Val.(*list).len() == 0 {
		c.db.dbDel(c.args[1])
	}
	server.incrDirtyCount(c, 1)
}

func lPopCommand(c *SRedisClient) {
	popGenericCommand(c, REDIS_HEAD)
}

func rPopCommand(c *SRedisClient) {
	popGenericCommand(c, REDIS_TAIL)
}
