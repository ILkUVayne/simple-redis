package src

import "math"

//-----------------------------------------------------------------------------
// String commands
//-----------------------------------------------------------------------------

func incrDecrCommand(c *SRedisClient, incr int64) {
	o := c.db.lookupKeyWrite(c.args[1])
	if o != nil && !o.checkType(c, SR_STR) {
		return
	}
	var value int64
	if o.getLongLongFromObjectOrReply(c, &value, nil) == REDIS_ERR {
		return
	}
	oldValue := value
	if (incr < 0 && oldValue < 0 && incr < (math.MinInt64-oldValue)) ||
		(incr > 0 && oldValue > 0 && incr > (math.MaxInt64-oldValue)) {
		c.addReplyError("increment or decrement would overflow")
		return
	}
	value += incr
	if o == nil {
		c.db.dictSet(c.args[1], createFromInt(value))
	} else {
		o.Val = value
	}
	server.incrDirtyCount(c, 1)
	c.addReplyLongLong(value)
}

// get key
func getCommand(c *SRedisClient) {
	val := c.db.lookupKeyReadOrReply(c, c.args[1], nil)
	if val != nil && val.checkType(c, SR_STR) {
		c.addReplyBulk(val)
	}
}

// set key value
func setCommand(c *SRedisClient) {
	key, val := c.args[1], c.args[2]
	if !val.checkType(c, SR_STR) {
		return
	}
	val.tryObjectEncoding()
	c.db.dictSet(key, val)
	c.db.expireDel(key)
	c.addReply(shared.ok)
	server.incrDirtyCount(c, 1)
}

// incr key
func incrCommand(c *SRedisClient) {
	incrDecrCommand(c, 1)
}

// decr key
func decrCommand(c *SRedisClient) {
	incrDecrCommand(c, -1)
}
