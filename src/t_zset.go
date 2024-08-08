package src

//-----------------------------------------------------------------------------
// Sorted set commands
//-----------------------------------------------------------------------------

func zAddGenericCommand(c *SRedisClient, incr bool) {
	var score float64
	var added int

	key := c.args[1]
	elements := len(c.args[2:]) / 2

	if len(c.args)%2 == 1 {
		c.addReply(shared.syntaxErr)
		return
	}

	scores := make([]float64, elements)
	for i := 0; i < elements; i++ {
		if c.args[2+i*2].getFloat64FromObjectOrReply(c, &scores[i], "") == REDIS_ERR {
			return
		}
	}

	zObj := c.db.lookupKeyWrite(key)
	if zObj != nil && !zObj.checkType(c, SR_ZSET) {
		return
	}
	if zObj == nil {
		zObj = createZsetSRobj()
		c.db.dictSet(key, zObj)
	}

	for i := 0; i < elements; i++ {
		score = scores[i]
		ele := c.args[3+i*2]
		ele.tryObjectEncoding()
		zs := assertZSet(zObj)
		_, de := zs.d.dictFind(ele)
		if de == nil {
			zNode := zs.zsl.insert(score, ele)
			ele.incrRefCount()
			zs.d.dictSet(ele, createFloatSRobj(SR_STR, zNode.score))
			ele.incrRefCount()
			server.incrDirtyCount(c, 1)
			if !incr {
				added++
			}
			continue
		}
		// de != nil
		curObj := de.key
		curScore, _ := de.val.floatVal()
		if incr {
			score += curScore
		}
		// unchanged
		if score == curScore {
			continue
		}
		zs.zsl.delete(curScore, curObj)
		zNode := zs.zsl.insert(score, curObj)
		curObj.incrRefCount()
		zs.d.dictSet(curObj, createFloatSRobj(SR_STR, zNode.score))
		server.incrDirtyCount(c, 1)
	}

	if incr {
		c.addReplyDouble(score)
		return
	}
	c.addReplyLongLong(int64(added))
}

// zadd key score member [score member ...]
func zAddCommand(c *SRedisClient) {
	zAddGenericCommand(c, false)
}

func zRangeGenericCommand(c *SRedisClient, reverse bool) {
	var start int64
	var end int64

	key := c.args[1]
	withscores := false

	if c.args[2].getLongLongFromObjectOrReply(c, &start, "") != REDIS_OK ||
		c.args[3].getLongLongFromObjectOrReply(c, &end, "") != REDIS_OK {
		return
	}
	if len(c.args) > 5 || (len(c.args) == 5 && c.args[4].strVal() != "withscores") {
		c.addReply(shared.syntaxErr)
		return
	}
	if len(c.args) == 5 && c.args[4].strVal() == "withscores" {
		withscores = true
	}

	zobj := c.db.lookupKeyReadOrReply(c, key, shared.emptyMultiBulk)
	if zobj == nil || !zobj.checkType(c, SR_ZSET) {
		return
	}

	zs := assertZSet(zobj)
	llen := int64(zs.zSetLength())
	if start < 0 {
		start = llen + start
	}
	if end < 0 {
		end = llen + end
	}
	if start < 0 {
		start = 0
	}

	if start > end || start > llen {
		c.addReply(shared.emptyMultiBulk)
		return
	}
	if end >= llen {
		end = llen - 1
	}
	rangeLen := (end - start) + 1

	var ln *zSkipListNode
	zsl := zs.zsl
	if reverse {
		ln = zsl.tail
		if start > 0 {
			ln = zsl.getElementByRank(uint(llen - start))
		}
	} else {
		ln = zsl.header.level[0].forward
		if start > 0 {
			ln = zsl.getElementByRank(uint(start + 1))
		}
	}
	arrayLen := rangeLen
	if withscores {
		arrayLen *= 2
	}
	c.addReplyMultiBulkLen(arrayLen, false)
	for ; rangeLen > 0; rangeLen-- {
		ele := ln.obj
		c.addReplyBulk(ele)
		if withscores {
			c.addReplyDouble(ln.score)
		}
		if reverse {
			ln = ln.backward
		} else {
			ln = ln.level[0].forward
		}
	}
}

// zrange key min max [withscores]
func zRangeCommand(c *SRedisClient) {
	zRangeGenericCommand(c, false)
}
