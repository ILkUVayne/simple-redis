package src

import (
	"simple-redis/utils"
	"strconv"
	"strings"
)

const (
	REDIS_ENCODING_RAW        uint8 = iota // Raw representation
	REDIS_ENCODING_INT                     // Encoded as integer
	REDIS_ENCODING_HT                      // Encoded as hash table
	REDIS_ENCODING_ZIPMAP                  // Encoded as zipmap
	REDIS_ENCODING_LINKEDLIST              // Encoded as regular linked list
	REDIS_ENCODING_ZIPLIST                 // Encoded as ziplist
	REDIS_ENCODING_INTSET                  // Encoded as intset
	REDIS_ENCODING_SKIPLIST                // Encoded as skiplist
)

type SRType uint8

// SR_STR 字符串类型
// SR_LIST 列表类型
// SR_SET 集合类型
// SR_ZSET 有序集合类型
// SR_DICT 字典类型
const (
	SR_STR SRType = iota
	SR_LIST
	SR_SET
	SR_ZSET
	SR_DICT
)

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

func (s *SRobj) intVal() int64 {
	if s.Typ != SR_STR {
		return 0
	}
	i, _ := strconv.ParseInt(s.Val.(string), 10, 64)
	return i
}

func (s *SRobj) strEncoding() string {
	switch s.encoding {
	case REDIS_ENCODING_RAW:
		return "raw"
	case REDIS_ENCODING_INT:
		return "int"
	case REDIS_ENCODING_HT:
		return "hashtable"
	case REDIS_ENCODING_LINKEDLIST:
		return "linkedlist"
	case REDIS_ENCODING_ZIPLIST:
		return "ziplist"
	case REDIS_ENCODING_INTSET:
		return "intset"
	case REDIS_ENCODING_SKIPLIST:
		return "skiplist"
	default:
		return "unknown"
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

func createFromInt(val int64) *SRobj {
	return &SRobj{
		Typ:      SR_STR,
		Val:      strconv.FormatInt(val, 10),
		refCount: 1,
	}
}

// return 0 obj1 == obj2, 1 obj1 > obj2, -1 obj1 < obj2
func compareStringObjects(obj1, obj2 *SRobj) int {
	if obj1.Typ != SR_STR || obj2.Typ != SR_STR {
		utils.ErrorF("compareStringObjects err: type fail, obj1.Typ = %d obj2.Typ = %d", obj1.Typ, obj2.Typ)
	}
	return strings.Compare(obj1.strVal(), obj2.strVal())
}
