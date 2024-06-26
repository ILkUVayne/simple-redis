package src

import (
	"simple-redis/utils"
	"strconv"
	"strings"
)

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

type SRType uint8

type SRVal any

type SRobj struct {
	Typ      SRType
	Val      SRVal
	encoding uint8
	refCount int
}

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
	s.refCount++
}

func (s *SRobj) decrRefCount() {
	s.refCount--
	// gc 自动回收
	if s.refCount == 0 {
		s.Val = nil
	}
}

func (s *SRobj) intVal() (int64, int) {
	if s.Typ != SR_STR {
		return 0, REDIS_ERR
	}
	if s.encoding == REDIS_ENCODING_INT {
		return s.Val.(int64), REDIS_OK
	}
	if s.encoding == REDIS_ENCODING_RAW {
		var i int64
		str := s.strVal()
		return i, utils.String2Int64(&str, &i)
	}
	panic("Unknown string encoding")
}

func (s *SRobj) floatVal() (float64, int) {
	if s.Typ != SR_STR {
		return 0, REDIS_ERR
	}
	if s.encoding == REDIS_ENCODING_INT {
		return s.Val.(float64), REDIS_OK
	}
	if s.encoding == REDIS_ENCODING_RAW {
		var i float64
		str := s.strVal()
		return i, utils.String2Float64(&str, &i)
	}
	panic("Unknown string encoding")
}

func (s *SRobj) strEncoding() string {
	encoding, ok := encodingMaps[s.encoding]
	if !ok {
		return "unknown"
	}
	return encoding
}

func (s *SRobj) strType() string {
	typ, ok := TypeMaps[s.Typ]
	if !ok {
		return "unknown"
	}
	return typ
}

func (s *SRobj) getEncoding() *SRobj {
	return createSRobj(SR_STR, s.strEncoding())
}

func (s *SRobj) getType() *SRobj {
	return createSRobj(SR_STR, s.strType())
}

func (s *SRobj) checkType(c *SRedisClient, typ SRType) bool {
	if s.Typ != typ {
		c.addReply(shared.wrongTypeErr)
		return false
	}
	return true
}

func (s *SRobj) tryObjectEncoding() {
	if s.encoding != REDIS_ENCODING_RAW || s.refCount > 1 || s.Typ != SR_STR {
		return
	}
	// Check if we can represent this string as a long integer
	var i int64
	str := s.strVal()
	if utils.String2Int64(&str, &i) == REDIS_ERR {
		return
	}
	s.encoding = REDIS_ENCODING_INT
	s.Val = i
}

func (s *SRobj) getDecodedObject() *SRobj {
	if s.encoding == REDIS_ENCODING_RAW {
		s.incrRefCount()
		return s
	}
	if s.Typ == SR_STR && s.encoding == REDIS_ENCODING_INT {
		var intVal int64
		str := s.strVal()
		if utils.String2Int64(&str, &intVal) == REDIS_ERR {
			utils.Error("getDecodedObject err")
		}
		return createFromInt(intVal)
	}
	panic("Unknown encoding type")
}

func (s *SRobj) getLongLongFromObject(target *int64) int {
	if s == nil {
		*target = 0
		return REDIS_OK
	}
	if s.Typ != SR_STR {
		return REDIS_ERR
	}
	intVal, res := s.intVal()
	*target = intVal
	return res
}

func (s *SRobj) getFloat64FromObject(target *float64) int {
	if s.Typ != SR_STR {
		return REDIS_ERR
	}
	i, res := s.floatVal()
	*target = i
	return res
}

func (s *SRobj) getFloat64FromObjectOrReply(c *SRedisClient, target *float64, msg *string) int {
	var value float64
	if s.getFloat64FromObject(&value) == REDIS_ERR {
		if msg != nil {
			c.addReplyError(*msg)
			return REDIS_ERR
		}
		c.addReplyError("value is not an float or out of range")
		return REDIS_ERR
	}
	*target = value
	return REDIS_OK
}

func (s *SRobj) getLongLongFromObjectOrReply(c *SRedisClient, target *int64, msg *string) int {
	var value int64
	if s.getLongLongFromObject(&value) == REDIS_ERR {
		if msg != nil {
			c.addReplyError(*msg)
			return REDIS_ERR
		}
		c.addReplyError("value is not an integer or out of range")
		return REDIS_ERR
	}
	*target = value
	return REDIS_OK
}

func (s *SRobj) isObjectRepresentableAsInt64(intVal *int64) int {
	if s.Typ != SR_STR {
		utils.ErrorF("isObjectRepresentableAsLongLong err: type fail, value.Typ = %d", s.Typ)
	}
	i, res := s.intVal()
	if intVal != nil {
		*intVal = i
	}
	return res
}

//-----------------------------------------------------------------------------
// object func
//-----------------------------------------------------------------------------

// return 0 obj1 == obj2, 1 obj1 > obj2, -1 obj1 < obj2
func compareStringObjects(obj1, obj2 *SRobj) int {
	if obj1.Typ != SR_STR || obj2.Typ != SR_STR {
		utils.ErrorF("compareStringObjects err: type fail, obj1.Typ = %d obj2.Typ = %d", obj1.Typ, obj2.Typ)
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
	zs := new(zSet)
	zs.zsl = zslCreate()
	zs.d = dictCreate(&zSetDictType)
	o := createSRobj(SR_ZSET, zs)
	o.encoding = REDIS_ENCODING_SKIPLIST
	return o
}

func createIntSetObject() *SRobj {
	is := intSetNew()
	o := createSRobj(SR_SET, is)
	o.encoding = REDIS_ENCODING_INTSET
	return o
}

func createSetObject() *SRobj {
	d := dictCreate(&setDictType)
	o := createSRobj(SR_SET, d)
	o.encoding = REDIS_ENCODING_HT
	return o
}

func createListObject() *SRobj {
	l := listCreate(&lType)
	o := createSRobj(SR_LIST, l)
	o.encoding = REDIS_ENCODING_LINKEDLIST
	return o
}

func createHashObject() *SRobj {
	h := dictCreate(&dbDictType)
	o := createSRobj(SR_DICT, h)
	o.encoding = REDIS_ENCODING_HT
	return o
}
