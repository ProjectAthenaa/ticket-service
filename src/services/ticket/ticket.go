package ticket

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/ProjectAthenaa/database-module/models"
	"main/aes"
	deob "main/deobufscator"
	"main/services"
	ticket "main/services/protos"
	"strings"
)

var dbs = services.DB
var rdb = dbs.RedisClient

type Server struct {
	ticket.UnimplementedTicketServer
}

func (s Server) Deobfuscate(ctx context.Context, request *ticket.DeobfsucateRequest) (*ticket.Hash, error) {
	ticketjs, err := getTicketJS(request.Proxy)
	if err != nil{
		return nil, errors.New("ERROR_RETRIEVING")
	}

	var agent = "browser"

	if request.Agent == ticket.AGENT_MOBILE{
		agent = "mobile"
	}

	version := deob.GetVersion(ticketjs, agent)

	hash := fmt.Sprint(md5.Sum([]byte(ticketjs)))

	_ = version.Save(hash, rdb)

	return &ticket.Hash{Value: hash}, nil
}

func (s Server) GenerateCookie(ctx context.Context, request *ticket.GenerateCookieRequest) (*ticket.Cookie, error) {
	livejson, _ := getLiveJSON(request.Proxy)

	version := models.TicketVersion{}
	err := version.Retrieve(request.Hash, rdb)

	if err != nil{
		return &ticket.Cookie{Value: "version_not_found"}, nil
	}

	cookie := ""

	switch strings.ToLower(version.Agent) {
	case "browser":
		f := aes.Decrypt(livejson, version.Dec)[:48]
		f = append(f, version.Flag...)
		f = append(f, 8, 8, 8, 8, 8, 8, 8, 8)
		cookie =  aes.Encrypt(f, aes.GenIV(), version.Enc)
	case "mobile":
		f := aes.Decrypt(livejson, version.Dec)[:48]
		f = append(f, version.Flag[:64]...)
		f = append(f, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16)
		cookie =  aes.Encrypt(f, aes.GenIV(), version.Enc)
	default:
		cookie =  "wrong_agent"
	}
	return &ticket.Cookie{Value: cookie}, nil
}