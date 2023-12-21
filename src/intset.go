package src

import (
	"math"
	"simple-redis/utils"
)

// contents []int16 []int32 []int64
type intSet struct {
	encoding uint8
	length   uint32
	contents any
}

func (is *intSet) _intSetGetEncoded(pos int, enc uint8) int64 {
	if enc == uint8(utils.SizeofIn64()) {
		return is.contents.([]int64)[pos]
	}
	if enc == uint8(utils.SizeofIn32()) {
		return int64(is.contents.([]int32)[pos])
	}
	return int64(is.contents.([]int16)[pos])
}

func (is *intSet) _intSetGet(pos int) {
	is._intSetGetEncoded(pos, is.encoding)
}

func (is *intSet) _intSetSet(pos int, value int64) {
	if is.encoding == uint8(utils.SizeofIn64()) {
		is.contents.([]int64)[pos] = value
		return
	}
	if is.encoding == uint8(utils.SizeofIn32()) {
		is.contents.([]int32)[pos] = int32(value)
		return
	}
	is.contents.([]int16)[pos] = int16(value)
}

func _intSetValueEncoding(v int64) uint8 {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return uint8(utils.SizeofIn64())
	}
	if v < math.MinInt16 || v > math.MaxInt16 {
		return uint8(utils.SizeofIn32())
	}
	return uint8(utils.SizeofIn16())
}

// Create an empty intSet
func intSetNew() *intSet {
	is := new(intSet)
	is.encoding = uint8(utils.SizeofIn16())
	return is
}
