# collector-tcp
TCP collector for the qframe framework.

## cmd/qframe-collector-tcp/main.go

The example script will instantiate the collector and wait for a message send to it.


```bash
go run main.go
2017/04/21 12:37:29 [II] Dispatch broadcast for Data and Tick
2017/04/21 12:37:29 [  INFO] test >> Listening on 127.0.0.1:11001
```

Once send...

```bash
$ echo "Test-$(date +%s)" | nc -w1  127.0.0.1 11001
```

... the message will be displayed and the script exits:

```
#### Received (remote:127.0.0.1:60846): Test-1492771635
```


## Deveolpment

### Start Dev-Container

```bash
$ docker run -ti --name qframe-collector-tcp --rm -e SKIP_ENTRYPOINTS=1 -p 11001:11001 \
           -v ${GOPATH}/src/github.com/qnib/qframe-collector-tcp:/usr/local/src/github.com/qnib/qframe-collector-tcp \
           -v ${GOPATH}/src/github.com/qnib/qframe-collector-docker-events/lib:/usr/local/src/github.com/qnib/qframe-collector-docker-events/lib \
           -v ${GOPATH}/src/github.com/qnib/qframe-types:/usr/local/src/github.com/qnib/qframe-types \
           -v ${GOPATH}/src/github.com/qnib/qframe-utils:/usr/local/src/github.com/qnib/qframe-utils \
           -v ${GOPATH}/src/github.com/qnib/qframe-inventory/lib:/usr/local/src/github.com/qnib/qframe-inventory/lib \
           -v ${GOPATH}/src/github.com/qnib/qframe-filter-inventory/lib:/usr/local/src/github.com/qnib/qframe-filter-inventory/lib \
           -v /var/run/docker.sock:/var/run/docker.sock \
           -w /usr/local/src/github.com/qnib/qframe-collector-tcp \
            qnib/uplain-golang bash

$ govendor update github.com/qnib/qframe-filter-inventory/lib \
                  github.com/qnib/qframe-inventory/lib \
                  github.com/qnib/qframe-collector-docker-events/lib \
                  github.com/qnib/qframe-collector-tcp/lib \
                  github.com/qnib/qframe-types github.com/qnib/qframe-utils 
```

### Start collector

```bash
$ go run main.go
2017/05/01 01:05:07 [II] Dispatch broadcast for Back, Data and Tick
2017/05/01 01:05:07 [  INFO] docker-events >> Start docker-events collector v0.2.1
2017/05/01 01:05:07 [  INFO] inventory >> Start inventory v0.1.1
2017/05/01 01:05:07 [  INFO] docker-events >> Connected to 'moby' / v'17.05.0-ce-rc1'
2017/05/01 01:05:09 [  INFO] tcp >> Listening on 0.0.0.0:11001
2017/05/01 01:05:41 [  INFO] tcp >> Received TCP message 'cee{"data": "test 123", "event_code": "001.001"}' from '172.17.0.3'
```

When a message is send...

```bash
$ docker run -ti --rm --name event-sender \
         qnib/qframe-collector-tcp-sender /usr/local/bin/send-event.sh \
         $(docker inspect -f '{{ .NetworkSettings.Networks.bridge.IPAddress }}' qframe-collector-tcp)
```

... it will be received like this: 

```bash
2017/05/01 01:05:41 [  INFO] inventory >> Received InventoryRequest for {2017-05-01 01:05:41.241390067 +0000 UTC   172.17.0.3 0xc4201c7b00}
2017/05/01 01:15:07 [  INFO] tcp >> Got inventory response for msg: 'cee{"data": "test 123", "event_code": "001.001"}'
2017/05/01 01:16:10 [  INFO] tcp >>         Container{Name:/event-sender, Image: sha256:00e3f5e01ec09673e36e477e522b6aefc2c17f969f266f3217e090f8d1941d69}
```
