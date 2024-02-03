package src

//-----------------------------------------------------------------------------
// List commands API
//-----------------------------------------------------------------------------

func listTypePush(subject, value *SRobj, where int) {
	if subject.encoding == REDIS_ENCODING_LINKEDLIST {
		l := assertList(subject)
		if where == REDIS_HEAD {
			l.lPush(value)
		} else {
			l.rPush(value)
		}
		value.incrRefCount()
		return
	}
	panic("Unknown list encoding")
}

func listTypePop(subject *SRobj, where int) *SRobj {
	var value *SRobj
	if subject.encoding == REDIS_ENCODING_LINKEDLIST {
		l := assertList(subject)
		var ln *node
		if where == REDIS_HEAD {
			ln = l.first()
		} else {
			ln = l.last()
		}
		if ln != nil {
			value = ln.data
			value.incrRefCount()
			l.delNode(ln)
		}
		return value
	}
	panic("Unknown list encoding")
}

func listTypeLength(subject *SRobj) int {
	if subject.encoding == REDIS_ENCODING_LINKEDLIST {
		return assertList(subject).len()
	}
	panic("Unknown list encoding")
}
