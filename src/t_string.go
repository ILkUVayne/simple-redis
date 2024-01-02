package src

import "fmt"

//-----------------------------------------------------------------------------
// String commands
//-----------------------------------------------------------------------------

func getCommand(c *SRedisClient) {
	key := c.args[1]
	val := findVal(key)
	if val == nil {
		c.addReply(shared.nullBulk)
		return
	}
	if val.Typ != SR_STR {
		c.addReply(shared.typeErr)
		return
	}
	str := val.strVal()
	c.addReplyStr(fmt.Sprintf(RESP_BULK, len(str), str))
}

func setCommand(c *SRedisClient) {
	key := c.args[1]
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReply(shared.typeErr)
		return
	}
	val.tryObjectEncoding()
	server.db.data.dictSet(key, val)
	server.db.expire.dictDelete(key)
	c.addReply(shared.ok)
}
