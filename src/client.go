package src

import (
	"errors"
	"simple-redis/utils"
	"strconv"
	"strings"
)

// SRedisClient 客户端结构
type SRedisClient struct {
	fd         int       // 客户端fd, 当fd等于 FAKE_CLIENT_FD 是fakeClient
	db         *SRedisDB // 数据库指针
	args       []*SRobj  // command args
	reply      *list     // reply data
	replyReady bool      // 响应数据是否准备完毕
	queryBuf   []byte
	queryLen   int
	sentLen    int
	cmd        *SRedisCommand // 客户端需要执行的命令
	cmdTyp     CmdType        // unknown inline bulk
	bulkNum    int
	bulkLen    int
}

// return true if is fake client
func (c *SRedisClient) isFake() bool {
	return c.fd == FAKE_CLIENT_FD
}

// getQueryLine
// e.g. "get name\r\n"
// idx == 8
// e.g. "$3\r\nget\r\n$4\r\nname\r\n"
// idx == 2
func (c *SRedisClient) getQueryLine() (int, error) {
	idx := strings.Index(string(c.queryBuf[:c.queryLen]), "\r\n")
	if idx < 0 && c.queryLen > SREDIS_MAX_INLINE {
		return idx, errors.New("inline cmd is too long")
	}
	return idx, nil
}

// getQueryNum
// e.g. "$3\r\nget\r\n$4\r\nname\r\n"
// n == 3
// string(queryBuf) == "get\r\n$4\r\nname\r\n"
func (c *SRedisClient) getQueryNum(start, end int) (int, error) {
	n, err := strconv.Atoi(string(c.queryBuf[start:end]))
	c.queryBuf = c.queryBuf[end+2:]
	c.queryLen -= end + 2
	return n, err
}

// append reply data to client.reply,when fake client while do nothing
//
// typ: AL_START_HEAD  AL_START_TAIL
func (c *SRedisClient) pushReply(data *SRobj, where int) {
	if c.isFake() || data == nil {
		return
	}
	switch where {
	case AL_START_TAIL:
		c.reply.rPush(data)
	case AL_START_HEAD:
		c.reply.lPush(data)
	default:
		utils.Error("invalid push type: ", where)
	}
	data.incrRefCount()
}

// return SRClient
//
// fd: client accept fd
func createSRClient(fd int) *SRedisClient {
	c := new(SRedisClient)
	c.fd = fd
	c.db = server.db
	c.cmdTyp = CMD_UNKNOWN
	c.reply = listCreate(&lType)
	c.queryBuf = make([]byte, SREDIS_IO_BUF)
	c.replyReady = true
	return c
}

func freeArgs(c *SRedisClient) {
	for _, arg := range c.args {
		arg.decrRefCount()
	}
}

func freeReplyList(c *SRedisClient) {
	for c.reply.length != 0 {
		n := c.reply.head
		c.reply.delNode(n)
		n.data.decrRefCount()
	}
}

func freeClient(c *SRedisClient) {
	freeArgs(c)
	delete(server.clients, c.fd)
	server.el.removeFileEvent(c.fd, AE_READABLE)
	server.el.removeFileEvent(c.fd, AE_WRITEABLE)
	freeReplyList(c)
	if !c.isFake() {
		Close(c.fd)
	}
}

func resetClient(c *SRedisClient) {
	freeArgs(c)
	c.cmd = nil
	c.cmdTyp = CMD_UNKNOWN
	c.bulkLen = 0
	c.bulkNum = 0
}

// 检查并设置cmdTyp
func checkCmdType(c *SRedisClient) {
	if c.cmdTyp != CMD_UNKNOWN {
		return
	}
	c.cmdTyp = CMD_INLINE
	if c.queryBuf[0] == '*' {
		c.cmdTyp = CMD_BULK
	}
}

// inline command handle
// e.g. "get name\r\n"
func inlineBufHandle(c *SRedisClient) (bool, error) {
	idx, err := c.getQueryLine()
	if idx < 0 {
		return false, err
	}
	// 通过空格分割字符串
	strs := strings.Split(string(c.queryBuf[:idx]), " ")
	c.queryBuf = c.queryBuf[idx+2:]
	c.queryLen -= idx + 2
	c.args = make([]*SRobj, len(strs))
	for i, v := range strs {
		c.args[i] = createSRobj(SR_STR, v)
	}
	return true, nil
}

// bulk command handle
// e.g. "*2\r\n$3\r\nget\r\n$4\r\nname\r\n"
// bulkNum == 2
func bulkBufHandle(c *SRedisClient) (bool, error) {
	if c.bulkNum == 0 {
		idx, err := c.getQueryLine()
		if idx < 0 {
			return false, err
		}
		// get bulkNum
		n, err := c.getQueryNum(1, idx)
		if err != nil {
			return false, err
		}
		if n == 0 {
			return true, nil
		}
		c.bulkNum = n
		c.args = make([]*SRobj, n)
	}
	// get command
	for c.bulkNum > 0 {
		// get bulkLen
		if c.bulkLen == 0 {
			idx, err := c.getQueryLine()
			if idx < 0 {
				return false, err
			}
			if c.queryBuf[0] != '$' {
				return false, errors.New("expect $ for bulk")
			}
			bulkLen, err := c.getQueryNum(1, idx)
			if bulkLen == 0 || err != nil {
				return false, err
			}
			if bulkLen > SREDIS_MAX_BULK {
				return false, errors.New("bulk is too long")
			}
			c.bulkLen = bulkLen
		}
		// read bulk string
		if c.queryLen < c.bulkLen+2 {
			return false, nil
		}
		idx := c.bulkLen
		if c.queryBuf[idx] != '\r' || c.queryBuf[idx+1] != '\n' {
			return false, errors.New("expect CRLF for bulk")
		}

		c.args[len(c.args)-c.bulkNum] = createSRobj(SR_STR, string(c.queryBuf[:idx]))
		c.queryBuf = c.queryBuf[idx+2:]
		c.queryLen -= idx + 2
		c.bulkLen = 0
		c.bulkNum -= 1
	}
	return true, nil
}

// query to args and processCommand
func processQueryBuf(c *SRedisClient) error {
	for c.queryLen > 0 {
		ok, err := cmdBufHandle(c)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		if len(c.args) == 0 {
			resetClient(c)
			continue
		}
		processCommand(c)
	}
	return nil
}
