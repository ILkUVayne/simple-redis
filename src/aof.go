package src

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"simple-redis/utils"
)

//-----------------------------------------------------------------------------
// aof loading
//-----------------------------------------------------------------------------

func createFakeClient() *SRedisClient {
	c := new(SRedisClient)
	c.fd = -1
	c.db = server.db
	c.cmdTyp = CMD_UNKNOWN
	c.reply = listCreate(&listType{keyCompare: SRStrCompare})
	c.queryBuf = make([]byte, SREDIS_IO_BUF)
	c.replyReady = true
	return c
}

func checkExpire(args []*SRobj) int {
	if args[0].strVal() != EXPIRE {
		return REDIS_OK
	}
	intVal, _ := args[2].intVal()
	if intVal < utils.GetMsTime() {
		server.db.data.dictDelete(args[1])
		server.db.expire.dictDelete(args[1])
		return REDIS_ERR
	}
	return REDIS_OK
}

func loadAppendOnlyFile(name string) {
	fp, err := os.Open(name)
	if err != nil {
		utils.Error("Can't open the append-only file: ", err)
	}
	defer func() { _ = fp.Close() }()

	scanner := bufio.NewScanner(fp)
	fakeClient := createFakeClient()
	var args []*SRobj
	aLen := int64(-1)
	for scanner.Scan() {
		if aLen <= 0 {
			args = make([]*SRobj, 0)
		}

		str := scanner.Text()
		if str[0] == '*' {
			str = str[1:]
			if utils.String2Int64(&str, &aLen) == REDIS_ERR {
				utils.Error("Bad file format reading the append only file")
			}
			continue
		}
		if str[0] == '$' {
			continue
		}
		args = append(args, createSRobj(SR_STR, str))
		aLen--
		if aLen == 0 {
			if checkExpire(args) == REDIS_ERR {
				args = nil
				continue
			}
			fakeClient.args = args
			processCommand(fakeClient)
			fakeClient.args = nil
			args = nil
		}
	}
}

//-----------------------------------------------------------------------------
// AOF file implementation
//-----------------------------------------------------------------------------

func flushAppendOnlyFile() {
	if len(server.aofBuf) == 0 {
		return
	}
	_, err := io.WriteString(server.aofFd, server.aofBuf)
	if err != nil {
		utils.Error("flushAppendOnlyFile err: ", err)
	}
	server.aofBuf = ""
}

func catAppendOnlyGenericCommand(argc int, args []*SRobj) string {
	buf := fmt.Sprintf(RESP_ARRAY, argc)
	for i := 0; i < argc; i++ {
		o := args[i].getDecodedObject()
		buf += fmt.Sprintf(RESP_BULK, len(o.strVal()), o.strVal())
		o.decrRefCount()
	}
	return buf
}

func (cmd *SRedisCommand) catAppendOnlyExpireAtCommand(buf string, key *SRobj) string {
	//seconds = seconds.getDecodedObject()
	//when, _ := seconds.intVal()
	val := server.db.expireGet(key)
	// key expire
	if val == nil {
		server.db.dbDel(key)
		return ""
	}
	args := make([]*SRobj, 3)
	args[0] = createSRobj(SR_STR, "expire")
	args[1] = key
	args[2] = createSRobj(SR_STR, val.strVal())
	buf = catAppendOnlyGenericCommand(3, args)
	args[0].decrRefCount()
	args[2].decrRefCount()
	return buf
}

func (cmd *SRedisCommand) feedAppendOnlyFile(args []*SRobj, argc int) {
	var buf string

	if cmd.name == EXPIRE {
		buf = cmd.catAppendOnlyExpireAtCommand(buf, args[1])
	} else {
		buf = catAppendOnlyGenericCommand(argc, args)
	}

	if server.aofState == REDIS_AOF_ON {
		server.aofBuf += buf
	}
}
