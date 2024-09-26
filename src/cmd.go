package src

import (
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"strings"
)

// CmdType CMD_UNKNOWN CMD_INLINE CMD_BULK
type CmdType = byte

// CommandProc 命令处理函数，每个命令对应一个处理函数
type CommandProc func(c *SRedisClient)

// SRedisCommand 命令结构
type SRedisCommand struct {
	name  string      // 命令名称，例如： set
	proc  CommandProc // 命令处理函数
	arity int         // command args mums,if < 0 like -3 means args >= 3
}

func (cmd *SRedisCommand) propagate(args []*SRobj) {
	if server.aofState == REDIS_AOF_ON {
		cmd.feedAppendOnlyFile(args, len(args))
	}
}

// 查询需要执行的命令
func lookupCommand(cmdStr string) *SRedisCommand {
	c, ok := server.commands[cmdStr]
	if !ok {
		return nil
	}
	return &c
}

// 执行命令
func processCommand(c *SRedisClient) {
	cmdStr := c.args[0].strVal()
	if !c.isFake() {
		ulog.Info("process command: ", cmdStr)
	}

	// Case-insensitive
	cmdStr = strings.ToLower(cmdStr)
	if cmdStr == "quit" {
		freeClient(c)
		return
	}
	c.cmd = lookupCommand(cmdStr)
	// non-existent
	if c.cmd == nil {
		c.addReply(shared.unknowErr)
		resetClient(c)
		return
	}
	// check arity
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
	// try doReply,if c.reply.len > 0
	c.doReply()
	// aof
	dirty = server.dirty - dirty
	if dirty > 0 {
		c.cmd.propagate(c.args)
	}
}

// =================================== command ====================================

// 初始化命令列表
func initCommands() map[string]SRedisCommand {
	return map[string]SRedisCommand{
		//server
		PING: {PING, pingCommand, -1},
		// db
		EXPIRE:    {EXPIRE, expireCommand, 3},
		OBJECT:    {OBJECT, objectCommand, 3},
		KEYS:      {KEYS, keysCommand, 2},
		PERSIST:   {PERSIST, persistCommand, 2},
		TTL:       {TTL, ttlCommand, 2},
		PTTL:      {PTTL, pTtlCommand, 2},
		DEL:       {DEL, delCommand, -2},
		EXISTS:    {EXISTS, existsCommand, -2},
		RANDOMKEY: {RANDOMKEY, randomKeyCommand, 1},
		FLUSHDB:   {FLUSHDB, flushDbCommand, 1},
		TYPE:      {TYPE, typeCommand, 2},
		// aof
		BGREWRITEAOF: {BGREWRITEAOF, bgRewriteAofCommand, 1},
		// rdb
		SAVE:   {SAVE, saveCommand, 1},
		BGSAVE: {BGSAVE, bgSaveCommand, 1},
		// string
		GET:  {GET, getCommand, 2},
		SET:  {SET, setCommand, 3},
		INCR: {INCR, incrCommand, 2},
		DECR: {DECR, decrCommand, 2},
		// zset
		Z_ADD:   {Z_ADD, zAddCommand, -4},
		Z_RANGE: {Z_RANGE, zRangeCommand, -4},
		Z_CARD:  {Z_CARD, zCardCommand, 2},
		// set
		S_ADD:        {S_ADD, sAddCommand, -3},
		SMEMBERS:     {SMEMBERS, sinterCommand, 2},
		SINTER:       {SINTER, sinterCommand, -2},
		SINTER_STORE: {SINTER_STORE, sinterStoreCommand, -2},
		S_POP:        {S_POP, sPopCommand, -2},
		S_REM:        {S_REM, sRemCommand, -3},
		S_UNION:      {S_UNION, sUnionCommand, -2},
		S_UNIONSTORE: {S_UNIONSTORE, sUnionStoreCommand, -3},
		S_DIFF:       {S_DIFF, sDiffCommand, -2},
		S_DIFFSTORE:  {S_DIFFSTORE, sDiffStoreCommand, -3},
		S_CARD:       {S_DIFFSTORE, sCardCommand, 2},
		// list
		R_PUSH: {R_PUSH, rPushCommand, -3},
		L_PUSH: {L_PUSH, lPushCommand, -3},
		R_POP:  {R_POP, rPopCommand, 2},
		L_POP:  {L_POP, lPopCommand, 2},
		L_LEN:  {L_LEN, lLenCommand, 2},
		// hash
		H_SET:    {H_SET, hSetCommand, 4},
		H_GET:    {H_GET, hGetCommand, 3},
		H_DEL:    {H_DEL, hDelCommand, -3},
		H_EXISTS: {H_EXISTS, hExistsCommand, 3},
		H_LEN:    {H_LEN, hLenCommand, 2},
		H_KEYS:   {H_KEYS, hKeysCommand, 2},
		H_VALS:   {H_VALS, hValsCommand, 2},
		H_GETALL: {H_GETALL, hGetAllCommand, 2},
		// to be continued ! ! !
	}
}
