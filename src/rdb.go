package src

func rdbSave(filename *string) int {
	// todo
	return REDIS_OK
}

func rdbSaveBackground(filename *string) int {
	// todo
	return REDIS_OK
}

func backgroundSaveDoneHandler() {

}

//-----------------------------------------------------------------------------
// rdb commands
//-----------------------------------------------------------------------------

func saveCommand(c *SRedisClient) {
	if server.rdbChildPid != -1 {
		c.addReplyError("Background save already in progress")
		return
	}
	if rdbSave(&server.rdbFilename) == REDIS_OK {
		c.addReply(shared.ok)
		return
	}
	c.addReply(shared.err)
}

func bgSaveCommand(c *SRedisClient) {
	if server.rdbChildPid != -1 {
		c.addReplyError("Background save already in progress")
		return
	}
	if server.aofChildPid != -1 {
		c.addReplyError("Can't BGSAVE while AOF log rewriting is in progress")
		return
	}
	if rdbSaveBackground(&server.rdbFilename) == REDIS_OK {
		c.addReplyStatus("Background saving started")
		return
	}
	c.addReply(shared.err)
}
