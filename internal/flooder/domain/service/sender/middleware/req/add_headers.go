package reqmiddleware

import (
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"net/http"
)

type AddHeadersMiddleware struct {
	headers map[string]string
	logger  logservice.Logger
}

func NewAddHeadersMiddlewareMiddleware(headers map[string]string, logger logservice.Logger) *AddHeadersMiddleware {
	return &AddHeadersMiddleware{headers: headers, logger: logger}
}

func (m *AddHeadersMiddleware) AddHeaders(next httpclientmiddleware.RequestModifier) httpclientmiddleware.RequestModifier {
	return httpclientmiddleware.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
		if req != nil {
			for headerKey, headerValue := range m.headers {
				req.Header.Add(headerKey, headerValue)
			}
		}

		return next.Do(req)
	})
}
