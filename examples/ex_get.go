package main

import (
	req "github.com/go-reqs/requests"
	log "github.com/sirupsen/logrus"
)

func testGet() {
	sess := req.NewSession()
	res, err := sess.Get(
		"http://httpbin.org/get?name1=value1&name2=value2",
		req.Header{"ok": "true"},
		req.Params{"s1": "some"},
		req.ClientConfig{
			TimeOut: 2,
		},
	)
	if err != nil {
		log.Error("err:", err)
		return
	}
	log.Printf("status: %s %d", res.R.Status, res.R.StatusCode)
	log.Printf("content: %s", res.GetBody())
	log.Printf("%s, %v", res, err)
	//log.Printf("%s", res.Request.Curl())
}

func main() {
	testGet()
}
