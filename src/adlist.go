package src

type Pusher interface {
	rPush(data *SRobj)
	lPush(data *SRobj)
}

type listIter struct {
	next      *node
	direction int
}

func (li *listIter) listNext() *node {
	curr := li.next
	if curr != nil {
		if li.direction == AL_START_HEAD {
			li.next = curr.next
		} else {
			li.next = curr.prev
		}
	}
	return curr
}

type listType struct {
	keyCompare func(key1, key2 *SRobj) bool
}

var lType = listType{keyCompare: SRStrCompare}

type node struct {
	data *SRobj
	prev *node
	next *node
}

func (n *node) nodeValue() *SRobj {
	return n.data
}

func (n *node) nodeNext() *node {
	return n.next
}

func (n *node) nodePrev() *node {
	return n.prev
}

type list struct {
	lType  *listType
	head   *node
	tail   *node
	length int
}

var _ Pusher = (*list)(nil)

func (l *list) len() int {
	return l.length
}

func (l *list) first() *node {
	return l.head
}

func (l *list) last() *node {
	return l.tail
}

func (l *list) _push(data *SRobj) (bool, *node) {
	n := new(node)
	n.data = data
	l.length++
	// list empty
	if l.head == nil {
		l.head = n
		l.tail = n
		return true, nil
	}
	return false, n
}

func (l *list) rPush(data *SRobj) {
	var n *node
	var res bool
	if res, n = l._push(data); res {
		return
	}
	n.prev = l.tail
	l.tail.next = n
	l.tail = n
}

func (l *list) lPush(data *SRobj) {
	var n *node
	var res bool
	if res, n = l._push(data); res {
		return
	}
	n.next = l.head
	l.head.prev = n
	l.head = n
}

func (l *list) find(data *SRobj) *node {
	p := l.head
	for ; p != nil; p = p.next {
		if l.lType.keyCompare(data, p.data) {
			break
		}
	}
	return p
}

func (l *list) delNode(n *node) {
	if n == nil {
		return
	}
	if l.length == 0 {
		return
	}
	l.length--

	if l.head == n {
		if n.next != nil {
			n.next.prev = nil
		}
		l.head = n.next
		n.next = nil
		return
	}

	if l.tail == n {
		if n.prev != nil {
			n.prev.next = nil
		}
		l.tail = n.prev
		n.prev = nil
		return
	}

	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	n.next = nil
	n.prev = nil
}

func (l *list) del(data *SRobj) {
	l.delNode(l.find(data))
}

func (l *list) listRewind() *listIter {
	li := new(listIter)
	li.next = l.head
	li.direction = AL_START_HEAD
	return li
}

func (l *list) listRewindTail() *listIter {
	li := new(listIter)
	li.next = l.tail
	li.direction = AL_START_TAIL
	return li
}

func listCreate(lType *listType) *list {
	l := new(list)
	l.lType = lType
	return l
}
