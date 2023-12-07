package src

import (
	"golang.org/x/sys/unix"
	glog "simple-redis/utils"
)

const BACKLOG int = 64

func Accept(fd int) int {
	nfd, _, err := unix.Accept(fd)
	if err != nil {
		glog.Error("simple-redis server: Accept err: ", err)
	}
	return nfd
}

func Connect(host [4]byte, port int) int {
	sfd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		glog.Error("simple-redis server: init socket err: ", err)
	}
	var addr unix.SockaddrInet4
	addr.Addr = host
	addr.Port = port
	err = unix.Connect(sfd, &addr)
	if err != nil {
		glog.Error("simple-redis server: connect err: ", err)
	}
	return sfd
}

func Write(fd int, buf []byte) (int, error) {
	return unix.Write(fd, buf)
}

func Read(fd int, buf []byte) (int, error) {
	return unix.Read(fd, buf)
}

func Close(fd int) {
	if err := unix.Close(fd); err != nil {
		glog.ErrorP("simple-redis server: close err: ", err)
	}
}

func TcpServer(port int) int {
	sfd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		glog.Error("simple-redis server: init socket err: ", err)
	}
	err = unix.SetsockoptInt(sfd, unix.SOL_SOCKET, unix.SO_REUSEPORT, port)
	if err != nil {
		glog.Error("simple-redis server: set SO_REUSEPORT err: ", err)
	}
	var addr unix.SockaddrInet4
	addr.Port = port
	err = unix.Bind(sfd, &addr)
	if err != nil {
		glog.Error("simple-redis server: bind err: ", err)
	}
	err = unix.Listen(sfd, BACKLOG)
	if err != nil {
		glog.Error("simple-redis server: listen err: ", err)
	}
	return sfd
}
