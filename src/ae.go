// Package src
//
// lib ae provides the creation and management of file events and time events based on epoll IO reuse
package src

import (
	"github.com/ILkUVayne/utlis-go/v2/time"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"golang.org/x/sys/unix"
)

// FeType fileEvent type
type FeType int

// TeType timeEvent type
type TeType int

// file event func
type aeFileProc func(el *aeEventLoop, fd int, clientData any)

// time event func
type aeTimeProc func(el *aeEventLoop, id int, clientData any)

// ae 文件事件
type aeFileEvent struct {
	mask       FeType     // ae file event type, AE_READABLE and AE_WRITEABLE
	proc       aeFileProc // 文件事件处理函数
	fd         int        // server 或者 client 文件描述符
	clientData any
}

// ae 时间事件
type aeTimeEvent struct {
	id         int
	when       int64      //ms
	interval   int64      //ms
	proc       aeTimeProc // 时间事件处理函数
	clientData any
	next       *aeTimeEvent
	mask       TeType // ae time event type, AE_NORMAL and AE_ONCE
}

// ae 事件循环
type aeEventLoop struct {
	fileEvent       map[int]*aeFileEvent // file event maps
	timeEvent       *aeTimeEvent         // time event list
	ffd             int                  // epoll fd
	timeEventNextId int                  // 下一个被创建时间事件的id
	stop            bool                 // default false
}

// fileEvent to epoll
var fe2ep = [3]uint32{0x0, unix.EPOLLIN, unix.EPOLLOUT}

// get fileEvent key
func feKey(fd int, mask FeType) int {
	switch mask {
	case AE_READABLE:
		return fd
	case AE_WRITEABLE:
		return fd * -1
	}
	return -1
}

// 获取 fd 对应的 epoll Mask. 默认返回0，表示未绑定事件.
//
// em == unix.EPOLLIN 时，表示已绑定AE_READABLE事件.
//
// em == unix.EPOLLOUT 时，表示已绑定AE_WRITEABLE事件.
//
// em == unix.EPOLLIN | unix.EPOLLOUT 时，表示已绑定AE_READABLE和AE_WRITEABLE事件.
func (el *aeEventLoop) epollMask(fd int) uint32 {
	var em uint32
	// 该fd上已存在AE_READABLE事件
	if el.fileEvent[feKey(fd, AE_READABLE)] != nil {
		// em |= unix.EPOLLIN
		em |= fe2ep[AE_READABLE]
	}
	// 该fd上已存在AE_WRITEABLE事件
	if el.fileEvent[feKey(fd, AE_WRITEABLE)] != nil {
		// em |= unix.EPOLLOUT
		em |= fe2ep[AE_WRITEABLE]
	}
	return em
}

// epoll_ctl and add file event to aeEventLoop.fileEvent
func (el *aeEventLoop) addFileEvent(fd int, mask FeType, proc aeFileProc, clientData any) {
	// epoll_ctl
	em := el.epollMask(fd)
	// 该fd对应的mask事件已注册
	if em&fe2ep[mask] != 0x0 {
		return
	}
	op := unix.EPOLL_CTL_ADD
	if em != 0x0 {
		op = unix.EPOLL_CTL_MOD
	}
	em |= fe2ep[mask]
	if err := unix.EpollCtl(el.ffd, op, fd, &unix.EpollEvent{Events: em, Fd: int32(fd)}); err != nil {
		ulog.Error("simple-redis server: ae epoll_ctl err: ", err)
	}
	// ae ctl
	el.fileEvent[feKey(fd, mask)] = &aeFileEvent{fd: fd, proc: proc, mask: mask, clientData: clientData}
}

// epoll_ctl and remove file event from aeEventLoop.fileEvent
func (el *aeEventLoop) removeFileEvent(fd int, mask FeType) {
	// epoll_ctl
	em := el.epollMask(fd)
	// 该fd对应的mask事件未注册，无需删除
	if em&fe2ep[mask] == 0x0 {
		return
	}
	op := unix.EPOLL_CTL_DEL
	em &= ^fe2ep[mask]
	if em != 0 {
		op = unix.EPOLL_CTL_MOD
	}
	if err := unix.EpollCtl(el.ffd, op, fd, &unix.EpollEvent{Events: em, Fd: int32(fd)}); err != nil {
		ulog.Error("simple-redis server: ae epoll_ctl err: ", err)
	}
	// ae ctl
	el.fileEvent[feKey(fd, mask)] = nil
}

// add time event to aeEventLoop.timeEvent
func (el *aeEventLoop) addTimeEvent(mask TeType, interval int64, proc aeTimeProc, clientData any) int {
	te := new(aeTimeEvent)
	te.id = el.timeEventNextId
	el.timeEventNextId++
	te.proc = proc
	te.clientData = clientData
	te.interval = interval
	te.mask = mask
	te.when = time.GetMsTime() + interval
	te.next = el.timeEvent
	el.timeEvent = te
	return te.id
}

// remove time event from aeEventLoop.timeEvent
func (el *aeEventLoop) removeTimeEvent(id int) {
	var pre *aeTimeEvent
	for p := el.timeEvent; p != nil; p = p.next {
		if p.id == id {
			if pre == nil {
				el.timeEvent = p.next
			} else {
				pre.next = p.next
			}
			// remove timeEvent
			// 解除需要删除的timeEvent的next引用，便于gc回收
			p.next = nil
			break
		}
		pre = p
	}
}

// return nearest timeEvent time
func (el *aeEventLoop) nearestTime() int64 {
	nearestTime := time.GetMsTime() + 1000
	for p := el.timeEvent; p != nil; p = p.next {
		if p.when < nearestTime {
			nearestTime = p.when
		}
	}
	return nearestTime
}

// ae 事件处理
func (el *aeEventLoop) aeProcessEvents() {
	// epoll_wait timeout
	timeout := el.nearestTime() - time.GetMsTime()
	if timeout <= 0 {
		timeout = 10
	}
	var events [128]unix.EpollEvent
	n, err := unix.EpollWait(el.ffd, events[:], int(timeout))
	if err != nil {
		if err != unix.EINTR {
			ulog.ErrorP("simple-redis server: ae epoll_wait err: ", err)
		}
	}
	// 收集可执行事件
	// 收集文件事件
	var fileEvents []*aeFileEvent
	for i := 0; i < n; i++ {
		e := events[i]
		// 读事件
		if e.Events&unix.EPOLLIN != 0 {
			if fileEvent := el.fileEvent[feKey(int(e.Fd), AE_READABLE)]; fileEvent != nil {
				fileEvents = append(fileEvents, fileEvent)
			}
			continue
		}
		// 写事件
		if fileEvent := el.fileEvent[feKey(int(e.Fd), AE_WRITEABLE)]; fileEvent != nil {
			fileEvents = append(fileEvents, fileEvent)
		}
	}
	// 收集时间事件
	var timeEvents []*aeTimeEvent
	now := time.GetMsTime()
	for p := el.timeEvent; p != nil; p = p.next {
		if p.when <= now {
			timeEvents = append(timeEvents, p)
		}
	}
	// 事件处理
	// 处理时间事件
	for _, te := range timeEvents {
		te.proc(el, te.id, te.clientData)
		if te.mask == AE_ONCE {
			// remove te
			el.removeTimeEvent(te.id)
			continue
		}
		te.when = te.when + te.interval
	}
	// 处理文件事件
	for _, fe := range fileEvents {
		fe.proc(el, fe.fd, fe.clientData)
	}
}

// create ae loop
func aeCreateEventLoop() *aeEventLoop {
	el := new(aeEventLoop)
	el.fileEvent = make(map[int]*aeFileEvent)
	el.timeEventNextId = 1
	efd, err := unix.EpollCreate1(0)
	if err != nil {
		ulog.Error("simple-redis server: aeCreateEventLoop err: ", err)
	}
	el.ffd = efd
	return el
}

// ae main loop
func aeMain(el *aeEventLoop) {
	for !el.stop {
		el.aeProcessEvents()
	}
}
