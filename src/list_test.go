package src

import (
	"testing"
)

func TestLen(t *testing.T) {
	l := listCreate(&listType{keyCompare: SRStrCompare})
	data := createSRobj(SR_STR, "name1")
	if l.len() != 0 {
		t.Error("get len err: len == ", l.len())
	}
	l.lPush(data)
	if l.len() != 1 {
		t.Error("get len err: len == ", l.len())
	}
}

func TestFirst(t *testing.T) {
	l := listCreate(&listType{keyCompare: SRStrCompare})
	data := createSRobj(SR_STR, "name1")
	l.lPush(data)
	data1 := createSRobj(SR_STR, "name1")
	l.lPush(data1)
	n := l.first()
	if !l.lType.keyCompare(n.data, data1) {
		t.Error("get first err: n == ", n.data.strVal())
	}
}

func TestLast(t *testing.T) {
	l := listCreate(&listType{keyCompare: SRStrCompare})
	data := createSRobj(SR_STR, "name1")
	l.lPush(data)
	data1 := createSRobj(SR_STR, "name1")
	l.lPush(data1)
	n := l.first()
	if !l.lType.keyCompare(n.data, data) {
		t.Error("get first err: n == ", n.data.strVal())
	}
}

func TestFind(t *testing.T) {
	l := listCreate(&listType{keyCompare: SRStrCompare})
	data := createSRobj(SR_STR, "name1")
	l.lPush(data)
	data1 := createSRobj(SR_STR, "name1")
	l.lPush(data1)
	data2 := createSRobj(SR_STR, "name2")
	n := l.find(data)
	if !l.lType.keyCompare(n.data, data) {
		t.Error("get first err: n == ", n.data.strVal())
	}
	n = l.find(data2)
	if n != nil {
		t.Error("find err: n == ", n)
	}
}

func TestRPush(t *testing.T) {
	l := listCreate(&listType{keyCompare: SRStrCompare})
	data := createSRobj(SR_STR, "name1")
	l.rPush(data)
	data1 := createSRobj(SR_STR, "name1")
	l.rPush(data1)
	n := l.first()
	if !l.lType.keyCompare(n.data, data) {
		t.Error("rPush err: n == ", n.data.strVal())
	}
}

func TestLPush(t *testing.T) {
	l := listCreate(&listType{keyCompare: SRStrCompare})
	data := createSRobj(SR_STR, "name1")
	l.rPush(data)
	data1 := createSRobj(SR_STR, "name1")
	l.rPush(data1)
	n := l.last()
	if !l.lType.keyCompare(n.data, data) {
		t.Error("rPush err: n == ", n.data.strVal())
	}
}

func TestDelNode(t *testing.T) {
	l := listCreate(&listType{keyCompare: SRStrCompare})
	data := createSRobj(SR_STR, "name1")
	l.rPush(data)
	data1 := createSRobj(SR_STR, "name2")
	l.rPush(data1)
	n := l.find(data)
	if !l.lType.keyCompare(n.data, data) {
		t.Error("rPush err: n == ", n.data.strVal())
	}
	l.delNode(n)
	n = l.find(data)
	if n != nil {
		t.Error("find err: n == ", n)
	}
}
