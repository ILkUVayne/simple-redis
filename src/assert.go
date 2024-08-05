package src

import (
	"github.com/ILkUVayne/utlis-go/v2/ulog"
)

func assertList(o *SRobj) *list {
	l, ok := o.Val.(*list)
	if !ok {
		ulog.Error("assertList err: ", o.Typ)
	}
	return l
}

func assertClient(o any) *SRedisClient {
	c, ok := o.(*SRedisClient)
	if !ok {
		ulog.Error("assertClient err")
	}
	return c
}

func assertDict(o *SRobj) *dict {
	d, ok := o.Val.(*dict)
	if !ok {
		ulog.Error("assertDict err: ", o.Typ)
	}
	return d
}

func assertIntSet(o *SRobj) *intSet {
	is, ok := o.Val.(*intSet)
	if !ok {
		ulog.Error("assertIntSet err: ", o.Typ)
	}
	return is
}

func assertZSet(o *SRobj) *zSet {
	zs, ok := o.Val.(*zSet)
	if !ok {
		ulog.Error("assertZSet err: ", o.Typ)
	}
	return zs
}
