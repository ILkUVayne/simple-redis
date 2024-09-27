package src

import (
	"errors"
	str2 "github.com/ILkUVayne/utlis-go/v2/str"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"strconv"
	"strings"
)

type SRType uint8

type SRVal any

// SRobj simple-redis 的基本存储对象结构
type SRobj struct {
	Typ      SRType // simple-redis定义的type，例如： SR_STR
	Val      SRVal  // 实际存储的值，可以是任意类型
	encoding uint8  // val实际的编码格式，例如： REDIS_ENCODING_INT
	refCount int
}

var encodingMaps = map[uint8]string{
	REDIS_ENCODING_RAW:        "raw",
	REDIS_ENCODING_INT:        "int",
	REDIS_ENCODING_HT:         "hashtable",
	REDIS_ENCODING_ZIPMAP:     "zipmap",
	REDIS_ENCODING_LINKEDLIST: "linkedlist",
	REDIS_ENCODING_ZIPLIST:    "ziplist",
	REDIS_ENCODING_INTSET:     "intset",
	REDIS_ENCODING_SKIPLIST:   "skiplist",
}

var TypeMaps = map[SRType]string{
	SR_STR:  "string",
	SR_LIST: "list",
	SR_SET:  "set",
	SR_ZSET: "zset",
	SR_DICT: "hash",
}

var NotStringTypeErr = errors.New("intVal err: type is not string")

// 尝试将 SRobj.Val 转换为字符串并返会
func (s *SRobj) strVal() string {
	if s.Typ != SR_STR {
		return ""
	}
	if s.encoding == REDIS_ENCODING_INT {
		iVal, _ := s.intVal()
		return strconv.FormatInt(iVal, 10)
	}
	return s.Val.(string)
}

func (s *SRobj) incrRefCount() {
	if s == nil {
		return
	}
	s.refCount++
}

func (s *SRobj) decrRefCount() {
	if s == nil {
		return
	}
	s.refCount--
	// gc 自动回收
	if s.refCount == 0 {
		s.Val = nil
	}
}

// 尝试将 SRobj.Val 转换为数字并返会
func (s *SRobj) intVal() (int64, error) {
	if s.Typ != SR_STR {
		return 0, NotStringTypeErr
	}
	if s.encoding == REDIS_ENCODING_INT {
		return s.Val.(int64), nil
	}
	if s.encoding == REDIS_ENCODING_RAW {
		var i int64
		return i, str2.String2Int64(s.strVal(), &i)
	}
	panic("Unknown string encoding")
}

// 尝试将 SRobj.Val 转换为浮点数并返会
func (s *SRobj) floatVal() (float64, error) {
	if s.Typ != SR_STR {
		return 0, NotStringTypeErr
	}
	if s.encoding == REDIS_ENCODING_INT {
		return s.Val.(float64), nil
	}
	if s.encoding == REDIS_ENCODING_RAW {
		var i float64
		return i, str2.String2Float64(s.strVal(), &i)
	}
	panic("Unknown string encoding")
}

func (s *SRobj) strEncoding() string {
	if encoding, ok := encodingMaps[s.encoding]; ok {
		return encoding
	}
	return UNKNOWN
}

func (s *SRobj) strType() string {
	if typ, ok := TypeMaps[s.Typ]; ok {
		return typ
	}
	return UNKNOWN
}

func (s *SRobj) getEncoding() *SRobj {
	return createSRobj(SR_STR, s.strEncoding())
}

func (s *SRobj) getType() *SRobj {
	return createSRobj(SR_STR, s.strType())
}

// 检查 SRobj.Typ 是否与传入的typ相等
func (s *SRobj) checkType(c *SRedisClient, typ SRType) bool {
	if s.Typ != typ {
		c.addReply(shared.wrongTypeErr)
		return false
	}
	return true
}

// try to get SRobj.Val encoding
func (s *SRobj) tryObjectEncoding() {
	if s.encoding != REDIS_ENCODING_RAW || s.refCount > 1 || s.Typ != SR_STR {
		return
	}
	// Check if we can represent this string as a long integer
	var i int64
	if str2.String2Int64(s.strVal(), &i) == nil {
		s.encoding = REDIS_ENCODING_INT
		s.Val = i
	}
}

func (s *SRobj) getDecodedObject() *SRobj {
	if s.encoding == REDIS_ENCODING_RAW {
		s.incrRefCount()
		return s
	}
	if s.Typ == SR_STR && s.encoding == REDIS_ENCODING_INT {
		var intVal int64
		if err := str2.String2Int64(s.strVal(), &intVal); err != nil {
			ulog.Error("getDecodedObject err: ", err)
		}
		return createFromInt(intVal)
	}
	panic("Unknown encoding type")
}

func (s *SRobj) getLongLongFromObject(target *int64) error {
	if s == nil {
		*target = 0
		return nil
	}
	if s.Typ != SR_STR {
		return NotStringTypeErr
	}
	intVal, err := s.intVal()
	*target = intVal
	return err
}

func (s *SRobj) getFloat64FromObject(target *float64) error {
	if s.Typ != SR_STR {
		return NotStringTypeErr
	}
	i, err := s.floatVal()
	*target = i
	return err
}

func (s *SRobj) getFloat64FromObjectOrReply(c *SRedisClient, target *float64, msg string) int {
	var value float64
	if err := s.getFloat64FromObject(&value); err != nil {
		if msg == "" {
			msg = "value is not an float or out of range"
		}
		ulog.ErrorP(err)
		c.addReplyError(msg)
		return REDIS_ERR
	}
	*target = value
	return REDIS_OK
}

func (s *SRobj) getLongLongFromObjectOrReply(c *SRedisClient, target *int64, msg string) int {
	var value int64
	if err := s.getLongLongFromObject(&value); err != nil {
		if msg == "" {
			msg = "value is not an integer or out of range"
		}
		ulog.ErrorP(err)
		c.addReplyError(msg)
		return REDIS_ERR
	}
	*target = value
	return REDIS_OK
}

func (s *SRobj) isObjectRepresentableAsInt64(intVal *int64) error {
	if s.Typ != SR_STR {
		ulog.ErrorF("isObjectRepresentableAsLongLong err: type fail, value.Typ = %d", s.Typ)
	}
	i, err := s.intVal()
	if intVal != nil {
		*intVal = i
	}
	return err
}

//-----------------------------------------------------------------------------
// zSet tool func
//-----------------------------------------------------------------------------

func (s *SRobj) isWithScores() bool {
	return s != nil && strings.ToLower(s.strVal()) == "withscores"
}

//-----------------------------------------------------------------------------
// object func
//-----------------------------------------------------------------------------

// return 0 obj1 == obj2, 1 obj1 > obj2, -1 obj1 < obj2
func compareStringObjects(obj1, obj2 *SRobj) int {
	if obj1.Typ != SR_STR || obj2.Typ != SR_STR {
		ulog.ErrorF("compareStringObjects err: type fail, obj1.Typ = %d obj2.Typ = %d", obj1.Typ, obj2.Typ)
	}
	return strings.Compare(obj1.strVal(), obj2.strVal())
}

//-----------------------------------------------------------------------------
// create object
//-----------------------------------------------------------------------------

func createFromInt(val int64) *SRobj {
	return &SRobj{
		Typ:      SR_STR,
		Val:      val,
		refCount: 1,
		encoding: REDIS_ENCODING_INT,
	}
}

func createFromFloat(val float64) *SRobj {
	return &SRobj{
		Typ:      SR_STR,
		Val:      val,
		refCount: 1,
		encoding: REDIS_ENCODING_INT,
	}
}

func createSRobj(typ SRType, ptr any) *SRobj {
	return &SRobj{
		Typ:      typ,
		Val:      ptr,
		refCount: 1,
		encoding: REDIS_ENCODING_RAW,
	}
}

func createFloatSRobj(typ SRType, ptr any) *SRobj {
	return &SRobj{
		Typ:      typ,
		Val:      ptr,
		refCount: 1,
		encoding: REDIS_ENCODING_INT,
	}
}

func createZsetSRobj() *SRobj {
	o := createSRobj(SR_ZSET, zSetCreate())
	o.encoding = REDIS_ENCODING_SKIPLIST
	return o
}

func createIntSetObject() *SRobj {
	o := createSRobj(SR_SET, intSetNew())
	o.encoding = REDIS_ENCODING_INTSET
	return o
}

func createSetObject() *SRobj {
	o := createSRobj(SR_SET, dictCreate(&setDictType))
	o.encoding = REDIS_ENCODING_HT
	return o
}

func createListObject() *SRobj {
	o := createSRobj(SR_LIST, listCreate())
	o.encoding = REDIS_ENCODING_LINKEDLIST
	return o
}

func createHashObject() *SRobj {
	o := createSRobj(SR_DICT, dictCreate(&dbDictType))
	o.encoding = REDIS_ENCODING_HT
	return o
}
