package src

import (
	"testing"
)

func TestIsFake(t *testing.T) {
	c := createSRClient(1)
	if c.isFake() {
		t.Error("isFake err: c.fd = ", c.fd)
	}
	fakeC := createFakeClient()
	if !fakeC.isFake() {
		t.Error("isFake err: c.fd = ", fakeC.fd)
	}
}

func TestCreateClient(t *testing.T) {
	c := createSRClient(4)
	if c.fd != 4 {
		t.Error("create client fail,fd == ", c.fd)
	}
}

func TestGetQueryLine(t *testing.T) {
	c := createSRClient(5)
	buf := []byte("set name 5\r\n")
	c.queryBuf = buf
	c.queryLen = len(buf)
	idx, err := c.getQueryLine()
	if err != nil {
		t.Error("getQueryLine error: ", err)
	}
	if idx != 10 {
		t.Error("getQueryLine error: idx ==", idx)
	}
}

func TestGetQueryNum(t *testing.T) {
	c := createSRClient(6)
	buf := []byte("$3\r\nset\r\n")
	c.queryBuf = buf
	c.queryLen = len(buf)
	idx, _ := c.getQueryLine()
	num, err := c.getQueryNum(1, idx)
	if err != nil {
		t.Error("getQueryNum error: ", err)
	}
	if num != 3 {
		t.Error("getQueryNum error: num ==", num)
	}
}

func TestFreeArgs(t *testing.T) {
	c := createSRClient(7)
	c.args = make([]*SRobj, 2)
	c.args[0] = createSRobj(SR_STR, "get")
	c.args[1] = createSRobj(SR_STR, "name")
	if c.args[0].strVal() != "get" || c.args[1].strVal() != "name" {
		t.Error("createSRobj error")
	}
	freeArgs(c)
	if c.args[0].Val != nil || c.args[1].Val != nil {
		t.Error("freeArgs error")
	}
}

func TestInlineBufHandle(t *testing.T) {
	c := createSRClient(8)
	buf := []byte("get name\r\n")
	c.queryBuf = buf
	c.queryLen = len(buf)
	_, err := inlineBufHandle(c)
	if err != nil {
		t.Error("inlineBufHandle err: ", err)
	}
	if len(c.args) != 2 || c.args[0].strVal() != "get" || c.args[1].strVal() != "name" {
		t.Error("c.args set error")
	}
}

func TestBulkBufHandle(t *testing.T) {
	c := createSRClient(9)
	buf := []byte("*2\r\n$3\r\nget\r\n$4\r\nname\r\n")
	c.queryBuf = buf
	c.queryLen = len(buf)
	_, err := bulkBufHandle(c)
	if err != nil {
		t.Error("bulkBufHandle err: ", err)
	}
	if len(c.args) != 2 || c.args[0].strVal() != "get" || c.args[1].strVal() != "name" {
		t.Error("c.args set error")
	}
}
