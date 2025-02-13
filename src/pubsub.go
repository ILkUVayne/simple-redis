package src

//-----------------------------------------------------------------------------
// PubSub API
//-----------------------------------------------------------------------------

func clientSubscriptionsCount(c *SRedisClient) int64 {
	return sLen(c.pubSubChannels)
}

func pubSubSubscribeChannel(c *SRedisClient, channel *SRobj) (ok bool) {
	var clients *list

	if c.pubSubChannels.dictAdd(channel, nil) {
		ok = true
		channel.incrRefCount()
		_, de := server.pubSubChannels.dictFind(channel)
		if de == nil {
			clients = listCreate()
			o := createSRobj(SR_LIST, clients)
			o.encoding = REDIS_ENCODING_LINKEDLIST
			server.pubSubChannels.dictAdd(channel, o)
		} else {
			clients = assertList(de.getVal())
		}
		clients.rPush(createSRobj(SR_STR, c))
	}
	c.addReplyMultiBulkLen(3, false)
	c.addReplyBulk(shared.subScribeBulk)
	c.addReplyBulk(channel)
	c.addReplyBulkInt(clientSubscriptionsCount(c))
	return
}

func pubSubPublishMessage(channel, message *SRobj) (receivers int64) {
	// Send to clients listening for that channel
	_, de := server.pubSubChannels.dictFind(channel)
	if de != nil {
		l := assertList(de.getVal())
		li := l.listRewind()
		for ln := li.listNext(); ln != nil; ln = li.listNext() {
			client := assertClient(ln.nodeValue().Val)
			client.addReplyMultiBulkLen(3, false)
			client.addReplyBulk(shared.messageBulk)
			client.addReplyBulk(channel)
			client.addReplyBulk(message)
			client.doReply()
			receivers++
		}
	}
	// Send to clients listening to matching channels

	return
}

//-----------------------------------------------------------------------------
// PubSub commands implementation
//-----------------------------------------------------------------------------

// usage: SUBSCRIBE channel [channel ...]
func subscribeCommand(c *SRedisClient) {
	for i := 1; i < len(c.args); i++ {
		pubSubSubscribeChannel(c, c.args[i])
	}
}

// usage: UNSUBSCRIBE [channel ...]
func unsubscribeCommand(c *SRedisClient) {

}

// usage: PUBLISH channel messages
func publishCommand(c *SRedisClient) {
	receivers := pubSubPublishMessage(c.args[1], c.args[2])
	c.addReplyLongLong(receivers)
}
