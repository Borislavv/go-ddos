package req

import (
	"ddos/internal/ddos/infrastructure/httpclient/middleware"
	"net/http"
	"strconv"
	"time"
)

const timestamp = "timestamp"

func AddTimestamp(next httpclientmiddleware.RequestModifier) httpclientmiddleware.RequestModifier {
	return httpclientmiddleware.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
		copyValues := req.URL.Query()
		if copyValues.Has(timestamp) {
			copyValues.Del(timestamp)
		}
		copyValues.Add(timestamp, strconv.Itoa(time.Now().Nanosecond()))
		req.URL.RawQuery = copyValues.Encode()
		return next.Do(req)
	})
}
