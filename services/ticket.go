package services

import (
	"context"
	"errors"
	protos "github.com/ProjectAthenaa/sonic-core/sonic/antibots/ticket"
	"github.com/ProjectAthenaa/ticket-service/aes"
	deob "github.com/ProjectAthenaa/ticket-service/deobufscator"
)

type Server struct {
	protos.UnimplementedTicketServer
}

func (s Server) Deobfuscate(ctx context.Context, request *protos.DeobfuscateRequest) (*protos.Hash, error) {
	ticketjs, ticketHash, err := getTicketJS(&request.Proxy)
	if err != nil {
		return nil, errors.New("ERROR_RETRIEVING")
	}

	if v, _ := deob.GetVersion(ticketHash); v != nil {
		return &protos.Hash{Value: v.Hash}, nil
	}

	version := deob.Process(ticketjs, ticketHash)

	version.Save()

	return &protos.Hash{Value: version.Hash}, nil
}

func (s Server) GenerateCookie(ctx context.Context, hash *protos.GenerateCookieRequest) (*protos.Cookie, error) {
	version, err := deob.GetVersion(hash.Hash)
	if err != nil {
		return nil, err
	}

	livejson, err := version.GetLiveJSON(&hash.Proxy, client)
	if err != nil {
		return nil, err
	}

	f := aes.Decrypt(livejson, version.DecKeys)[:48]
	f = append(f, version.Flags[:64]...)
	f = append(f, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16)
	cookie := aes.Encrypt(f, aes.GenIV(), version.EncKeys)

	return &protos.Cookie{Value: cookie}, nil
}
