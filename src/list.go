package src

//-----------------------------------------------------------------------------
// List commands API
//-----------------------------------------------------------------------------

func checkListEncoding(subject *SRobj) {
	if subject.encoding != REDIS_ENCODING_LINKEDLIST {
		panic("Unknown list encoding")
	}
}

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

func listTypeLength(subject *SRobj) int {
	checkListEncoding(subject)
	return assertList(subject).len()
}
