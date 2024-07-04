package src

import (
	"errors"
	"fmt"
)

type sRedisReply struct {
	typ    int    // respType
	buf    []byte // server reply buff
	str    string // server reply string
	fStr   string // server reply format string
	length int
}

func (r *sRedisReply) strFormat() {
	r.fStr = strFormatHandle(r)
}

type sRedisContext struct {
	fd     int    // cli connect fd
	oBuf   []byte // cli send args buff
	reader any
	err    error
}

func sRedisContextInit() *sRedisContext {
	c := new(sRedisContext)
	return c
}

func __sRedisAppendCommand(c *sRedisContext, cmd *string) {
	c.oBuf = []byte(*cmd)
}

func sRedisAppendCommandArg(c *sRedisContext, args []string) {
	var cmd string
	sRedisFormatCommandArg(&cmd, args)
	__sRedisAppendCommand(c, &cmd)
}

// Format args,Compliant with resp specifications
func sRedisFormatCommandArg(target *string, args []string) {
	cmd := fmt.Sprintf("*%d\r\n", len(args))
	for _, v := range args {
		cmd += fmt.Sprintf("$%d\r\n", len(v))
		cmd += fmt.Sprintf("%s\r\n", v)
	}
	*target = cmd
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
