package src

import (
	"fmt"
	time2 "github.com/ILkUVayne/utlis-go/v2/time"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/encoder"
	"github.com/hdt3213/rdb/model"
	"github.com/hdt3213/rdb/parser"
	"os"
	"simple-redis/utils"
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

func rdbBeforeWrite(enc *core.Encoder) int {
	err := enc.WriteHeader()
	if err != nil {
		utils.ErrorP("rdbSave err: ", err)
		return REDIS_ERR
	}
	for k, v := range auxMap {
		err = enc.WriteAux(k, v)
		if err != nil {
			utils.ErrorP("rdbSave err: ", err)
			return REDIS_ERR
		}
	}
	// set db index,keyCount,expireCount
	err = enc.WriteDBHeader(0, uint64(server.db.dbDataSize()), uint64(server.db.dbExpireSize()))
	if err != nil {
		utils.ErrorP("rdbSave err: ", err)
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

func rdbLoadExpire(key *SRobj, expire int64) {
	if expire == 0 {
		return
	}
	expireObj := createFromInt(expire)
	server.db.expireSet(key, expireObj)
	expireObj.decrRefCount()
}

func rdbLoadStringObject(obj parser.RedisObject) {
	expire := rdbCheckExpire(obj)
	if expire == -1 {
		return
	}
	o, ok := obj.(*parser.StringObject)
	if !ok {
		utils.Error("rdbLoadStringObject err: invalid obj type")
	}
	// add key value
	key, val := createSRobj(SR_STR, o.Key), createSRobj(SR_STR, string(o.Value))
	// maybe int, try encoding
	val.tryObjectEncoding()
	server.db.dictSet(key, val)
	// add expire
	rdbLoadExpire(key, expire)
}

func rdbLoadListObject(obj parser.RedisObject) {
	expire := rdbCheckExpire(obj)
	if expire == -1 {
		return
	}
	o, ok := obj.(*parser.ListObject)
	if !ok {
		utils.Error("rdbLoadListObject err: invalid obj type")
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

func rdbLoadHashObject(obj parser.RedisObject) {
	expire := rdbCheckExpire(obj)
	if expire == -1 {
		return
	}
	o, ok := obj.(*parser.HashObject)
	if !ok {
		utils.Error("rdbLoadHashObject err: invalid obj type")
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

func rdbLoadZSetObject(obj parser.RedisObject) {
	expire := rdbCheckExpire(obj)
	if expire == -1 {
		return
	}
	o, ok := obj.(*parser.ZSetObject)
	if !ok {
		utils.Error("rdbLoadZSetObject err: invalid obj type")
	}
	key := createSRobj(SR_STR, o.Key)

	ZSobj := server.db.lookupKeyWrite(key)
	if ZSobj != nil && ZSobj.Typ != SR_ZSET {
		return
	}
	if ZSobj == nil {
		ZSobj = createZsetSRobj()
		server.db.dictSet(key, ZSobj)
	}
	zs := assertZSet(ZSobj)
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

func rdbLoadSetObject(obj parser.RedisObject) {
	expire := rdbCheckExpire(obj)
	if expire == -1 {
		return
	}
	o, ok := obj.(*parser.SetObject)
	if !ok {
		utils.Error("rdbLoadSetObject err: invalid obj type")
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
func rdbLoad(filename *string) {
	fd, err := os.OpenFile(*filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		utils.Error("Can't open the rdb file: ", err)
	}
	defer func() { _ = fd.Close() }()

	fInfo, err := fd.Stat()
	if err != nil {
		utils.Error("Unable to obtain the AOF file length. stat: ", err)
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
		utils.Error("rdbLoad err: ", err)
	}
}

// -----------------------------------------------------------------------------
// rdb file implementation
// -----------------------------------------------------------------------------

// ================================ write rdb data to disk =================================

func _writeStringObject(enc *core.Encoder, key string, value any, options ...any) error {
	return enc.WriteStringObject(key, value.([]byte), options)
}

func _writeListObject(enc *core.Encoder, key string, value any, options ...any) error {
	return enc.WriteListObject(key, value.([][]byte), options)
}

func _writeSetObject(enc *core.Encoder, key string, value any, options ...any) error {
	return enc.WriteSetObject(key, value.([][]byte), options)
}

func _writeZSetObject(enc *core.Encoder, key string, value any, options ...any) error {
	return enc.WriteZSetObject(key, value.([]*model.ZSetEntry), options)
}

func _writeDictObject(enc *core.Encoder, key string, value any, options ...any) error {
	return enc.WriteHashMapObject(key, value.(map[string][]byte), options)
}

// ================================ build rdb save data =================================

func writeStringObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	return _writeObjectHandle(val.Typ, enc, key.strVal(), []byte(val.strVal()), expire)
}

func writeListObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	values := make([][]byte, 0)

	checkListEncoding(val)
	// encoding is linked list
	l := assertList(val)
	li := l.listRewind()
	for ln := li.listNext(); ln != nil; ln = li.listNext() {
		eleObj := ln.nodeValue()
		values = append(values, []byte(eleObj.strVal()))
	}
	return _writeObjectHandle(val.Typ, enc, key.strVal(), values, expire)
}

func writeSetObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	values := make([][]byte, 0)

	checkSetEncoding(val)

	if val.encoding == REDIS_ENCODING_INTSET {
		var intVal int64
		for ii := 0; assertIntSet(val).intSetGet(uint32(ii), &intVal); ii++ {
			values = append(values, []byte(strconv.FormatInt(intVal, 10)))
		}
	}
	if val.encoding == REDIS_ENCODING_HT {
		di := assertDict(val).dictGetIterator()
		for de := di.dictNext(); de != nil; de = di.dictNext() {
			eleObj := de.getKey()
			values = append(values, []byte(eleObj.strVal()))
		}
		di.dictReleaseIterator()
	}
	return _writeObjectHandle(val.Typ, enc, key.strVal(), values, expire)
}

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
func rdbSave(filename *string) int {
	if server.db.dbDataSize() == 0 {
		_ = os.Remove(*filename)
		_, _ = os.Create(*filename)
		utils.Info("database is empty")
		return REDIS_OK
	}

	tmpFile := utils.PersistenceFile(fmt.Sprintf("temp-%d.rdb", os.Getpid()))
	f, err := os.Create(tmpFile)
	if err != nil {
		utils.ErrorP("Failed opening .rdb for saving: ", err)
		return REDIS_ERR
	}
	defer func() { _ = f.Close() }()

	enc := encoder.NewEncoder(f)
	if rdbBeforeWrite(enc) == REDIS_ERR {
		return REDIS_ERR
	}

	di := server.db.dbDataDi()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		key, val := de.getKey(), de.getVal()
		expireTime := server.db.expireTime(key)
		if rdbWriteObject(enc, key, val, expireTime) == REDIS_ERR {
			goto werr
		}
	}
	di.dictReleaseIterator()

	err = enc.WriteEnd()
	if err != nil {
		utils.ErrorP("rdbSave err: ", err)
		return REDIS_ERR
	}
	if err = os.Rename(tmpFile, *filename); err != nil {
		utils.ErrorP("Error moving temp DB file on the final destination: ", err)
		_ = os.Remove(tmpFile)
		return REDIS_ERR
	}

	utils.Info("DB saved on disk")
	server.dirty = 0
	server.lastSave = time2.GetMsTime()
	server.lastBgSaveStatus = REDIS_OK
	return REDIS_OK

werr:
	_ = os.Remove(tmpFile)
	utils.ErrorP("Write error saving DB on disk: ", err)
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
		if rdbSave(&server.rdbFilename) == REDIS_OK {
			utils.Exit(0)
		}
		utils.Exit(1)
	} else {
		utils.Info("Background saving started by pid %d", childPid)
		server.rdbChildPid = childPid
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
	server.changeLoadFactor(LOAD_FACTOR)
	utils.Info("Background RDB finished successfully")
}

//-----------------------------------------------------------------------------
// rdb commands
//-----------------------------------------------------------------------------

// SAVE
func saveCommand(c *SRedisClient) {
	if server.rdbChildPid != -1 {
		c.addReplyError("Background save already in progress")
		return
	}
	if rdbSave(&server.rdbFilename) == REDIS_OK {
		c.addReply(shared.ok)
		return
	}
	c.addReply(shared.err)
}

// BGSAVE
func bgSaveCommand(c *SRedisClient) {
	if server.rdbChildPid != -1 {
		c.addReplyError("Background save already in progress")
		return
	}
	if server.aofChildPid != -1 {
		c.addReplyError("Can't BGSAVE while AOF log rewriting is in progress")
		return
	}
	if rdbSaveBackground() == REDIS_OK {
		c.addReplyStatus("Background saving started")
		return
	}
	c.addReply(shared.err)
}
