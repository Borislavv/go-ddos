package sender

import "net/http"

type Sender interface {
	Send(req *http.Request)
}
