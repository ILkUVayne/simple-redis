package src

import (
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
		c.addReply(shared.unknowErr)
		resetClient(c)
		return
	}
	if cmd.arity > 0 && cmd.arity != len(c.args) {
		c.addReply(shared.argsNumErr)
		resetClient(c)
		return
	}
	cmd.proc(c)
	resetClient(c)
}

// =================================== command ====================================

// commandTable 命令列表
var commandTable = []SRedisCommand{
	{"expire", expireCommand, 3},
	{"object", objectCommand, 3},
	// string
	{"get", getCommand, 2},
	{"set", setCommand, 3},
	// zset
	{"zadd", zAddCommand, -4},
	{"zrange", zRangeCommand, -4},
	// more
}

func expireCommand(c *SRedisClient) {
	key := c.args[1]
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReply(shared.typeErr)
		return
	}
	expire := utils.GetMsTime() + (val.intVal() * 1000)
	expireObj := createFromInt(expire)
	server.db.expire.dictSet(key, expireObj)
	expireObj.decrRefCount()
	c.addReply(shared.ok)
}

func objectCommand(c *SRedisClient) {
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReply(shared.typeErr)
		return
	}
	value := server.db.data.dictGet(val)
	if value == nil {
		c.addReply(shared.nullBulk)
		return
	}
	c.addReplyBulk(value.getEncoding())
}
