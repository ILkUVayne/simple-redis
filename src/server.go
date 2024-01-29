package src

import (
	"fmt"
	"os"
	"simple-redis/utils"
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

type SRedisServer struct {
	port           int
	fd             int
	db             *SRedisDB
	clients        map[int]*SRedisClient
	el             *aeEventLoop
	loadFactor     int64
	rehashNullStep int64
	// AOF persistence
	aofFd               *os.File
	aofChildPid         int
	aofFilename         string
	aofBuf              string
	aofState            int
	aofCurrentSize      int64
	aofRewriteBaseSize  int64
	aofRewriteBufBlocks string
	aofRewritePerc      int
	aofRewriteMinSize   int64
	// RDB persistence
	dirty      int64
	saveParams []*saveParam
}

func (s *SRedisServer) incrDirtyCount(c *SRedisClient, num int64) {
	if c.fd > 0 {
		s.dirty += num
	}
}

func (s *SRedisServer) changeLoadFactor(lf int) {
	if s.loadFactor == int64(lf) {
		return
	}
	if lf == LOAD_FACTOR {
		if s.aofChildPid == -1 {
			s.loadFactor = int64(lf)
			return
		}
	}
	if lf == BG_PERSISTENCE_LOAD_FACTOR {
		if s.aofChildPid != -1 {
			s.loadFactor = int64(lf)
			return
		}
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
	// aof
	server.aofState = REDIS_AOF_OFF
	if config.AppendOnly {
		server.aofState = REDIS_AOF_ON
	}
	// rdb
	if config.saveParams != nil && len(config.saveParams) != 0 {
		server.saveParams = config.saveParams
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
	server.aofChildPid = -1
	if server.aofState == REDIS_AOF_ON {
		server.aofFilename = aofFile(REDIS_AOF_DEFAULT)
		fd, err := os.OpenFile(server.aofFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			utils.Error("Can't open the append-only file: ", err)
		}
		server.aofFd = fd
		server.aofRewritePerc = REDIS_AOF_REWRITE_PERC
		server.aofRewriteMinSize = REDIS_AOF_REWRITE_MIN_SIZE
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

	SetupSignalHandler(serverShutdown)
	// aeMain
	utils.InfoF("* server started, The server is now ready to accept connections on port %d", server.port)
	aeMain(server.el)
}
