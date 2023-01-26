# udp2ws

This is a toy tool that will setup a udp server and a websocket server, and broadcast the data it received from the UDP port to all connected websocket clients.

## Run server

```bash
$ go run main.go -wsaddr 127.0.0.1:6080 -udpaddr 127.0.0.1:6081 -data text
# use -h for help
```

## To test it locally

1, Run udp2ws server:
```bash
$ go run main.go -wsaddr 127.0.0.1:6080 -udpaddr 127.0.0.1:6081 -data text
```

2, Use `wscat` to connect to the websocket address
```bash
$ wscat --connect 127.0.0.1:6080
```

3, Use a udp client connect to the udp address and send some data. 
```bash
$ nc -u 127.0.0.1 6081 
hello world
```

Then in the 2nd terminal which runs `wscat`, you should be able to see the data you just sent.