package src

import (
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"golang.org/x/sys/unix"
)

// fork child process
func fork() int {
	id, _, err := unix.Syscall(unix.SYS_FORK, 0, 0, 0)
	ulog.ErrorP("fork err: ", err)
	return int(id)
}

// Query whether the child process terminates.
//
// pid < -1 等待进程组 ID 为 pid 绝对值的进程组中的任何子进程.
// pid = -1 等待任何子进程.
// pid = 0 等待进程组 ID 与当前进程相同的任何子进程（也就是等待同一个进程组中的任何子进程）.
// pid > 0 等待任何子进程 ID 为 pid 的子进程，只要指定的子进程还没有结束，wait4() 就会一直等下去.
//
// options
// unix.WNOHANG 如果没有任何已经结束了的子进程，则马上返回，不等待
// unix.WUNTRACED 如果子进程进入暂停执行的情况，则马上返回，但结束状态不予理会
// options 设为0，则会一直等待，直到有进程退出
//
// return 0 如果设置了 WNOHANG，而调用 wait4() 时，没有发现已退出的子进程可收集，则返回0.
// return > 0 正常返回时，wait4() 返回收集到的子进程的PID.
// return -1 如果调用出错，则返回 -1，这时error 会被设置为相应的值以指示错误所在。（当 pid 所指示的子进程不存在，或此进程存在，
// 但不是调用进程的子进程， wait4() 就会返回出错，这时 error 被设置为 ECHILD）
func wait4(pid int, options int) (int, error) {
	var status unix.WaitStatus
	id, err := unix.Wait4(pid, &status, options, nil)
	return id, err
}
