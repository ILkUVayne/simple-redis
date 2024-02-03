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
	server.incrDirtyCount(c, int64(added))
}

func sinterGenericCommand(c *SRedisClient, setKeys []*SRobj, setNum int64, dstKey *SRobj) {
	var dstSet *SRobj
	var eleObj *SRobj
	var encoding int
	var intObj int64
	var j int64
	var replyLen *node
	var cardinality int
	set := make([]*SRobj, setNum)

	for i := int64(0); i < setNum; i++ {
		var setObj *SRobj
		if dstKey != nil {
			setObj = c.db.lookupKeyWrite(setKeys[i])
		} else {
			setObj = c.db.lookupKeyRead(setKeys[i])
		}
		if setObj == nil {
			set = nil
			if dstKey != nil {
				if c.db.dbDel(dstKey) == REDIS_OK {
					server.incrDirtyCount(c, 1)
				}
				c.addReply(shared.czero)
				return
			}
			c.addReply(shared.emptyMultiBulk)
			return
		}
		if !setObj.checkType(c, SR_SET) {
			set = nil
			return
		}
		set[i] = setObj
	}

	sortSet(set)

	if dstKey == nil {
		replyLen = c.addDeferredMultiBulkLength()
	} else {
		dstSet = createIntSetObject()
	}

	si := setTypeInitIterator(set[0])
	for encoding = si.setTypeNext(&eleObj, &intObj); encoding != -1; encoding = si.setTypeNext(&eleObj, &intObj) {
		for j = 1; j < setNum; j++ {
			if set[j] == set[0] {
				continue
			}
			if uint8(encoding) == REDIS_ENCODING_INTSET {
				if set[j].encoding == REDIS_ENCODING_INTSET && !assertIntSet(set[j]).intSetFind(intObj) {
					break
				}
				if set[j].encoding == REDIS_ENCODING_HT {
					eleObj = createFromInt(intObj)
					if !setTypeIsMember(set[j], eleObj) {
						eleObj.decrRefCount()
						break
					}
					eleObj.decrRefCount()
				}
			}
			if uint8(encoding) == REDIS_ENCODING_HT {
				if eleObj.encoding == REDIS_ENCODING_INT && set[j].encoding == REDIS_ENCODING_INTSET {
					iVal, _ := eleObj.intVal()
					if !assertIntSet(set[j]).intSetFind(iVal) {
						break
					}
				}
				if !setTypeIsMember(set[j], eleObj) {
					break
				}
			}
		}

		if j == setNum {
			if dstKey == nil {
				if uint8(encoding) == REDIS_ENCODING_HT {
					c.addReplyBulk(eleObj)
				} else {
					c.addReplyBulkInt(intObj)
				}
				cardinality++
			} else {
				if uint8(encoding) == REDIS_ENCODING_INTSET {
					eleObj = createFromInt(intObj)
					setTypeAdd(dstSet, eleObj)
					eleObj.decrRefCount()
				} else {
					setTypeAdd(dstSet, eleObj)
				}
			}
		}
	}
	si.setTypeReleaseIterator()
	set = nil
	if dstKey == nil {
		c.setDeferredMultiBulkLength(replyLen, cardinality)
		return
	}
	c.db.dbDel(dstKey)
	if setTypeSize(dstSet) > 0 {
		c.db.dictSet(dstKey, dstSet)
		c.addReplyLongLong(int(setTypeSize(dstSet)))
	} else {
		dstSet.decrRefCount()
		c.addReply(shared.czero)
	}
	server.incrDirtyCount(c, 1)
}

func sinterCommand(c *SRedisClient) {
	sinterGenericCommand(c, c.args[1:], int64(len(c.args[1:])), nil)
}

func sinterStoreCommand(c *SRedisClient) {
	sinterGenericCommand(c, c.args[2:], int64(len(c.args[2:])), c.args[1])
}
