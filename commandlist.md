# command list

## db

- expire

**expire key value**

~~~bash
127.0.0.1:6379> expire name 5
OK
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

// TODO MORE COMMAND