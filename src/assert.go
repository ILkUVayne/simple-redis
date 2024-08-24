package src

import (
	"github.com/ILkUVayne/utlis-go/v2/ulog"
)

// 断言 *list 类型
func assertList(o *SRobj) *list {
	l, ok := o.Val.(*list)
	if !ok {
		ulog.Error("assertList err: ", o.Typ)
	}
	return l
}

// 断言 *SRedisClient 类型
func assertClient(o any) *SRedisClient {
	c, ok := o.(*SRedisClient)
	if !ok {
		ulog.Error("assertClient err")
	}
	return c
}

// 断言 *dict 类型
func assertDict(o *SRobj) *dict {
	d, ok := o.Val.(*dict)
	if !ok {
		ulog.Error("assertDict err: ", o.Typ)
	}
	return d
}

// 断言 *intSet 类型
func assertIntSet(o *SRobj) *intSet {
	is, ok := o.Val.(*intSet)
	if !ok {
		ulog.Error("assertIntSet err: ", o.Typ)
	}
	return is
}

// 断言 *zSet 类型
func assertZSet(o *SRobj) *zSet {
	zs, ok := o.Val.(*zSet)
	if !ok {
		ulog.Error("assertZSet err: ", o.Typ)
	}
	return zs
}
