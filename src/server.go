package src

const SREDIS_MAX_BULK = 1024 * 4
const SREDIS_MAX_INLINE = 1024 * 4

type SRedisDB struct {
	data   *dict
	expire *dict
}

type SRedisServer struct {
	port    int
	fd      int
	db      *SRedisDB
	clients map[int]*SRedisClient
	el      *aeEventLoop
}

var server SRedisServer

func initServerConfig() {
	server.port = config.Port
	server.fd = -1
}

func initServer() {
	server.db = &SRedisDB{
		data:   dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare}),
		expire: dictCreate(&dictType{hashFunc: SRStrHash, keyCompare: SRStrCompare}),
	}
	server.clients = make(map[int]*SRedisClient)
	server.fd = TcpServer(server.port)
	server.el = aeCreateEventLoop()
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
