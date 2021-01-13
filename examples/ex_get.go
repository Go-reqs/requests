package main

import (
	req "github.com/go-reqs/requests"
	log "github.com/sirupsen/logrus"
)

func testGet() {
	sess := req.NewSession()
	res, err := sess.Get(
		"http://httpbin.org/cookies/set?name1=value1&name2=value2",
		req.Header{"ok": "true"},
		req.Params{"s1": "some"},
	)
	log.Printf("%s, %v", res, err)
	log.Printf("%s", res.Request.Curl())
}

func main() {
	testGet()
}
