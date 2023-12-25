package model

import "net/http"

type Response struct {
	Resp *http.Response
	Err  error
}
