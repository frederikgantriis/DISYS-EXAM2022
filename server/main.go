package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	auction "github.com/frederikgantriis/DISYS-EXAM2022/gRPC"
	"google.golang.org/grpc"
)

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

	grpcServer := grpc.NewServer()
	auction.RegisterAuctionServer(grpcServer, &Server{
		id:                int32(convOwnId),
		highestBid:        0,
		timeLeft:          -1,
		currentWinnerUser: "",
	})

	log.Printf("server listening at %v", listen.Addr())

	grpcServer.Serve(listen)
}

func (s *Server) ServerPut(ctx context.Context, req *auction.PutRequest) (*auction.OutcomeReply, error) {
	// if a bid is made when timeLeft is -1, a new auction starts
	log.Printf("server %v: recieved a bid from %v. Amount: %v", s.id, req.GetUser(), req.GetBid())
	if s.timeLeft == -1 {
		s.highestBid = 0
		s.currentWinnerUser = ""
		s.timeLeft = 60

		go func() {
			// timer stops at 0 so a new auction cannot be started immediately by a new bid
			for s.timeLeft > 0 {
				s.timeLeft--
				if s.timeLeft%10 == 0 {
					log.Printf("time left of auction at server %v: %v seconds", s.id, s.timeLeft)
				}
				time.Sleep(time.Second)
			}
		}()

		log.Printf("server %v: started a new auction", s.id)
	}

	if (req.Bid > s.highestBid) && (s.timeLeft > 0) {
		s.highestBid = req.Bid
		s.currentWinnerUser = req.User
		return &auction.OutcomeReply{Outcome: auction.Outcomes(SUCCESS)}, nil
	} else {
		return &auction.OutcomeReply{Outcome: auction.Outcomes(FAIL)}, nil
	}
}

func (s *Server) ServerResult(ctx context.Context, resReq *auction.Request) (*auction.ResultReply, error) {
	return &auction.ResultReply{User: s.currentWinnerUser, HighestBid: s.highestBid, TimeLeft: s.timeLeft}, nil
}

func (s *Server) ServerReset(ctx context.Context, resReq *auction.Request) (*auction.OutcomeReply, error) {
	// Don't reset if auction isn't over
	if s.timeLeft > 0 {
		return &auction.OutcomeReply{Outcome: auction.Outcomes_FAIL}, nil
	}

	// timeLeft == -1 means that a new auction will be started when a bid is made
	s.timeLeft = -1
	log.Printf("server %v: resetted the auction", s.id)
	return &auction.OutcomeReply{}, nil
}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

type Server struct {
	auction.UnimplementedAuctionServer
	id                int32
	highestBid        int32
	currentWinnerUser string
	timeLeft          int32
}

type Outcomes int32

const (
	FAIL    Outcomes = 0
	SUCCESS Outcomes = 1
)
