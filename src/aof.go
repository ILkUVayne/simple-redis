// Package src
//
// Lib AOF provides AOF file loading and persistence methods
package src

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"simple-redis/utils"
	"strconv"
)

// Write the contents of the AOF rewrite buffer into a file
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

// return a fake client,client.fd == -1
func createFakeClient() *SRedisClient {
	return createSRClient(FAKE_CLIENT_FD)
}

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
		utils.Error("Can't open the append-only file: ", err)
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
		utils.Error("flushAppendOnlyFile err: ", err)
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
		utils.Error("Unable to obtain the AOF file length. stat: ", err)
	}
	server.aofCurrentSize = fInfo.Size()
}

// rewrite command to file.
//
// s is command string
func rewrite(f *os.File, s *string) {
	if _, err := io.WriteString(f, *s); err != nil {
		utils.Error("rewriteStringObject err: ", err)
	}
}

// e.g. if val == "get" while write "$3\r\nget\r\n"
func rewriteBulkObject(f *os.File, val *SRobj) {
	strVal := val.strVal()
	cmd := fmt.Sprintf(RESP_BULK, len(strVal), strVal)
	rewrite(f, &cmd)
}

// e.g. if cmd == "*3\r\n$3\r\nSET\r\n"
// val == ["name", "hello world"]
//
// while write "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$11\r\nhello world\r\n"
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

// Obtain the number of items, which cannot exceed REDIS_AOF_REWRITE_ITEMS_PER_CMD
//
// e.g. rpush key value [value ...],items is numbers of values
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

// rewrite string object to file
func rewriteStringObject(f *os.File, key, val *SRobj) {
	cmd := RESP_STR
	// set key value
	rewriteObject(f, &cmd, key, val)
}

// rewrite expire object to file
func rewriteExpireObject(f *os.File, key, val *SRobj) {
	cmd := RESP_EXPIRE
	// expire key expireTime
	rewriteObject(f, &cmd, key, val)
}

// rewrite list object to file
func rewriteListObject(f *os.File, key, val *SRobj) {
	count, items := 0, listTypeLength(val)
	// encoding is linked list
	checkListEncoding(val)
	l := assertList(val)
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
}

// rewrite set object to file
func rewriteSetObject(f *os.File, key, val *SRobj) {
	count, items := 0, int(setTypeSize(val))
	// encoding is intSet
	if val.encoding == REDIS_ENCODING_INTSET {
		var intVal int64
		for ii := 0; assertIntSet(val).intSetGet(uint32(ii), &intVal); ii++ {
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
	// encoding is hash table
	di := assertDict(val).dictGetIterator()
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
}

// rewrite zSet object to file
func rewriteZSetObject(f *os.File, key, val *SRobj) {
	count, items := 0, int(zSetLength(val))
	// encoding is skip list
	zs := assertZSet(val)
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
}

// rewrite hash object to file
func rewriteDictObject(f *os.File, key, val *SRobj) {
	checkHashEncoding(val)
	// encoding is hash table
	di := assertDict(val).dictGetIterator()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		cmd := fmt.Sprintf(RESP_HASH_HSET, 4)
		// add key hashKey hashVal
		rewriteObject(f, &cmd, key, de.getKey(), de.getVal())
	}
	di.dictReleaseIterator()
}

// Iterator dict and append rewrite command to temp aof file
func rewriteAppendOnlyFile(filename string) int {
	tmpFile := utils.PersistenceFile(fmt.Sprintf("temp-rewriteaof-%d.aof", os.Getpid()))
	now := utils.GetMsTime()
	f, err := os.Create(tmpFile)
	if err != nil {
		utils.ErrorP("Opening the temp file for AOF rewrite in rewriteAppendOnlyFile(): ", err)
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
		utils.ErrorP("Error moving temp append only file on the final destination: ", err)
		_ = os.Remove(tmpFile)
		return REDIS_ERR
	}

	utils.Info("SYNC append only file rewrite performed")
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
		tmpFile := utils.PersistenceFile(fmt.Sprintf("temp-rewriteaof-bg-%d.aof", os.Getpid()))
		if rewriteAppendOnlyFile(tmpFile) == REDIS_OK {
			utils.Exit(0)
		}
		utils.Exit(1)
	} else {
		utils.Info("Background append only file rewriting started by pid %d", childPid)
		server.aofChildPid = childPid
		server.changeLoadFactor(BG_PERSISTENCE_LOAD_FACTOR)
		return REDIS_OK
	}
	return REDIS_OK
}

// aof child process success handler
func backgroundRewriteDoneHandler() {
	// Get the temporary AOF file name generated by the child process
	tmpFile := utils.PersistenceFile(fmt.Sprintf("temp-rewriteaof-bg-%d.aof", server.aofChildPid))
	newFd, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		utils.ErrorP("Unable to open the temporary AOF produced by the child: ", err)
		goto cleanup
	}
	if aofRewriteBufferWrite(newFd) == -1 {
		utils.ErrorP("Error trying to flush the parent diff to the rewritten AOF: ", err)
		goto cleanup
	}
	// Replace temporary AOF file name to AOF file name
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
	// reset aofChildPid == -1
	server.aofChildPid = -1
	// Recovery load factor
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
