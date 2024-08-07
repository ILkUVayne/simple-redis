package src

//-----------------------------------------------------------------------------
// Set commands
//-----------------------------------------------------------------------------

func getSets(c *SRedisClient, setKeys []*SRobj, setNum int64, dstKey *SRobj) []*SRobj {
	sets := make([]*SRobj, setNum)
	for i := int64(0); i < setNum; i++ {
		setObj := c.db.lookupKeyRead(setKeys[i])
		if setObj != nil {
			if !setObj.checkType(c, SR_SET) {
				return nil
			}
			sets[i] = setObj
			continue
		}
		// setObj == nil
		if dstKey == nil {
			c.addReply(shared.emptyMultiBulk)
			return nil
		}
		if c.db.dbDel(dstKey) == REDIS_OK {
			server.incrDirtyCount(c, 1)
		}
		c.addReply(shared.czero)
		return nil
	}
	return sets
}

func sinterGenericCommand(c *SRedisClient, setKeys []*SRobj, setNum int64, dstKey *SRobj) {
	var dstSet *SRobj
	var eleObj *SRobj
	var encoding int
	var intObj int64
	var replyLen *node
	var cardinality int
	// 通过key查找对应的无序集合，并放入sets切片
	sets := getSets(c, setKeys, setNum, dstKey)
	if sets == nil {
		return
	}
	// 根据sets中的每个集合的size大小，从小到大排序，方便后续取交集
	sortSet(sets)

	if dstKey == nil {
		replyLen = c.addDeferredMultiBulkLength()
	} else {
		dstSet = createIntSetObject()
	}
	// 获取交集
	// 迭代最小的集合sets[0]，然后遍历剩余的集合，若都存在则表示交集
	// 例如：sets[0] = {1,2} sets[1] = {1,3,4} sets[2] = {1,4,5}
	// 第一次迭代：intObj（eleObj是当set encoding是hash table的时候的值）= 1，依次遍历sets[1] sets[2]，均存在1，是交集
	// 第二次迭代：intObj = 2，依次遍历sets[1] sets[2]，只要有一个集合不存2，则不是交集
	// 最终得出交集为 1
	si := setTypeInitIterator(sets[0])
	var j int64
	for encoding = si.setTypeNext(&eleObj, &intObj); encoding != -1; encoding = si.setTypeNext(&eleObj, &intObj) {
		// 遍历剩余集合，查询当前迭代值是否存在该集合中，若不存在直接break，表示当前值不是交集，迭代sets[0]中的下一个
		for j = 1; j < setNum; j++ {
			if sets[j] == sets[0] {
				continue
			}
			if uint8(encoding) == REDIS_ENCODING_INTSET {
				if sets[j].encoding == REDIS_ENCODING_INTSET && !assertIntSet(sets[j]).intSetFind(intObj) {
					break
				}
				if sets[j].encoding == REDIS_ENCODING_HT {
					eleObj = createFromInt(intObj)
					if !setTypeIsMember(sets[j], eleObj) {
						eleObj.decrRefCount()
						break
					}
					eleObj.decrRefCount()
				}
			}
			if uint8(encoding) == REDIS_ENCODING_HT {
				if eleObj.encoding == REDIS_ENCODING_INT && sets[j].encoding == REDIS_ENCODING_INTSET {
					iVal, _ := eleObj.intVal()
					if !assertIntSet(sets[j]).intSetFind(iVal) {
						break
					}
				}
				if !setTypeIsMember(sets[j], eleObj) {
					break
				}
			}
		}
		// j != setNum 表示当前值不是交集
		if j != setNum {
			continue
		}
		// j == setNum 当前值是交集
		if dstKey == nil {
			cardinality++
			if uint8(encoding) == REDIS_ENCODING_HT {
				c.addReplyBulk(eleObj)
				continue
			}
			c.addReplyBulkInt(intObj)
			continue
		}
		if uint8(encoding) == REDIS_ENCODING_INTSET {
			eleObj = createFromInt(intObj)
			setTypeAdd(dstSet, eleObj)
			eleObj.decrRefCount()
			continue
		}
		setTypeAdd(dstSet, eleObj)
	}
	// 返回响应
	si.setTypeReleaseIterator()
	sets = nil
	if dstKey == nil {
		c.setDeferredMultiBulkLength(replyLen, cardinality)
		return
	}
	c.db.dbDel(dstKey)
	server.incrDirtyCount(c, 1)
	if setTypeSize(dstSet) > 0 {
		c.db.dictSet(dstKey, dstSet)
		c.addReplyLongLong(setTypeSize(dstSet))
		return
	}
	dstSet.decrRefCount()
	c.addReply(shared.czero)
}

// sadd key member [member ...]
func sAddCommand(c *SRedisClient) {
	key := c.args[1]
	set := c.db.lookupKeyWrite(key)
	if set != nil && !set.checkType(c, SR_SET) {
		return
	}
	if set == nil {
		set = setTypeCreate(c.args[2])
		c.db.dictSet(key, set)
	}
	// add...
	added := 0
	for j := 2; j < len(c.args); j++ {
		c.args[j].tryObjectEncoding()
		if setTypeAdd(set, c.args[j]) {
			added++
		}
	}
	c.addReplyLongLong(int64(added))
	server.incrDirtyCount(c, int64(added))
}

// smembers key
//
// sinter key [key ...]
func sinterCommand(c *SRedisClient) {
	sinterGenericCommand(c, c.args[1:], int64(len(c.args[1:])), nil)
}

// sinterstore key [key ...]
func sinterStoreCommand(c *SRedisClient) {
	sinterGenericCommand(c, c.args[2:], int64(len(c.args[2:])), c.args[1])
}
