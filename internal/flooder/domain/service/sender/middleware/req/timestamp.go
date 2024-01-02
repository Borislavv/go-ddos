package reqmiddleware

import (
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	"net/http"
	"strconv"
	"time"
)

const Timestamp = "Timestamp"

type TimestampMiddleware struct {
}

func NewTimestampMiddleware() *TimestampMiddleware {
	return &TimestampMiddleware{}
}

func (m *TimestampMiddleware) AddTimestamp(next httpclientmiddleware.RequestModifier) httpclientmiddleware.RequestModifier {
	return httpclientmiddleware.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
		if req != nil {
			copyValues := req.URL.Query()
			copyValues.Set(Timestamp, strconv.FormatInt(time.Now().UnixMilli(), 10))
			req.URL.RawQuery = copyValues.Encode()
		}
		return next.Do(req)
	})
}
