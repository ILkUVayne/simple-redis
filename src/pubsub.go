package src

//-----------------------------------------------------------------------------
// PubSub API
//-----------------------------------------------------------------------------

func pubSubSubscribeChannel(c *SRedisClient, channel *SRobj) {

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
