package main

import (
	"github.com/ProjectAthenaa/certificate-module"
	"go.elastic.co/apm/module/apmgrpc"
	"google.golang.org/grpc"
	"log"
	ticket "main/services/protos"
	ticket2 "main/services/ticket"
	"net"
)

func main() {
	listener, _ := net.Listen("tcp", ":4000")

	tlsCredentials, err := certificate_module.LoadCertificate()
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(tlsCredentials),
		grpc.UnaryInterceptor(
			apmgrpc.NewUnaryServerInterceptor()),
	)

	ticket.RegisterTicketServer(grpcServer, ticket2.Server{})

	log.Println("gRPC server started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
