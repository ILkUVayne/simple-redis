# simple-redis

Simple Redis written in Go, mainly for learning, for reference only. Inspiration comes from [godis](https://github.com/archeryue/godis).

![](https://github.com/ILkUVayne/simple-redis/blob/dev/demonstrate.gif)

## Requirements

- Linux
- Go 1.21 or above

## Running

- mode 1

~~~bash
# step 1. build

# server
go build ./sredis.go
# cli
go build ./sredis-cli.go

# step 2. run

# server
./sredis
# or with conf
./sredis -c ./sredis.conf

# cli
./sredis-cli
# cli with args
./sredis-cli -host 127.0.0.1 -p 6379
~~~

- mode 2

~~~bash
# server
go run ./sredis.go
# server with config
go run ./sredis.go -c ./sredis.conf

# cli
go run ./sredis-cli.go
# cli with args
go run ./sredis-cli.go -host 127.0.0.1 -p 6379
~~~

## Playing

You can use the following client to play with Simple-Redis. Start a sredis instance, then in another terminal try the following:

- sredis-cli

~~~bash
./sredis-cli
127.0.0.1:6379> set name helloworld
OK
127.0.0.1:6379> get name
"helloworld"
127.0.0.1:6379>

~~~

- telnet

~~~bash
telnet 127.0.0.1 6379
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.
get name
*1
$2ly
~~~

- redis-cli

~~~bash
./redis-cli
127.0.0.1:6379> set name sadasda
OK
127.0.0.1:6379> get name
"sadasda"
127.0.0.1:6379> quit
~~~

You can find the list of all the available commands at https://github.com/ILkUVayne/simple-redis/blob/master/commandlist.md.

## project layout

- `cgo`: contains the cgo implementation.
- `src`: contains the Simple-Redis implementation, written in Go.
- `sredis.conf` is the configuration file of Simple-Redis.
- `sredis.go` is the entry point of the Simple-Redis server.
- `sredis-cli.go` is the entry point of the Simple-Redis client cli.

Enjoy!