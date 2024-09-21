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

func sUnionDiffGenericCommand(c *SRedisClient, setKeys []*SRobj, setNum int64, dstKey *SRobj, op int) {
	sets := make([]*SRobj, setNum)
	for i := int64(0); i < setNum; i++ {
		setObj := c.db.lookupKeyRead(setKeys[i])
		if setObj == nil {
			sets[i] = nil
			continue
		}
		if !setObj.checkType(c, SR_SET) {
			return
		}
		sets[i] = setObj
	}

	// select DIFF algorithm
	//
	// Algorithm 1 is O(N*M) where N is the size of the element first set
	// and M the total number of sets.
	//
	// Algorithm 2 is O(N) where N is the total number of elements in all
	// the sets.
	diffAlgo := 1
	if op == SET_OP_DIFF && sets[0] != nil {
		algoOne, algoTwo := int64(0), int64(0)
		for i := int64(0); i < setNum; i++ {
			if sets[i] == nil {
				continue
			}
			algoOne += setTypeSize(sets[0])
			algoTwo += setTypeSize(sets[i])
		}
		algoOne /= 2
		if algoOne > algoTwo {
			diffAlgo = 2
		}
		if diffAlgo == 1 && setNum > 1 {
			sortSet(sets[1:])
		}
	}

	dstset := createIntSetObject()
	cardinality := int64(0)

	// sunion
	if op == SET_OP_UNION {
		for i := int64(0); i < setNum; i++ {
			if sets[i] == nil {
				continue
			}
			si := setTypeInitIterator(sets[i])
			for ele := si.setTypeNextObject(); ele != nil; ele = si.setTypeNextObject() {
				if setTypeAdd(dstset, ele) {
					cardinality++
				}
				ele.decrRefCount()
			}
			si.setTypeReleaseIterator()
		}
	}

	// sdiff Algorithm 1
	if op == SET_OP_DIFF && sets[0] != nil && diffAlgo == 1 {
		si := setTypeInitIterator(sets[0])
		for ele := si.setTypeNextObject(); ele != nil; ele = si.setTypeNextObject() {
			var i int64
			for i = int64(1); i < setNum; i++ {
				if sets[i] == nil {
					continue
				}
				if sets[i] == sets[0] || setTypeIsMember(sets[i], ele) {
					break
				}
			}
			if i == setNum {
				setTypeAdd(dstset, ele)
				cardinality++
			}
			ele.decrRefCount()
		}
		si.setTypeReleaseIterator()
	}

	// sdiff Algorithm 2
	if op == SET_OP_DIFF && sets[0] != nil && diffAlgo == 2 {
		for i := int64(0); i < setNum; i++ {
			if sets[i] == nil {
				continue
			}
			si := setTypeInitIterator(sets[i])
			for ele := si.setTypeNextObject(); ele != nil; ele = si.setTypeNextObject() {
				if i == 0 {
					if setTypeAdd(dstset, ele) {
						cardinality++
					}
				} else {
					if setTypeRemove(dstset, ele) {
						cardinality--
					}
				}
				ele.decrRefCount()
			}
			si.setTypeReleaseIterator()

			if cardinality == 0 {
				break
			}
		}
	}

	// 返回响应
	if dstKey == nil {
		c.addReplyMultiBulkLen(cardinality, false)
		si := setTypeInitIterator(dstset)
		for ele := si.setTypeNextObject(); ele != nil; ele = si.setTypeNextObject() {
			c.addReplyBulk(ele)
			ele.decrRefCount()
		}
		si.setTypeReleaseIterator()
		dstset.decrRefCount()
		return
	}
	// dstKey != nil
	c.db.dbDel(dstKey)
	server.incrDirtyCount(c, 1)
	if setTypeSize(dstset) == 0 {
		dstset.decrRefCount()
		c.addReply(shared.czero)
		return
	}
	// dstSet is not empty
	c.db.dictSet(dstKey, dstset)
	c.addReplyLongLong(setTypeSize(dstset))
}

func sPopWithCountCommand(c *SRedisClient) {
	// get count
	var count int64
	if c.args[2].getLongLongFromObjectOrReply(c, &count, "") != REDIS_OK {
		return
	}

	if count < 0 {
		c.addReply(shared.outOfRangeErr)
		return
	}

	key := c.args[1]
	set := c.db.lookupKeyReadOrReply(c, key, shared.emptyMultiBulk)
	if set == nil || !set.checkType(c, SR_SET) {
		return
	}

	if count == 0 {
		c.addReply(shared.emptyMultiBulk)
		return
	}

	size := setTypeSize(set)

	if count >= size {
		sUnionDiffGenericCommand(c, c.args[1:2], 1, nil, SET_OP_UNION)
		// del the set
		c.rewriteClientCommandVector(shared.del, key)
		c.db.dbDel(key)
		server.incrDirtyCount(c, 1)
		return
	}

	rewriteArgs := make([]*SRobj, 0)
	rewriteArgs = append(rewriteArgs, shared.sRem, key)
	c.addReplyMultiBulkLen(count, false)
	remaining := size - count
	server.incrDirtyCount(c, 1)

	if remaining*SPOP_MOVE_STRATEGY_MUL > count {
		for ; count > 0; count-- {
			encoding, objEle, intEle := setTypeRandomElement(set)
			if encoding == REDIS_ENCODING_INTSET {
				objEle = createFromInt(intEle)
			}
			objEle.incrRefCount()
			c.addReplyBulk(objEle)
			setTypeRemove(set, objEle)
			rewriteArgs = append(rewriteArgs, objEle)
		}
		c.rewriteClientCommandVector(rewriteArgs...)
		return
	}

	var newSet *SRobj
	for ; remaining > 0; remaining-- {
		encoding, objEle, intEle := setTypeRandomElement(set)
		if encoding == REDIS_ENCODING_INTSET {
			objEle = createFromInt(intEle)
		}
		objEle.incrRefCount()
		if newSet == nil {
			newSet = setTypeCreate(objEle)
		}
		setTypeAdd(newSet, objEle)
		setTypeRemove(set, objEle)
	}

	set.incrRefCount()
	c.db.dictSet(key, newSet)

	si := setTypeInitIterator(set)
	var eleObj *SRobj
	var intObj int64
	for encoding := si.setTypeNext(&eleObj, &intObj); encoding != -1; encoding = si.setTypeNext(&eleObj, &intObj) {
		if uint8(encoding) == REDIS_ENCODING_INTSET {
			eleObj = createFromInt(intObj)
		}
		eleObj.incrRefCount()
		c.addReplyBulk(eleObj)
		rewriteArgs = append(rewriteArgs, eleObj)
	}
	si.setTypeReleaseIterator()
	c.rewriteClientCommandVector(rewriteArgs...)
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

// SUNION key [key ...]
func sUnionCommand(c *SRedisClient) {
	sUnionDiffGenericCommand(c, c.args[1:], int64(len(c.args[1:])), nil, SET_OP_UNION)
}

// SUNIONSTORE destination key [key ...]
func sUnionStoreCommand(c *SRedisClient) {
	sUnionDiffGenericCommand(c, c.args[2:], int64(len(c.args[2:])), c.args[1], SET_OP_UNION)
}

// SDIFF key [key ...]
func sDiffCommand(c *SRedisClient) {
	sUnionDiffGenericCommand(c, c.args[1:], int64(len(c.args[1:])), nil, SET_OP_DIFF)
}

// SDIFFSTORE destination key [key ...]
func sDiffStoreCommand(c *SRedisClient) {
	sUnionDiffGenericCommand(c, c.args[2:], int64(len(c.args[2:])), c.args[1], SET_OP_DIFF)
}

// spop key [count]
func sPopCommand(c *SRedisClient) {
	if len(c.args) > 3 {
		c.addReply(shared.syntaxErr)
		return
	}
	// with count
	if len(c.args) == 3 {
		sPopWithCountCommand(c)
		return
	}
	// single pop
	key := c.args[1]
	set := c.db.lookupKeyReadOrReply(c, key, shared.nullBulk)
	if set == nil || !set.checkType(c, SR_SET) {
		return
	}

	// Get a random element
	encoding, objEle, intEle := setTypeRandomElement(set)

	// remove the element
	if encoding == REDIS_ENCODING_INTSET {
		objEle = createFromInt(intEle)
	}
	objEle.incrRefCount()
	setTypeRemove(set, objEle)

	// Replicate/AOF this command as an SREM operation
	c.rewriteClientCommandVector(shared.sRem, key, objEle)

	// Add the element to the reply
	c.addReplyBulk(objEle)

	// Delete the set if it's empty
	if setTypeSize(set) == 0 {
		c.db.dbDel(key)
	}

	// update dirty
	server.incrDirtyCount(c, 1)
}

// srem key member [member ...]
func sRemCommand(c *SRedisClient) {
	key := c.args[1]
	set := c.db.lookupKeyReadOrReply(c, key, shared.nullBulk)
	if set == nil || !set.checkType(c, SR_SET) {
		return
	}

	deleted := int64(0)
	for i := 2; i < len(c.args); i++ {
		if setTypeRemove(set, c.args[i]) {
			deleted++
			if setTypeSize(set) == 0 {
				c.db.dbDel(key)
				break
			}
		}
	}
	server.incrDirtyCount(c, deleted)
	c.addReplyLongLong(deleted)
}
