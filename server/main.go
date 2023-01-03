package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	dictionary "github.com/frederikgantriis/DISYS-EXAM2022/gRPC"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var hashtable map[string]string
var follower dictionary.DictionaryClient
var ownPort int32

func main() {
	file, _ := openLogFile("./logs/serverlog.log")

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFlags(2 | 3)

	if len(os.Args) != 2 {
		log.Printf("Please input a number to run the server on. Fx. inputting 3 would run the server on port 3003")
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

	log.Printf("server listening at %v", listen.Addr())

	grpcServer.Serve(listen)
}

func (s *Server) FollowerPut(ctx context.Context, req *dictionary.PutRequest) (*dictionary.PutReply, error) {
	hashtable[req.Key] = req.Value

	return &dictionary.PutReply{Message: true}, nil
}

func (s *Server) FollowerGet(ctx context.Context, req *dictionary.GetRequest) (*dictionary.GetReply, error) {
	result := hashtable[req.Key]

	return &dictionary.GetReply{Value: result}, nil
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

/* func main() {
	ownPort, _ := strconv.Atoi(os.Args[1])

	file, _ := openLogFile("./logs/frontendlog.log")

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFlags(2 | 3)

	servers = make([]dictionary.DictionaryClient, 3)

	for i := 0; i < 3; i++ {

	}

	listen, _ := net.Listen("tcp", "localhost:"+fmt.Sprint(ownPort))

	grpcServer := grpc.NewServer()
	dictionary.RegisterDictionaryServer(grpcServer, &FrontEnd{
		id: int32(ownPort),
	})

	log.Printf("Front end listening at %v", listen.Addr())

	grpcServer.Serve(listen)

	log.Printf("Front end served")
} */

func (server *Server) LeaderPut(ctx context.Context, req *dictionary.PutRequest) (*dictionary.PutReply, error) {
	w := 0
	var reply *dictionary.PutReply
	connect(ownPort)

	//TODO: Put into own hashmap

	res, err := follower.FollowerPut(ctx, req)
	if err != nil {
		log.Printf("Front end %v: ERROR - %v", server.id, err)
	}
	w++

	if reply == nil || (reply.GetMessage() != res.GetMessage() && reply.GetMessage() == true) {
		reply = res
	}

	return &dictionary.PutReply{Message: reply.GetMessage()}, nil
}

func (server *Server) LeaderGet(ctx context.Context, req *dictionary.GetRequest) (*dictionary.GetReply, error) {
	r := 0
	var reply *dictionary.GetReply
	connect(ownPort)

	res, err := follower.FollowerGet(ctx, req)
	if err != nil {
		log.Printf("Front end %v: ERROR - %v", server.id, err)
	}
	r++

	//TODO: Get own value from hashmap

	// Take the first result
	if reply == nil {
		reply = res
	}

	return &dictionary.GetReply{Value: reply.GetValue()}, nil
}

func connect(ownPort int32) {
	port := ownPort

	fmt.Printf("Trying to dial: %v\n", port)
	conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Front end %v: Could not connect: %s", ownPort, err)
	}
	follower = dictionary.NewDictionaryClient(conn)
	defer conn.Close()
}
