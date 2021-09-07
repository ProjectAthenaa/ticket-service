package deob

import (
	"context"
	"fmt"
	"github.com/ProjectAthenaa/sonic-core/fasttls"
	"github.com/ProjectAthenaa/sonic-core/sonic/core"
	"github.com/ProjectAthenaa/ticket-service/aes"
	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/common/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"regexp"
	"time"
)

var (
	liveJsonRe = regexp.MustCompile(`"v":"(\w+)"`)
	json       = jsoniter.ConfigFastest
)

type Version struct {
	EncKeys [][]int32 `json:"enc_keys"`
	DecKeys [][]int32 `json:"dec_keys"`
	Flags   []int     `json:"flags"`
	CParam  string    `json:"c_param"`
	Hash    string    `json:"hash"`
}

func (v *Version) Save() {
	data, _ := json.Marshal(&v)
	core.Base.GetRedis("cache").SetNX(context.Background(), fmt.Sprintf("ticket:versions:%s", v.Hash), string(data), time.Hour*24)
}

func (v *Version) GetLiveJSON(proxy *string, c *fasttls.Client) (string, error) {
	cparamValue := aes.Encrypt(Cparamarray(), aes.GenIV(), v.EncKeys)

	req, err := c.NewRequest("GET", fmt.Sprintf("https://www.supremenewyork.com/live.json?%s=%s", v.CParam, cparamValue), nil)
	if err != nil {
		return "", nil
	}

	res, err := c.Do(req)
	if err != nil {
		return "", err
	}

	return liveJsonRe.FindStringSubmatch(string(res.Body))[1], nil
}

func GetVersion(hash string) (*Version, error) {
	v := core.Base.GetRedis("cache").Get(context.Background(), fmt.Sprintf("ticket:versions:%s", hash)).Val()
	if v == "" {
		return nil, status.Error(codes.NotFound, "version_not_found")
	}

	var version *Version

	if err := json.Unmarshal([]byte(v), &version); err != nil {
		log.Error("error unmarshalling: ", err)
		return nil, status.Error(codes.Internal, "error_parsing_version")
	}

	return version, nil
}
