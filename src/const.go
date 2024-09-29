package src

//-----------------------------------------------------------------------------
// base
//-----------------------------------------------------------------------------

const (
	UNKNOWN = "unknown"
)

//-----------------------------------------------------------------------------
// errors
//-----------------------------------------------------------------------------

const (
	CONN_DISCONNECTED = srError(0x01) /* connect disconnected */
)

//-----------------------------------------------------------------------------
// server
//-----------------------------------------------------------------------------

const (
	REDIS_VERSION = "3.2.0"

	DEFAULT_PORT       = 6379
	DEFAULT_RH_NN_STEP = 10
	REDIS_OK           = 0
	REDIS_ERR          = 1

	REDIS_AOF_OFF = 0 /* AOF is off */
	REDIS_AOF_ON  = 1

	REDIS_AOF_DEFAULT = "appendonly.aof"
	REDIS_RDB_DEFAULT = "dump.rdb"

	SREDIS_MAX_BULK   = 1024 * 4
	SREDIS_MAX_INLINE = 1024 * 4
	SREDIS_IO_BUF     = 1024 * 16

	REDIS_BGSAVE_RETRY_DELAY = 5

	CONFIG = "./sredis.conf"

	// splitArgs status

	SPA_CONTINUE   = 1
	SPA_DONE       = 2
	SPA_TERMINATED = 3

	// bgSave 或者 BGREWRITEAOF (child process) 执行时间阈值

	C_PROC_MAX_TIME int64 = 1000 * 60 * 5 // 单位毫秒,默认5分钟
)

//-----------------------------------------------------------------------------
// client
//-----------------------------------------------------------------------------

const (
	FAKE_CLIENT_FD = -2
)

//-----------------------------------------------------------------------------
// server-cli
//-----------------------------------------------------------------------------

const (
	CLI_OK  = 0
	CLI_ERR = 1

	NIL_STR = "(nil)"

	REDIS_CLI_HISTFILE_DEFAULT = ".srediscli_history"
)

//-----------------------------------------------------------------------------
// resp
//-----------------------------------------------------------------------------

const (
	RESP_NIL_VAL = "$-1\r\n"

	RESP_OK  = "+OK\r\n"
	RESP_ERR = "-ERR: %s\r\n"

	RESP_INT        = ":%d\r\n"
	RESP_BULK       = "$%d\r\n%v\r\n"
	RESP_ARRAY      = "*%d\r\n"
	RESP_STR        = "*3\r\n$3\r\nSET\r\n"
	RESP_EXPIRE     = "*3\r\n$6\r\nEXPIRE\r\n"
	RESP_LIST_RPUSH = "*%d\r\n$5\r\nRPUSH\r\n"
	RESP_SET        = "*%d\r\n$4\r\nSADD\r\n"
	RESP_ZSET       = "*%d\r\n$4\r\nZADD\r\n"
	RESP_HASH_HSET  = "*%d\r\n$4\r\nHSET\r\n"
)

const (
	SIMPLE_STR   = iota + 1 // +OK\r\n
	SIMPLE_ERROR            // -Error message\r\n
	INTEGERS                // :[<+|->]<value>\r\n
	BULK_STR                // $<length>\r\n<data>\r\n
	ARRAYS                  // *<number-of-elements>\r\n<element-1>...<element-n>
	NULLS                   // _\r\n
	BOOLEANS                // #<t|f>\r\n
	DOUBLE                  // ,[<+|->]<integral>[.<fractional>][<E|e>[sign]<exponent>]\r\n e.g. ,1.23\r\n
	BIG_NUMBERS             // ([+|-]<number>\r\n
	BULK_ERR                // !<length>\r\n<error>\r\n
	VERBATIM_STR            // =<length>\r\n<encoding>:<data>\r\n
	MAPS                    // %<number-of-entries>\r\n<key-1><value-1>...<key-n><value-n>
	SETS                    // ~<number-of-elements>\r\n<element-1>...<element-n>
	PUSHES                  // ><number-of-elements>\r\n<element-1>...<element-n>
	// more
)

//-----------------------------------------------------------------------------
// ae
//-----------------------------------------------------------------------------

const (
	AE_ERR = -1
)
const (
	AE_READABLE FeType = iota + 1
	AE_WRITEABLE
)

// AE_NORMAL 周期执行的事件事件
// AE_ONCE 只执行一次
const (
	AE_NORMAL TeType = iota
	AE_ONCE
)

//-----------------------------------------------------------------------------
// cmd
//-----------------------------------------------------------------------------

const (
	CMD_UNKNOWN CmdType = iota
	CMD_INLINE
	CMD_BULK

	MAX_EXPIRE = 1000000000

	// server command

	PING = "ping"
	INFO = "info"

	// db command

	EXPIRE       = "expire"
	OBJECT       = "object"
	DEL          = "del"
	EXISTS       = "exists"
	RANDOMKEY    = "randomkey"
	FLUSHDB      = "flushdb"
	TYPE         = "type"
	PERSIST      = "persist"
	TTL          = "ttl"
	PTTL         = "pttl"
	KEYS         = "keys"
	BGREWRITEAOF = "bgrewriteaof"
	BGSAVE       = "bgsave"
	SAVE         = "save"
	DBSIZE       = "dbsize"
	SCAN         = "scan"
	SELECT       = "select"

	// string command

	GET  = "get"
	SET  = "set"
	INCR = "incr"
	DECR = "decr"

	// zset command

	Z_ADD   = "zadd"
	Z_RANGE = "zrange"
	Z_CARD  = "zcard"

	// set command

	S_ADD        = "sadd"
	SMEMBERS     = "smembers"
	SINTER       = "sinter"
	SINTER_STORE = "sinterstore"
	S_POP        = "spop"
	S_REM        = "srem"
	S_UNION      = "sunion"
	S_UNIONSTORE = "sunionstore"
	S_DIFF       = "sdiff"
	S_DIFFSTORE  = "sdiffstore"
	S_CARD       = "scard"

	// list command

	R_PUSH = "rpush"
	L_PUSH = "lpush"
	R_POP  = "rpop"
	L_POP  = "lpop"
	L_LEN  = "llen"

	// hash command

	H_SET    = "hset"
	H_GET    = "hget"
	H_DEL    = "hdel"
	H_EXISTS = "hexists"
	H_LEN    = "hlen"
	H_KEYS   = "hkeys"
	H_VALS   = "hvals"
	H_GETALL = "hgetall"
)

//-----------------------------------------------------------------------------
// snet
//-----------------------------------------------------------------------------

const (
	BACKLOG int = 64
)

//-----------------------------------------------------------------------------
// aof
//-----------------------------------------------------------------------------

const (
	REDIS_AOF_REWRITE_ITEMS_PER_CMD = 64
	REDIS_AOF_REWRITE_PERC          = 100
	REDIS_AOF_REWRITE_MIN_SIZE      = 102 * 1024
)

//-----------------------------------------------------------------------------
// rdb
//-----------------------------------------------------------------------------

const (
	REDIS_RDB_VERSION = "7"
	REDIS_RDB_BITS    = "64"
)

//-----------------------------------------------------------------------------
// srobj
//-----------------------------------------------------------------------------

const (
	REDIS_ENCODING_RAW        uint8 = iota // Raw representation
	REDIS_ENCODING_INT                     // Encoded as integer
	REDIS_ENCODING_HT                      // Encoded as hash table
	REDIS_ENCODING_ZIPMAP                  // Encoded as zipmap
	REDIS_ENCODING_LINKEDLIST              // Encoded as regular linked list
	REDIS_ENCODING_ZIPLIST                 // Encoded as ziplist
	REDIS_ENCODING_INTSET                  // Encoded as intset
	REDIS_ENCODING_SKIPLIST                // Encoded as skiplist
)

// SR_STR 字符串类型
// SR_LIST 列表类型
// SR_SET 集合类型
// SR_ZSET 有序集合类型
// SR_DICT 字典类型
const (
	SR_STR SRType = iota
	SR_LIST
	SR_SET
	SR_ZSET
	SR_DICT
)

//-----------------------------------------------------------------------------
// intSet
//-----------------------------------------------------------------------------

const (
	DEFAULT_INTSET_BUF = 4
)

//-----------------------------------------------------------------------------
// set
//-----------------------------------------------------------------------------

const (
	SET_OP_UNION = 0
	SET_OP_DIFF  = 1

	SPOP_MOVE_STRATEGY_MUL = 5
)

//-----------------------------------------------------------------------------
// dict
//-----------------------------------------------------------------------------

const (
	DICT_SET = 0
	DICT_REP = 1

	EXPIRE_CHECK_COUNT         int   = 100
	DICT_OK                          = 0
	DICT_ERR                         = 1
	DEFAULT_REHASH_STEP              = 1
	DICT_HT_INITIAL_SIZE       int64 = 4
	EXPEND_RATIO               int64 = 2
	LOAD_FACTOR                      = 1 // LOAD_FACTOR 负载因子
	BG_PERSISTENCE_LOAD_FACTOR       = 5 // BG_PERSISTENCE_LOAD_FACTOR bgsave或者bgrewriteaof 的负载因子

	HT_MIN_FILL = 10

	OBJ_HASH_KEY   = 1
	OBJ_HASH_VALUE = 2
)

//-----------------------------------------------------------------------------
// zset
//-----------------------------------------------------------------------------

const (
	ZSKIPLIST_MAXLEVEL = 32
	ZSKIPLIST_P        = 0.25
)

//-----------------------------------------------------------------------------
// list
//-----------------------------------------------------------------------------

const (
	AL_START_HEAD = 0
	AL_START_TAIL = 1
)
