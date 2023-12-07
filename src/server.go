package src

type SRedisDB struct {
	data   *dict
	expire *dict
}

type SRedisServer struct {
	port    int
	fd      int
	db      *SRedisDB
	clients map[int]*SRedisClient
	aeLoop  *aeEventLoop
}

type SRedisClient struct {
	fd int
	db *SRedisDB
	// TODO
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
}

func ServerStart() {
	// init config
	initServerConfig()
	// init server
	initServer()
	// aeMain
	// TODO
}
