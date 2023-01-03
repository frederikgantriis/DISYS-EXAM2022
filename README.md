# AuctionSystem-DISYS

## How to run

### Servers

The system will have 3 servers. You therefore need to run:

```shl
go run server/main.go 0
go run server/main.go 1
go run server/main.go 2
```

each in a seperate terminal.

### Clients

To run a client:

```shl
go run client/main.go [username] [port of front end]
```

With `[username]` a chosen username and `[port of front end]` the port of the front end for each client.

Example of use:

```shl
go run client/main.go Foo 5000
```

(Having run a front end on port `5000` in another terminal.)

### Front ends

To run a front end:

```shl
go run frontend/main.go [port]
```

With `[port]` a chosen port from which the front end can be accessed (user for connecting a client to the port).

Example of use:

```shl
go run frontend/main.go 5000
```
