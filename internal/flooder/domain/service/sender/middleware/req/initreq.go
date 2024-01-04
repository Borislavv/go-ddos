package reqmiddleware

import (
	middleware "ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "ddos/internal/log/domain/service"
	"net/http"
)

type InitRequestMiddleware struct {
	logger logservice.Logger
}

func NewInitRequestMiddleware(logger logservice.Logger) *InitRequestMiddleware {
	return &InitRequestMiddleware{logger: logger}
}

func (m *InitRequestMiddleware) InitRequest(next middleware.RequestModifier) middleware.RequestModifier {
	return middleware.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
		req, err := http.NewRequest("GET", "", nil)
		if err != nil {
			return nil, err
		}

		return next.Do(req)
	})
}
