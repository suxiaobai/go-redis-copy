# go-redis-copy
Script to copy data by keys pattern from one redis instance to another.
Thanks for [go-redis-migrate](https://github.com/obukhov/go-redis-migrate)


### Usage
```
Usage of go-redis-copy:
  -check_pool_size int
        check redis exists exector pool size (default 30)
  -dest string
        destination redis connection uris
  -dest_pass string
        destination redis password
  -pattern string
        Only transfer matching keys (default *) (default "*")
  -pull_pool_size int
        get redis info exector pool size (default 30)
  -push_pool_size int
        load redis key exector pool size (default 30)
  -report_count int
        Migrate report count option (default: 10000)  (default 10000)
  -scan_count int
        Redis scan count option (default: 1000)  (default 1000)
  -src string
        source redis connection uris
  -src_pass string
        source redis password

# redis to redis cluster
./go-redis-copy -src 127.0.0.1:6379 -dest 127.0.0.1:7000,127.0.0.1:7001 
```

### Support Redis Type
| Type   | export Command | import command        |
| :------| -------------: | :-------------------: |
| string | get key        | set key               |
| set    | smembers key   | sadd key ...values    |
| zset   | zrangebyscore key -inf +inf withscores | zadd key ...scores ...values | 
| hash   | hgetall key    | hmset key ...fields ...values |
| list   | lrange key 0 -1 | lpush key ...values | 
