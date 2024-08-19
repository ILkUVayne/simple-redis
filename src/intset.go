package src

import (
	"math/rand"
)

type intSet struct {
	length   uint32
	contents []int64
}

func (is *intSet) _intSetGet(pos int) int64 {
	return is.contents[pos]
}

func (is *intSet) _intSetSet(pos int, value int64) {
	if pos > int(is.length) {
		is.contents[pos] = value
		return
	}
	copy(is.contents[pos+1:], is.contents[pos:])
	is.contents[pos] = value
}

func (is *intSet) _intSetRemove(pos int) {
	is.contents = append(is.contents[:pos], is.contents[pos+1:]...)
}

func (is *intSet) intSetSearch(value int64, pos *uint32) bool {
	minIdx, midIdx, maxIdx := 0, -1, int(is.length)
	cur := int64(-1)
	if is.length == 0 {
		*pos = 0
		return false
	}
	if value > is._intSetGet(int(is.length)-1) {
		*pos = is.length
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
		*pos = uint32(midIdx)
		return true
	}
	*pos = uint32(minIdx)
	return false
}

// return true if existed,false non-existent
func (is *intSet) intSetFind(value int64) bool {
	var pos uint32
	return is.intSetSearch(value, &pos)
}

func (is *intSet) intSetResize() {
	length := is.length * 2
	newContents := make([]int64, length)
	copy(newContents, is.contents)
	is.contents = newContents
}

func (is *intSet) intSetAdd(value int64, success *bool) *intSet {
	if is.length == uint32(len(is.contents)) {
		is.intSetResize()
	}
	var pos uint32
	if is.intSetSearch(value, &pos) {
		*success = false
		return is
	}
	is._intSetSet(int(pos), value)
	is.length++
	*success = true
	return is
}

func (is *intSet) intSetRemove(value int64) {
	var pos uint32
	if !is.intSetSearch(value, &pos) {
		return
	}
	is._intSetRemove(int(pos))
	is.length--
}

func (is *intSet) intSetRandom() int64 {
	return is._intSetGet(rand.Intn(int(is.length)))
}

func (is *intSet) intSetGet(pos uint32, value *int64) bool {
	if pos < is.length {
		*value = is._intSetGet(int(pos))
		return true
	}
	return false
}

func (is *intSet) intSetLen() uint32 {
	return is.length
}

func (is *intSet) isEmpty() bool {
	return is.length == 0
}

// Create an empty intSet
func intSetNew() *intSet {
	return &intSet{contents: make([]int64, DEFAULT_INTSET_BUF)}
}
