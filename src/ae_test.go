package src

import (
	"fmt"
	"golang.org/x/sys/unix"
	"testing"
	"time"
)

func testFeProc(*aeEventLoop, int, any) {
	fmt.Println("run testFeProc")
}

func testTeProc(*aeEventLoop, int, any) {
	fmt.Println("run testTeProc")
}

func testOnceTeProc(*aeEventLoop, int, any) {
	fmt.Println("run once testTeProc")
}

func TestAeFeKey(t *testing.T) {
	fd := 6
	k := feKey(fd, AE_READABLE)
	if k != 6 {
		t.Error("AE_READABLE feKey k ==:", k)
	}
	k = feKey(fd, AE_WRITEABLE)
	if k != -6 {
		t.Error("AE_READABLE feKey k ==:", k)
	}
}

func TestEpollMask(t *testing.T) {
	fd := 7
	el := aeCreateEventLoop()
	em := el.epollMask(fd)
	if em != 0 {
		t.Error("el.epollMask err: em==", em)
	}
	el.fileEvent[feKey(fd, AE_READABLE)] = new(aeFileEvent)
	em = el.epollMask(fd)
	if em != unix.EPOLLIN {
		t.Error("el.epollMask err: em==", em)
	}
	el.fileEvent[feKey(fd, AE_WRITEABLE)] = new(aeFileEvent)
	em = el.epollMask(fd)
	if em != (unix.EPOLLIN | unix.EPOLLOUT) {
		t.Error("el.epollMask err: em==", em)
	}
}

// test addFileEvent and removeFileEvent
func TestAddFe(t *testing.T) {
	fd := TcpServer(9999)
	el := aeCreateEventLoop()
	el.addFileEvent(fd, AE_READABLE, testFeProc, nil)
	if len(el.fileEvent) != 1 {
		t.Error("addFileEvent err: fileEvent number ==", len(el.fileEvent))
	}
	el.addFileEvent(fd, AE_WRITEABLE, testFeProc, nil)
	if len(el.fileEvent) != 2 {
		t.Error("addFileEvent err: fileEvent number ==", len(el.fileEvent))
	}
	el.addFileEvent(fd, AE_READABLE, testFeProc, nil)
	el.addFileEvent(fd, AE_WRITEABLE, testFeProc, nil)
	el.fileEvent[feKey(fd, AE_READABLE)].proc(el, fd, nil)
	el.fileEvent[feKey(fd, AE_WRITEABLE)].proc(el, fd, nil)
	// remove
	el.removeFileEvent(fd, AE_READABLE)
	if el.fileEvent[feKey(fd, AE_READABLE)] != nil {
		t.Error("addFileEvent err: fileEvent remove err")
	}
	el.removeFileEvent(fd, AE_WRITEABLE)
	if el.fileEvent[feKey(fd, AE_READABLE)] != nil {
		t.Error("addFileEvent err: fileEvent remove err")
	}
	el.removeFileEvent(fd, AE_READABLE)
	el.removeFileEvent(fd, AE_WRITEABLE)
}

// test addTimeEvent and removeTimeEvent
func TestAddTe(t *testing.T) {
	fd := TcpServer(8888)
	el := aeCreateEventLoop()
	el.addTimeEvent(AE_ONCE, 10, testTeProc, nil)
	el.addTimeEvent(AE_NORMAL, 11, testTeProc, nil)

	var timeEvents []*aeTimeEvent
	for p := el.timeEvent; p != nil; p = p.next {
		p.proc(el, fd, nil)
		timeEvents = append(timeEvents, p)
	}
	if len(timeEvents) != 2 {
		t.Error("addTimeEvent err: timeEvent num ==", len(timeEvents))
	}

	for _, te := range timeEvents {
		el.removeTimeEvent(te.id)
	}
	if el.timeEvent != nil {
		t.Error("removeTimeEvent err: timeEvent ==", el.timeEvent)
	}
}

func TestAeMain(t *testing.T) {
	fd := TcpServer(7777)
	el := aeCreateEventLoop()
	el.addFileEvent(fd, AE_READABLE, testFeProc, nil)
	el.addFileEvent(fd, AE_WRITEABLE, testFeProc, nil)
	el.addTimeEvent(AE_ONCE, 10, testOnceTeProc, nil)
	el.addTimeEvent(AE_NORMAL, 11, testTeProc, nil)
	go aeMain(el)
	time.Sleep(100 * time.Millisecond)
	el.stop = true
	time.Sleep(100 * time.Millisecond)
}
