package src

import (
	"fmt"
	"simple-redis/utils"
	"strconv"
)

// accept client connect
func acceptTcpHandler(el *aeEventLoop, fd int, clientData any) {
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
func readQueryFromClient(el *aeEventLoop, fd int, clientData any) {
	c := clientData.(*SRedisClient)
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
func SendReplyToClient(el *aeEventLoop, fd int, clientData any) {
	c := clientData.(*SRedisClient)
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

func activeExpireCycle() {
	for i := 0; i < EXPIRE_CHECK_COUNT; i++ {
		if server.db.expire.dictSize() == 0 {
			break
		}

		entry := server.db.expire.dictGetRandomKey()
		if entry == nil {
			break
		}
		intVal, _ := entry.val.intVal()
		if intVal < utils.GetMsTime() {
			server.db.data.dictDelete(entry.key)
			server.db.expire.dictDelete(entry.key)
		}
	}
}

// run cronjob, default 100ms
func serverCron(el *aeEventLoop, id int, clientData any) {
	// check expire key
	activeExpireCycle()
	// flush aof_buf on disk
	flushAppendOnlyFile()
}

// ================================ addReply =================================

func (c *SRedisClient) doReply() {
	c.replyReady = true
	c.addReply(nil)
}

// 将查询结果添加到c.reply中,并创建SendReplyToClient事件
func (c *SRedisClient) addReply(data *SRobj) {
	c.pushReply(data, "r")
	if c.replyReady && c.fd > 0 {
		server.el.addFileEvent(c.fd, AE_WRITEABLE, SendReplyToClient, c)
	}
}

// 查询结果添加到c.reply中
func (c *SRedisClient) addReplyStr(s string) {
	data := createSRobj(SR_STR, s)
	c.addReply(data)
	data.decrRefCount()
}

func (c *SRedisClient) addReplyError(err string) {
	if err == "" {
		return
	}
	c.addReplyStr(fmt.Sprintf(RESP_ERR, err))
}

func (c *SRedisClient) addReplyDouble(f float64) {
	str := strconv.FormatFloat(f, 'f', 2, 64)
	c.addReplyStr(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str))
}

func (c *SRedisClient) addReplyLongLong(ll int) {
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
	c.pushReply(createSRobj(SR_STR, nil), "r")
	return c.reply.first()
}

func (c *SRedisClient) setDeferredMultiBulkLength(n *node, length int) {
	if n == nil {
		return
	}
	n.data = createSRobj(SR_STR, fmt.Sprintf("*%d\r\n", length))
	c.doReply()
}
