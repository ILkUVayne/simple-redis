package src

import (
	"errors"
	"fmt"
)

const (
	NIL_STR = "(nil)"
)

type sRedisReply struct {
	typ    int
	buf    []byte
	str    string
	fStr   string
	length int
}

func (r *sRedisReply) formatStr() {
	switch r.typ {
	case BULK_STR:
		if r.str != NIL_STR {
			r.fStr = fmt.Sprintf("\"%s\"", r.str)
		}
	case SIMPLE_ERROR:
		r.fStr = fmt.Sprintf("(error) %s", r.str)
	case INTEGERS:
		r.fStr = fmt.Sprintf("(integer) %s", r.str)
	}
}

type sRedisContext struct {
	fd     int
	obuf   []byte
	reader any
	err    error
}

func sRedisContextInit() *sRedisContext {
	c := new(sRedisContext)
	return c
}

func __sRedisAppendCommand(c *sRedisContext, cmd *string) {
	c.obuf = []byte(*cmd)
}

func sRedisAppendCommandArg(c *sRedisContext, args []string) {
	var cmd string
	sRedisFormatCommandArg(&cmd, args)
	__sRedisAppendCommand(c, &cmd)
}

func sRedisFormatCommandArg(target *string, args []string) {
	var cmd string
	cmd = fmt.Sprintf("*%d\r\n", len(args))
	for _, v := range args {
		cmd += fmt.Sprintf("$%d\r\n", len(v))
		cmd += fmt.Sprintf("%s\r\n", v)
	}
	*target = cmd
}

func getReply(c *sRedisContext, reply *sRedisReply) int {
	if reply.typ == 0 {
		typ := getRespType(reply.buf[0])
		if typ == -1 {
			c.err = errors.New("invalid resp type")
			return CLI_ERR
		}
		reply.typ = typ
	}
	//
	str, err := respParseFuncs[reply.typ](reply.buf, reply.length)
	if err != nil {
		c.err = err
		return CLI_ERR
	}
	if str != "" {
		reply.str = str
		reply.formatStr()
	}
	return CLI_OK
}

func sRedisGetReply(c *sRedisContext, reply *sRedisReply) int {
	// Write
	wLen := len(c.obuf)
	sendLen := 0
	for {
		n, err := Write(c.fd, c.obuf[sendLen:])
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
		getReply(c, reply)
	}
	return CLI_OK
}
