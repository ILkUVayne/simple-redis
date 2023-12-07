package src

type SRedisDB struct {
	//data   *Dict
	//expire *Dict
}

type SRedisServer struct {
	port    int
	fd      int
	db      *SRedisDB
	clients map[int]*SRedisClient
	//aeLoop  *AeLoop
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
