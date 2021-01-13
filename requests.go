/* Copyright 2021 by danielliwd86.
Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package requests

import (
	"bytes"
	"encoding/json"
	_ "fmt"
	"github.com/alessio/shellescape"
	"github.com/go-yaml/yaml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

var VERSION string = "0.1"

type Dict map[string]interface{}
type List []interface{}
type Header Dict
type Params Dict
type Datas Dict
type Forms Dict
type RawData interface{}
type Files Dict
type Cookies Dict

type Json struct {
	Data   interface{}
	stream io.ReadCloser
}

type Auth struct {
	username string
	password string
}

type Session struct {
	Request  *Request
	Response *Response
	Client   *http.Client
}

type Request struct {
	R       *http.Request
	Session *Session
	Jar     http.CookieJar
	Params  Dict
	Datas   Dict
	Body    []byte
}

type Response struct {
	R       *http.Response
	Request *Request
}

func paramsMakeKey(prefix string, k string) string {
	if prefix == "" {
		return k
	} else {
		return prefix + "[" + k + "]"
	}
}

func (r *Request) String() string {
	return string(YamlEncode(r).([]byte))
}

func (r *Request) MarshalYAML() (interface{}, error) {
	return Dict{
		"Url":    r.R.URL,
		"Header": r.R.Header,

		// yaml encode字符串时如果存在\n则使用yaml_LITERAL_SCALAR_STYLE风格
		"Body": string(r.Body) + "\n",
	}, nil
}

type _httpResponse http.Response

func (r *_httpResponse) MarshalYAML() (interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		body = []byte("!!read body error:" + err.Error())
	} else {
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
	}
	return Dict{
		"Status":     r.Status,
		"StatusCode": r.StatusCode,
		"Proto":      r.Proto,
		"ProtoMajor": r.ProtoMajor,
		"ProtoMinor": r.ProtoMinor,
		"Header":     r.Header,

		// yaml encode字符串时如果存在\n则使用yaml_LITERAL_SCALAR_STYLE风格
		"Body": string([]byte(body)) + "\n",
	}, nil
}

func (r *Response) String() string {
	h := Dict{
		"R":       (*_httpResponse)(r.R),
		"Request": r.Request,
	}
	return string(YamlEncode(h).([]byte))
}

func (r *Request) Curl() string {
	var s = []string{"curl -X " + r.R.Method}
	s = append(s, `"`+r.R.URL.String()+`"`)
	//var cs []string
	// for _, c := range r.Jar.Cookies(r.R.URL) {
	// 	cs = append(cs, c.Name+"="+c.Value)
	// }
	// if len(cs) > 0 {
	// 	s = append(s, `-H "Cookie: `+strings.Join(cs, "; ")+`"`)
	// }
	for k, v := range r.R.Header {
		s = append(s, `-H "`+k+`: `+strings.Join(v, "")+`"`)
	}

	if len(r.Body) > 0 {
		s = append(s, `-d `+shellescape.Quote(string(r.Body)))
	}

	return strings.Join(s, " \\\n")
}

func buildURLParams_(prefix string, params interface{}) []string {
	var out []string
	switch v := params.(type) {
	case string:
		out = append(out, url.QueryEscape(prefix)+"="+url.QueryEscape(v))
	case List:
		for i, val := range v {
			key := paramsMakeKey(prefix, strconv.Itoa(i))
			out = append(out, buildURLParams_(key, val)...)
		}
	case Dict:
		for k, val := range v {
			key := paramsMakeKey(prefix, k)
			out = append(out, buildURLParams_(key, val)...)
		}
	}
	return out
}

// handle URL params
func buildURLParams(prefix string, params Dict, args ...interface{}) string {
	var out []string
	var sorter func([]string)
	for _, v := range args {
		switch innerv := v.(type) {
		case func([]string):
			sorter = innerv
		}
	}
	out = buildURLParams_(prefix, params)
	if sorter != nil {
		sorter(out)
	}
	return strings.Join(out, "&")
}

func NewSession(args ...interface{}) *Session {

	sess := new(Session)

	for _, arg := range args {
		switch t := arg.(type) {
		case *http.Client:
			sess.Client = t
		}
	}

	if sess.Client == nil {
		c := &http.Client{}
		jar, _ := cookiejar.New(nil)
		c.Jar = jar
		sess.Client = c
	}

	return sess
}

func dictUpdate(d *Dict, u *Dict) {
	if (*d) == nil {
		*d = make(Dict)
	}
	for k, v := range *u {
		(*d)[k] = v
	}
}

func YamlEncode(data interface{}) RawData {
	d, err := yaml.Marshal(data)
	if err != nil {
		return nil
	}
	return RawData([]byte(d))
}

func JsonEncode(data interface{}) RawData {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(data)
	if err != nil {
		return nil
	}
	return RawData(b.Bytes())
}

func NewRequest(method string, urlstr string, args ...interface{}) (*Request, error) {
	req := new(Request)
	var err error

	req.R, err = http.NewRequest(method, urlstr, nil)
	if err != nil {
		return nil, err
	}
	req.R.Header.Set("User-Agent", "go-requests."+VERSION)
	var (
		rawData interface{}
	)
	for _, arg := range args {
		switch val := arg.(type) {
		case Header:
			for k, v := range val {
				headerval, ok := v.(string)
				if !ok {
					continue
				}
				req.R.Header.Set(k, headerval)
			}
		case Params:
			dictUpdate(&req.Params, (*Dict)(&val))
		case Datas:
			dictUpdate(&req.Datas, (*Dict)(&val))
		case RawData:
			rawData = val
		}
	}
	if rawData != nil {
		raw, ok := rawData.([]byte)
		if ok {
			req.Body = raw
		} else {
			raw, ok := rawData.(string)
			if ok {
				req.Body = []byte(raw)
			}
		}
		req.Datas = nil
	}

	if req.R.Header.Get("Content-Type") == "" {
		req.R.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if len(req.Datas) > 0 {
		if strings.Contains(req.R.Header.Get("Content-Type"), "x-www-form-urlencoded") {
			formdata := buildURLParams("", req.Datas)
			req.Body = []byte(formdata)
		} else if strings.Contains(req.R.Header.Get("Content-Type"), "json") {
			b := new(bytes.Buffer)
			err = json.NewEncoder(b).Encode(req.Datas)
			if err != nil {
				return nil, err
			}
			req.Body = b.Bytes()
		}
		req.R.Body = ioutil.NopCloser(bytes.NewReader(req.Body))
	}

	params := buildURLParams("", req.Params)
	if req.R.URL.RawQuery == "" {
		req.R.URL.RawQuery = params
	} else {
		req.R.URL.RawQuery += "&" + params
	}
	return req, nil
}

func (s *Session) Get(urlstr string, args ...interface{}) (*Response, error) {
	req, err := NewRequest("GET", urlstr, args...)
	if err != nil {
		return nil, err
	}

	res, err := s.Client.Do(req.R)

	if err != nil {
		return nil, err
	}

	resp := &Response{}
	resp.R = res
	resp.Request = req
	return resp, nil
}

func (s *Session) Post(urlstr string, args ...interface{}) (*Response, error) {
	req, err := NewRequest("POST", urlstr, args...)
	if err != nil {
		return nil, err
	}

	res, err := s.Client.Do(req.R)

	if err != nil {
		return nil, err
	}

	resp := &Response{}
	resp.R = res
	resp.Request = req
	return resp, nil
}
