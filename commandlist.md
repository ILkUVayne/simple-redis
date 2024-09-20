# command list

> command Case-insensitive, but key and values Case-sensitive

## db

### EXPIRE

**EXPIRE key seconds**

The EXPIRE command is used to set the expiration time of a key in seconds. When the key expires, it cannot be used again.

~~~bash
127.0.0.1:6379> set name "hello world"
OK
127.0.0.1:6379> expire name 5
OK
127.0.0.1:6379> ttl name
(integer) 3
~~~

### object encoding

**object encoding key**

Returns the internal encoding for the Redis object stored at <key>

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
127.0.0.1:6379> exists names
(integer) 0
~~~

### TTL

**TTL key**

The TTL command checks the remaining expiration time of the Key, in seconds.If the key does not exist, return -2; If the key exists but the remaining survival time is not set, return -1.

~~~bash
127.0.0.1:6379> set url www
OK
127.0.0.1:6379> expire url 50
OK
127.0.0.1:6379> TTL url
(integer) 40
~~~

### PTTL

**PTTL key**

The PTTL command checks the remaining expiration time of a Key in milliseconds.When the key does not exist, return -2. When the key exists but the remaining survival time is not set, return -1. Otherwise, return the remaining lifetime of the key.

~~~bash
127.0.0.1:6379> set url www
OK
127.0.0.1:6379> set s1 asd
OK
127.0.0.1:6379> expire url 50
OK
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
127.0.0.1:6379> set url www
OK
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

Instruct Redis to start an Append Only File rewrite process.

~~~bash
127.0.0.2:6379> BGREWRITEAOF
Background append only file rewriting started
~~~

## RDB

### BGSAVE

Save the DB in background.

~~~bash
127.0.0.1:6379> bgsave
Background saving started
~~~

### SAVE

Save the DB.

~~~bash
127.0.0.1:6379> save
OK
~~~

## string

### set

**set key value**

Set the value stored in the key. When the key has already stored other values, the SET command will overwrite the original value and reset the new value.

~~~bash
127.0.0.1:6379> set k1 hello
OK
127.0.0.1:6379> set k2 "hello world"
OK
~~~

### get 

**get key**

Returns the string value stored by the key. If the key does not exist, it returns null. If the key stores a value that is not a string type, it returns an error because the GET command can only handle strings.

~~~bash
127.0.0.1:6379> get k1
"hello"
127.0.0.1:6379> get k2
"hello world"
~~~

### incr

**incr key**

Add 1 to the value stored in the key.

If the key does not exist, the value of the key will be initialized to 0 before executing the INCR operation. If the value contains an incorrect type, or if a value of string type cannot be represented as a number, then an error is returned.

~~~bash
127.0.0.1:6379> set n1 20
OK
127.0.0.1:6379> incr n1
(integer) 21
127.0.0.1:6379> get n1
"21"
127.0.0.1:6379> incr n2
(integer) 1
127.0.0.1:6379> get n2
"1"
127.0.0.1:6379> sadd n3 sds fs
(integer) 2
127.0.0.1:6379> incr n3
(error) ERR: Operation against a key holding the wrong kind of value
~~~

### decr

**decr key**

Subtract the value stored in the key by 1.

If the key does not exist, the value of the key will be initialized to 0 before executing the DECR operation. If the value contains an incorrect type, or if a value of string type cannot be represented as a number, then an error is returned.

~~~bash
127.0.0.1:6379> set n1 30
OK
127.0.0.1:6379> decr n1
(integer) 29
127.0.0.1:6379> get n1
"29"
127.0.0.1:6379> decr n2
(integer) -1
127.0.0.1:6379> get n2
"-1"
127.0.0.1:6379> sadd n3 baidu google
(integer) 2
127.0.0.1:6379> decr n3
(error) ERR: Operation against a key holding the wrong kind of value
~~~

## list

### lpush

**lpush key value [value ...]**

Insert one or more values into the header of the list (starting from the left), and if there are multiple value values, insert them in order from left to right.

If the key does not exist, an empty list will be automatically created and LPUSH operations will be executed; When the key exists but is not a list type, an error is returned.

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

Remove and return the header element of the list key.

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

### llen

**llen key**

Returns the length of the list stored at key. If key does not exist, it is interpreted as an empty list and 0 is returned. An error is returned when the value stored at key is not a list.

~~~bash
127.0.0.1:6379> rpush list1 foo
(integer) 1
127.0.0.1:6379> rpush list1 bar
(integer) 2
127.0.0.1:6379> llen list1
(integer) 2
127.0.0.1:6379> llen list2
(integer) 0
~~~

## hash

### hset

**hset key field value**

Set the value of the field in the hash table key to value. If the key does not exist, a new hash table will be automatically created and HSET operations will be performed. If the field has already been saved, the old value will be overwritten.

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

Returns the field value of the given field in the hash table key.

~~~bash
127.0.0.1:6379> hget h1 k1
"baidu.com"
127.0.0.1:6379> hget h1 k2
"100"
~~~

### hdel

**hget key field [field ...]**

Delete one or more specified fields in the hash table key, and non-existent fields will be ignored.

~~~bash
127.0.0.1:6379> hset h1 k1 aaa
(integer) 1
127.0.0.1:6379> hset h1 k2 bbb
(integer) 1
127.0.0.1:6379> hset h1 k3 ccc
(integer) 1
127.0.0.1:6379> hdel h1 k1
(integer) 1
127.0.0.1:6379> hget h1 k1
(nil)
127.0.0.1:6379> hdel h1 k1 k2 k3
(integer) 2
127.0.0.1:6379> hdel s4 k1
(error) ERR: Operation against a key holding the wrong kind of value
~~~

### hexists

**hexists key field**

Check if the given field exists in the hash table key.

~~~bash
127.0.0.1:6379> hset h2 k1 baidu
(integer) 1
127.0.0.1:6379> hexists h2 k1
(integer) 1
127.0.0.1:6379> hexists h2 k2
(integer) 0
127.0.0.1:6379> hexists h3 k2
(integer) 0
127.0.0.1:6379> sadd s1 as asdas
(integer) 2
127.0.0.1:6379> type s1
set
127.0.0.1:6379> hexists s1 k2
(error) ERR: Operation against a key holding the wrong kind of value
~~~

### hlen

**hlen key**

Get the number of fields in the hash table.

If the key does not exist, return 0.

When the key exists but is not a hash type, an error is returned.

~~~bash
127.0.0.1:6379> hlen h1
(integer) 0
127.0.0.1:6379> hset h1 k1 aaa
(integer) 1
127.0.0.1:6379> hset h1 k2 bbb
(integer) 1
127.0.0.1:6379> hlen h1
(integer) 2
127.0.0.1:6379> set name aaa
OK
127.0.0.1:6379> hlen name
(error) ERR: Operation against a key holding the wrong kind of value
~~~

### hkeys

**hkeys key**

Return all field in the hash table, When the key does not exist, return an empty list

~~~bash
127.0.0.1:6379> hkeys h1
(empty array)
127.0.0.1:6379> hset h1 k1 aaa
(integer) 1
127.0.0.1:6379> hset h1 k2 bbb
(integer) 1
127.0.0.1:6379> hkeys h1
1) "k1"
2) "k2"
~~~

### hvals

**hvals key**

Return all values in the hash table, When the key does not exist, return an empty list

~~~bash
127.0.0.1:6379> hvals h1
(empty array)   
127.0.0.1:6379> hset h1 k1 aaa
(integer) 1     
127.0.0.1:6379> hset h1 k2 bbb
(integer) 1     
127.0.0.1:6379> hvals h1
1) "aaa"
2) "bbb"
~~~

### hgetall

**hgetall key**

Return all field and values in the hash table, When the key does not exist, return an empty list

~~~bash
127.0.0.1:6379> hgetall h1
(empty array)
127.0.0.1:6379> hset h1 k1 aaa
(integer) 1
127.0.0.1:6379> hset h1 k2 bbb
(integer) 1
127.0.0.1:6379> hgetall h1
1) "k1"
2) "aaa"
3) "k2"
4) "bbb"
~~~

## set

### sadd

**sadd key member [member ...]**

Adding one or more member elements to the set key will ignore member elements that already exist in the set. If the key does not exist, automatically create a collection containing member elements. When the key is not a collection type, an error is returned.

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

Returns all members in the set key.

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

Returns all members of a set, which is the intersection of all given sets. For non-existent keys, they are considered as empty sets. If there is an empty set in a given set, the result is also an empty set.

~~~bash
127.0.0.1:6379> sinter s1 s2
1) "12"
127.0.0.1:6379> sinter s1 s4
(empty array)
~~~

### sinterstore

**sinterstore key [key ...]**

Save the results to the destination set instead of simply returning the result set. If the destination set already exists, overwrite it.

The destination can be the key itself.

~~~bash
127.0.0.1:6379> sinterstore s6 s3 s4
(integer) 2
127.0.0.1:6379> smembers s6
1) "c"
2) "php"
~~~

### sunion

**sunion key [key ...]**

Returns all members of a set that is the union of all given sets.

~~~bash
127.0.0.1:6379> SADD website www.biancheng.net www.baidu.com
(integer) 2
127.0.0.1:6379> SADD site git python svn docker
(integer) 4
127.0.0.1:6379> SUNION website site
1) "docker"
2) "www.biancheng.net"
3) "python"
4) "www.baidu.com"
5) "svn"
6) "git"
~~~

### sunionstore

**sunionstore destination key [key ...]**

Similar to the SUNION command, but it saves the results to the destination set instead of simply returning the result set. If the destination already exists, overwrite it.

The destination can be the key itself.

~~~bash
127.0.0.1:6379> SADD website www.biancheng.net www.baidu.com
(integer) 2
127.0.0.1:6379> SADD site git python svn docker
(integer) 4
127.0.0.1:6379> SUNIONSTORE mysite site website
(integer) 6
127.0.0.1:6379> SMEMBERS mysite
1) "docker"
2) "git"
3) "python"
4) "svn"
5) "www.baidu.com"
6) "www.biancheng.net"
~~~

### sdiff

**sdiff key [key ...]**

Returns the difference set between the first set and other sets, which can also be considered as an element unique to the first set. A non-existent set key will be considered an empty set. For non-existent keys, they will be treated as empty sets.

~~~bash
127.0.0.1:6379> SADD website www.biancheng.net www.baidu.com www.jd.com
(integer) 3
127.0.0.1:6379> SADD site www.biancheng.net www.baidu.com stackoverflow.com
(integer) 3
127.0.0.1:6379> SDIFF website site
1) "www.jd.com"
127.0.0.1:6379> SDIFF site website
1) "stackoverflow.com"
~~~

### sdiffstore

**sdiffstore destination key [key ...]**

Similar to the SDIFF command, but the former saves the results to the destination set instead of simply returning the result set. If the destination set already exists, overwrite it.

The destination can be the key itself.

~~~bash
127.0.0.1:6379> SADD website www.biancheng.net www.baidu.com www.jd.com
(integer) 3
127.0.0.1:6379> SADD site www.biancheng.net www.baidu.com stackoverflow.com
(integer) 3
127.0.0.1:6379> SDIFFSTORE mysite website site
(integer) 1
127.0.0.1:6379> SDIFFSTORE mysite1 site website
(integer) 1
127.0.0.1:6379> smembers mysite
1) "www.jd.com"
127.0.0.1:6379> smembers mysite1
1) "stackoverflow.com"
~~~

### srem

**SREM key member [member ...]**

Remove one or more member elements from the set key, non-existent member elements will be ignored. When the key is not a collection type, return an error.

~~~bash
127.0.0.1:6379> sadd s1 11 aaa 5 bbb 66 ccc
(integer) 6
127.0.0.1:6379> srem s1 aaa
(integer) 1
127.0.0.1:6379> srem s1 aaa 11 ccc
(integer) 2
127.0.0.1:6379> smembers s1
1) "5"
2) "bbb"
3) "66"
~~~

### spop

**SPOP key [count]**

Remove and return a random element from the collection.

~~~bash
127.0.0.1:6379> sadd s1 11 5 88 baidu google
(integer) 5
127.0.0.1:6379> spop s1
"google"
127.0.0.1:6379> smembers s1
1) "88"
2) "baidu"
3) "5"
4) "11"
127.0.0.1:6379> spop s1 3
1) "88"
2) "baidu"
3) "5"
127.0.0.1:6379> smembers s1
1) "11"
~~~

## zset

### zadd

**zadd key score member [score member ...]**

Add one or more member elements and their score values to the ordered set key.

~~~bash
127.0.0.1:6379> zadd zs 50 z1 40 z2 60 z3 45.5 z4
(integer) 4
~~~

### zrange

**zrange key min max [withscores]**

Returns the members within a specified interval in an ordered set key, with their positions sorted in ascending order of score (from smallest to largest).

By using the withscores option, return the member and its score value together.

~~~bash
127.0.0.1:6379> zrange z1 0 5 withscores
(empty array)
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