package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	auction "github.com/frederikgantriis/DISYS-EXAM2022/gRPC"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FrontEnd struct {
	auction.UnimplementedAuctionServer
	id int32
}

var servers []auction.AuctionClient
var ownPort int

func main() {
	ownPort, _ := strconv.Atoi(os.Args[1])

	file, _ := openLogFile("./logs/frontendlog.log")

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFlags(2 | 3)

	servers = make([]auction.AuctionClient, 3)

	for i := 0; i < 3; i++ {
		port := int32(3000) + int32(i)

		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Front end %v: Could not connect: %s", ownPort, err)
		}
		servers[i] = auction.NewAuctionClient(conn)
		defer conn.Close()
	}

	listen, _ := net.Listen("tcp", "localhost:"+fmt.Sprint(ownPort))

	grpcServer := grpc.NewServer()
	auction.RegisterAuctionServer(grpcServer, &FrontEnd{
		id: int32(ownPort),
	})

	log.Printf("Front end listening at %v", listen.Addr())

	grpcServer.Serve(listen)

	log.Printf("Front end served")
}

func (fe *FrontEnd) Bid(ctx context.Context, req *auction.BidRequest) (*auction.ClientReply, error) {
	w := 0
	var reply *auction.OutcomeReply
	for _, server := range servers {
		res, err := server.ServerBid(ctx, req)
		if err != nil {
			log.Printf("Front end %v: ERROR - %v", fe.id, err)
			continue
		}
		w++

		if reply == nil || (reply.GetOutcome() != res.GetOutcome() && reply.GetOutcome() == auction.Outcomes_SUCCESS) {
			reply = res
		}
	}

	var message string
	if w >= 2 {
		if reply.GetOutcome() == auction.Outcomes_SUCCESS {
			message = fmt.Sprintf("Made a succesfull bid, for amount: %v", req.GetBid())
		} else if reply.GetOutcome() == auction.Outcomes_FAIL {
			message = "Bid was either too low or auction has ended"
		}
	} else {
		message = fmt.Sprintf("Call successful server writes and delete write (not implemented), Wrote to %v servers", w)
	}
	return &auction.ClientReply{Message: message}, nil
}

func (fe *FrontEnd) Result(ctx context.Context, req *auction.Request) (*auction.ClientReply, error) {
	r := 0
	var reply *auction.ResultReply
	for _, server := range servers {
		res, err := server.ServerResult(ctx, req)
		if err != nil {
			log.Printf("Front end %v: ERROR - %v", fe.id, err)
			continue
		}
		r++

		// Take the first result
		if reply == nil {
			reply = res
			continue
		}

		// Update highestBid
		if res.GetHighestBid() > reply.GetHighestBid() {
			reply.HighestBid = res.GetHighestBid()
			reply.User = res.GetUser()
		}

		// Update timeleft
		if res.GetTimeLeft() < reply.GetTimeLeft() {
			reply.TimeLeft = reply.GetTimeLeft()
		}
	}

	var message string
	if r >= 2 {
		if reply.GetTimeLeft() > 0 {
			message = fmt.Sprintf("Current highest bid: %v by %v\n The auction ends in: %v seconds", reply.GetHighestBid(), reply.GetUser(), reply.GetTimeLeft())
		} else {
			message = fmt.Sprintf("The auction has ended. The winner was %v with the bid: %v", reply.GetUser(), reply.GetHighestBid())
		}
	} else {
		log.Printf("Front end %v: only %v answers was received from the servers and therefore couldn't produce a result for the client", fe.id, r)
		message = "Didn't receive enough answers to get a usefull result"
	}
	return &auction.ClientReply{Message: message}, nil
}

func (fe *FrontEnd) Reset(ctx context.Context, req *auction.Request) (*auction.ClientReply, error) {
	w := 0
	for _, server := range servers {
		_, err := server.ServerReset(ctx, req)
		if err != nil {
			log.Printf("Front end %v: ERROR - %v", fe.id, err)
			continue
		}
		w++
	}
	var message string
	if w >= 2 {
		message = "Reset successfull"
	} else {
		message = "Couldn't write to enough server to complete a reset"
	}

	return &auction.ClientReply{Message: message}, nil
}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}
