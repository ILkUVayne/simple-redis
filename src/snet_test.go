package src

import (
	"fmt"
	"testing"
	"time"
)

func EchoServer(s, c, e chan struct{}) {
	host := [4]byte{127, 0, 0, 1}
	sfd := TcpServer(9999, host)
	fmt.Println("server started")
	fmt.Println("tcpserver sfd:", sfd)
	s <- struct{}{}
	<-c
	cfd := Accept(sfd)
	fmt.Printf("accepted cfd: %v\n", cfd)
	buf := make([]byte, 10)
	n, err := Read(cfd, buf)
	if err != nil {
		fmt.Printf("server read error: %v\n", err)
	}
	fmt.Printf("read %v bytes\n", n)
	n, err = Write(cfd, buf)
	if err != nil {
		fmt.Printf("server write error: %v\n", err)
	}
	fmt.Printf("write %v bytes\n", n)
	e <- struct{}{}
}

func TestTcpServer(t *testing.T) {
	fmt.Println("test snet lib")
	s := make(chan struct{})
	c := make(chan struct{})
	e := make(chan struct{})
	go EchoServer(s, c, e)
	<-s
	host := [4]byte{127, 0, 0, 1}
	cfd, _ := Connect(host, 9999)
	fmt.Printf("connected cfd: %v\n", cfd)
	time.Sleep(100 * time.Millisecond)
	c <- struct{}{}
	msg := "helloworld"
	n, err := Write(cfd, []byte(msg))
	if err != nil {
		t.Error("write  err:", err)
	}

	if n != len(msg) {
		t.Error("write n not equal 10, n=", n)
	}
	<-e
	buf := make([]byte, 10)
	n, err = Read(cfd, buf)
	if err != nil {
		t.Error("read  err:", err)
	}
	if n != len(msg) {
		t.Error("read n not equal 10, n=", n)
	}
}
