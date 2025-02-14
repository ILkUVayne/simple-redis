package src

import (
	"fmt"
	time2 "github.com/ILkUVayne/utlis-go/v2/time"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/encoder"
	"github.com/hdt3213/rdb/model"
	"github.com/hdt3213/rdb/parser"
	"os"
	"strconv"
	"time"
)

// -----------------------------------------------------------------------------
// rdb api
// -----------------------------------------------------------------------------

var auxMap = map[string]string{
	"redis-ver":    REDIS_RDB_VERSION,
	"redis-bits":   REDIS_RDB_BITS,
	"aof-preamble": "0",
}

// rdb前置准备工作
func rdbBeforeWrite(enc *core.Encoder) int {
	err := enc.WriteHeader()
	if err != nil {
		ulog.ErrorP("rdbSave err: ", err)
		return REDIS_ERR
	}
	for k, v := range auxMap {
		err = enc.WriteAux(k, v)
		if err != nil {
			ulog.ErrorP("rdbSave err: ", err)
			return REDIS_ERR
		}
	}
	// set db index,keyCount,expireCount
	err = enc.WriteDBHeader(0, uint64(server.db.dbDataSize()), uint64(server.db.dbExpireSize()))
	if err != nil {
		ulog.ErrorP("rdbSave err: ", err)
		return REDIS_ERR
	}
	return REDIS_OK
}

//-----------------------------------------------------------------------------
// rdb loading
//-----------------------------------------------------------------------------

// return Ms time,return -1 when Expired, return 0 when persistent object
func rdbCheckExpire(obj parser.RedisObject) int64 {
	expire := obj.GetExpiration()
	// persistent object
	if expire == nil {
		return 0
	}
	// Expired
	if expire.Before(time.Now()) {
		return -1
	}
	return time2.GetMsTimeByTime(expire)
}

// load expireTime if it has
func rdbLoadExpire(key *SRobj, expire int64) {
	if expire == 0 {
		return
	}
	expireObj := createFromInt(expire)
	server.db.expireSet(key, expireObj)
	expireObj.decrRefCount()
}

// load string obj from rdb
func rdbLoadStringObject(obj parser.RedisObject) {
	expire := rdbCheckExpire(obj)
	if expire == -1 {
		return
	}
	o, ok := obj.(*parser.StringObject)
	if !ok {
		ulog.Error("rdbLoadStringObject err: invalid obj type")
	}
	// add key value
	key, val := createSRobj(SR_STR, o.Key), createSRobj(SR_STR, string(o.Value))
	// maybe int, try encoding
	val.tryObjectEncoding()
	server.db.dictSet(key, val)
	// add expire
	rdbLoadExpire(key, expire)
}

// load string list from rdb
func rdbLoadListObject(obj parser.RedisObject) {
	expire := rdbCheckExpire(obj)
	if expire == -1 {
		return
	}
	o, ok := obj.(*parser.ListObject)
	if !ok {
		ulog.Error("rdbLoadListObject err: invalid obj type")
	}
	key := createSRobj(SR_STR, o.Key)
	lObj := server.db.lookupKeyWrite(key)

	if lObj != nil && lObj.Typ != SR_LIST {
		return
	}
	for _, v := range o.Values {
		if lObj == nil {
			lObj = createListObject()
			server.db.dictSet(key, lObj)
		}
		listTypePush(lObj, createSRobj(SR_STR, string(v)), AL_START_TAIL)
	}
	// add expire
	rdbLoadExpire(key, expire)
}

// load hash obj from rdb
func rdbLoadHashObject(obj parser.RedisObject) {
	expire := rdbCheckExpire(obj)
	if expire == -1 {
		return
	}
	o, ok := obj.(*parser.HashObject)
	if !ok {
		ulog.Error("rdbLoadHashObject err: invalid obj type")
	}
	key := createSRobj(SR_STR, o.Key)

	hashObj := server.db.lookupKeyWrite(key)
	if hashObj != nil && hashObj.Typ != SR_DICT {
		return
	}
	if hashObj == nil {
		hashObj = createHashObject()
		server.db.dictSet(key, hashObj)
	}
	for k, v := range o.Hash {
		hashTypeSet(hashObj, createSRobj(SR_STR, k), createSRobj(SR_STR, string(v)))
	}
	// add expire
	rdbLoadExpire(key, expire)
}

// load zSet obj from rdb
func rdbLoadZSetObject(obj parser.RedisObject) {
	expire := rdbCheckExpire(obj)
	if expire == -1 {
		return
	}
	o, ok := obj.(*parser.ZSetObject)
	if !ok {
		ulog.Error("rdbLoadZSetObject err: invalid obj type")
	}
	key := createSRobj(SR_STR, o.Key)

	ZSObj := server.db.lookupKeyWrite(key)
	if ZSObj != nil && ZSObj.Typ != SR_ZSET {
		return
	}
	if ZSObj == nil {
		ZSObj = createZsetSRobj()
		server.db.dictSet(key, ZSObj)
	}
	zs := assertZSet(ZSObj)
	for _, v := range o.Entries {
		ele := createSRobj(SR_STR, v.Member)
		zNode := zs.zsl.insert(v.Score, ele)
		ele.incrRefCount()
		zs.d.dictSet(ele, createFloatSRobj(SR_STR, zNode.score))
		ele.incrRefCount()
	}
	// add expire
	rdbLoadExpire(key, expire)
}

// load set obj from rdb
func rdbLoadSetObject(obj parser.RedisObject) {
	expire := rdbCheckExpire(obj)
	if expire == -1 {
		return
	}
	o, ok := obj.(*parser.SetObject)
	if !ok {
		ulog.Error("rdbLoadSetObject err: invalid obj type")
	}
	key := createSRobj(SR_STR, o.Key)

	set := server.db.lookupKeyWrite(key)
	if set != nil && set.Typ != SR_SET {
		return
	}
	if set == nil {
		set = setTypeCreate(createSRobj(SR_STR, string(o.Members[0])))
		server.db.dictSet(key, set)
	}
	for _, v := range o.Members {
		val := createSRobj(SR_STR, string(v))
		val.tryObjectEncoding()
		setTypeAdd(set, val)
	}
	// add expire
	rdbLoadExpire(key, expire)
}

// 加载rdb数据到内存中
func rdbLoad(filename string) {
	fd, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		ulog.Error("Can't open the rdb file: ", err)
	}
	defer func() { _ = fd.Close() }()

	fInfo, err := fd.Stat()
	if err != nil {
		ulog.Error("Unable to obtain the AOF file length. stat: ", err)
	}
	if fInfo.Size() == 0 {
		return
	}

	decoder := parser.NewDecoder(fd)
	err = decoder.Parse(func(o parser.RedisObject) bool {
		rdbLoadObject(o)
		// return true to continue, return false to stop the iteration
		return true
	})

	if err != nil {
		ulog.Error("rdbLoad err: ", err)
	}
}

// -----------------------------------------------------------------------------
// rdb file implementation
// -----------------------------------------------------------------------------

// ================================ write rdb data to disk =================================

// write string obj to disk (dump.rdb)
func _writeStringObject(enc *core.Encoder, key string, value any, options ...any) error {
	return enc.WriteStringObject(key, value.([]byte), options)
}

// write list obj to disk (dump.rdb)
func _writeListObject(enc *core.Encoder, key string, value any, options ...any) error {
	return enc.WriteListObject(key, value.([][]byte), options)
}

// write set obj to disk (dump.rdb)
func _writeSetObject(enc *core.Encoder, key string, value any, options ...any) error {
	return enc.WriteSetObject(key, value.([][]byte), options)
}

// write zSet obj to disk (dump.rdb)
func _writeZSetObject(enc *core.Encoder, key string, value any, options ...any) error {
	return enc.WriteZSetObject(key, value.([]*model.ZSetEntry), options)
}

// write hash obj to disk (dump.rdb)
func _writeDictObject(enc *core.Encoder, key string, value any, options ...any) error {
	return enc.WriteHashMapObject(key, value.(map[string][]byte), options)
}

// ================================ build rdb save data =================================

// build string obj and write to rdb
func writeStringObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	return _writeObjectHandle(val.Typ, enc, key.strVal(), []byte(val.strVal()), expire)
}

// build list obj and write to rdb
func writeListObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	values := make([][]byte, 0)

	checkListEncoding(val)
	// encoding is linked list
	li := assertList(val).listRewind()
	for ln := li.listNext(); ln != nil; ln = li.listNext() {
		eleObj := ln.nodeValue()
		values = append(values, []byte(eleObj.strVal()))
	}
	return _writeObjectHandle(val.Typ, enc, key.strVal(), values, expire)
}

// build set obj and write to rdb
func writeSetObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	var eleObj *SRobj
	var intObj int64
	values := make([][]byte, 0)

	checkSetEncoding(val)
	si := setTypeInitIterator(val)
	for encoding := si.setTypeNext(&eleObj, &intObj); encoding != -1; encoding = si.setTypeNext(&eleObj, &intObj) {
		if uint8(encoding) == REDIS_ENCODING_INTSET {
			values = append(values, []byte(strconv.FormatInt(intObj, 10)))
			continue
		}
		// REDIS_ENCODING_HT
		values = append(values, []byte(eleObj.strVal()))
	}
	si.setTypeReleaseIterator()
	return _writeObjectHandle(val.Typ, enc, key.strVal(), values, expire)
}

// build hash obj and write to rdb
func writeDictObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	checkHashEncoding(val)
	values := make(map[string][]byte)
	// encoding is hash table
	di := assertDict(val).dictGetIterator()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		values[de.getKey().strVal()] = []byte(de.getVal().strVal())
	}
	di.dictReleaseIterator()
	return _writeObjectHandle(val.Typ, enc, key.strVal(), values, expire)
}

// build zSet obj and write to rdb
func writeZSetObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	checkZSetEncoding(val)

	values := make([]*model.ZSetEntry, 0)
	// encoding is skip list
	zs := assertZSet(val)
	di := zs.d.dictGetIterator()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		eleObj, score := de.getKey(), de.getVal()
		zn := new(model.ZSetEntry)
		sf, _ := score.floatVal()
		zn.Score = sf
		zn.Member = eleObj.strVal()
		values = append(values, zn)
	}
	di.dictReleaseIterator()
	return _writeObjectHandle(val.Typ, enc, key.strVal(), values, expire)
}

// 保存当前内存中的数据到rdb
func rdbSave() int {
	if server.db.dbDataSize() == 0 {
		_ = os.Remove(server.rdbFilename)
		_, _ = os.Create(server.rdbFilename)
		ulog.Info("database is empty")
		return REDIS_OK
	}

	tmpFile := PersistenceFile(server.dir, fmt.Sprintf("temp-%d.rdb", os.Getpid()))
	f, err := os.Create(tmpFile)
	if err != nil {
		_ = os.Remove(tmpFile)
		ulog.ErrorP("Failed opening .rdb for saving: ", err)
		return REDIS_ERR
	}
	defer func() { _ = f.Close() }()

	enc := encoder.NewEncoder(f)
	di := server.db.dbDataDi()
	if rdbBeforeWrite(enc) == REDIS_ERR {
		goto wErr
	}

	for de := di.dictNext(); de != nil; de = di.dictNext() {
		key, val := de.getKey(), de.getVal()
		expireTime := server.db.expireTime(key)
		if rdbWriteObject(enc, key, val, expireTime) == REDIS_ERR {
			goto wErr
		}
	}

	err = enc.WriteEnd()
	if err != nil {
		ulog.ErrorP("rdbSave err: ", err)
		goto wErr
	}
	if err = os.Rename(tmpFile, server.rdbFilename); err != nil {
		ulog.ErrorP("Error moving temp DB file on the final destination: ", err)
		goto wErr
	}

	di.dictReleaseIterator()
	server.dirty = 0
	server.lastSave = time2.GetMsTime()
	server.lastBgSaveStatus = REDIS_OK
	ulog.Info("DB saved on disk")
	return REDIS_OK

wErr:
	_ = os.Remove(tmpFile)
	di.dictReleaseIterator()
	return REDIS_ERR
}

// 后台rdbSave
func rdbSaveBackground() int {
	var childPid int

	if server.rdbChildPid != -1 {
		return REDIS_ERR
	}

	server.dirtyBeforeBgSave = server.dirty
	server.lastBgSaveTry = time2.GetMsTime()

	if childPid = fork(); childPid == 0 {
		if server.fd > 0 {
			Close(server.fd)
		}
		if rdbSave() == REDIS_OK {
			os.Exit(0)
		}
		os.Exit(1)
	} else {
		ulog.Info("Background saving started by pid %d", childPid)
		server.rdbChildPid = childPid
		server.rdbStartTime = time2.GetMsTime()
		server.changeLoadFactor(BG_PERSISTENCE_LOAD_FACTOR)
		updateDictResizePolicy()
		return REDIS_OK
	}
	return REDIS_OK
}

// rdb完成后的收尾工作
func backgroundSaveDoneHandler() {
	server.dirty = server.dirty - server.dirtyBeforeBgSave
	server.lastSave = time2.GetMsTime()
	server.lastBgSaveStatus = REDIS_OK
	server.rdbChildPid = -1
	server.rdbStartTime = 0
	server.changeLoadFactor(LOAD_FACTOR)
	ulog.Info("Background RDB finished successfully")
}

//-----------------------------------------------------------------------------
// rdb commands
//-----------------------------------------------------------------------------

// Save the DB.
//
// usage: SAVE
func saveCommand(c *SRedisClient) {
	if server.rdbChildPid != -1 {
		c.addReplyError("Background save already in progress")
		return
	}
	if rdbSave() == REDIS_OK {
		c.addReply(shared.ok)
		return
	}
	c.addReplyError("save failed")
}

// Save the DB in background.
//
// usage: BgSave
func bgSaveCommand(c *SRedisClient) {
	if server.rdbChildPid != -1 {
		c.addReplyError("Background save already in progress")
		return
	}
	if server.aofChildPid != -1 {
		c.addReplyError("Can't BgSave while AOF log rewriting is in progress")
		return
	}
	if rdbSaveBackground() == REDIS_OK {
		c.addReplyStatus("Background saving started")
		return
	}
	c.addReplyError("bgSave failed")
}
