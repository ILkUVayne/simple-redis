package src

//-----------------------------------------------------------------------------
// List commands API
//-----------------------------------------------------------------------------

// 检查list对象encoding是否合法
func checkListEncoding(subject *SRobj) {
	if subject.encoding != REDIS_ENCODING_LINKEDLIST {
		panic("Unknown list encoding")
	}
}

// push data to list
//
// AL_START_HEAD lpush
// AL_START_TAIL rpush
func listTypePush(subject, value *SRobj, where int) {
	checkListEncoding(subject)
	l := assertList(subject)
	value.incrRefCount()
	if where == AL_START_HEAD {
		l.lPush(value)
		return
	}
	l.rPush(value)
}

// pop data from list
//
// AL_START_HEAD lpop
// AL_START_TAIL rpop
func listTypePop(subject *SRobj, where int) *SRobj {
	checkListEncoding(subject)
	l := assertList(subject)
	ln := l.last()
	if where == AL_START_HEAD {
		ln = l.first()
	}
	if ln == nil {
		return nil
	}
	value := ln.data
	value.incrRefCount()
	l.delNode(ln)
	return value
}

// return list length
func listTypeLength(subject *SRobj) int64 {
	checkListEncoding(subject)
	return sLen(assertList(subject))
}
