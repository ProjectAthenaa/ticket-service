package services

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/ProjectAthenaa/sonic-core/fasttls"
	"github.com/ProjectAthenaa/sonic-core/fasttls/tls"
	"github.com/prometheus/common/log"
)

var client = fasttls.NewClient(tls.HelloChrome_91, nil)

func getTicketJS(proxy *string) (string, string, error) {
	req, err := client.NewRequest("GET", "https://www.supremenewyork.com/ticket.js", nil)
	if err != nil {
		log.Error("create req error: ", err)
		return "", "", err
	}

	req.Proxy = proxy

	resp, err := client.Do(req)
	if err != nil {
		log.Error("do req error: ", err)
		return "", "", err
	}

	return string(resp.Body), hash(string(resp.Body)), nil
}

func hash(text string) string {
	algorithm := sha1.New()
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}
