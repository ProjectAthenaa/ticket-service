package main

import (
	protos "github.com/ProjectAthenaa/sonic-core/sonic/antibots/ticket/protos"
	ticket "github.com/ProjectAthenaa/ticket-service/services"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()

	protos.RegisterTicketServer(grpcServer, ticket.Server{})

	log.Println("gRPC server started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
