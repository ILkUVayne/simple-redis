package src

//-----------------------------------------------------------------------------
// Hash type commands API
//-----------------------------------------------------------------------------

// return hash obj,create new hash obj if null
func hashTypeLookupWriteOrCreate(c *SRedisClient, key *SRobj) *SRobj {
	o := c.db.lookupKeyWrite(key)
	if o == nil {
		o = createHashObject()
		c.db.dictSet(key, o)
	}

	if o != nil && o.Typ != SR_DICT {
		c.addReply(shared.wrongTypeErr)
		return nil
	}

	return o
}

func hashTypeTryObjectEncoding(subject, o1, o2 *SRobj) {
	if subject.encoding == REDIS_ENCODING_HT {
		if o1 != nil {
			o1.tryObjectEncoding()
		}
		if o2 != nil {
			o2.tryObjectEncoding()
		}
	}
}

// hash obj set by encoding type
func hashTypeSet(o, field, value *SRobj) int {
	if o.encoding == REDIS_ENCODING_HT {
		return assertDict(o).dictSet(field, value)
	}
	panic("Unknown hash encoding")
}

func hashTypeGetFromHashTable(o, field *SRobj, value **SRobj) bool {
	if o.encoding != REDIS_ENCODING_HT {
		panic("Unknown hash encoding")
	}
	v := assertDict(o).dictGet(field)
	if v == nil {
		return false
	}
	*value = v
	return true
}

func addHashFieldToReply(c *SRedisClient, o, field *SRobj) {
	if o == nil {
		c.addReply(shared.nullBulk)
		return
	}

	if o.encoding == REDIS_ENCODING_HT {
		var value *SRobj
		if hashTypeGetFromHashTable(o, field, &value) {
			c.addReplyBulk(value)
			return
		}
		c.addReply(shared.nullBulk)
		return
	}

	panic("Unknown hash encoding")
}

// return hash ogj length by encoding
func hashTypeLength(o *SRobj) int64 {
	if o.encoding == REDIS_ENCODING_HT {
		return assertDict(o).dictSize()
	}
	panic("Unknown hash encoding")
}
