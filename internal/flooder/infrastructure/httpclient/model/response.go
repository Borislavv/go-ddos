package httpclientmodel

import (
	"net/http"
	"time"
)

type Response struct {
	resp      *http.Response
	err       error
	isFailed  bool
	timestamp time.Time
}

func NewResponse() *Response {
	return &Response{timestamp: time.Now()}
}

func (r *Response) Resp() (resp *http.Response) {
	return r.resp
}

func (r *Response) Err() (err error) {
	return r.err
}

func (r *Response) SetErr(err error) {
	r.err = err
}

func (r *Response) Response() (resp *http.Response, err error) {
	return r.resp, r.err
}

func (r *Response) SetResponse(resp *http.Response, err error) *Response {
	r.resp = resp
	r.err = err
	return r
}

func (r *Response) IsFailed() bool {
	return r.isFailed
}

func (r *Response) SetFailed() {
	r.isFailed = true
}

func (r *Response) Timestamp() time.Time {
	return r.timestamp
}

func (r *Response) Duration() time.Duration {
	return time.Since(r.timestamp)
}
