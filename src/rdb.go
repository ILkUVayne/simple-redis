package src

import (
	"fmt"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/encoder"
	"github.com/hdt3213/rdb/model"
	"os"
	"simple-redis/utils"
	"strconv"
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
	err = enc.WriteDBHeader(0, uint64(server.db.data.dictSize()), uint64(server.db.expire.dictSize()))
	if err != nil {
		utils.ErrorP("rdbSave err: ", err)
		return REDIS_ERR
	}
	return REDIS_OK
}

//-----------------------------------------------------------------------------
// rdb loading
//-----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// rdb file implementation
// -----------------------------------------------------------------------------

func writeStringObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	var err error
	strKey, valStr := key.strVal(), val.strVal()
	if expire != -1 {
		err = enc.WriteStringObject(strKey, []byte(valStr), encoder.WithTTL(uint64(expire)))
	} else {
		err = enc.WriteStringObject(strKey, []byte(valStr))
	}

	if err != nil {
		utils.ErrorP("rdbSave writeStringObject: ", err)
		return REDIS_ERR
	}
	return REDIS_OK
}

func writeListObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	var err error
	values := make([][]byte, 0)

	if val.encoding != REDIS_ENCODING_LINKEDLIST {
		panic("Unknown list encoding")
	}

	if val.encoding == REDIS_ENCODING_LINKEDLIST {
		l := val.Val.(*list)
		li := l.listRewind()
		for ln := li.listNext(); ln != nil; ln = li.listNext() {
			eleObj := ln.nodeValue()
			values = append(values, []byte(eleObj.strVal()))
		}
	}
	if expire != -1 {
		err = enc.WriteListObject(key.strVal(), values, encoder.WithTTL(uint64(expire)))
	} else {
		err = enc.WriteListObject(key.strVal(), values)
	}
	// gc
	values = nil
	if err != nil {
		utils.ErrorP("rdbSave writeListObject: ", err)
		return REDIS_ERR
	}
	return REDIS_OK
}

func writeSetObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	var err error
	values := make([][]byte, 0)

	if val.encoding != REDIS_ENCODING_INTSET && val.encoding != REDIS_ENCODING_HT {
		panic("Unknown set encoding")
	}

	if val.encoding == REDIS_ENCODING_INTSET {
		var intVal int64
		for ii := 0; val.Val.(*intSet).intSetGet(uint32(ii), &intVal); ii++ {
			values = append(values, []byte(strconv.FormatInt(intVal, 10)))
		}
	}
	if val.encoding == REDIS_ENCODING_HT {
		di := val.Val.(*dict).dictGetIterator()
		for de := di.dictNext(); de != nil; de = di.dictNext() {
			eleObj := de.getKey()
			values = append(values, []byte(eleObj.strVal()))
		}
		di.dictReleaseIterator()
	}

	if expire != -1 {
		err = enc.WriteSetObject(key.strVal(), values, encoder.WithTTL(uint64(expire)))
	} else {
		err = enc.WriteSetObject(key.strVal(), values)
	}
	// gc
	values = nil
	if err != nil {
		utils.ErrorP("rdbSave writeSetObject: ", err)
		return REDIS_ERR
	}
	return REDIS_OK
}

func writeDictObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	var err error
	if val.encoding != REDIS_ENCODING_HT {
		panic("Unknown hash encoding")
	}
	values := make(map[string][]byte)
	if val.encoding == REDIS_ENCODING_HT {
		di := val.Val.(*dict).dictGetIterator()
		for de := di.dictNext(); de != nil; de = di.dictNext() {
			values[de.getKey().strVal()] = []byte(de.getVal().strVal())
		}
		di.dictReleaseIterator()
	}

	if expire != -1 {
		err = enc.WriteHashMapObject(key.strVal(), values, encoder.WithTTL(uint64(expire)))
	} else {
		err = enc.WriteHashMapObject(key.strVal(), values)
	}
	// gc
	values = nil
	if err != nil {
		utils.ErrorP("rdbSave writeDictObject: ", err)
		return REDIS_ERR
	}
	return REDIS_OK
}

func writeZSetObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	var err error
	values := make([]*model.ZSetEntry, 0)

	if val.encoding != REDIS_ENCODING_SKIPLIST {
		panic("Unknown sorted zset encoding")
	}
	if val.encoding == REDIS_ENCODING_SKIPLIST {
		zs := val.Val.(*zSet)
		di := zs.d.dictGetIterator()
		for de := di.dictNext(); de != nil; de = di.dictNext() {
			eleObj := de.getKey()
			score := de.getVal()
			zn := new(model.ZSetEntry)
			sf, _ := score.floatVal()
			zn.Score = sf
			zn.Member = eleObj.strVal()
			values = append(values, zn)
		}
		di.dictReleaseIterator()
	}

	if expire != -1 {
		err = enc.WriteZSetObject(key.strVal(), values, encoder.WithTTL(uint64(expire)))
	} else {
		err = enc.WriteZSetObject(key.strVal(), values)
	}
	// gc
	values = nil
	if err != nil {
		utils.ErrorP("rdbSave writeZSetObject: ", err)
		return REDIS_ERR
	}
	return REDIS_OK
}

func rdbSave(filename *string) int {
	tmpFile := persistenceFile(fmt.Sprintf("temp-%d.rdb", os.Getpid()))
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

	di := server.db.data.dictGetIterator()
	for de := di.dictNext(); de != nil; de = di.dictNext() {
		key := de.getKey()
		val := de.getVal()
		expireTime := server.db.expireTime(key)

		switch val.Typ {
		case SR_STR:
			if writeStringObject(enc, key, val, expireTime) == REDIS_ERR {
				goto werr
			}
		case SR_LIST:
			if writeListObject(enc, key, val, expireTime) == REDIS_ERR {
				goto werr
			}
		case SR_SET:
			if writeSetObject(enc, key, val, expireTime) == REDIS_ERR {
				goto werr
			}
		case SR_ZSET:
			if writeZSetObject(enc, key, val, expireTime) == REDIS_ERR {
				goto werr
			}
		case SR_DICT:
			if writeDictObject(enc, key, val, expireTime) == REDIS_ERR {
				goto werr
			}
		default:
			panic("Unknown object type")
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
	server.lastSave = utils.GetMsTime()
	return REDIS_OK

werr:
	_ = os.Remove(tmpFile)
	utils.ErrorP("Write error saving DB on disk: ", err)
	di.dictReleaseIterator()
	return REDIS_ERR
}

func rdbSaveBackground(filename *string) int {
	var childPid int

	if server.rdbChildPid != -1 {
		return REDIS_ERR
	}
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
		return REDIS_OK
	}
	return REDIS_OK
}

func backgroundSaveDoneHandler() {

}

//-----------------------------------------------------------------------------
// rdb commands
//-----------------------------------------------------------------------------

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

func bgSaveCommand(c *SRedisClient) {
	if server.rdbChildPid != -1 {
		c.addReplyError("Background save already in progress")
		return
	}
	if server.aofChildPid != -1 {
		c.addReplyError("Can't BGSAVE while AOF log rewriting is in progress")
		return
	}
	if rdbSaveBackground(&server.rdbFilename) == REDIS_OK {
		c.addReplyStatus("Background saving started")
		return
	}
	c.addReply(shared.err)
}
