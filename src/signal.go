package src

import (
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"os"
	"os/signal"
	"syscall"
)

//-----------------------------------------------------------------------------
// recv
//-----------------------------------------------------------------------------

type signalHandler func(sig os.Signal)

// SetupSignalHandler 信号处理
func SetupSignalHandler(shutdownFunc signalHandler) {
	closeSignalChan := make(chan os.Signal, 1)
	signal.Notify(closeSignalChan,
		os.Kill,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	)
	go func() {
		sig := <-closeSignalChan
		shutdownFunc(sig)
	}()
}

// ================================ server =================================

// server 退出信号处理
func serverShutdown(sig os.Signal) {
	sendKill(server.rdbChildPid, server.aofChildPid)
	ulog.InfoF("signal-handler Received %s scheduling shutdown...", sig.String())

	if server.saveParams != nil {
		ulog.Info("SYNC rdb save start...")
		rdbSave()
	}
	if server.aofState == REDIS_AOF_ON && server.saveParams == nil {
		ulog.Info("SYNC append only file rewrite start...")
		rewriteAppendOnlyFileSync()
	}
	ulog.Info("Simple-Redis is now ready to exit, bye bye !!!")
	os.Exit(0)
}

//-----------------------------------------------------------------------------
// send
//-----------------------------------------------------------------------------

// 向指定的pid发送kill信号
func sendKill(PIDs ...int) {
	for _, pid := range PIDs {
		if pid == -1 {
			continue
		}
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
}
