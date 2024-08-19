// Package src
//
// lib list provides methods for creating linked lists and adding/deleting queries
package src

// list dictType
var lType = dictType{
	hashFunc:      nil,
	keyCompare:    SRStrCompare,
	keyDestructor: nil,
	valDestructor: nil,
}

// -----------------------------------------------------------------------------
// list iterators
// -----------------------------------------------------------------------------

type listIter struct {
	next      *node
	direction int
}

// return next node
func (li *listIter) listNext() *node {
	curr := li.next
	if curr == nil {
		return curr
	}
	// default AL_START_TAIL
	li.next = curr.prev
	if li.direction == AL_START_HEAD {
		li.next = curr.next
	}
	return curr
}

// -----------------------------------------------------------------------------
// list node
// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------
// list
// -----------------------------------------------------------------------------

type list struct {
	lType  *dictType
	head   *node
	tail   *node
	length int
}

var _ Pusher = (*list)(nil)

func (l *list) len() int {
	return l.length
}

func (l *list) isEmpty() bool {
	return l.len() == 0
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

// return head list iterators
func (l *list) listRewind() *listIter {
	return &listIter{next: l.head, direction: AL_START_HEAD}
}

// return tail list iterators
func (l *list) listRewindTail() *listIter {
	return &listIter{next: l.tail, direction: AL_START_TAIL}
}

// -----------------------------------------------------------------------------
// list API
// -----------------------------------------------------------------------------

// create new list
func listCreate(lType *dictType) *list {
	return &list{lType: lType}
}
