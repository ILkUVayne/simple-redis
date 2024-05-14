package src

import "simple-redis/utils"

func assertList(o *SRobj) *list {
	l, ok := o.Val.(*list)
	if !ok {
		utils.Error("assertList err: ", o.Typ)
	}
	return l
}

func assertClient(o any) *SRedisClient {
	c, ok := o.(*SRedisClient)
	if !ok {
		utils.Error("assertClient err")
	}
	return c
}

func assertDict(o *SRobj) *dict {
	d, ok := o.Val.(*dict)
	if !ok {
		utils.Error("assertDict err: ", o.Typ)
	}
	return d
}

func assertIntSet(o *SRobj) *intSet {
	is, ok := o.Val.(*intSet)
	if !ok {
		utils.Error("assertIntSet err: ", o.Typ)
	}
	return is
}

func assertZSet(o *SRobj) *zSet {
	zs, ok := o.Val.(*zSet)
	if !ok {
		utils.Error("assertZSet err: ", o.Typ)
	}
	return zs
}
