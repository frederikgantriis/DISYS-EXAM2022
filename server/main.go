package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	dictionary "github.com/frederikgantriis/DISYS-EXAM2022/gRPC"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var hashtable map[string]string
var ownPort int32

func main() {
	file, _ := openLogFile("./logs/serverlog.log")

	hashtable = make(map[string]string)

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFlags(2 | 3)

	if len(os.Args) != 2 {
		log.Printf("Please inAdd a number to run the server on. Fx. inAddting 3 would run the server on port 3003")
		return
	}

	ownId := os.Args[1]

	listen, _ := net.Listen("tcp", "localhost:300"+ownId)

	convOwnId, _ := strconv.ParseInt(ownId, 10, 32)
	ownPort = int32(3000) + int32(convOwnId)

	grpcServer := grpc.NewServer()
	dictionary.RegisterDictionaryServer(grpcServer, &Server{
		id: int32(convOwnId),
	})

	log.Printf("%v: server listening at %v", ownPort, listen.Addr())

	grpcServer.Serve(listen)
}

func (s *Server) FollowerAdd(ctx context.Context, req *dictionary.AddRequest) (*dictionary.AddReply, error) {
	hashtable[req.Key] = req.Value

	return &dictionary.AddReply{Message: true}, nil
}

func (s *Server) FollowerRead(ctx context.Context, req *dictionary.ReadRequest) (*dictionary.ReadReply, error) {
	result := hashtable[req.Key]

	return &dictionary.ReadReply{Value: result}, nil
}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

type Server struct {
	dictionary.UnimplementedDictionaryServer
	id int32
}

func (server *Server) LeaderAdd(ctx context.Context, req *dictionary.AddRequest) (*dictionary.AddReply, error) {
	var reply *dictionary.AddReply
	var res *dictionary.AddReply

	follower, conn, err := connect(ownPort)

	if err != nil || follower == nil || conn == nil {
		log.Printf("Server %v: ERROR - %v\n", server.id, err)
	} else {
		defer conn.Close()
		res, _ = follower.FollowerAdd(ctx, req)
		if reply == nil {
			reply = res
		}
	}

	hashtable[req.Key] = req.Value
	res = &dictionary.AddReply{Message: true}

	if reply == nil {
		reply = res
	}

	return &dictionary.AddReply{Message: reply.GetMessage()}, nil
}

func (server *Server) LeaderRead(ctx context.Context, req *dictionary.ReadRequest) (*dictionary.ReadReply, error) {

	var reply *dictionary.ReadReply
	var res *dictionary.ReadReply

	follower, conn, err := connect(ownPort)

	if err != nil || follower == nil || conn == nil {
		log.Printf("Server %v: ERROR - %v\n", server.id, err)
	} else {
		defer conn.Close()
		res, _ = follower.FollowerRead(ctx, req)
		if reply == nil {
			reply = res
		}
	}

	result := hashtable[req.Key]
	res = &dictionary.ReadReply{Value: result}

	if reply == nil {
		reply = res
	}

	return &dictionary.ReadReply{Value: reply.GetValue()}, nil
}

func connect(ownPort int32) (dictionary.DictionaryClient, *grpc.ClientConn, error) {
	port := ownPort + 1
	var conn *grpc.ClientConn
	var err error

	go func() {
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err = grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	}()

	time.Sleep(2 * time.Second)

	if err != nil || conn == nil {
		log.Printf("Server %v: Could not connect: %s", ownPort, err)
		return nil, nil, err
	}

	follower := dictionary.NewDictionaryClient(conn)
	return follower, conn, err
}
