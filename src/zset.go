package src

import (
	"github.com/ILkUVayne/utlis-go/v2/str"
	"math/rand"
)

//-----------------------------------------------------------------------------
// Sorted set commands API
//-----------------------------------------------------------------------------

// ================================ Parse Range =================================

type zRangeSpec struct {
	min, max     float64
	minex, maxex int
}

func parseParentheses(s string) (float64, int, error) {
	var i float64
	if err := str.String2Float64(s, &i); err != nil {
		return 0, 0, err
	}
	return i, 1, nil
}

func zslParseRange(min *SRobj, max *SRobj) (*zRangeSpec, error) {
	var err error
	spec := new(zRangeSpec)
	spec.min, spec.minex, err = _parseRange(min)
	if err != nil {
		return nil, err
	}
	spec.max, spec.maxex, err = _parseRange(max)
	return spec, err
}

var zSetDictType = dictType{
	hashFunc:      SRStrHash,
	keyCompare:    SRStrCompare,
	keyDestructor: nil,
	valDestructor: nil,
}

type zSkipListNodeLevel struct {
	forward *zSkipListNode // 前进指针
	span    int64          // 跨度
}

// ================================ skipList node =================================

// 跳表节点
type zSkipListNode struct {
	obj      *SRobj                // 成员对象
	score    float64               // 分数
	backward *zSkipListNode        // 后退指针
	level    []*zSkipListNodeLevel // 层
}

func (zn *zSkipListNode) freeNode() {
	zn.obj.decrRefCount()
	zn.obj = nil
	zn.backward = nil
	zn.level = nil
}

// create zSkipListNode
func zslCreateNode(level int, score float64, obj *SRobj) *zSkipListNode {
	zslNode := new(zSkipListNode)
	zslNode.obj = obj
	zslNode.score = score
	zslNode.level = make([]*zSkipListNodeLevel, level)
	for i := 0; i < level; i++ {
		zslNode.level[i] = new(zSkipListNodeLevel)
	}
	return zslNode
}

// ================================== skipList ===================================

// 跳表
type zSkipList struct {
	header, tail *zSkipListNode // 表头、表尾节点指针
	length       int64          // 节点数量
	level        int            // 表中节点最高的层数
}

func (z *zSkipList) free() {
	var next *zSkipListNode
	zslNode := z.header.level[0].forward
	// free
	z.header = nil
	for zslNode != nil {
		next = zslNode.level[0].forward
		zslNode.freeNode()
		zslNode = next
	}
	z.tail = nil
}

func (z *zSkipList) _getUpdateAndRank(score float64, obj *SRobj) (*[32]*zSkipListNode, *[32]int64, *zSkipListNode) {
	var update [ZSKIPLIST_MAXLEVEL]*zSkipListNode
	var rank [ZSKIPLIST_MAXLEVEL]int64

	x := z.header
	for i := z.level - 1; i >= 0; i-- {
		rank[i] = 0
		if i != (z.level - 1) {
			rank[i] = rank[i+1]
		}

		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score &&
					compareStringObjects(x.level[i].forward.obj, obj) < 0)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}
	return &update, &rank, x
}

func (z *zSkipList) insert(score float64, obj *SRobj) *zSkipListNode {
	var i, level int
	update, rank, x := z._getUpdateAndRank(score, obj)

	level = zslRandomLevel()
	if level > z.level {
		for i = z.level; i < level; i++ {
			rank[i] = 0
			update[i] = z.header
			update[i].level[i].span = z.length
		}
		z.level = level
	}
	x = zslCreateNode(level, score, obj)
	for i = 0; i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	for i = level; i < z.level; i++ {
		update[i].level[i].span++
	}

	x.backward = update[0]
	if update[0] == z.header {
		x.backward = nil
	}

	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		z.tail = x
	}

	z.length++
	return x
}

func (z *zSkipList) deleteNode(node *zSkipListNode, update *[ZSKIPLIST_MAXLEVEL]*zSkipListNode) {
	for i := 0; i < z.level; i++ {
		if update[i].level[i].forward != node {
			update[i].level[i].span -= 1
			continue
		}
		update[i].level[i].span += node.level[i].span - 1
		update[i].level[i].forward = node.level[i].forward
	}
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node.backward
	} else {
		z.tail = node.backward
	}
	for z.level > 1 && (z.header.level[z.level-1].forward == nil) {
		z.level--
	}
	z.length--
}

func (z *zSkipList) delete(score float64, obj *SRobj) bool {
	update, _, x := z._getUpdateAndRank(score, obj)
	x = x.level[0].forward
	if x != nil && score == x.score && compareStringObjects(x.obj, obj) == 0 {
		z.deleteNode(x, update)
		x.freeNode()
		return true
	}
	return false
}

func (z *zSkipList) getElementByRank(rank int64) *zSkipListNode {
	var traversed int64
	x := z.header
	for i := z.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}

// create zSkipList
func zslCreate() *zSkipList {
	zsl := new(zSkipList)
	zsl.level = 1
	zsl.header = zslCreateNode(ZSKIPLIST_MAXLEVEL, 0, nil)
	for i := 0; i < ZSKIPLIST_MAXLEVEL; i++ {
		zsl.header.level[i].forward = nil
		zsl.header.level[i].span = 0
	}
	zsl.header.backward = nil
	zsl.tail = nil
	return zsl
}

// ==================================== zSet =====================================

// 有序集合
type zSet struct {
	zsl *zSkipList // 跳表，存储有序的元素及分数
	d   *dict      // 冗余的dict，存储元素和分数的映射，用于快速查询元素对应的分数
}

// return skipList node numbers
func (z *zSet) len() int64 {
	return z.zsl.length
}

// return a random skipList level
func zslRandomLevel() int {
	level := 1
	for float64(rand.Int63()&0xFFFF) < (ZSKIPLIST_P * 0xFFFF) {
		level++
	}
	if level > ZSKIPLIST_MAXLEVEL {
		level = ZSKIPLIST_MAXLEVEL
	}
	return level
}

// create a new zSet
func zSetCreate() *zSet {
	return &zSet{zsl: zslCreate(), d: dictCreate(&zSetDictType)}
}

// 检查有序集合的encoding是否正确，不正确时会抛出panic
func checkZSetEncoding(subject *SRobj) {
	if subject.encoding != REDIS_ENCODING_SKIPLIST {
		panic("Unknown sorted zSet encoding")
	}
}

// 验证有序集合encoding，并返回有序集合元素数量
func zSetLength(o *SRobj) int64 {
	checkZSetEncoding(o)
	return sLen(assertZSet(o))
}
