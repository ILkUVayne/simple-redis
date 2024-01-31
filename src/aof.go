package src

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"simple-redis/utils"
	"strconv"
)

func aofRewriteBufferWrite(f *os.File) int {
	if len(server.aofRewriteBufBlocks) == 0 {
		return 0
	}
	n, err := io.WriteString(f, server.aofRewriteBufBlocks)
	if err != nil {
		utils.ErrorP("aofRewriteBufferWrite err: ", err)
		return -1
	}
	server.aofRewriteBufferReset()
	return n
}

func (s *SRedisServer) aofBufReset() {
	if len(s.aofBuf) == 0 {
		return
	}
	s.aofBuf = ""
}

func (s *SRedisServer) aofRewriteBufferReset() {
	if len(s.aofRewriteBufBlocks) == 0 {
		return
	}
	s.aofRewriteBufBlocks = ""
}

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
	aofUpdateCurrentSize()
	server.aofRewriteBaseSize = server.aofCurrentSize
	freeClient(fakeClient)
}

//-----------------------------------------------------------------------------
// AOF file implementation
//-----------------------------------------------------------------------------

func flushAppendOnlyFile() {
	if len(server.aofBuf) == 0 {
		return
	}
	n, err := io.WriteString(server.aofFd, server.aofBuf)
	if err != nil {
		utils.Error("flushAppendOnlyFile err: ", err)
	}
	server.aofCurrentSize += int64(n)
	server.aofBufReset()
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

	if server.aofChildPid != -1 {
		server.aofRewriteBufBlocks += buf
	}
}

// ----------------------------------------------------------------------------
// AOF rewrite
// ----------------------------------------------------------------------------

func aofUpdateCurrentSize() {
	fInfo, err := server.aofFd.Stat()
	if err != nil {
		utils.Error("Unable to obtain the AOF file length. stat: ", err)
	}
	server.aofCurrentSize = fInfo.Size()
}

func rewrite(f *os.File, s *string) {
	_, err := io.WriteString(f, *s)
	if err != nil {
		utils.Error("rewriteStringObject err: ", err)
	}
}

func rewriteBulkObject(f *os.File, val *SRobj) {
	strVal := val.strVal()
	cmd := fmt.Sprintf(RESP_BULK, len(strVal), strVal)
	rewrite(f, &cmd)
}

func rewriteObject(f *os.File, cmd *string, val ...*SRobj) {
	if cmd != nil {
		rewrite(f, cmd)
	}

	if len(val) == 0 {
		return
	}
	for _, v := range val {
		rewriteBulkObject(f, v)
	}
}

func getItems(items int) int {
	cmdItems := items
	if items > REDIS_AOF_REWRITE_ITEMS_PER_CMD {
		cmdItems = REDIS_AOF_REWRITE_ITEMS_PER_CMD
	}
	return cmdItems
}

func checkItems(count, items *int) {
	*count++
	if *count == REDIS_AOF_REWRITE_ITEMS_PER_CMD {
		*count = 0
	}
	*items--
}

func rewriteStringObject(f *os.File, key, val *SRobj) {
	cmd := RESP_STR
	// set key value
	rewriteObject(f, &cmd, key, val)
}

func rewriteExpireObject(f *os.File, key, val *SRobj) {
	cmd := RESP_EXPIRE
	// expire key expireTime
	rewriteObject(f, &cmd, key, val)
}

func rewriteListObject(f *os.File, key, val *SRobj) {
	count, items := 0, listTypeLength(val)

	if val.encoding == REDIS_ENCODING_LINKEDLIST {
		l := val.Val.(*list)
		li := l.listRewind()
		for ln := li.listNext(); ln != nil; ln = li.listNext() {
			eleObj := ln.nodeValue()
			if count == 0 {
				cmd := fmt.Sprintf(RESP_LIST_RPUSH, 2+getItems(items))
				// add key
				rewriteObject(f, &cmd, key)
			}
			// add val
			rewriteObject(f, nil, eleObj)
			checkItems(&count, &items)
		}
		return
	}
	panic("Unknown list encoding")
}

func rewriteSetObject(f *os.File, key, val *SRobj) {
	count, items := 0, int(setTypeSize(val))

	if val.encoding == REDIS_ENCODING_INTSET {
		var intVal int64
		for ii := 0; val.Val.(*intSet).intSetGet(uint32(ii), &intVal); ii++ {
			if count == 0 {
				cmd := fmt.Sprintf(RESP_SET, 2+getItems(items))
				// add key
				rewriteObject(f, &cmd, key)
			}
			// add val
			rewriteObject(f, nil, createFromInt(intVal))
			checkItems(&count, &items)
		}
		return
	}
	if val.encoding == REDIS_ENCODING_HT {
		di := val.Val.(*dict).dictGetIterator()
		for de := di.dictNext(); de != nil; de = di.dictNext() {
			eleObj := de.getKey()
			if count == 0 {
				cmd := fmt.Sprintf(RESP_SET, 2+getItems(items))
				// add key
				rewriteObject(f, &cmd, key)
			}
			// add val
			rewriteObject(f, nil, eleObj)
			checkItems(&count, &items)
		}
		di.dictReleaseIterator()
		return
	}
	panic("Unknown set encoding")
}

func rewriteZSetObject(f *os.File, key, val *SRobj) {
	count, items := 0, int(zSetLength(val))

	if val.encoding == REDIS_ENCODING_SKIPLIST {
		zs := val.Val.(*zSet)
		di := zs.d.dictGetIterator()
		for de := di.dictNext(); de != nil; de = di.dictNext() {
			eleObj := de.getKey()
			score := de.getVal()

			if count == 0 {
				cmd := fmt.Sprintf(RESP_ZSET, 2+getItems(items)*2)
				// add key
				rewriteObject(f, &cmd, key)
			}
			sf, _ := score.floatVal()
			str := strconv.FormatFloat(sf, 'f', 2, 64)
			// add zSetScore and zSetVal
			rewriteObject(f, nil, createSRobj(SR_STR, str), eleObj)
			checkItems(&count, &items)
		}
		di.dictReleaseIterator()
		return
	}
	panic("Unknown sorted zset encoding")
}

func rewriteDictObject(f *os.File, key, val *SRobj) {
	if val.encoding == REDIS_ENCODING_HT {
		di := val.Val.(*dict).dictGetIterator()
		for de := di.dictNext(); de != nil; de = di.dictNext() {
			cmd := fmt.Sprintf(RESP_HASH_HSET, 4)
			// add key hashKey hashVal
			rewriteObject(f, &cmd, key, de.getKey(), de.getVal())
		}
		di.dictReleaseIterator()
		return
	}
	panic("Unknown hash encoding")
}

func rewriteAppendOnlyFile(filename string) int {
	tmpFile := persistenceFile(fmt.Sprintf("temp-rewriteaof-%d.aof", os.Getpid()))
	now := utils.GetMsTime()
	f, err := os.Create(tmpFile)
	if err != nil {
		utils.Error("Opening the temp file for AOF rewrite in rewriteAppendOnlyFile(): ", err)
	}
	defer func() { _ = f.Close() }()

	di := server.db.data.dictGetIterator()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		key := de.getKey()
		val := de.getVal()
		expireTime := server.db.expireTime(key)
		if expireTime != -1 && expireTime < now {
			continue
		}
		switch val.Typ {
		case SR_STR:
			rewriteStringObject(f, key, val)
		case SR_LIST:
			rewriteListObject(f, key, val)
		case SR_SET:
			rewriteSetObject(f, key, val)
		case SR_ZSET:
			rewriteZSetObject(f, key, val)
		case SR_DICT:
			rewriteDictObject(f, key, val)
		default:
			panic("Unknown object type")
		}
		// Save the expireTime
		if expireTime != -1 {
			rewriteExpireObject(f, key, server.db.expireGet(key))
		}
	}
	di.dictReleaseIterator()

	if err = os.Rename(tmpFile, filename); err != nil {
		utils.ErrorP("Error moving temp append only file on the final destination: ", err)
		_ = os.Remove(tmpFile)
		return REDIS_ERR
	}

	utils.Info("SYNC append only file rewrite performed")
	return REDIS_OK
}

func rewriteAppendOnlyFileBackground() int {
	var childPid int

	if server.aofChildPid != -1 {
		return REDIS_ERR
	}
	if childPid = fork(); childPid == 0 {
		// child process
		if server.fd > 0 {
			Close(server.fd)
		}
		tmpFile := persistenceFile(fmt.Sprintf("temp-rewriteaof-bg-%d.aof", os.Getpid()))
		if rewriteAppendOnlyFile(tmpFile) == REDIS_OK {
			os.Exit(0)
		}
		os.Exit(1)
	} else {
		utils.Info("Background append only file rewriting started by pid %d", childPid)
		server.aofChildPid = childPid
		server.changeLoadFactor(BG_PERSISTENCE_LOAD_FACTOR)
		return REDIS_OK
	}
	return REDIS_OK
}

func backgroundRewriteDoneHandler() {
	tmpFile := persistenceFile(fmt.Sprintf("temp-rewriteaof-bg-%d.aof", server.aofChildPid))
	newFd, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		utils.ErrorP("Unable to open the temporary AOF produced by the child: ", err)
		goto cleanup
	}
	if aofRewriteBufferWrite(newFd) == -1 {
		utils.ErrorP("Error trying to flush the parent diff to the rewritten AOF: ", err)
		goto cleanup
	}
	if err = os.Rename(tmpFile, server.aofFilename); err != nil {
		utils.ErrorP("Error trying to rename the temporary AOF file: ", err)
		_ = newFd.Close()
		goto cleanup
	}

	if server.aofFd == nil {
		_ = newFd.Close()
	} else {
		server.aofFd = newFd
		aofUpdateCurrentSize()
		server.aofRewriteBaseSize = server.aofCurrentSize
		server.aofBufReset()
	}
	utils.Info("Background AOF rewrite finished successfully")

cleanup:
	server.aofRewriteBufferReset()
	_ = os.Remove(tmpFile)
	server.aofChildPid = -1
	server.changeLoadFactor(LOAD_FACTOR)
}

// aof rewrite command
func bgRewriteAofCommand(c *SRedisClient) {
	if server.aofChildPid != -1 {
		c.addReplyError("Background append only file rewriting already in progress")
		return
	}
	if server.rdbChildPid != -1 {
		c.addReplyError("Background save already in progress")
		return
	}
	if rewriteAppendOnlyFileBackground() == REDIS_OK {
		c.addReplyStatus("Background append only file rewriting started")
		return
	}
	c.addReply(shared.err)
}
