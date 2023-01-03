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

	auction "github.com/frederikgantriis/DISYS-EXAM2022/gRPC"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	username := os.Args[1]
	i, _ := strconv.Atoi(os.Args[2])
	fePort := int32(i)

	file, _ := openLogFile("./logs/clientlog.log")

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFlags(2 | 3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connection to front end
	fmt.Printf("Trying to dial: %v\n", fePort)
	conn, err := grpc.Dial(fmt.Sprintf(":%v", fePort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("User %v: Could not connect. Error: %s", username, err)
	}
	fe := auction.NewAuctionClient(conn)
	defer conn.Close()

	log.Printf("User %v: Connected to front end %v", username, fePort)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := strings.Split(scanner.Text(), " ")
		command[0] = strings.ToLower(command[0])

		if command[0] == "bid" {
			bidAmount, _ := strconv.Atoi(command[1])
			bid := &auction.BidRequest{User: username, Bid: int32(bidAmount)}

			res, err := fe.Bid(ctx, bid)
			if err != nil {
				log.Printf("ERROR: %v", err)
				continue
			}
			fmt.Printf("User %v: %v", username, res.Message)
		} else if command[0] == "result" {
			res, err := fe.Result(ctx, &auction.Request{})
			if err != nil {
				log.Printf("ERROR: %v", err)
				continue
			}

			fmt.Printf("User %v: %v", username, res.Message)
		} else if command[0] == "reset" {
			res, err := fe.Reset(ctx, &auction.Request{})
			if err != nil {
				log.Printf("ERROR: %v", err)
				continue
			}
			fmt.Printf("User %v: %v", username, res.Message)
		}
	}
}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}
