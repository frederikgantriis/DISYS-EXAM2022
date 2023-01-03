package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	dictionary "github.com/frederikgantriis/DISYS-EXAM2022/gRPC"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	username := os.Args[1]
	i, _ := strconv.Atoi(os.Args[2])
	leaderPort := int32(i)

	file, _ := openLogFile("./logs/clientlog.log")

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFlags(2 | 3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	leader, conn := connect(username, leaderPort)
	defer conn.Close()

	log.Printf("User %v: Connected to front end %v\n", username, leaderPort)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := strings.Split(scanner.Text(), " ")
		command[0] = strings.ToLower(command[0])

		if command[0] == "put" {
			Key := command[1]
			Value := command[2]
			fmt.Printf("Key: %v, Value: %v\n", Key, Value)
			put := &dictionary.PutRequest{Key: Key, Value: Value}

			var res *dictionary.PutReply
			var err error

			res, err = leader.LeaderPut(ctx, put)
			for err != nil {
				log.Printf("ERROR: %v\n", err)
				leaderPort++
				log.Printf("LeaderPort: %v\n", leaderPort)
				leader, conn = connect(username, leaderPort)
				defer conn.Close()
				res, err = leader.LeaderPut(ctx, put)
				continue
			}

			fmt.Println("result from putrequest:", res)

		} else if command[0] == "get" {
			Key := command[1]
			get := &dictionary.GetRequest{Key: Key}

			var res *dictionary.GetReply
			var err error

			res, err = leader.LeaderGet(ctx, get)
			for err != nil {
				log.Printf("ERROR: %v\n", err)
				leaderPort++
				log.Printf("LeaderPort: %v\n", leaderPort)
				leader, conn = connect(username, leaderPort)
				defer conn.Close()
				res, err = leader.LeaderGet(ctx, get)
				continue
			}

			fmt.Println("result from getrequest:", res.Value)
		}
	}
}

func connect(username string, leaderPort int32) (dictionary.DictionaryClient, *grpc.ClientConn) {
	var leader dictionary.DictionaryClient
	var conn *grpc.ClientConn
	var err error

	// Connection to front end
	fmt.Printf("Trying to dial: %v\n", leaderPort)
	conn, err = grpc.Dial(fmt.Sprintf(":%v", leaderPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	for err != nil {
		log.Fatalf("User %v: Could not connect. Error: %s\n", username, err)
		leaderPort = leaderPort + 1
		log.Printf("User %v: Trying again with port %v\n", username, leaderPort)
		conn, err = grpc.Dial(fmt.Sprintf(":%v", leaderPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	}
	leader = dictionary.NewDictionaryClient(conn)

	return leader, conn
}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}
