package src

import (
	"fmt"
	"github.com/ILkUVayne/utlis-go/v2/time"
	"os"
	"simple-redis/utils"
)

// 全局共享SRobj对象结构体，用以复用常用的命令返回对象
type sharedObjects struct {
	crlf, ok, err, czero, cone, emptyMultiBulk, nullBulk, syntaxErr, typeErr, unknowErr, argsNumErr, wrongTypeErr,
	none *SRobj
}

// 全局共享SRobj对象
var shared sharedObjects

// 初始化全局共享SRobj对象
func initSharedObjects() {
	shared.crlf = createSRobj(SR_STR, "\r\n")
	shared.ok = createSRobj(SR_STR, RESP_OK)
	shared.err = createSRobj(SR_STR, RESP_ERR)
	shared.czero = createSRobj(SR_STR, ":0\r\n")
	shared.cone = createSRobj(SR_STR, ":1\r\n")
	shared.emptyMultiBulk = createSRobj(SR_STR, "*0\r\n")
	shared.none = createSRobj(SR_STR, "+none\r\n")
	shared.nullBulk = createSRobj(SR_STR, RESP_NIL_VAL)
	shared.syntaxErr = createSRobj(SR_STR, fmt.Sprintf(RESP_ERR, "syntax error"))
	shared.typeErr = createSRobj(SR_STR, fmt.Sprintf(RESP_ERR, "wrong type"))
	shared.unknowErr = createSRobj(SR_STR, fmt.Sprintf(RESP_ERR, "unknow command"))
	shared.argsNumErr = createSRobj(SR_STR, fmt.Sprintf(RESP_ERR, "wrong number of args"))
	shared.wrongTypeErr = createSRobj(SR_STR, fmt.Sprintf(RESP_ERR, "Operation against a key holding the wrong kind of value"))
}

// SRedisServer server 结构体
//
// 定义server所需的所有基本信息
type SRedisServer struct {
	port           int
	fd             int // server 监听的fd
	db             *SRedisDB
	clients        map[int]*SRedisClient
	el             *aeEventLoop
	loadFactor     int64 // 负载因子
	rehashNullStep int64 // 每次rehash最多遍历rehashNullStep步为nil的数据

	// AOF persistence

	aofFd               *os.File // aof文件fd
	aofChildPid         int
	aofFilename         string
	aofBuf              string // AOF buffer
	aofState            int
	aofCurrentSize      int64
	aofRewriteBaseSize  int64
	aofRewriteBufBlocks string // AOF rewrite buffer
	aofRewritePerc      int
	aofRewriteMinSize   int64

	// RDB persistence

	dirty             int64
	dirtyBeforeBgSave int64
	lastBgSaveTry     int64
	lastBgSaveStatus  int
	saveParams        []*saveParam
	rdbChildPid       int
	rdbFilename       string
	lastSave          int64
}

func (s *SRedisServer) incrDirtyCount(c *SRedisClient, num int64) {
	if !c.isFake() {
		s.dirty += num
	}
}

// 更新负载因子，负载因子越小，越容易发生rehash
//
// 正常情况下为1，当进行BGREWRITEAOF或者BGSAVE时为了尽量避免rehash,会更新为5
func (s *SRedisServer) changeLoadFactor(lf int) {
	if s.loadFactor == int64(lf) {
		return
	}
	if lf == LOAD_FACTOR && s.aofChildPid == -1 && s.rdbChildPid == -1 {
		s.loadFactor = int64(lf)
		return
	}
	if lf == BG_PERSISTENCE_LOAD_FACTOR && (s.aofChildPid != -1 || s.rdbChildPid != -1) {
		s.loadFactor = int64(lf)
		return
	}
}

// 更新调整dict容量权限
func updateDictResizePolicy() {
	if server.rdbChildPid == -1 && server.aofChildPid == -1 {
		dictEnableResize()
		return
	}
	dictDisableResize()
}

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

var server SRedisServer

func initServer() {
	server.db = &SRedisDB{
		data:   dictCreate(&dbDictType),
		expire: dictCreate(&keyPtrDictType),
	}
	server.clients = make(map[int]*SRedisClient)
	server.fd = TcpServer(server.port)
	server.el = aeCreateEventLoop()
	server.loadFactor = LOAD_FACTOR
	// add fileEvent
	server.el.addFileEvent(server.fd, AE_READABLE, acceptTcpHandler, nil)
	// add timeEvent
	server.el.addTimeEvent(AE_NORMAL, 100, serverCron, nil)
	// AOF fd
	server.aofChildPid = -1
	if server.aofState == REDIS_AOF_ON {
		server.aofFilename = utils.PersistenceFile(REDIS_AOF_DEFAULT)
		fd, err := os.OpenFile(server.aofFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			utils.Error("Can't open the append-only file: ", err)
		}
		server.aofFd = fd
		server.aofRewritePerc = REDIS_AOF_REWRITE_PERC
		server.aofRewriteMinSize = REDIS_AOF_REWRITE_MIN_SIZE
	}
	// rdb
	server.rdbChildPid = -1
	server.rdbFilename = utils.PersistenceFile(REDIS_RDB_DEFAULT)
}

func loadDataFromDisk() {
	start := time.GetMsTime()
	if server.aofState == REDIS_AOF_ON {
		loadAppendOnlyFile(server.aofFilename)
		utils.InfoF("DB loaded from append only file: %.3f seconds", float64(time.GetMsTime()-start)/1000)
		return
	}
	rdbLoad(&server.rdbFilename)
	utils.InfoF("DB loaded from disk: %.3f seconds", float64(time.GetMsTime()-start)/1000)
}

func ServerStart() {
	// load config
	SetupConf(ServerArgs.confPath)
	// init config
	initServerConfig()
	// init Shared Objects
	initSharedObjects()
	// init server
	initServer()
	utils.Info("* Server initialized")
	// load data from rdb or aof
	loadDataFromDisk()
	// set signal handle
	SetupSignalHandler(serverShutdown)
	// aeMain loop
	utils.InfoF("* server started, The server is now ready to accept connections on port %d", server.port)
	aeMain(server.el)
}
