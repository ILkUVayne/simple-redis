package src

import (
	"simple-redis/utils"
	"strings"
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
	// Case-insensitive
	cmdStr = strings.ToLower(cmdStr)
	if cmdStr == "quit" {
		freeClient(c)
		return
	}
	c.cmd = lookupCommand(cmdStr)
	if c.cmd == nil {
		c.addReply(shared.unknowErr)
		resetClient(c)
		return
	}
	if (c.cmd.arity > 0 && c.cmd.arity != len(c.args)) || -c.cmd.arity > len(c.args) {
		c.addReply(shared.argsNumErr)
		resetClient(c)
		return
	}
	call(c)
	resetClient(c)
}

// call is the core of Redis execution of a command
func call(c *SRedisClient) {
	c.cmd.proc(c)
	// aof
}

// =================================== command ====================================

// commandTable 命令列表
var commandTable = []SRedisCommand{
	{"expire", expireCommand, 3},
	{"object", objectCommand, 3},
	{"del", delCommand, -2},
	{"keys", keysCommand, 2},
	// string
	{"get", getCommand, 2},
	{"set", setCommand, 3},
	// zset
	{"zadd", zAddCommand, -4},
	{"zrange", zRangeCommand, -4},
	// set
	{"sadd", sAddCommand, -3},
	{"smembers", sinterCommand, 2},
	{"sinter", sinterCommand, -2},
	{"sinterstore", sinterStoreCommand, -2},
	// list
	{"rpush", rPushCommand, -3},
	{"lpush", lPushCommand, -3},
	{"rpop", rPopCommand, 2},
	{"lpop", lPopCommand, 2},
	// hash
	{"hset", hSetCommand, 4},
	{"hget", hGetCommand, 3},
	// more
}

//-----------------------------------------------------------------------------
// db commands
//-----------------------------------------------------------------------------

func expireCommand(c *SRedisClient) {
	key := c.args[1]
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReply(shared.typeErr)
		return
	}
	eval, _ := val.intVal()
	expire := utils.GetMsTime() + (eval * 1000)
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

func delCommand(c *SRedisClient) {
	deleted := 0
	for i := 1; i < len(c.args); i++ {
		if server.db.dbDel(c.args[i]) == REDIS_OK {
			deleted++
		}
	}
	c.addReplyLongLong(deleted)
}

func keysCommand(c *SRedisClient) {
	pattern := c.args[1].strVal()
	numKeys := 0
	allKeys := false
	if pattern[0] == '*' && len(pattern) == 1 {
		allKeys = true
	}
	replyLen := c.addDeferredMultiBulkLength()
	di := server.db.data.dictGetIterator()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		key := de.getKey()
		if allKeys || utils.StringMatch(pattern, key.strVal(), false) {
			if !server.db.expireIfNeeded(key) {

				c.addReplyBulk(key)
				numKeys++
			}
		}
	}
	di.dictReleaseIterator()
	c.setDeferredMultiBulkLength(replyLen, numKeys)
}
