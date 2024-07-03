package reqmiddleware

import (
	"context"
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"net/http"
)

type InitRequestMiddleware struct {
	ctx    context.Context
	logger logservice.Logger
}

func NewInitRequestMiddleware(ctx context.Context, logger logservice.Logger) *InitRequestMiddleware {
	return &InitRequestMiddleware{ctx: ctx, logger: logger}
}

func (m *InitRequestMiddleware) InitRequest(next middleware.RequestModifier) middleware.RequestModifier {
	return middleware.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
		req, err := http.NewRequestWithContext(m.ctx, "GET", "", nil)
		if err != nil {
			return nil, err
		}
		return next.Do(req)
	})
}
