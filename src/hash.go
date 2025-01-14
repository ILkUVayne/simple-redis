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

// tryObjectEncoding o1 and o2
func hashTypeTryObjectEncoding(subject, o1, o2 *SRobj) {
	checkHashEncoding(subject)
	if o1 != nil {
		o1.tryObjectEncoding()
	}
	if o2 != nil {
		o2.tryObjectEncoding()
	}
}

// 检查hash对象的encoding是否合法
func checkHashEncoding(subject *SRobj) {
	if subject.encoding != REDIS_ENCODING_HT {
		panic("Unknown hash encoding")
	}
}

// hash obj set by encoding type
func hashTypeSet(o, field, value *SRobj) int {
	checkHashEncoding(o)
	return assertDict(o).dictSet(field, value)
}

// get value form hashTable encoding, return true when field Exists
func hashTypeGetFromHashTable(o, field *SRobj, value **SRobj) bool {
	checkHashEncoding(o)
	v := assertDict(o).dictGet(field)
	if v == nil {
		return false
	}
	*value = v
	return true
}

// 获取哈希对象 o 中的关键字 field 对应的值 value ，并添加到reply中。若不存在则返回nil
func addHashFieldToReply(c *SRedisClient, o, field *SRobj) {
	if o == nil {
		c.addReply(shared.nullBulk)
		return
	}

	checkHashEncoding(o)
	var value *SRobj
	if hashTypeGetFromHashTable(o, field, &value) {
		c.addReplyBulk(value)
		return
	}
	c.addReply(shared.nullBulk)
}

// return hash ogj length by encoding
func hashTypeLength(o *SRobj) int64 {
	checkHashEncoding(o)
	return sLen(assertDict(o))
}

// del hash data by field
func hashTypeDel(o, field *SRobj) bool {
	checkHashEncoding(o)
	deleted := false
	d := assertDict(o)
	if d.dictDelete(field) == REDIS_OK {
		deleted = true
	}
	if d.htNeedResize() {
		d.dictResize()
	}
	return deleted
}

// 检查field对应的值是否存在
func hashTypeExists(o, field *SRobj) bool {
	checkHashEncoding(o)
	var value *SRobj
	return hashTypeGetFromHashTable(o, field, &value)
}
