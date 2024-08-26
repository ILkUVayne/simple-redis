package src

import (
	"math/rand"
)

// intSet set集合的底层存储结构
type intSet struct {
	length   int64
	contents []int64
}

// return value by pos
func (is *intSet) _intSetGet(pos int64) int64 {
	return is.contents[pos]
}

// set value in pos
func (is *intSet) _intSetSet(pos int64, value int64) {
	if pos > sLen(is) {
		is.contents[pos] = value
		return
	}
	copy(is.contents[pos+1:], is.contents[pos:])
	is.contents[pos] = value
}

// remove value by pos
func (is *intSet) _intSetRemove(pos int64) {
	is.contents = append(is.contents[:pos], is.contents[pos+1:]...)
}

// 获取value在intSet中的位置pos，存在时，pos为对应的索引并返回ture，不存在时，pos为插入value的索引并返回false
func (is *intSet) intSetSearch(value int64, pos *int64) bool {
	minIdx, midIdx, maxIdx := int64(0), int64(-1), sLen(is)
	cur := int64(-1)
	if sLen(is) == 0 {
		*pos = 0
		return false
	}
	if value > is._intSetGet(sLen(is)-1) {
		*pos = sLen(is)
		return false
	}
	if value < is._intSetGet(0) {
		*pos = 0
		return false
	}
	for maxIdx >= minIdx {
		midIdx = (minIdx + maxIdx) / 2
		cur = is._intSetGet(midIdx)
		if value == cur {
			break
		}
		if value > cur {
			minIdx = midIdx + 1
			continue
		}
		maxIdx = midIdx - 1
	}

	if value == cur {
		*pos = midIdx
		return true
	}
	*pos = minIdx
	return false
}

// return true if existed,false non-existent
func (is *intSet) intSetFind(value int64) bool {
	var pos int64
	return is.intSetSearch(value, &pos)
}

// 调整intSet容量
func (is *intSet) intSetResize() {
	newContents := make([]int64, sLen(is)*2)
	copy(newContents, is.contents)
	is.contents = newContents
}

// add value to intSet
func (is *intSet) intSetAdd(value int64, success *bool) *intSet {
	if sLen(is) == int64(len(is.contents)) {
		is.intSetResize()
	}
	var pos int64
	if is.intSetSearch(value, &pos) {
		*success = false
		return is
	}
	is._intSetSet(pos, value)
	is.length++
	*success = true
	return is
}

// remove value from intSet if exist
func (is *intSet) intSetRemove(value int64) {
	var pos int64
	if !is.intSetSearch(value, &pos) {
		return
	}
	is._intSetRemove(pos)
	is.length--
}

// return a Random value from intSet
func (is *intSet) intSetRandom() int64 {
	return is._intSetGet(rand.Int63n(sLen(is)))
}

// set value to intSet
func (is *intSet) intSetGet(pos int64, value *int64) bool {
	if pos < sLen(is) {
		*value = is._intSetGet(pos)
		return true
	}
	return false
}

// return intSet len
func (is *intSet) len() int64 {
	return is.length
}

// return ture if it is Empty
func (is *intSet) isEmpty() bool {
	return sLen(is) == 0
}

// Create an empty intSet
func intSetNew() *intSet {
	return &intSet{contents: make([]int64, DEFAULT_INTSET_BUF)}
}
