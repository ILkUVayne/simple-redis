package src

import (
	"golang.org/x/sys/unix"
	"simple-redis/utils"
)

func Accept(fd int) int {
	nfd, _, err := unix.Accept(fd)
	if err != nil {
		utils.Error("simple-redis server: Accept err: ", err)
	}
	return nfd
}

func Connect(host [4]byte, port int) int {
	sfd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		utils.Error("simple-redis server: init socket err: ", err)
	}
	err = unix.Connect(sfd, &unix.SockaddrInet4{Addr: host, Port: port})
	if err != nil {
		utils.Error("simple-redis server: connect err: ", err)
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
		utils.ErrorP("simple-redis server: close err: ", err)
	}
}

func TcpServer(port int) int {
	sfd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		utils.Error("simple-redis server: init socket err: ", err)
	}
	//err = unix.SetsockoptInt(sfd, unix.SOL_SOCKET, unix.SO_REUSEPORT, port)
	err = unix.SetsockoptInt(sfd, unix.SOL_SOCKET, unix.SO_REUSEADDR, port)
	if err != nil {
		utils.Error("simple-redis server: set SO_REUSEPORT err: ", err)
	}
	addr := unix.SockaddrInet4{Port: port}
	err = unix.Bind(sfd, &addr)
	if err != nil {
		utils.ErrorF("simple-redis server: %s:%d bind err: %s", string(addr.Addr[:]), addr.Port, err)
	}
	err = unix.Listen(sfd, BACKLOG)
	if err != nil {
		utils.Error("simple-redis server: listen err: ", err)
	}
	return sfd
}
