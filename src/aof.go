package src

import (
	"bufio"
	"os"
	"simple-redis/utils"
)

func createFakeClient() *SRedisClient {
	c := new(SRedisClient)
	c.fd = -1
	c.db = server.db
	c.cmdTyp = CMD_UNKNOWN
	c.reply = listCreate(&listType{keyCompare: SRStrCompare})
	c.queryBuf = make([]byte, SREDIS_IO_BUF)
	c.replyReady = true
	return c
}

func loadAppendOnlyFile(name string) {
	fp, err := os.Open(name)
	if err != nil {
		utils.Error("Can't open the append-only file: ", err)
	}
	defer func() { _ = fp.Close() }()

	scanner := bufio.NewScanner(fp)
	fakeClient := createFakeClient()
	var args []*SRobj
	aLen := int64(-1)
	for scanner.Scan() {
		if aLen <= 0 {
			args = make([]*SRobj, 0)
		}

		str := scanner.Text()
		if str[0] == '*' {
			str = str[1:]
			if utils.String2Int64(&str, &aLen) == REDIS_ERR {
				utils.Error("Bad file format reading the append only file")
			}
			continue
		}
		if str[0] == '$' {
			continue
		}
		args = append(args, createSRobj(SR_STR, str))
		aLen--
		if aLen == 0 {
			fakeClient.args = args
			processCommand(fakeClient)
			fakeClient.args = nil
			args = nil
		}
	}
}
