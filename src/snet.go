package src

import (
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"golang.org/x/sys/unix"
)

func checkEOF(buf []byte, length int) bool {
	return length >= 2 && string(buf[length-2:length]) == "\r\n"
}

func Accept(fd int) int {
	nfd, _, err := unix.Accept(fd)
	if err != nil {
		ulog.Error("simple-redis server: Accept err: ", err)
	}
	return nfd
}

func Connect(host [4]byte, port int) int {
	sfd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		ulog.Error("simple-redis server: init socket err: ", err)
	}
	err = unix.Connect(sfd, &unix.SockaddrInet4{Addr: host, Port: port})
	if err != nil {
		ulog.Error("simple-redis server: connect err: ", err)
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
		ulog.ErrorP("simple-redis server: close err: ", err)
	}
}

func TcpServer(port int) int {
	sfd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		ulog.Error("simple-redis server: init socket err: ", err)
	}
	// SO_REUSEPORT可以让你将多个socket绑定在同一个监听端口，然后让内核给你自动做负载均衡，将请求平均地让多个线程进行处理。
	//err = unix.SetsockoptInt(sfd, unix.SOL_SOCKET, unix.SO_REUSEPORT, port)
	err = unix.SetsockoptInt(sfd, unix.SOL_SOCKET, unix.SO_REUSEADDR, port)
	if err != nil {
		ulog.Error("simple-redis server: set SO_REUSEPORT err: ", err)
	}
	addr := unix.SockaddrInet4{Port: port}
	err = unix.Bind(sfd, &addr)
	if err != nil {
		ulog.ErrorF("simple-redis server: %s:%d bind err: %s", string(addr.Addr[:]), addr.Port, err)
	}
	err = unix.Listen(sfd, BACKLOG)
	if err != nil {
		ulog.Error("simple-redis server: listen err: ", err)
	}
	return sfd
}
