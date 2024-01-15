# command list

## db

- expire

**expire key value**

~~~bash
127.0.0.1:6379> expire name 5
OK
~~~

- object encoding

**object encoding key**

~~~bash
127.0.0.1:6379> set k1 asdas
OK
127.0.0.1:6379> object encoding k1
"raw"
~~~

- del

**del key [key ...]**

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

- keys

**keys pattern**

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

## string

- set

**set key value**

~~~bash
127.0.0.1:6379> set k1 hello
OK
~~~

- get 

**get key**

~~~bash
127.0.0.1:6379> get k1
"hello"
~~~

## set

- sadd

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

- smembers

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

- sinter

**sinter key [key ...]**

~~~bash
127.0.0.1:6379> sinter s1 s2
1) "12"
127.0.0.1:6379> sinter s1 s4
(empty array)
~~~

- sinterstore

**sinterstore key [key ...]**

~~~bash
127.0.0.1:6379> sinterstore s6 s3 s4
(integer) 2
127.0.0.1:6379> smembers s6
1) "c"
2) "php"
~~~

## zset

- zadd

**zadd key score member [score member ...]**

~~~bash
127.0.0.1:6379> zadd zs 50 z1 40 z2 60 z3 45.5 z4
(integer) 4
~~~

- zrange

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