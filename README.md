# DISYS-EXAM2022

## How to run

### Servers

The system will have 2 servers. You therefore need to run:

```shl
go run server/main.go 0
go run server/main.go 1
```

each in a seperate terminal.

### Clients

To run a client:

```shl
go run client/main.go [username]
```

With `[username]` a chosen username

Example of use:

```shl
go run client/main.go Foo 
```

### Input format
To add a word to the dictionary from the client:
```
add [word] [defintion]
```
With `[word]` being a single word and `[definition]` being n-amount of words seperated by spaces

To read a word to the dictionary from the client:
```
read [word]
```
With `[word]` being a single word

REMINDER: The `[word]` is case sensitive
