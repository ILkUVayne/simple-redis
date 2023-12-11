package src

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

func createSRobj(typ SRType, ptr any) *SRobj {
	return &SRobj{
		Typ:      typ,
		Val:      ptr,
		refCount: 1,
	}
}