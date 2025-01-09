package src

import (
	"errors"
	"fmt"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/encoder"
	"github.com/hdt3213/rdb/parser"
	"os"
	"strings"
)

// -----------------------------------------------------------------------------
// args
// -----------------------------------------------------------------------------

// ============================= splitArgs handle ==============================

type splitArgsHandleFunc func(line string, i int) (string, int, int)

var splitArgsHandleMap = map[string]splitArgsHandleFunc{
	"normal": normalArgs,
	"\"":     quotesArgs,
	"'":      singleQuotesArgs,
}

func splitArgsHandle(initial, line string, i int) (string, int, int) {
	fn, ok := splitArgsHandleMap[initial]
	if !ok {
		fn = splitArgsHandleMap["normal"]
	}
	return fn(line, i)
}

//-----------------------------------------------------------------------------
// client
//-----------------------------------------------------------------------------

// ================================ cmd buff handle =================================

type cmdBufHandleFunc func(c *SRedisClient) (bool, error)

var cmdBufHandleFuncMaps = map[CmdType]cmdBufHandleFunc{
	CMD_INLINE: inlineBufHandle,
	CMD_BULK:   bulkBufHandle,
}

func cmdBufHandle(c *SRedisClient) (bool, error) {
	checkCmdType(c)
	fn, ok := cmdBufHandleFuncMaps[c.cmdTyp]
	if !ok {
		return false, errors.New("unknow cmd type")
	}
	return fn(c)
}

// -----------------------------------------------------------------------------
// aof
// -----------------------------------------------------------------------------

// ================================ AOF rewrite =================================

// aof rewriteObjectFunc type
type aofRWObjectFunc func(f *os.File, key, val *SRobj)

// aof rewriteObjectFunc maps
var aofRWObjectMaps = map[SRType]aofRWObjectFunc{
	SR_STR:  rewriteStringObject,
	SR_LIST: rewriteListObject,
	SR_SET:  rewriteSetObject,
	SR_ZSET: rewriteZSetObject,
	SR_DICT: rewriteDictObject,
}

// aof rewriteObjectFunc factory
func aofRWObject(f *os.File, key, val *SRobj) int {
	fn, ok := aofRWObjectMaps[val.Typ]
	if !ok {
		ulog.ErrorP("Unknown object type: ", val.Typ)
		return REDIS_ERR
	}
	// call
	fn(f, key, val)
	return REDIS_OK
}

// -----------------------------------------------------------------------------
// rdb
// -----------------------------------------------------------------------------

// ================================ rdb loading =================================

// load rdb data

type rdbLoadObjectFunc func(obj parser.RedisObject)

var rdbLoadObjectMaps = map[string]rdbLoadObjectFunc{
	parser.StringType: rdbLoadStringObject,
	parser.ListType:   rdbLoadListObject,
	parser.HashType:   rdbLoadHashObject,
	parser.ZSetType:   rdbLoadZSetObject,
	parser.SetType:    rdbLoadSetObject,
}

func rdbLoadObject(obj parser.RedisObject) {
	fn, ok := rdbLoadObjectMaps[obj.GetType()]
	if !ok {
		ulog.Error("Unknown object type: ", obj.GetType())
	}
	fn(obj)
}

// ================================ rdb file implementation =================================

// write rdb data to disk

type writeObjectFunc func(enc *core.Encoder, key string, value any, options ...any) error

var writeObjectMaps = map[SRType]writeObjectFunc{
	SR_STR:  _writeStringObject,
	SR_LIST: _writeListObject,
	SR_SET:  _writeSetObject,
	SR_ZSET: _writeZSetObject,
	SR_DICT: _writeDictObject,
}

func _writeObjectHandle(typ SRType, enc *core.Encoder, key string, values any, expire int64) int {
	var err error
	fn, ok := writeObjectMaps[typ]
	if !ok {
		ulog.Error("Unknown object type: ", typ)
	}
	if expire != -1 {
		err = fn(enc, key, values, encoder.WithTTL(uint64(expire)))
	} else {
		err = fn(enc, key, values)
	}

	if err != nil {
		ulog.ErrorP("rdbSave writeObject err: ", err)
		return REDIS_ERR
	}
	return REDIS_OK
}

// build rdb save data

type rdbSaveObjectFunc func(enc *core.Encoder, key, val *SRobj, expire int64) int

var rdbSaveMaps = map[SRType]rdbSaveObjectFunc{
	SR_STR:  writeStringObject,
	SR_LIST: writeListObject,
	SR_SET:  writeSetObject,
	SR_ZSET: writeZSetObject,
	SR_DICT: writeDictObject,
}

func rdbWriteObject(enc *core.Encoder, key, val *SRobj, expire int64) int {
	fn, ok := rdbSaveMaps[val.Typ]
	if !ok {
		ulog.ErrorP("Unknown object type: ", val.Typ)
		return REDIS_ERR
	}
	return fn(enc, key, val, expire)
}

// -----------------------------------------------------------------------------
// config
// -----------------------------------------------------------------------------

// complex config parse function
type complexConfFunc func(val string)

var complexConfFuncMaps = map[string]complexConfFunc{
	"save": appendServerSaveParams,
}

// return true complexConf,or false simpleConf
func complexConfHandle(key, val string) (ok bool) {
	var fn complexConfFunc
	if fn, ok = complexConfFuncMaps[strings.ToLower(key)]; ok {
		fn(val)
	}
	return
}

// -----------------------------------------------------------------------------
// resp
// -----------------------------------------------------------------------------

// ================================ Parse resp data =================================

type respParseFunc func([]byte, int, int) (string, error)

var respParseFuncMaps = map[int]respParseFunc{
	SIMPLE_STR:   respParseSimpleStr,
	SIMPLE_ERROR: respParseSimpleStr,
	BULK_STR:     respParseBulk,
	INTEGERS:     respParseIntegers,
	ARRAYS:       respParseArrays,
}

func respParseHandle(reply *sRedisReply) (string, error) {
	if fn, ok := respParseFuncMaps[reply.typ]; ok {
		idx, err := getQueryLine(reply.buf, reply.length)
		if err != nil || idx < 0 {
			return "", err
		}
		return fn(reply.buf, reply.length, idx)
	}
	return "", errors.New(fmt.Sprintf("type %d respParseFunc not found", reply.typ))
}

// ================================ format response string =================================

type strFormatFunc func(s string) string

var strFormatFuncMaps = map[int]strFormatFunc{
	BULK_STR:     bulkStrFormat,
	SIMPLE_ERROR: simpleErrStrFormat,
	INTEGERS:     intStrFormat,
}

func strFormatHandle(reply *sRedisReply) string {
	if fn, ok := strFormatFuncMaps[reply.typ]; ok {
		return fn(reply.str)
	}
	return ""
}

//-----------------------------------------------------------------------------
// Sorted set
//-----------------------------------------------------------------------------

// ================================ Parse Range =================================

type parseRangeFunc func(s string) (float64, int, error)

var parseRangeFuncMaps = map[uint8]parseRangeFunc{
	'(': parseParentheses,
}

// return (min,minex) or (max,maxnx) and error
func _parseRange(obj *SRobj) (float64, int, error) {
	str := obj.strVal()
	fn, ok := parseRangeFuncMaps[str[0]]
	if !ok {
		val, _ := obj.floatVal()
		return val, 0, nil
	}
	str = str[1:]
	return fn(str)
}
