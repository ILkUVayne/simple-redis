package src

//-----------------------------------------------------------------------------
// Sorted set commands
//-----------------------------------------------------------------------------

func zAddGenericCommand(c *SRedisClient, incr int) {
	//nanErr := errors.New("resulting score is not a number (NaN)")
	//key := c.args[1]

	elements := len(c.args[2:]) / 2

	if len(c.args)%2 == 1 {
		c.addReply(shared.syntaxErr)
		return
	}

	scores := make([]float64, elements)
	for i := 0; i < elements; i++ {
		if c.args[2+i*2].getFloat64FromObjectOrReply(c, &scores[i], nil) == REDIS_ERR {
			return
		}
	}
}
