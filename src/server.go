package src

import (
	"fmt"
	"os"
	"os/signal"
	"simple-redis/utils"
	"syscall"
)

type sharedObjects struct {
	crlf, ok, err, czero, cone, emptyMultiBulk, nullBulk, syntaxErr, typeErr, unknowErr, argsNumErr, wrongTypeErr *SRobj
}

var shared sharedObjects

func createSharedObjects() {
	shared.crlf = createSRobj(SR_STR, "\r\n")
	shared.ok = createSRobj(SR_STR, RESP_OK)
	shared.err = createSRobj(SR_STR, RESP_ERR)
	shared.czero = createSRobj(SR_STR, ":0\r\n")
	shared.cone = createSRobj(SR_STR, ":1\r\n")
	shared.emptyMultiBulk = createSRobj(SR_STR, "*0\r\n")
	shared.nullBulk = createSRobj(SR_STR, RESP_NIL_VAL)
	shared.syntaxErr = createSRobj(SR_STR, fmt.Sprintf(RESP_ERR, "syntax error"))
	shared.typeErr = createSRobj(SR_STR, fmt.Sprintf(RESP_ERR, "wrong type"))
	shared.unknowErr = createSRobj(SR_STR, fmt.Sprintf(RESP_ERR, "unknow command"))
	shared.argsNumErr = createSRobj(SR_STR, fmt.Sprintf(RESP_ERR, "wrong number of args"))
	shared.wrongTypeErr = createSRobj(SR_STR, fmt.Sprintf(RESP_ERR, "Operation against a key holding the wrong kind of value"))
}

func aofFile(file string) string {
	return getHome() + "/" + file
}

func loadDataFromDisk() {
	start := utils.GetMsTime()
	if server.aofState == REDIS_AOF_ON {
		loadAppendOnlyFile(server.aofFilename)
		utils.InfoF("DB loaded from append only file: %.3f seconds", float64(utils.GetMsTime()-start)/1000)
	}
}

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

func shutdown(sig os.Signal) {
	utils.InfoF("signal-handler Received %s scheduling shutdown...", sig.String())
	// todo do something before exit
	utils.Info("Simple-Redis is now ready to exit, bye bye...")
	os.Exit(0)
}

type SRedisServer struct {
	port           int
	fd             int
	db             *SRedisDB
	clients        map[int]*SRedisClient
	el             *aeEventLoop
	loadFactor     int64
	rehashNullStep int64
	// AOF persistence
	aofFd              *os.File
	aofFilename        string
	aofBuf             string
	aofState           int
	aofCurrentSize     int64
	aofRewriteBaseSize int64
	// RDB persistence
	dirty int64
}

func (s *SRedisServer) incrDirtyCount(c *SRedisClient, num int64) {
	if c.fd > 0 {
		s.dirty += num
	}
}

var server SRedisServer

func initServerConfig() {
	server.port = DEFAULT_PORT
	if config.Port > 0 {
		server.port = config.Port
	}
	server.fd = -1
	server.rehashNullStep = DEFAULT_RH_NN_STEP
	if config.RehashNullStep > 0 {
		server.rehashNullStep = config.RehashNullStep
	}
	server.aofState = REDIS_AOF_OFF
	if config.AppendOnly {
		server.aofState = REDIS_AOF_ON
	}
}

func initServer() {
	server.db = &SRedisDB{
		data:   dictCreate(&dbDictType),
		expire: dictCreate(&keyPtrDictType),
	}
	server.clients = make(map[int]*SRedisClient)
	server.fd = TcpServer(server.port)
	server.el = aeCreateEventLoop()
	server.loadFactor = LOAD_FACTOR
	createSharedObjects()
	// add fileEvent
	server.el.addFileEvent(server.fd, AE_READABLE, acceptTcpHandler, nil)
	// add timeEvent
	server.el.addTimeEvent(AE_NORMAL, 100, serverCron, nil)
	// AOF fd
	if server.aofState == REDIS_AOF_ON {
		server.aofFilename = aofFile(REDIS_AOF_DEFAULT)
		fd, err := os.OpenFile(server.aofFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			utils.Error("Can't open the append-only file: ", err)
		}
		server.aofFd = fd
	}
}

func ServerStart() {
	// load config
	SetupConf(ServerArgs.confPath)
	// init config
	initServerConfig()
	// init server
	initServer()
	utils.Info("* Server initialized")
	// load data
	loadDataFromDisk()

	SetupSignalHandler(shutdown)
	// aeMain
	utils.InfoF("* server started, The server is now ready to accept connections on port %d", server.port)
	aeMain(server.el)
}
