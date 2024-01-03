package src

import (
	"errors"
	"strconv"
	"strings"
)

type SRedisClient struct {
	fd       int
	db       *SRedisDB
	args     []*SRobj
	reply    *list
	queryBuf []byte
	queryLen int
	sentLen  int
	cmdTyp   CmdType
	bulkNum  int
	bulkLen  int
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

func createSRClient(fd int) *SRedisClient {
	c := new(SRedisClient)
	c.fd = fd
	c.db = server.db
	c.cmdTyp = CMD_UNKNOWN
	c.reply = listCreate(&listType{keyCompare: SRStrCompare})
	c.queryBuf = make([]byte, SREDIS_IO_BUF)
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
	Close(c.fd)
}

func resetClient(c *SRedisClient) {
	freeArgs(c)
	c.cmdTyp = CMD_UNKNOWN
	c.bulkLen = 0
	c.bulkNum = 0
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
		// get cmd type
		if c.cmdTyp == CMD_UNKNOWN {
			c.cmdTyp = CMD_INLINE
			if c.queryBuf[0] == '*' {
				c.cmdTyp = CMD_BULK
			}
		}
		var ok bool
		var err error
		switch c.cmdTyp {
		case CMD_INLINE:
			ok, err = inlineBufHandle(c)
		case CMD_BULK:
			ok, err = bulkBufHandle(c)
		default:
			return errors.New("unknow cmd type")
		}
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
