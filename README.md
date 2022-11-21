# redis_mirror_proxy
This redis protocol service will act as a redis service and mirror the commands between two redis servers. Used for zero downtime migration.

Through mirroring we can also do capacity setting of new redis servers.

# Demo

Note: you will only need docker to test the setup

```
# Start the main redis server
$ docker run --rm --name redis-main -p 6379:6379 redis
# Start the mirror redis server
$ docker run --rm --name redis-mirror -p 6381:6379 redis

# complile and run the mirror service in docker
$ docker run -ti --rm -v $PWD/:/ws --workdir /ws -p 6380:6380 golang:1 go run *.go \
     -main host.docker.internal:6379 -mirror host.docker.internal:6381 -addr :6380


# connect to the redis-mirror service (6380)
$ docker run -ti --rm --entrypoint /usr/local/bin/redis-cli redis -h host.docker.internal -p 6380 SET KEY VAL
OK
$ docker run -ti --rm --entrypoint /usr/local/bin/redis-cli redis -h host.docker.internal -p 6380 GET KEY
"VAL"

# check the value in the main server
$ docker run -ti --rm --entrypoint /usr/local/bin/redis-cli redis -h host.docker.internal -p 6379 GET KEY
"VAL"
# check the value in the mirror server
$ docker run -ti --rm --entrypoint /usr/local/bin/redis-cli redis -h host.docker.internal -p 6381 GET KEY
"VAL"
```