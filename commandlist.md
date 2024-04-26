# command list

> command Case-insensitive, but key and values Case-sensitive

## db

### EXPIRE

**EXPIRE key seconds**

The EXPIRE command is used to set the expiration time of a key in seconds. When the key expires, it cannot be used again.

~~~bash
127.0.0.1:6379> expire name 5
OK
~~~

### object encoding

**object encoding key**

~~~bash
127.0.0.1:6379> set k1 asdas
OK
127.0.0.1:6379> object encoding k1
"raw"
~~~

### DEL

**DEL key [key ...]**

The DEL command is used to delete one or more keys that already exist in the database. If they do not exist, they are automatically ignored.

~~~bash
127.0.0.1:6379> set k1 aaa
OK
127.0.0.1:6379> set k2 bbb
OK
127.0.0.1:6379> sadd s1 baidu.com google.com
(integer) 2
127.0.0.1:6379> zadd zs 12 hello 5 world
(integer) 2
127.0.0.1:6379> del k3
(integer) 0
127.0.0.1:6379> del k1
(integer) 1
127.0.0.1:6379> del k1 k2 s1 zs
(integer) 3
~~~

### KEYS

**KEYS pattern**

The KEYS command is used to find all keys that match the specified pattern.

~~~bash
127.0.0.1:6379> set name asdsa
OK
127.0.0.1:6379> set nnme asda
OK
127.0.0.1:6379> keys *
1) "name"       
2) "nnme"       
127.0.0.1:6379> keys n*
1) "name"
2) "nnme"
127.0.0.1:6379> keys nn*
1) "nnme"
127.0.0.1:6379> keys n*e
1) "name"
2) "nnme"
127.0.0.1:6379> keys n[af]me
1) "name"
127.0.0.1:6379> keys na?e
1) "name"
127.0.0.1:6379> keys nv?e
(empty array)
~~~

### EXISTS

**EXISTS key [key ...]**

The EXISTS command is used to check whether a specified key or multiple keys exist.

~~~bash
127.0.0.1:6379> set kk v1
OK
127.0.0.1:6379> exists kk
(integer) 1
127.0.0.1:6379> exists kk kkk
(integer) 1
~~~

### TTL

**TTL key**

The TTL command checks the remaining expiration time of the Key, in seconds

~~~bash
127.0.0.1:6379> expire url 50
OK
127.0.0.1:6379> TTL url
(integer) 40
~~~

### PTTL

**PTTL key**

The PTTL command checks the remaining expiration time of a Key in milliseconds.

~~~bash
127.0.0.1:6379> pTTL url
(integer) 36804
127.0.0.1:6379> pTTL url1
(integer) -2
127.0.0.1:6379> pTTL s1
(integer) -1
~~~

### PERSIST

**PERSIST key**

The PERSIST command is used to remove the expiration time of a specified key, so that the key will never expire

~~~bash
127.0.0.1:6379> expire url 50
OK
127.0.0.1:6379> ttl url
(integer) 44
127.0.0.1:6379> persist url
(integer) 1
127.0.0.1:6379> ttl url
(integer) -1
~~~

### RANDOMKEY

**RANDOMKEY**

The Redis RANDOMKEY command is used to randomly return a key from the current database. When the database is not empty, return a key; Returns nil when the database is empty.

~~~bash
127.0.0.1:6379> keys *
(empty array)
127.0.0.1:6379> RANDOMKEY
(nil)
127.0.0.1:6379> set name hello
OK
127.0.0.1:6379> set key world
OK
127.0.0.1:6379> RANDOMKEY
"key"
~~~

### FLUSHDB

**FLUSHDB**

The Redis FLUSHDB command will clear all data in Redis

~~~bash
127.0.0.1:6379> keys *
1) "k1"
2) "k2"
3) "k3"
127.0.0.1:6379> FLUSHDB
OK
127.0.0.1:6379> keys *
(empty array)
~~~

### TYPE

**TYPE key**

The Redis FLUSHDB command Returns the data type of the key, such as string, list, set, hash, zset, etc. If it returns none, it indicates that the key does not exist.

~~~bash
127.0.0.1:6379> SET webname www.baidu.com
OK
127.0.0.1:6379> TYPE webname
string
127.0.0.1:6379> LPUSH weburl www.baidu.com
(integer) 1
127.0.0.1:6379> TYPE weburl
list
127.0.0.1:6379> Hset name url www.baidu.com
(integer) 1
127.0.0.1:6379> TYPE name
hash
127.0.0.1:6379> SADD web www.taobao.com www.jd.com www.baidu.com
(integer) 3
127.0.0.1:6379> TYPE web
set
127.0.0.1:6379> ZADD zs 100 math
(integer) 1
127.0.0.1:6379> TYPE zs
zset
127.0.0.1:6379> TYPE zs11
none
~~~

## AOF

### BGREWRITEAOF

~~~bash
127.0.0.2:6379> BGREWRITEAOF
Background append only file rewriting started
~~~

## RDB

### BGSAVE

~~~bash
127.0.0.1:6379> bgsave
Background saving started
~~~

### SAVE

~~~bash
127.0.0.1:6379> save
OK
~~~

## string

### set

**set key value**

~~~bash
127.0.0.1:6379> set k1 hello
OK
~~~

### get 

**get key**

~~~bash
127.0.0.1:6379> get k1
"hello"
~~~

## list

### lpush

**lpush key value [value ...]**

~~~bash
127.0.0.1:6379> lpush l1 baidu.com google.com 4399.com 7k7k.com
(integer) 4
127.0.0.1:6379> object encoding l1
"linkedlist"
~~~

### rpush

**rpush key value [value ...]**

~~~bash
127.0.0.1:6379> rpush l2 baidu.com google.com 4399.com 7k7k.com
(integer) 4
127.0.0.1:6379> object encoding l2
"linkedlist"
~~~

### lpop

**lpop key**

~~~bash
127.0.0.1:6379> lpop l1
"7k7k.com"
127.0.0.1:6379> lpop l2
"baidu.com"
~~~

### rpop

**rpop key**

~~~bash
127.0.0.1:6379> rpop l1
"baidu.com"
127.0.0.1:6379> rpop l2
"7k7k.com"
~~~

## hash

### hset

**hset key field value**

~~~bash
127.0.0.1:6379> hset h1 k1 baidu.com
(integer) 1
127.0.0.1:6379> hset h1 k2 100
(integer) 1
127.0.0.1:6379> object encoding h1
"hashtable"
~~~

### hget

**hget key field**

~~~bash
127.0.0.1:6379> hget h1 k1
"baidu.com"
127.0.0.1:6379> hget h1 k2
"100"
~~~

## set

### sadd

**add key member [member ...]**

~~~bash
127.0.0.1:6379> sadd s1 12
(integer) 1
127.0.0.1:6379> sadd s2 12 c
(integer) 2
127.0.0.1:6379> sadd s3 12 c php
(integer) 3
127.0.0.1:6379> sadd s4 c php java
(integer) 3
127.0.0.1:6379> sadd s10 12 51 789 13 456
(integer) 5
127.0.0.1:6379> object encoding s10
"intset"
127.0.0.1:6379> object encoding s4
"hashtable"
~~~

### smembers

**smembers key**

~~~bash
127.0.0.1:6379> smembers s10
1) "12"
2) "13"
3) "51"
4) "456"
5) "789"
127.0.0.1:6379> smembers s4
1) "c"
2) "java"
3) "php"
~~~

### sinter

**sinter key [key ...]**

~~~bash
127.0.0.1:6379> sinter s1 s2
1) "12"
127.0.0.1:6379> sinter s1 s4
(empty array)
~~~

### sinterstore

**sinterstore key [key ...]**

~~~bash
127.0.0.1:6379> sinterstore s6 s3 s4
(integer) 2
127.0.0.1:6379> smembers s6
1) "c"
2) "php"
~~~

## zset

### zadd

**zadd key score member [score member ...]**

~~~bash
127.0.0.1:6379> zadd zs 50 z1 40 z2 60 z3 45.5 z4
(integer) 4
~~~

### zrange

**zrange key min max [withscores]**

~~~bash
127.0.0.1:6379> zrange z1 0 5 withscores
(nil)
127.0.0.1:6379> zrange zs 0 5 withscores
1) "z2"
2) "40.00"
3) "z4"
4) "45.50"
5) "z1"
6) "50.00"
7) "z3"
8) "60.00"
127.0.0.1:6379> object encoding zs
"skiplist"
~~~

// TODO MORE COMMAND