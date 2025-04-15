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

}
