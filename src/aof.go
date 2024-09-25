// Package src
//
// Lib AOF provides AOF file loading and persistence methods
package src

import (
	"bufio"
	"fmt"
	str2 "github.com/ILkUVayne/utlis-go/v2/str"
	"github.com/ILkUVayne/utlis-go/v2/time"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"io"
	"os"
)

// Write the contents of the AOF rewrite buffer into a file
func aofRewriteBufferWrite(f *os.File) int {
	if len(server.aofRewriteBufBlocks) == 0 {
		return 0
	}
	n, err := io.WriteString(f, server.aofRewriteBufBlocks)
	if err != nil {
		ulog.ErrorP("aofRewriteBufferWrite err: ", err)
		return -1
	}
	server.aofRewriteBufferReset()
	return n
}

// reset aof buffer
func (s *SRedisServer) aofBufReset() {
	if len(s.aofBuf) != 0 {
		s.aofBuf = ""
	}
}

// reset aof rewrite buffer
func (s *SRedisServer) aofRewriteBufferReset() {
	if len(s.aofRewriteBufBlocks) != 0 {
		s.aofRewriteBufBlocks = ""
	}
}

// -----------------------------------------------------------------------------
// aof loading
// -----------------------------------------------------------------------------

// if expired,del key
func aofCheckExpire(args []*SRobj) bool {
	if args[0].strVal() != EXPIRE {
		return false
	}
	when, _ := args[2].intVal()
	return server.db.expireIfNeeded1(when, args[1])
}

// read aof file and load data
func loadAppendOnlyFile(name string) {
	fp, err := os.Open(name)
	if err != nil {
		ulog.Error("Can't open the append-only file: ", err)
	}
	defer func() { _ = fp.Close() }()

	// read aof file
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
			if str2.String2Int64(str, &aLen) != nil {
				ulog.Error("Bad file format reading the append only file")
			}
			continue
		}
		if str[0] == '$' {
			continue
		}
		args = append(args, createSRobj(SR_STR, str))
		aLen--
		// call command
		if aLen == 0 {
			if aofCheckExpire(args) {
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

// -----------------------------------------------------------------------------
// AOF file implementation
// -----------------------------------------------------------------------------

// Flush the data from the AOF buffer to the AOF persistence file
func flushAppendOnlyFile() {
	if len(server.aofBuf) == 0 {
		return
	}
	n, err := io.WriteString(server.aofFd, server.aofBuf)
	if err != nil {
		ulog.Error("flushAppendOnlyFile err: ", err)
	}
	server.aofCurrentSize += int64(n)
	server.aofBufReset()
}

// build Persistence Command.
//
// argc is numbers of args.
// args is Command arrays
func catAppendOnlyGenericCommand(argc int, args []*SRobj) string {
	buf := fmt.Sprintf(RESP_ARRAY, argc)
	for i := 0; i < argc; i++ {
		o := args[i].getDecodedObject()
		buf += fmt.Sprintf(RESP_BULK, len(o.strVal()), o.strVal())
		o.decrRefCount()
	}
	return buf
}

// build expire Persistence Command.
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

// append aof Persistence Command to aof buffer
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
	// If AOF rewriting is in progress, append the AOF rewriting buffer
	if server.aofChildPid != -1 {
		server.aofRewriteBufBlocks += buf
	}
}

// ----------------------------------------------------------------------------
// AOF rewrite
// ----------------------------------------------------------------------------

// update current aof file size (server.aofCurrentSize)
func aofUpdateCurrentSize() {
	fInfo, err := server.aofFd.Stat()
	if err != nil {
		ulog.Error("Unable to obtain the AOF file length. stat: ", err)
	}
	server.aofCurrentSize = fInfo.Size()
}

// rewrite command to file.
//
// s is command string
func rewrite(f *os.File, s string) {
	if _, err := io.WriteString(f, s); err != nil {
		ulog.Error("rewriteStringObject err: ", err)
	}
}

// e.g. if val == "get" while write "$3\r\nget\r\n"
func rewriteBulkObject(f *os.File, val *SRobj) {
	strVal := val.strVal()
	rewrite(f, fmt.Sprintf(RESP_BULK, len(strVal), strVal))
}

// e.g. if cmd == "*3\r\n$3\r\nSET\r\n"
// val == ["name", "hello world"]
//
// while write "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$11\r\nhello world\r\n"
func rewriteObject(f *os.File, cmd string, val ...*SRobj) {
	if cmd != "" {
		rewrite(f, cmd)
	}

	if len(val) == 0 {
		return
	}
	for _, v := range val {
		rewriteBulkObject(f, v)
	}
}

// Obtain the number of items, which cannot exceed REDIS_AOF_REWRITE_ITEMS_PER_CMD
//
// e.g. rpush key value [value ...],items is numbers of values
func getItems(items int64) int64 {
	cmdItems := items
	if items > REDIS_AOF_REWRITE_ITEMS_PER_CMD {
		cmdItems = REDIS_AOF_REWRITE_ITEMS_PER_CMD
	}
	return cmdItems
}

func checkItems(count, items *int64) {
	*count++
	if *count == REDIS_AOF_REWRITE_ITEMS_PER_CMD {
		*count = 0
	}
	*items--
}

// rewrite string object to file
func rewriteStringObject(f *os.File, key, val *SRobj) {
	// set key value
	rewriteObject(f, RESP_STR, key, val)
}

// rewrite expire object to file
func rewriteExpireObject(f *os.File, key, val *SRobj) {
	// expire key expireTime
	rewriteObject(f, RESP_EXPIRE, key, val)
}

// rewrite list object to file
func rewriteListObject(f *os.File, key, val *SRobj) {
	count, items := int64(0), listTypeLength(val)
	// encoding is linked list
	checkListEncoding(val)
	l := assertList(val)
	li := l.listRewind()
	for ln := li.listNext(); ln != nil; ln = li.listNext() {
		eleObj := ln.nodeValue()
		if count == 0 {
			cmd := fmt.Sprintf(RESP_LIST_RPUSH, 2+getItems(items))
			// add key
			rewriteObject(f, cmd, key)
		}
		// add val
		rewriteObject(f, "", eleObj)
		checkItems(&count, &items)
	}
}

// rewrite set object to file
func rewriteSetObject(f *os.File, key, val *SRobj) {
	var eleObj *SRobj
	var intObj int64

	count, items := int64(0), setTypeSize(val)
	si := setTypeInitIterator(val)
	for encoding := si.setTypeNext(&eleObj, &intObj); encoding != -1; encoding = si.setTypeNext(&eleObj, &intObj) {
		if count == 0 {
			cmd := fmt.Sprintf(RESP_SET, 2+getItems(items))
			// add key
			rewriteObject(f, cmd, key)
		}
		// add val
		var setVal *SRobj
		if uint8(encoding) == REDIS_ENCODING_INTSET {
			setVal = createFromInt(intObj)
		}
		if uint8(encoding) == REDIS_ENCODING_HT {
			setVal = eleObj
		}
		rewriteObject(f, "", setVal)
		checkItems(&count, &items)
	}
	si.setTypeReleaseIterator()
}

// rewrite zSet object to file
func rewriteZSetObject(f *os.File, key, val *SRobj) {
	count, items := int64(0), zSetLength(val)
	// encoding is skip list
	zs := assertZSet(val)
	di := zs.d.dictGetIterator()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		eleObj, score := de.getKey(), de.getVal()

		if count == 0 {
			cmd := fmt.Sprintf(RESP_ZSET, 2+getItems(items)*2)
			// add key
			rewriteObject(f, cmd, key)
		}
		sf, _ := score.floatVal()
		str := formatFloat(sf, 10)
		// add zSetScore and zSetVal
		rewriteObject(f, "", createSRobj(SR_STR, str), eleObj)
		checkItems(&count, &items)
	}
	di.dictReleaseIterator()
}

// rewrite hash object to file
func rewriteDictObject(f *os.File, key, val *SRobj) {
	checkHashEncoding(val)
	// encoding is hash table
	di := assertDict(val).dictGetIterator()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		cmd := fmt.Sprintf(RESP_HASH_HSET, 4)
		// add key hashKey hashVal
		rewriteObject(f, cmd, key, de.getKey(), de.getVal())
	}
	di.dictReleaseIterator()
}

// Iterator dict and append rewrite command to temp aof file
func rewriteAppendOnlyFile(filename string) int {
	if server.aofState == REDIS_AOF_OFF {
		ulog.ErrorP("Background append only file rewriting is not enabled")
		return REDIS_ERR
	}
	tmpFile := PersistenceFile(fmt.Sprintf("temp-rewriteaof-%d.aof", os.Getpid()))
	now := time.GetMsTime()
	f, err := os.Create(tmpFile)
	if err != nil {
		ulog.ErrorP("Opening the temp file for AOF rewrite in rewriteAppendOnlyFile(): ", err)
		return REDIS_ERR
	}
	defer func() { _ = f.Close() }()

	di := server.db.dbDataDi()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		key, val := de.getKey(), de.getVal()
		expireTime := server.db.expireTime(key)
		// if expired, skip
		if expireTime != -1 && expireTime < now {
			continue
		}
		if aofRWObject(f, key, val) == REDIS_ERR {
			di.dictReleaseIterator()
			return REDIS_ERR
		}
		// Save the expireTime
		if expireTime != -1 {
			rewriteExpireObject(f, key, server.db.expireGet(key))
		}
	}
	di.dictReleaseIterator()

	if err = os.Rename(tmpFile, filename); err != nil {
		ulog.ErrorP("Error moving temp append only file on the final destination: ", err)
		_ = os.Remove(tmpFile)
		return REDIS_ERR
	}

	ulog.Info("SYNC append only file rewrite performed")
	return REDIS_OK
}

// fork child process to rewrite aof
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
		tmpFile := PersistenceFile(fmt.Sprintf("temp-rewriteaof-bg-%d.aof", os.Getpid()))
		if rewriteAppendOnlyFile(tmpFile) == REDIS_OK {
			os.Exit(0)
		}
		os.Exit(1)
	} else {
		ulog.Info("Background append only file rewriting started by pid %d", childPid)
		server.aofChildPid = childPid
		server.aofStartTime = time.GetMsTime()
		server.changeLoadFactor(BG_PERSISTENCE_LOAD_FACTOR)
		updateDictResizePolicy()
		return REDIS_OK
	}
	return REDIS_OK
}

// 同步执行aof重写，会阻塞程序执行
func rewriteAppendOnlyFileSync() {
	tmpFile := PersistenceFile(fmt.Sprintf("temp-rewriteaof-sync-%d.aof", time.GetMsTime()))
	if rewriteAppendOnlyFile(tmpFile) == REDIS_ERR {
		server.aofRewriteBufferReset()
		_ = os.Remove(tmpFile)
		return
	}

	newFd, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		ulog.ErrorP("Unable to open the temporary AOF produced: ", err)
		goto cleanup
	}

	// Replace temporary AOF file name to AOF file name
	if err = os.Rename(tmpFile, server.aofFilename); err != nil {
		ulog.ErrorP("Error trying to rename the temporary AOF file: ", err)
		goto cleanup
	}

	server.aofFd = newFd
	aofUpdateCurrentSize()
	server.aofRewriteBaseSize = server.aofCurrentSize
	server.aofBufReset()
	ulog.Info("AOF rewrite finished successfully")

cleanup:
	server.aofRewriteBufferReset()
	_ = os.Remove(tmpFile)
}

// aof child process success handler
func backgroundRewriteDoneHandler() {
	// Get the temporary AOF file name generated by the child process
	tmpFile := PersistenceFile(fmt.Sprintf("temp-rewriteaof-bg-%d.aof", server.aofChildPid))
	newFd, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		ulog.ErrorP("Unable to open the temporary AOF produced by the child: ", err)
		goto cleanup
	}
	if aofRewriteBufferWrite(newFd) == -1 {
		ulog.ErrorP("Error trying to flush the parent diff to the rewritten AOF: ", err)
		goto cleanup
	}
	// Replace temporary AOF file name to AOF file name
	if err = os.Rename(tmpFile, server.aofFilename); err != nil {
		ulog.ErrorP("Error trying to rename the temporary AOF file: ", err)
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
	ulog.Info("Background AOF rewrite finished successfully")

cleanup:
	server.aofRewriteBufferReset()
	_ = os.Remove(tmpFile)
	// reset aofChildPid == -1
	server.aofChildPid = -1
	server.aofStartTime = 0
	// Recovery load factor
	server.changeLoadFactor(LOAD_FACTOR)
}

//-----------------------------------------------------------------------------
// aof commands
//-----------------------------------------------------------------------------

// BGREWRITEAOF
func bgRewriteAofCommand(c *SRedisClient) {
	if server.aofState == REDIS_AOF_OFF {
		c.addReplyError("Background append only file rewriting is not enabled")
		return
	}
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
	c.addReplyError("bgRewriteAof failed")
}
