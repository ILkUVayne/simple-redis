package src

import (
	"simple-redis/utils"
	"strings"
)

type CmdType = byte

type CommandProc func(c *SRedisClient)

type SRedisCommand struct {
	name  string
	proc  CommandProc
	arity int
}

func (cmd *SRedisCommand) propagate(args []*SRobj) {
	if server.aofState == REDIS_AOF_ON {
		cmd.feedAppendOnlyFile(args, len(args))
	}
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
	if c.fd > 0 {
		utils.Info("process command: ", cmdStr)
	}

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
	dirty := server.dirty
	c.cmd.proc(c)
	// aof
	dirty = server.dirty - dirty
	if dirty > 0 {
		c.cmd.propagate(c.args)
	}
}

// =================================== command ====================================

// commandTable 命令列表
var commandTable = []SRedisCommand{
	{EXPIRE, expireCommand, 3},
	{OBJECT, objectCommand, 3},
	{DEL, delCommand, -2},
	{KEYS, keysCommand, 2},
	// string
	{GET, getCommand, 2},
	{SET, setCommand, 3},
	// zset
	{Z_ADD, zAddCommand, -4},
	{Z_RANGE, zRangeCommand, -4},
	// set
	{S_ADD, sAddCommand, -3},
	{SMEMBERS, sinterCommand, 2},
	{SINTER, sinterCommand, -2},
	{SINTER_STORE, sinterStoreCommand, -2},
	// list
	{R_PUSH, rPushCommand, -3},
	{L_PUSH, lPushCommand, -3},
	{R_POP, rPopCommand, 2},
	{L_POP, lPopCommand, 2},
	// hash
	{H_SET, hSetCommand, 4},
	{H_GET, hGetCommand, 3},
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

	if c.db.lookupKeyReadOrReply(c, key, nil) == nil {
		return
	}

	eval, _ := val.intVal()
	expire := eval
	if eval < MAX_EXPIRE {
		expire = utils.GetMsTime() + (eval * 1000)
	}

	expireObj := createFromInt(expire)
	server.db.expire.dictSet(key, expireObj)
	expireObj.decrRefCount()
	c.addReply(shared.ok)
	server.incrDirtyCount(c, 1)
}

func objectCommand(c *SRedisClient) {
	val := c.args[2]
	if val.Typ != SR_STR {
		c.addReply(shared.typeErr)
		return
	}
	value := c.db.lookupKeyReadOrReply(c, val, nil)
	if value == nil {
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
