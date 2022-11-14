# redis_mirror_proxy
This redis protocol service will act as a redis service and mirror the commands between two redis servers. Used for zero downtime migration.

Through mirroring we can also do capacity setting of new redis servers.

# Demo

Note: you will only need docker to test the setup

```
# complile and run the mirror service in docker
$ docker run -ti --rm -v $PWD/:/ws --workdir /ws -p 6380:6380 golang:1  go run *.go -main host.docker.internal:6379 -addr :6380

# Start the main redis server
$ docker run --rm --name some-redis -p 6379:6379 redis

# connect to the redis server directly and set a value (6379)
$  docker run -ti --rm --entrypoint /usr/local/bin/redis-cli redis -h host.docker.internal -p 6379 set KEY VAL
OK

# connect to the redis-mirror service (6380)
$  docker run -ti --rm --entrypoint /usr/local/bin/redis-cli redis -h host.docker.internal -p 6380 GET KEY 
"VAL"
$  docker run -ti --rm --entrypoint /usr/local/bin/redis-cli redis -h host.docker.internal -p 6380 SET KEY NEWVAL
OK
$  docker run -ti --rm --entrypoint /usr/local/bin/redis-cli redis -h host.docker.internal -p 6380 GET KEY
"NEWVAL"

# check the value in the main server
$ docker run -ti --rm --entrypoint /usr/local/bin/redis-cli redis -h host.docker.internal -p 6379 GET KEY
"NEWVAL"
```