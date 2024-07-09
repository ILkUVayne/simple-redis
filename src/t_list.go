package src

//-----------------------------------------------------------------------------
// List commands
//-----------------------------------------------------------------------------

func pushGenericCommand(c *SRedisClient, where int) {
	var pushed int64
	lObj := c.db.lookupKeyWrite(c.args[1])

	if lObj != nil && !lObj.checkType(c, SR_LIST) {
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
	c.addReplyLongLong(int64(assertList(lObj).len()))
	server.incrDirtyCount(c, pushed)
}

// lpush key value [value ...]
func lPushCommand(c *SRedisClient) {
	pushGenericCommand(c, AL_START_HEAD)
}

// rpush key value [value ...]
func rPushCommand(c *SRedisClient) {
	pushGenericCommand(c, AL_START_TAIL)
}

func popGenericCommand(c *SRedisClient, where int) {
	lObj := c.db.lookupKeyReadOrReply(c, c.args[1], nil)
	if lObj == nil || !lObj.checkType(c, SR_LIST) {
		return
	}

	value := listTypePop(lObj, where)
	if value == nil {
		c.addReply(shared.nullBulk)
		return
	}
	c.addReplyBulk(value)
	value.decrRefCount()
	if assertList(lObj).len() == 0 {
		c.db.dbDel(c.args[1])
	}
	server.incrDirtyCount(c, 1)
}

// lpop key
func lPopCommand(c *SRedisClient) {
	popGenericCommand(c, AL_START_HEAD)
}

// rpop key
func rPopCommand(c *SRedisClient) {
	popGenericCommand(c, AL_START_TAIL)
}
