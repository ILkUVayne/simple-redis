package src

import (
	"errors"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"strconv"
	"strings"
)

// SRedisClient 客户端结构
type SRedisClient struct {
	fd            int       // 客户端fd, 当fd等于 FAKE_CLIENT_FD 是fakeClient
	db            *SRedisDB // 数据库指针
	authenticated bool      // 当前客户端密码验证状态，默认未验证（false）
	args          []*SRobj  // command args
	reply         *list     // reply data
	replyReady    bool      // 响应数据是否准备完毕
	queryBuf      []byte
	queryLen      int
	sentLen       int
	cmd           *SRedisCommand // 客户端需要执行的命令
	cmdTyp        CmdType        // unknown inline bulk
	bulkNum       int
	bulkLen       int

	pubSubChannels *dict
}

// return true if is fake client
func (c *SRedisClient) isFake() bool {
	return c.fd == FAKE_CLIENT_FD
}

// return complete command slice
func (c *SRedisClient) completeCommand() []string {
	if len(c.args) == 0 {
		return nil
	}
	cmdStr := make([]string, 0)
	for _, v := range c.args {
		cmdStr = append(cmdStr, v.strVal())
	}
	return cmdStr
}

// e.g. "$3\r\nget\r\n$4\r\nname\r\n"
// n == 3
// string(queryBuf) == "get\r\n$4\r\nname\r\n"
func (c *SRedisClient) getQueryNum() (int, error) {
	idx, err := getQueryLine(c.queryBuf[:c.queryLen], c.queryLen)
	if idx < 0 {
		return 0, err
	}
	n, err := strconv.Atoi(string(c.queryBuf[1:idx]))
	c.queryBuf = c.queryBuf[idx+2:]
	c.queryLen -= idx + 2
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
		ulog.Error("invalid push type: ", where)
	}
	data.incrRefCount()
}

// 重写 client 的 args 和 cmd
func (c *SRedisClient) rewriteClientCommandVector(args ...*SRobj) {
	c.args = args
	args[0].incrRefCount()
	c.cmd = lookupCommand(strings.ToLower(args[0].strVal()))
}

func (c *SRedisClient) selectDb(id int64) bool {
	// only 1 db now
	ulog.InfoF("select db: %d now", id)
	return true
}

// return SRClient
//
// fd: client accept fd
func createSRClient(fd int) *SRedisClient {
	return &SRedisClient{
		fd:         fd,
		db:         server.db,
		cmdTyp:     CMD_UNKNOWN,
		reply:      listCreate(),
		queryBuf:   make([]byte, SREDIS_IO_BUF),
		replyReady: true,
	}
}

// return a fake client,client.fd == FAKE_CLIENT_FD
func createFakeClient() *SRedisClient {
	fakeClient := createSRClient(FAKE_CLIENT_FD)
	// fake客户端不需要验证密码，默认true
	fakeClient.authenticated = true
	return fakeClient
}

// free SRedisClient.reply
func freeReplyList(c *SRedisClient) {
	for c.reply.length != 0 {
		n := c.reply.head
		c.reply.delNode(n)
		n.data.decrRefCount()
	}
}

// free client
func freeClient(c *SRedisClient) {
	c.args = nil
	delete(server.clients, c.fd)
	server.el.removeFileEvent(c.fd, AE_READABLE)
	server.el.removeFileEvent(c.fd, AE_WRITEABLE)
	freeReplyList(c)
	if !c.isFake() {
		Close(c.fd)
	}
}

func resetClient(c *SRedisClient) {
	c.args = nil
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
	idx, err := getQueryLine(c.queryBuf[:c.queryLen], c.queryLen)
	if idx < 0 {
		return false, err
	}
	// 分割字符串
	strs := splitArgs(string(c.queryBuf[:idx]))
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
		// get bulkNum
		n, err := c.getQueryNum()
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
			if c.queryBuf[0] != '$' {
				return false, errors.New("expect $ for bulk")
			}
			bulkLen, err := c.getQueryNum()
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
