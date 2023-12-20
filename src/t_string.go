package src

import "fmt"

func getCommand(c *SRedisClient) {
	key := c.args[1]
	val := findVal(key)
	if val == nil {
		c.addReplyStr(RESP_NIL_VAL)
		return
	}
	if val.Typ != SR_STR {
		c.addReplyStr(RESP_TYP_ERR)
		return
	}
	str := val.strVal()
	c.addReplyStr(fmt.Sprintf(RESP_BULK, len(str), str))
}

func setCommand(c *SRedisClient) {
	key := c.args[1]
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReplyStr(RESP_TYP_ERR)
	}
	server.db.data.dictSet(key, val)
	server.db.expire.dictDelete(key)
	c.addReplyStr(RESP_OK)
}
