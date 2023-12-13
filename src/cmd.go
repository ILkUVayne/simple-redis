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

// commandTable 命令列表
var commandTable = []SRedisCommand{
	{"get", getCommand, 2},
	{"set", setCommand, 3},
}

func getCommand(c *SRedisClient) {
	key := c.args[1]
	val := findVal(key)
	if val == nil {
		c.addReplyStr("$-1\r\n")
		return
	}
	if val.Typ != SR_STR {
		c.addReplyStr("-ERR: wrong type\r\n")
		return
	}
	str := val.strVal()
	c.addReplyStr(fmt.Sprintf("$%d%v\r\n", len(str), str))
}

func setCommand(c *SRedisClient) {
	key := c.args[1]
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReplyStr("-ERR: wrong type\r\n")
	}
	server.db.data.dictSet(key, val)
	server.db.expire.dictDelete(key)
	c.addReplyStr("+OK\r\n")
}

func lookupCommand(cmdStr string) *SRedisCommand {
	for _, c := range commandTable {
		if c.name == cmdStr {
			return &c
		}
	}
	return nil
}

func processCommand(c *SRedisClient) {
	cmdStr := c.args[0].strVal()
	utils.Info("process command: ", cmdStr)
	if cmdStr == "quit" {
		freeClient(c)
		return
	}
	cmd := lookupCommand(cmdStr)
	if cmd == nil {
		c.addReplyStr("-ERR: unknow command")
		freeClient(c)
		return
	}
	if cmd.arity != len(c.args) {
		c.addReplyStr("-ERR: wrong number of args")
		freeClient(c)
		return
	}
	cmd.proc(c)
	resetClient(c)
}
