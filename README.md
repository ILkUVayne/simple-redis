# simple-redis

Simple Redis written in Go, mainly for learning, for reference only.

![](https://github.com/ILkUVayne/simple-redis/blob/dev/demonstrate.gif)

## Requirements

- Linux
- Go 1.21 or above

## Building

- build

~~~bash
# server
go build ./sredis.go

# cli
go build ./sredis-cli.go
~~~

- run

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

## Running

~~~bash
# server
./sredis
# or with conf
./sredis -c ./sredis.conf

# cli
./sredis-cli
# cli with args
./sredis-cli -host 127.0.0.1 -p 6379
~~~

## Playing

Please refer to  [command list](https://github.com/ILkUVayne/simple-redis/blob/master/commandlist.md) for the complete command documentation

- use sredis-cli

~~~bash
./sredis-cli
127.0.0.1:6379> set name helloworld
OK
127.0.0.1:6379> get name
"helloworld"
127.0.0.1:6379>

~~~

- use telnet

~~~bash
telnet 127.0.0.1 6379
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.
get name
*1
$2ly
~~~

- use redis-cli

~~~bash
./redis-cli
127.0.0.1:6379> set name sadasda
OK
127.0.0.1:6379> get name
"sadasda"
127.0.0.1:6379> quit
~~~