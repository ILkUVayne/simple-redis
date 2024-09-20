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
		rdbSaveBackground()
	}
	if server.aofState == REDIS_AOF_ON && server.saveParams == nil {
		ulog.Info("SYNC append only file rewrite start...")
		rewriteAppendOnlyFileBackground()
	}
	pid, err := wait4(-1, 0)
	if err != nil {
		ulog.ErrorP("wait4 err: ", err)
	}
	if pid != 0 && pid != -1 {
		if pid == server.aofChildPid {
			backgroundRewriteDoneHandler()
		}
		if pid == server.rdbChildPid {
			backgroundSaveDoneHandler()
		}
	}
	ulog.Info("Simple-Redis is now ready to exit, bye bye...")
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
