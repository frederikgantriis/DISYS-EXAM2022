package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	dictionary "github.com/frederikgantriis/DISYS-EXAM2022/gRPC"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	username := os.Args[1]
	leaderPort := int32(3000)

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

		if command[0] == "add" {
			Key := command[1]
			var Value string

			for _, v := range command[2:] {
				Value += v + " "
			}
			log.Printf("Key: %v, Value: %v\n", Key, Value)
			Add := &dictionary.AddRequest{Key: Key, Value: Value}

			var res *dictionary.AddReply
			var err error

			res, err = leader.LeaderAdd(ctx, Add)
			for err != nil {
				log.Printf("ERROR: %v\n", err)
				leaderPort++
				leader, conn = connect(username, leaderPort)
				defer conn.Close()
				res, err = leader.LeaderAdd(ctx, Add)
				continue
			}

			log.Println("result from Addrequest:", res)

		} else if command[0] == "read" {
			Key := command[1]
			Read := &dictionary.ReadRequest{Key: Key}

			var res *dictionary.ReadReply
			var err error

			res, err = leader.LeaderRead(ctx, Read)
			for err != nil {
				log.Printf("ERROR: %v\n", err)
				leaderPort++
				leader, conn = connect(username, leaderPort)
				defer conn.Close()
				res, err = leader.LeaderRead(ctx, Read)
				continue
			}

			log.Println("result from Readrequest:", res.Value)
		}
	}
}

func connect(username string, leaderPort int32) (dictionary.DictionaryClient, *grpc.ClientConn) {
	var leader dictionary.DictionaryClient
	var conn *grpc.ClientConn
	var err error

	// Connection to front end
	log.Printf("Trying to dial: %v\n", leaderPort)
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
