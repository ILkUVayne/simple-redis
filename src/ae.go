// Package src
//
// lib ae provides the creation and management of file events and time events based on epoll IO reuse
package src

import (
	"golang.org/x/sys/unix"
	"simple-redis/utils"
)

// FeType fileEvent type
type FeType int

// TeType timeEvent type
type TeType int

// file event func
type aeFileProc func(el *aeEventLoop, fd int, clientData any)

// time event func
type aeTimeProc func(el *aeEventLoop, id int, clientData any)

type aeFileEvent struct {
	mask       FeType // ae file event type, AE_READABLE and AE_WRITEABLE
	proc       aeFileProc
	fd         int
	clientData any
}

type aeTimeEvent struct {
	id         int
	when       int64 //ms
	interval   int64 //ms
	proc       aeTimeProc
	clientData any
	next       *aeTimeEvent
	mask       TeType // ae time event type, AE_NORMAL and AE_ONCE
}

type aeEventLoop struct {
	fileEvent       map[int]*aeFileEvent // file event maps
	timeEvent       *aeTimeEvent         // time event list
	ffd             int                  // epoll fd
	timeEventNextId int
	stop            bool
}

// fileEvent to epoll
var fe2ep = [3]uint32{0, unix.EPOLLIN, unix.EPOLLOUT}

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

func (el *aeEventLoop) epollMask(fd int) uint32 {
	var em uint32
	// 该fd上已存在AE_READABLE事件
	if el.fileEvent[feKey(fd, AE_READABLE)] != nil {
		// em == unix.EPOLLIN
		em |= fe2ep[AE_READABLE]
	}
	// 该fd上已存在AE_WRITEABLE事件
	if el.fileEvent[feKey(fd, AE_WRITEABLE)] != nil {
		// em == unix.EPOLLIN | unix.EPOLLOUT
		em |= fe2ep[AE_WRITEABLE]
	}
	// default em == 0
	return em
}

// epoll_ctl and add file event to aeEventLoop.fileEvent
func (el *aeEventLoop) addFileEvent(fd int, mask FeType, proc aeFileProc, clientData any) {
	// epoll_ctl
	em := el.epollMask(fd)
	// 该fd对应的mask事件已注册
	if em&fe2ep[mask] != 0 {
		return
	}
	op := unix.EPOLL_CTL_ADD
	if em != 0 {
		op = unix.EPOLL_CTL_MOD
	}
	em |= fe2ep[mask]
	if err := unix.EpollCtl(el.ffd, op, fd, &unix.EpollEvent{Events: em, Fd: int32(fd)}); err != nil {
		utils.Error("simple-redis server: ae epoll_ctl err: ", err)
	}
	// ae ctl
	fileEvent := new(aeFileEvent)
	fileEvent.fd = fd
	fileEvent.proc = proc
	fileEvent.mask = mask
	fileEvent.clientData = clientData
	el.fileEvent[feKey(fd, mask)] = fileEvent
	//utils.InfoF("simple-redis server: add fileEvent fd %d,mask %d: ", fd, mask)
}

// epoll_ctl and remove file event from aeEventLoop.fileEvent
func (el *aeEventLoop) removeFileEvent(fd int, mask FeType) {
	// epoll_ctl
	em := el.epollMask(fd)
	// 该fd对应的mask事件未注册，无需删除
	if em&fe2ep[mask] == 0 {
		return
	}
	op := unix.EPOLL_CTL_DEL
	em &= ^fe2ep[mask]
	if em != 0 {
		op = unix.EPOLL_CTL_MOD
	}
	if err := unix.EpollCtl(el.ffd, op, fd, &unix.EpollEvent{Events: em, Fd: int32(fd)}); err != nil {
		utils.Error("simple-redis server: ae epoll_ctl err: ", err)
	}
	// ae ctl
	el.fileEvent[feKey(fd, mask)] = nil
	//utils.InfoF("simple-redis server: remove fileEvent fd %d,mask %d: ", fd, mask)
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
	te.when = utils.GetMsTime() + interval
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
	nearestTime := utils.GetMsTime() + 1000
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
	timeout := el.nearestTime() - utils.GetMsTime()
	if timeout <= 0 {
		timeout = 10
	}
	var events [128]unix.EpollEvent
	n, err := unix.EpollWait(el.ffd, events[:], int(timeout))
	if err != nil {
		if err != unix.EINTR {
			utils.ErrorP("simple-redis server: ae epoll_wait err: ", err)
		}
	}
	//utils.InfoF("simple-redis server: ae epoll get %d events: ", n)
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
	now := utils.GetMsTime()
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
	el.stop = false
	el.timeEventNextId = 1
	efd, err := unix.EpollCreate1(0)
	if err != nil {
		utils.Error("simple-redis server: aeCreateEventLoop err: ", err)
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
