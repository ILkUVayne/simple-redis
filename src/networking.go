package src

import "simple-redis/utils"

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
