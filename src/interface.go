package src

// ----------------------------- base interface -------------------------

type empty interface {
	isEmpty() bool
}

// ----------------------------- list interface -------------------------

type Pusher interface {
	rPush(data *SRobj)
	lPush(data *SRobj)
}
