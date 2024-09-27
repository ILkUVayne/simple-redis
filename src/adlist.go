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
	direction int // list 下一个元素的迭代方向， AL_START_HEAD => node.next ， AL_START_TAIL => node.prev
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

// list node 是一个双向链表
type node struct {
	data *SRobj
	prev *node
	next *node
}

// 当前 node 结点数据
func (n *node) nodeValue() *SRobj {
	return n.data
}

// 当前 node 结点的下一个结点指针
func (n *node) nodeNext() *node {
	return n.next
}

// 当前 node 结点的上一个结点指针
func (n *node) nodePrev() *node {
	return n.prev
}

// -----------------------------------------------------------------------------
// list
// -----------------------------------------------------------------------------

type list struct {
	lType  *dictType
	head   *node // list 头结点指针
	tail   *node // list 尾结点指针
	length int64
}

var _ Pusher = (*list)(nil)

// return list length
func (l *list) len() int64 {
	return l.length
}

// 判断 list 是否为空（长度是否为零）
func (l *list) isEmpty() bool {
	return sLen(l) == 0
}

// 返回list头结点
func (l *list) first() *node {
	return l.head
}

// 返回list尾结点
func (l *list) last() *node {
	return l.tail
}

// 创建 node 结点，若 list 为空，则直接插入并返回true，反之则返回创建的 node 并返回false
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

// 在 node 链表的尾部插入新的数据结点
func (l *list) rPush(data *SRobj) {
	var n *node
	var res bool
	if res, n = l._push(data); res {
		return
	}

	// 假设当前链表结构：nil <- A <-> B <-> C -> nil，此时 n 为需要新加入的结点

	// nil <- A <-> B <-> C -> nil , C <- n , l.tail = C
	n.prev = l.tail
	// nil <- A <-> B <-> C <-> n , l.tail = C
	l.tail.next = n
	// nil <- A <-> B <-> C <-> n , l.tail = n
	l.tail = n
}

// // 在 node 链表的头部插入新的数据结点
func (l *list) lPush(data *SRobj) {
	var n *node
	var res bool
	if res, n = l._push(data); res {
		return
	}

	// 假设当前链表结构：nil <- A <-> B <-> C -> nil，此时 n 为需要新加入的结点

	// nil <- A <-> B <-> C -> nil , n -> A , l.head = A
	n.next = l.head
	// n <-> A <-> B <-> C -> nil , l.head = A
	l.head.prev = n
	// n <-> A <-> B <-> C -> nil , l.head = n
	l.head = n
}

// 根据数据查找 list 结点，不存在则返回nil
func (l *list) find(data *SRobj) *node {
	p := l.head
	for ; p != nil; p = p.next {
		if l.lType.keyCompare(data, p.data) {
			break
		}
	}
	return p
}

// 根据 node 结点删除
func (l *list) delNode(n *node) {
	if n == nil || l.isEmpty() {
		return
	}

	l.length--
	// 删除的结点为头结点
	if l.head == n {
		if n.next != nil {
			n.next.prev = nil
		}
		l.head = n.next
		n.next = nil
		return
	}
	// 删除的结点为尾结点
	if l.tail == n {
		if n.prev != nil {
			n.prev.next = nil
		}
		l.tail = n.prev
		n.prev = nil
		return
	}
	// 删除的结点为其他结点
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	n.next = nil
	n.prev = nil
}

// 删除结点
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
func listCreate() *list {
	return &list{lType: &lType}
}
