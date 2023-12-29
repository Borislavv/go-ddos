package reqmiddleware

import (
	"ddos/internal/ddos/infrastructure/httpclient/middleware"
	"fmt"
	"net/http"
	"time"
)

const Timestamp = "Timestamp"

func AddTimestamp(next httpclientmiddleware.RequestModifier) httpclientmiddleware.RequestModifier {
	return httpclientmiddleware.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
		copyValues := req.URL.Query()
		if copyValues.Has(Timestamp) {
			copyValues.Del(Timestamp)
		}
		//copyValues.Add(Timestamp, strconv.FormatInt(time.Now().UnixMilli(), 10))
		copyValues.Add(Timestamp, fmt.Sprintf("%d", time.Now().UnixMilli()))
		req.URL.RawQuery = copyValues.Encode()
		return next.Do(req)
	})
}
