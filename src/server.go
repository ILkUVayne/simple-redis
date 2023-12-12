package src

const (
	DEFAULT_PORT       = 6379
	DEFAULT_RH_NN_STEP = 10
)

const SREDIS_MAX_BULK = 1024 * 4
const SREDIS_MAX_INLINE = 1024 * 4

type SRedisDB struct {
	data   *dict
	expire *dict
}

type SRedisServer struct {
	port           int
	fd             int
	db             *SRedisDB
	clients        map[int]*SRedisClient
	el             *aeEventLoop
	loadFactor     int64
	rehashNullStep int64
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
}

func initServer() {
	server.db = &SRedisDB{
		data:   dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare}),
		expire: dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare}),
	}
	server.clients = make(map[int]*SRedisClient)
	server.fd = TcpServer(server.port)
	server.el = aeCreateEventLoop()
	server.loadFactor = LOAD_FACTOR
	// add fileEvent
	server.el.addFileEvent(server.fd, AE_READABLE, acceptTcpHandler, nil)
	// add timeEvent
	server.el.addTimeEvent(AE_NORMAL, 100, serverCron, nil)
}

func ServerStart() {
	// init config
	initServerConfig()
	// init server
	initServer()
	// aeMain
	aeMain(server.el)
}
