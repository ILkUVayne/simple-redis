package src

import "simple-redis/utils"

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

// run cronjob, default 100ms
func serverCron(el *aeEventLoop, id int, clientData any) {
	// TODO check expire key
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
			n, err := Write(c.fd, buf[c.sentLen:])
			if err != nil {
				freeClient(c)
				utils.ErrorP("simple-redis server: SendReplyToClient err: ", err)
				return
			}
			c.sentLen += n
			utils.InfoF("simple-redis server: send %v bytes to client:%v", n, c.fd)
			if c.sentLen != bufLen {
				break
			}
			c.reply.delNode(resp)
			resp.data.decrRefCount()
		}
	}
	if c.reply.len() == 0 {
		c.sentLen = 0
		el.removeFileEvent(c.fd, AE_WRITEABLE)
	}
}
