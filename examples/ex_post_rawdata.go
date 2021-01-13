package main

import (
	req "github.com/go-reqs/requests"
	log "github.com/sirupsen/logrus"
)

func testPostRaw() {
	sess := req.NewSession()
	res, err := sess.Post(
		"http://httpbin.org/post?name1=value1&name2=value2",
		req.Header{
			"Token":        "abcd",
			"content-type": "application/json",
		},
		req.Params{"time": "2021-01-01"},
		req.Datas{"event": "golang post"},
		// RawData will overide Datas
		req.RawData(`{"raw":"this is raw"}`),
	)
	log.Printf("%s, %v", res, err)
	log.Printf("%s", res.Request.Curl())
}

func main() {
	testPostRaw()
}
