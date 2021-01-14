package requests

import (
	"net/url"
	"sort"
	"testing"
)

func TestParamsBuild(t *testing.T) {
	// t.Fatal("not implemented")
	d := Dict{
		"k1": "V1",
		"k2": Dict{
			"sub1": "v2",
		},
		"k3": List{
			"abc",
			"def",
		},
	}
	res, err := url.QueryUnescape(buildURLParams("", d, sort.Strings))
	expect := `k1=V1&k2[sub1]=v2&k3[0]=abc&k3[1]=def`
	if err != nil {
		t.Errorf("err %v", err)
	}
	if res != expect {
		t.Errorf("expect :%v\nbut got:%v", expect, res)
	}
}

func TestGet(t *testing.T) {
	sess := NewSession()
	res, err := sess.Get(
		"http://httpbin.org/cookies/set?na%20me1=value1&name2=value2",
		Header{"ok": "true"},
		Params{"s1": "some"},
	)
	t.Logf("%v, %v, %v", res.R, res.Request, err)
	//for _, cs := range sess.Client.Jar.Cookies(res.Request.R.URL) {
	//	t.Logf("%v", cs)
	//}
	//res.Request.Jar = sess.Client.Jar
	t.Logf("%s", res.Request.Curl())
}

func TestPostForm(t *testing.T) {
	sess := NewSession()
	res, err := sess.Post(
		"http://httpbin.org/post?na%20me1=value1&name2=value2",
		Header{"ok": "true"},
		Params{"s1": "some"},
		Datas{"s2": "post"},
	)
	t.Logf("%v, %v, %v", res.R, res.Request, err)
	t.Logf("%s", res.Request.Curl())
}

func TestPostJson(t *testing.T) {
	sess := NewSession()
	res, err := sess.Post(
		"http://httpbin.org/post?na%20me1=value1&name2=value2",
		Header{"content-type": "application/json"},
		Params{"s1": "some"},
		Datas{"s2": "post"},
	)
	t.Logf("%v, %v, %v", res.R, res.Request, err)
	t.Logf("%s", res.Request.Curl())
}

func TestPostRawJson(t *testing.T) {
	sess := NewSession()
	res, err := sess.Post(
		"http://httpbin.org/post?na%20me1=value1&name2=value2",
		Header{"content-type": "application/json"},
		Params{"s1": "some"},
		Datas{"s2": "post"},
		RawData(`{"raw":1}`),
	)
	t.Logf("%v, %v, %v", res.R, res.Request, err)
	t.Logf("%s", res.Request.Curl())
}

func TestPostJsonEncode(t *testing.T) {
	type PostData_Data struct {
		Msg    string `json:"msg"`
		Status int    `json:"status"`
	}
	type PostData struct {
		Id   int           `json:"id"`
		Data PostData_Data `json:"data"`
	}
	sess := NewSession()
	res, err := sess.Post(
		"http://httpbin.org/post?na%20me1=value1&name2=value2",
		Header{"content-type": "application/json"},
		Params{"s1": "some"},
		Datas{"s2": "post"},
		JsonEncode(PostData{
			Id: 1,
			Data: PostData_Data{
				Msg:    "helo",
				Status: 1,
			},
		}),
	)
	t.Logf("%s, %v", res, err)
	t.Logf("%s", res.Request.Curl())
}
