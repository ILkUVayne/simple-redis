package src

import "testing"

func TestSRedisAppendCommandArg(t *testing.T) {
	args := []string{"set", "name", "tom"}
	c := new(sRedisContext)
	sRedisAppendCommandArg(c, args)
	if string(c.oBuf) != "*3\r\n$3\r\nset\r\n$4\r\nname\r\n$3\r\ntom\r\n" {
		t.Error("sRedisAppendCommandArg err: string(c.oBuf) = ", string(c.oBuf))
	}
}
