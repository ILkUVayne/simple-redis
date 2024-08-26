package src

import (
	"errors"
	"fmt"
)

// server 返回的数据结构
type sRedisReply struct {
	typ    int    // respType
	buf    []byte // server reply buff
	str    string // server reply string
	fStr   string // server reply format string
	length int
}

// Format server reply string by sRedisReply.typ
func (r *sRedisReply) strFormat() {
	r.fStr = strFormatHandle(r)
}

type sRedisContext struct {
	fd     int    // cli connect fd
	oBuf   []byte // cli send args buff
	reader any
	err    error
}

// build bulk command from args.
//
// e.g. args = ["get", "name"], sRedisAppendCommandArg(c, args) => "*2\r\n$3\r\nget\r\n$4\r\nname\r\n"
func sRedisAppendCommandArg(c *sRedisContext, args []string) {
	// Format args,Compliant with resp specifications
	cmd := fmt.Sprintf("*%d\r\n", len(args))
	for _, v := range args {
		cmd += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
	}
	c.oBuf = []byte(cmd)
}

// parse server response if complete
func getReply(c *sRedisContext, reply *sRedisReply) int {
	if reply.typ == 0 {
		typ := getRespType(reply.buf[0])
		if typ == -1 {
			c.err = errors.New("invalid response from server, invalid resp type")
			return CLI_ERR
		}
		reply.typ = typ
	}

	str, err := respParseHandle(reply)
	if err != nil {
		c.err = err
		return CLI_ERR
	}
	if str != "" {
		reply.str = str
		reply.strFormat()
	}
	return CLI_OK
}

// send command to server and read response
func sRedisGetReply(c *sRedisContext, reply *sRedisReply) int {
	// Write
	wLen := len(c.oBuf)
	sendLen := 0
	for {
		n, err := Write(c.fd, c.oBuf[sendLen:])
		if err != nil {
			c.err = err
			return CLI_ERR
		}
		if n == wLen {
			break
		}
		sendLen += n
	}
	// Read
	reply.buf = make([]byte, SREDIS_IO_BUF)
	for reply.str == "" {
		n, err := Read(c.fd, reply.buf[reply.length:])
		if err != nil {
			c.err = err
			return CLI_ERR
		}
		reply.length += n
		if (len(reply.buf) - reply.length) < SREDIS_MAX_BULK {
			reply.buf = append(reply.buf, make([]byte, SREDIS_MAX_BULK)...)
		}
		if getReply(c, reply) == CLI_ERR {
			return CLI_ERR
		}
	}
	return CLI_OK
}
