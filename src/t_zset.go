package src

//-----------------------------------------------------------------------------
// Sorted set commands
//-----------------------------------------------------------------------------

func zAddGenericCommand(c *SRedisClient, incr bool) {
	//nanErr := errors.New("resulting score is not a number (NaN)")
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
		if c.args[2+i*2].getFloat64FromObjectOrReply(c, &scores[i], nil) == REDIS_ERR {
			return
		}
	}

	zobj := server.db.lookupKeyWrite(key)
	if zobj != nil && zobj.Typ != SR_ZSET {
		c.addReply(shared.wrongTypeErr)
		return
	}
	if zobj == nil {
		zobj = createZsetSRobj()
		server.db.dictSet(key, zobj)
	}

	for i := 0; i < elements; i++ {
		score = scores[i]
		ele := c.args[3+i*2]
		ele.tryObjectEncoding()
		zs := zobj.Val.(*zSet)
		_, de := zs.d.dictFind(ele)
		if de != nil {
			curobj := de.key
			curscore, _ := de.val.floatVal()
			if incr {
				score += curscore
			}
			if score != curscore {
				zs.zsl.delete(curscore, curobj)
				zNode := zs.zsl.insert(curscore, curobj)
				curobj.incrRefCount()
				zs.d.dictSet(curobj, createFloatSRobj(SR_STR, zNode.score))
				server.dirty++
			}
		} else {
			zNode := zs.zsl.insert(score, ele)
			ele.incrRefCount()
			zs.d.dictSet(ele, createFloatSRobj(SR_STR, zNode.score))
			ele.incrRefCount()
			server.dirty++
			if !incr {
				added++
			}
		}
	}

	if incr {
		c.addReplyDouble(score)
		return
	}
	c.addReplyLongLong(added)
}

func zAddCommand(c *SRedisClient) {
	zAddGenericCommand(c, false)
}

func zRangeGenericCommand(c *SRedisClient, reverse bool) {
	var start int64
	var end int64
	var zobj *SRobj
	key := c.args[1]
	withscores := false

	if c.args[2].getLongLongFromObjectOrReply(c, &start, nil) != REDIS_OK ||
		c.args[3].getLongLongFromObjectOrReply(c, &end, nil) != REDIS_OK {
		return
	}
	if len(c.args) > 5 || (len(c.args) == 5 && c.args[4].strVal() != "withscores") {
		c.addReply(shared.syntaxErr)
		return
	}
	if len(c.args) == 5 && c.args[4].strVal() == "withscores" {
		withscores = true
	}

	zobj = c.db.lookupKeyReadOrReply(c, key, shared.emptyMultiBulk)
	if zobj == nil || zobj.checkType(c, SR_ZSET) == false {
		return
	}

	zs := zobj.Val.(*zSet)
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
	c.replyReady = false
	c.addReplyMultiBulkLen(arrayLen)
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
	c.doReply()
}

func zRangeCommand(c *SRedisClient) {
	zRangeGenericCommand(c, false)
}
