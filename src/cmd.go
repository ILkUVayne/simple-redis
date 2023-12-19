package src

import (
	"fmt"
	"simple-redis/utils"
)

const (
	CMD_UNKNOWN CmdType = iota
	CMD_INLINE
	CMD_BULK
)

type CmdType = byte

type CommandProc func(c *SRedisClient)

type SRedisCommand struct {
	name  string
	proc  CommandProc
	arity int
}

// 查询需要执行的命令
func lookupCommand(cmdStr string) *SRedisCommand {
	for _, c := range commandTable {
		if c.name == cmdStr {
			return &c
		}
	}
	return nil
}

// 执行命令
func processCommand(c *SRedisClient) {
	cmdStr := c.args[0].strVal()
	utils.Info("process command: ", cmdStr)
	if cmdStr == "quit" {
		freeClient(c)
		return
	}
	cmd := lookupCommand(cmdStr)
	if cmd == nil {
		c.addReplyStr(RESP_UNKOWN)
		resetClient(c)
		return
	}
	if cmd.arity != len(c.args) {
		c.addReplyStr(RESP_ARGS_NUM_ERR)
		resetClient(c)
		return
	}
	cmd.proc(c)
	resetClient(c)
}

// =================================== command ====================================

// commandTable 命令列表
var commandTable = []SRedisCommand{
	{"get", getCommand, 2},
	{"set", setCommand, 3},
	{"expire", expireCommand, 3},
	// more
}

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

func expireCommand(c *SRedisClient) {
	key := c.args[1]
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReplyStr(RESP_TYP_ERR)
	}
	expire := utils.GetMsTime() + (val.intVal() * 1000)
	expireObj := createFromInt(expire)
	server.db.expire.dictSet(key, expireObj)
	expireObj.decrRefCount()
	c.addReplyStr(RESP_OK)
}
