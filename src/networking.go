package src

import (
	"fmt"
	"github.com/ILkUVayne/utlis-go/v2/time"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"golang.org/x/sys/unix"
	"strconv"
)

// accept client connect
func acceptTcpHandler(el *aeEventLoop, fd int, _ any) {
	cfd := Accept(fd)
	if cfd == AE_ERR {
		ulog.Error("simple-redis server: Accepting client err: cfd == ", cfd)
	}
	client := createSRClient(cfd)
	server.clients[cfd] = client
	// 注册读取查询事件
	el.addFileEvent(cfd, AE_READABLE, readQueryFromClient, client)
}

// read client query and process
func readQueryFromClient(_ *aeEventLoop, fd int, clientData any) {
	c := assertClient(clientData)
	if (len(c.queryBuf) - c.queryLen) < SREDIS_MAX_BULK {
		c.queryBuf = append(c.queryBuf, make([]byte, SREDIS_MAX_BULK)...)
	}
	n, err := Read(fd, c.queryBuf)
	if err != nil {
		freeClient(c)
		ulog.ErrorPf("simple-redis server: client %v read err: %v", fd, err)
		return
	}
	c.queryLen += n
	err = processQueryBuf(c)
	if err != nil {
		freeClient(c)
		ulog.ErrorP("simple-redis server: process query buf err: ", err)
		return
	}
}

// SendReplyToClient send query result to client
func SendReplyToClient(el *aeEventLoop, _ int, clientData any) {
	c := assertClient(clientData)
	for !isEmpty(c.reply) {
		resp := c.reply.first()
		buf := []byte(resp.data.strVal())
		bufLen := len(buf)
		if c.sentLen < bufLen {
			//n, err := Write(c.fd, buf[c.sentLen:])
			_, err := Write(c.fd, buf[:])
			if err != nil {
				freeClient(c)
				ulog.ErrorP("simple-redis server: SendReplyToClient err: ", err)
				return
			}
			//c.sentLen += n
			//utils.InfoF("simple-redis server: send %v bytes to client:%v", n, c.fd)
			//if c.sentLen != bufLen {
			//	break
			//}
			c.reply.delNode(resp)
			resp.data.decrRefCount()
		}
	}
	if isEmpty(c.reply) {
		c.sentLen = 0
		el.removeFileEvent(c.fd, AE_WRITEABLE)
	}
}

// ======================= Cron: called every 100 ms ========================

// check some random expire key
func activeExpireCycle() {
	for i := 0; i < EXPIRE_CHECK_COUNT; i++ {
		if server.db.dbExpireSize() == 0 {
			break
		}

		entry := server.db.expireRandomKey()
		if entry == nil {
			break
		}
		when, _ := entry.val.intVal()
		server.db.expireIfNeeded1(when, entry.getKey())
	}
}

// check background persistence terminated
func checkPersistence() {
	if server.aofChildPid != -1 || server.rdbChildPid != -1 {
		pid, _ := wait4(-1, unix.WNOHANG)
		if pid != 0 && pid != -1 {
			if pid == server.aofChildPid {
				backgroundRewriteDoneHandler()
			}
			if pid == server.rdbChildPid {
				backgroundSaveDoneHandler()
			}
		}
		updateDictResizePolicy()
		return
	}

	// If there is not a background saving/rewrite in progress check if
	// we have to save/rewrite now
	for _, v := range server.saveParams {
		now := time.GetMsTime()
		if server.dirty > int64(v.changes) &&
			now-server.lastSave > int64(v.seconds) &&
			(now-server.lastBgSaveTry > REDIS_BGSAVE_RETRY_DELAY ||
				server.lastBgSaveStatus == REDIS_OK) {
			ulog.InfoF("%d changes in %d seconds. Saving...", v.changes, v.seconds)
			rdbSaveBackground()
			break
		}
	}

	// Trigger an AOF rewrite if needed
	if server.aofChildPid == -1 &&
		server.aofRewritePerc > 0 &&
		server.aofCurrentSize > server.aofRewriteMinSize {
		base := int64(1)
		if server.aofRewriteBaseSize > 0 {
			base = server.aofRewriteBaseSize
		}
		growth := (server.aofCurrentSize*100)/base - 100
		if growth > int64(server.aofRewritePerc) {
			ulog.InfoF("Starting automatic rewriting of AOF on %d% growth", growth)
			rewriteAppendOnlyFileBackground()
		}
	}
}

// database cronjob
//
// 1. check and delete some expire key.
// 2. try to resize hashtable.
// 3. try to Rehash.
func databaseCron() {
	// check expire key
	activeExpireCycle()
	if server.aofChildPid == -1 && server.rdbChildPid == -1 {
		// Resize
		tryResizeHashTables()
		// Rehash
		tryRehash()
	}
}

// 检查bgSave 或者 BGREWRITEAOF 执行事件是否超过阈值,
// 若超过了（可能是子进程死锁阻塞了），则手动发送中断信号kill子进程
func checkRdbOrAofExecTimeout() {
	// check rdb
	if server.rdbChildPid != -1 && (time.GetMsTime()-server.rdbStartTime) > C_PROC_MAX_TIME {
		sendKill(server.rdbChildPid)
		ulog.Info("rdb bgSave exec timeout, childPid = ", server.rdbChildPid)
		server.rdbChildPid = -1
	}
	// check aof
	if server.aofChildPid != -1 && (time.GetMsTime()-server.aofStartTime) > C_PROC_MAX_TIME {
		sendKill(server.aofChildPid)
		ulog.Info("aof bgRewriteAof exec timeout, childPid = ", server.aofChildPid)
		server.aofChildPid = -1
	}
}

// server cronjob, default 100ms
func serverCron(*aeEventLoop, int, any) {
	// database corn
	databaseCron()
	// flush aof_buf on disk
	flushAppendOnlyFile()
	// Check if a background saving or AOF rewrite in progress terminated.
	checkPersistence()
	// check bgSave or BGREWRITEAOF timeout
	checkRdbOrAofExecTimeout()
}

// ================================ addReply =================================

// set SRedisClient.replyReady = true and try to send reply
func (c *SRedisClient) doReply() {
	c.replyReady = true
	c.addReply(nil)
}

// 将查询结果添加到c.reply中,并创建SendReplyToClient事件
func (c *SRedisClient) addReply(data *SRobj) {
	c.pushReply(data, AL_START_TAIL)
	if c.replyReady && !c.isFake() && !isEmpty(c.reply) {
		server.el.addFileEvent(c.fd, AE_WRITEABLE, SendReplyToClient, c)
	}
}

// 查询结果添加到c.reply中
func (c *SRedisClient) addReplyStr(s string) {
	data := createSRobj(SR_STR, s)
	c.addReply(data)
	data.decrRefCount()
}

// 添加字符串错误返回
func (c *SRedisClient) addReplyError(err string) {
	if err != "" {
		c.addReplyStr(fmt.Sprintf(RESP_ERR, err))
	}
}

// 添加字符串错误(可格式化)返回
func (c *SRedisClient) addReplyErrorFormat(format string, a ...any) {
	err := fmt.Sprintf(format, a...)
	if err != "" {
		c.addReplyStr(fmt.Sprintf(RESP_ERR, err))
	}
}

// add bulk int to SRedisClient.reply and send reply.
//
// e.g. addReplyBulkInt(15) = "$2\r\n15\r\n"
func (c *SRedisClient) addReplyBulkInt(ll int64) {
	c.addReplyBulk(createSRobj(SR_STR, strconv.FormatInt(ll, 10)))
}

// 添加浮点数返回
func (c *SRedisClient) addReplyDouble(f float64) {
	str := formatFloat(f, 10)
	c.addReplyStr(fmt.Sprintf(RESP_BULK, len(str), str))
}

// 添加整数返回
func (c *SRedisClient) addReplyLongLong(ll int64) {
	c.addReplyStr(fmt.Sprintf(RESP_INT, ll))
}

// build length Prefix and add to SRedisClient.reply.
//
// e.g. addReplyLongLongWithPrefix(10, "*") = "*10\r\n"
func (c *SRedisClient) addReplyLongLongWithPrefix(ll int64, prefix string) {
	c.addReplyStr(prefix + strconv.FormatInt(ll, 10))
	c.addReply(shared.crlf)
}

// add array length to SRedisClient.reply, will not send reply if replyReady = false
func (c *SRedisClient) addReplyMultiBulkLen(length int64, replyReady bool) {
	c.replyReady = replyReady
	c.addReplyLongLongWithPrefix(length, "*")
}

// add bulk length to SRedisClient.reply
func (c *SRedisClient) addReplyBulkLen(data *SRobj) {
	c.addReplyLongLongWithPrefix(int64(len(data.strVal())), "$")
}

// add bulk message to SRedisClient.reply and send reply
func (c *SRedisClient) addReplyBulk(data *SRobj) {
	c.addReplyBulkLen(data)
	c.addReplyStr(fmt.Sprintf("%s", data.strVal()))
	c.addReply(shared.crlf)
}

// add status reply.
//
// e.g. addReplyStatus("success") = "+success\r\n"
func (c *SRedisClient) addReplyStatus(s string) {
	c.addReplyStr(fmt.Sprintf("+%s", s))
	c.addReply(shared.crlf)
}

// add deferred array length to SRedisClient.reply, will not send reply.
// use with setDeferredMultiBulkLength.
//
// e.g. addDeferredMultiBulkLength() => c.reply.first().data = nil
func (c *SRedisClient) addDeferredMultiBulkLength() *node {
	c.replyReady = false
	c.pushReply(createSRobj(SR_STR, nil), AL_START_TAIL)
	return c.reply.first()
}

// set deferred array length to SRedisClient.reply. use with addDeferredMultiBulkLength
func (c *SRedisClient) setDeferredMultiBulkLength(n *node, length int) {
	if n != nil {
		n.data = createSRobj(SR_STR, fmt.Sprintf("*%d\r\n", length))
	}
}
