package ticket

import (
	"fmt"
	"github.com/valyala/fasthttp/fasthttpproxy"
	ticket "main/services/protos"
	"github.com/valyala/fasthttp"
	)

func getTicketJS(proxy *ticket.Proxy) (string, error){
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI("http://www.supremenewyork.com/ticket.js")

	var c *fasthttp.Client

	if proxy.Username != nil && proxy.Password != nil {
		c = &fasthttp.Client{
			Dial: fasthttpproxy.FasthttpHTTPDialer(fmt.Sprintf("%s:%s@%s:%s", *proxy.Username, *proxy.Password, proxy.IP, proxy.Port)),
		}
	}else{
		c = &fasthttp.Client{
			Dial: fasthttpproxy.FasthttpHTTPDialer(fmt.Sprintf("%s:%s", proxy.IP, proxy.Port)),
		}
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := c.Do(req, resp); err != nil{
		return "", err
	}

	return string(resp.Body()), nil
}


func getLiveJSON(proxy *ticket.Proxy) (string, error){
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI("http://www.supremenewyork.com/live.json")

	var c *fasthttp.Client

	if proxy.Username != nil && proxy.Password != nil {
		c = &fasthttp.Client{
			Dial: fasthttpproxy.FasthttpHTTPDialer(fmt.Sprintf("%s:%s@%s:%s", *proxy.Username, *proxy.Password, proxy.IP, proxy.Port)),
		}
	}else{
		c = &fasthttp.Client{
			Dial: fasthttpproxy.FasthttpHTTPDialer(fmt.Sprintf("%s:%s", proxy.IP, proxy.Port)),
		}
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := c.Do(req, resp); err != nil{
		return "", nil
	}


	return string(resp.Body()), nil
}