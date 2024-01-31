package src

import (
	"os"
	"os/signal"
	"simple-redis/utils"
	"syscall"
)

func SetupSignalHandler(shutdownFunc func(os.Signal)) {
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
	utils.InfoF("signal-handler Received %s scheduling shutdown...", sig.String())
	utils.Info("SYNC append only file rewrite start...")
	if server.aofState == REDIS_AOF_ON && server.aofChildPid == -1 {
		rewriteAppendOnlyFileBackground()
	}
	pid, err := wait4(-1, 0)
	if err != nil {
		utils.ErrorP("wait4 err: ", err)
	}
	if pid == server.aofChildPid {
		backgroundRewriteDoneHandler()
	}
	utils.Info("Simple-Redis is now ready to exit, bye bye...")
	utils.Exit(0)
}

//-----------------------------------------------------------------------------
// cli
//-----------------------------------------------------------------------------
