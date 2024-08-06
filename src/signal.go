package src

import (
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"os"
	"os/signal"
	"simple-redis/utils"
	"syscall"
)

type signalHandler func(sig os.Signal)

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

//-----------------------------------------------------------------------------
// server
//-----------------------------------------------------------------------------

func serverShutdown(sig os.Signal) {
	ulog.InfoF("signal-handler Received %s scheduling shutdown...", sig.String())

	if server.saveParams != nil && server.rdbChildPid == -1 {
		ulog.Info("SYNC rdb save start...")
		rdbSaveBackground()
	}
	if server.aofState == REDIS_AOF_ON && server.aofChildPid == -1 {
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
	utils.Exit(0)
}

//-----------------------------------------------------------------------------
// cli
//-----------------------------------------------------------------------------
