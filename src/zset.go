package src

type zSkipListLevel struct {
	forward *zSkipListNode
	span    uint
}

type zSkipListNode struct {
	obj      *SRobj
	score    float64
	backward *zSkipListNode
	level    []*zSkipListLevel
}

type zSkipList struct {
	header, tail *zSkipListNode
	length       uint64
	level        int
}

type zSet struct {
	zsl *zSkipList
	d   *dict
}
