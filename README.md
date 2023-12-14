# simple-redis

## Building

- build

This command will build an executable file **sredis**

~~~bash
go build ./sredis.go
~~~

- run

This command will directly start the server

~~~bash
go run ./sredis.go
~~~

## Running

~~~bash
./sredis

# or with conf
./sredis -c ./sredis.conf
~~~

## Playing

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