package main_test

import (
	"context"
	protos "github.com/ProjectAthenaa/sonic-core/sonic/antibots/ticket"
	"github.com/ProjectAthenaa/ticket-service/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"testing"
)

const bufSize = 1024 * 1024

var (
	client protos.TicketClient
	lis    *bufconn.Listener
)

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func init() {
	lis = bufconn.Listen(bufSize)

	server := grpc.NewServer()

	protos.RegisterTicketServer(server, services.Server{})

	go func() {
		server.Serve(lis)
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}

	client = protos.NewTicketClient(conn)
}

func TestTicket(t *testing.T) {
	hash, err := client.Deobfuscate(context.Background(), &protos.DeobfuscateRequest{Proxy: "1moewci2:4k7cvljz@178.159.147.248:65112"})
	if err != nil {
		t.Fatal(err)
	}


	cookie, err := client.GenerateCookie(context.Background(), &protos.GenerateCookieRequest{Proxy: "1moewci2:4k7cvljz@178.159.147.248:65112", Hash: hash.Value})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("NTBCC: " + cookie.Value)
}
