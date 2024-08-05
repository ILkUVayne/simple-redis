package src

import (
	"fmt"
	"github.com/ILkUVayne/utlis-go/v2/time"
	"golang.org/x/sys/unix"
	"simple-redis/utils"
	"strconv"
)

// accept client connect
func acceptTcpHandler(el *aeEventLoop, fd int, _ any) {
	cfd := Accept(fd)
	if cfd == AE_ERR {
		utils.Error("simple-redis server: Accepting client err: cfd == ", cfd)
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
		utils.ErrorPf("simple-redis server: client %v read err: %v", fd, err)
		return
	}
	c.queryLen += n
	err = processQueryBuf(c)
	if err != nil {
		freeClient(c)
		utils.ErrorP("simple-redis server: process query buf err: ", err)
		return
	}
}

// SendReplyToClient send query result to client
func SendReplyToClient(el *aeEventLoop, _ int, clientData any) {
	c := assertClient(clientData)
	for c.reply.len() > 0 {
		resp := c.reply.first()
		buf := []byte(resp.data.strVal())
		bufLen := len(buf)
		if c.sentLen < bufLen {
			//n, err := Write(c.fd, buf[c.sentLen:])
			_, err := Write(c.fd, buf[:])
			if err != nil {
				freeClient(c)
				utils.ErrorP("simple-redis server: SendReplyToClient err: ", err)
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
	if c.reply.len() == 0 {
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
			utils.InfoF("%d changes in %d seconds. Saving...", v.changes, v.seconds)
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
			utils.InfoF("Starting automatic rewriting of AOF on %d% growth", growth)
			rewriteAppendOnlyFileBackground()
		}
	}
}

func databaseCorn() {
	// check expire key
	activeExpireCycle()
	if server.aofChildPid == -1 && server.rdbChildPid == -1 {
		// Resize
		tryResizeHashTables()
		// Rehash
		tryRehash()
	}
}

// run cronjob, default 100ms
func serverCron(*aeEventLoop, int, any) {
	// database corn
	databaseCorn()
	// flush aof_buf on disk
	flushAppendOnlyFile()
	// Check if a background saving or AOF rewrite in progress terminated.
	checkPersistence()
}

// ================================ addReply =================================

func (c *SRedisClient) doReply() {
	c.replyReady = true
	c.addReply(nil)
}

// 将查询结果添加到c.reply中,并创建SendReplyToClient事件
func (c *SRedisClient) addReply(data *SRobj) {
	c.pushReply(data, AL_START_TAIL)
	if c.replyReady && !c.isFake() {
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

// 添加浮点数返回
func (c *SRedisClient) addReplyDouble(f float64) {
	str := strconv.FormatFloat(f, 'f', 2, 64)
	c.addReplyStr(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str))
}

// 添加整数返回
func (c *SRedisClient) addReplyLongLong(ll int64) {
	if ll == 0 {
		c.addReply(shared.czero)
		return
	}
	if ll == 1 {
		c.addReply(shared.cone)
		return
	}
	c.addReplyStr(fmt.Sprintf(":%d\r\n", ll))
}

func (c *SRedisClient) addReplyLongLongWithPrefix(ll int64, prefix string) {
	str := prefix + strconv.FormatInt(ll, 10) + "\r\n"
	c.addReplyStr(str)
}

func (c *SRedisClient) addReplyMultiBulkLen(length int64) {
	c.addReplyLongLongWithPrefix(length, "*")
}

func (c *SRedisClient) addReplyBulkLen(data *SRobj) {
	c.addReplyLongLongWithPrefix(int64(len(data.strVal())), "$")
}

func (c *SRedisClient) addReplyBulk(data *SRobj) {
	c.addReplyBulkLen(data)
	c.addReplyStr(fmt.Sprintf("%s", data.strVal()))
	c.addReply(shared.crlf)
}

func (c *SRedisClient) addReplyBulkInt(ll int64) {
	c.addReplyBulk(createSRobj(SR_STR, strconv.FormatInt(ll, 10)))
}

func (c *SRedisClient) addReplyStatus(s string) {
	c.addReplyStr(fmt.Sprintf("+%s\r\n", s))
}

func (c *SRedisClient) addDeferredMultiBulkLength() *node {
	c.replyReady = false
	c.pushReply(createSRobj(SR_STR, nil), AL_START_TAIL)
	return c.reply.first()
}

func (c *SRedisClient) setDeferredMultiBulkLength(n *node, length int) {
	if n != nil {
		n.data = createSRobj(SR_STR, fmt.Sprintf("*%d\r\n", length))
		c.doReply()
	}
}
