package src

// ----------------------------- base interface -------------------------

type empty interface {
	isEmpty() bool
}

type length interface {
	len() int64
}

type capacity interface {
	cap() int64
}

// ----------------------------- list interface -------------------------

type Pusher interface {
	rPush(data *SRobj)
	lPush(data *SRobj)
}
