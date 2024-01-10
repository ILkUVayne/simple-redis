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
127.0.0.1:6379> sadd s1 12 51 789 13 456
(integer) 5
127.0.0.1:6379> object encoding s1
"intset"
127.0.0.1:6379> sadd s2 12 51 789 13 456 asd
(integer) 6
127.0.0.1:6379> object encoding s2
"hashtable"
~~~

## zset

- zadd

**zadd key score member [score member ...] **

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